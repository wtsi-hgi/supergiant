package core

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/supergiant/supergiant/pkg/models"
)

type Volumes struct {
	Collection
}

func (c *Volumes) Provision(id *int64, m *models.Volume) *Action {
	return &Action{
		Status: &models.ActionStatus{
			Description: "provisioning",

			// TODO
			// This resource has an issue with retryable provisioning -- which in this
			// context means creating an remote asset from the local record.
			//
			// Apps, for example, which use their user-set Name field as the actual
			// identifier for the provisioned Kubernetes Namespace. That makes the
			// creation of the Namespace retryable, because it is IDEMPOTENT.
			//
			// The problem here, is that WE CANNOT SET AN IDENTIFIER UP FRONT. The ID
			// is given to us upon successful creation of the remote asset.
			//
			// You can tag volumes after creation, but that means it is a 2-step
			// process, which means it fails to be atomic -- if tag creation fails,
			// retrying would re-create a volume, since our identifer (which is used
			// to load and check existence of the asset) was never set.
			//
			// We currently do not have a great solution in place for this problem.
			// In the meantime, MaxRetries is set low to prevent creating several
			// duplicate, billable assets in the user's cloud account. If there is an
			// error, the user will know about it quickly, instead of after 20 retries.
			MaxRetries: 1,
		},
		core:  c.core,
		scope: c.core.DB.Preload("Instance").Preload("Kube.CloudAccount"),
		model: m,
		id:    id,
		fn: func(_ *Action) error {
			return c.createVolume(m, nil)
		},
	}
}

func (c *Volumes) Delete(id *int64, m *models.Volume) *Action {
	return &Action{
		Status: &models.ActionStatus{
			Description: "deleting",
			MaxRetries:  5,
		},
		core:  c.core,
		scope: c.core.DB.Preload("Instance").Preload("Kube.CloudAccount"),
		model: m,
		id:    id,
		fn: func(_ *Action) error {
			if err := c.deleteVolume(m); err != nil {
				return err
			}
			return c.Collection.Delete(id, m)
		},
	}
}

// Resize the Volume
func (c *Volumes) Resize(id *int64, m *models.Volume) *Action {
	return &Action{
		Status: &models.ActionStatus{
			Description: "resizing",
		},
		core:  c.core,
		scope: c.core.DB.Preload("Instance").Preload("Kube.CloudAccount"),
		model: m,
		id:    id,
		fn: func(_ *Action) error {
			snapshot, err := c.createSnapshot(m)
			if err != nil {
				return err
			}
			if err := c.deleteVolume(m); err != nil {
				return err
			}
			if err := c.createVolume(m, snapshot.SnapshotId); err != nil {
				return err
			}
			if err := c.deleteSnapshot(m, snapshot); err != nil {
				c.core.Log.Errorf("Error deleting snapshot %s: %s", *snapshot.SnapshotId, err.Error())
			}
			return nil
		},
	}
}

////////////////////////////////////////////////////////////////////////////////
// Private methods                                                            //
////////////////////////////////////////////////////////////////////////////////

func (c *Volumes) createVolume(volume *models.Volume, snapshotID *string) error {
	volInput := &ec2.CreateVolumeInput{
		AvailabilityZone: aws.String(volume.Kube.Config.AvailabilityZone),
		VolumeType:       aws.String(volume.Type),
		Size:             aws.Int64(int64(volume.Size)),
		SnapshotId:       snapshotID,
	}
	awsVol, err := c.core.CloudAccounts.ec2(volume.Kube.CloudAccount, volume.Kube.Config.Region).CreateVolume(volInput)
	if err != nil {
		return err
	}

	volume.ProviderID = *awsVol.VolumeId
	volume.Size = int(*awsVol.Size)
	if err := c.core.DB.Save(volume); err != nil {
		return err
	}

	tagsInput := &ec2.CreateTagsInput{
		Resources: []*string{
			awsVol.VolumeId,
		},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(volume.Name),
			},
		},
	}
	_, err = c.core.CloudAccounts.ec2(volume.Kube.CloudAccount, volume.Kube.Config.Region).CreateTags(tagsInput)
	return err
}

func (c *Volumes) deleteVolume(volume *models.Volume) error {
	if err := c.waitForAvailable(volume); err != nil {
		return err
	}
	input := &ec2.DeleteVolumeInput{
		VolumeId: aws.String(volume.ProviderID),
	}
	if _, err := c.core.CloudAccounts.ec2(volume.Kube.CloudAccount, volume.Kube.Config.Region).DeleteVolume(input); isErrAndNotAWSNotFound(err) {
		return err
	}
	return nil
}

func (c *Volumes) waitForAvailable(volume *models.Volume) error {
	input := &ec2.DescribeVolumesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("volume-id"),
				Values: []*string{
					aws.String(volume.ProviderID),
				},
			},
		},
	}

	resp, err := c.core.CloudAccounts.ec2(volume.Kube.CloudAccount, volume.Kube.Config.Region).DescribeVolumes(input)
	if err != nil {
		return err
	}
	if len(resp.Volumes) == 0 {
		return nil
	}

	c.core.Log.Debugf("Waiting for EBS volume %s to be available", volume.Name)
	return c.core.CloudAccounts.ec2(volume.Kube.CloudAccount, volume.Kube.Config.Region).WaitUntilVolumeAvailable(input)
}

func (c *Volumes) forceDetachVolume(volume *models.Volume) error {
	input := &ec2.DetachVolumeInput{
		VolumeId: aws.String(volume.ProviderID),
		Force:    aws.Bool(true),
	}
	if _, err := c.core.CloudAccounts.ec2(volume.Kube.CloudAccount, volume.Kube.Config.Region).DetachVolume(input); isErrAndNotAWSNotFound(err) {
		return err
	}
	return nil
}

func (c *Volumes) createSnapshot(volume *models.Volume) (*ec2.Snapshot, error) {
	input := &ec2.CreateSnapshotInput{
		Description: aws.String(fmt.Sprintf("%s-%d", volume.Name, volume.Instance.ReleaseID)),
		VolumeId:    aws.String(volume.ProviderID),
	}
	snapshot, err := c.core.CloudAccounts.ec2(volume.Kube.CloudAccount, volume.Kube.Config.Region).CreateSnapshot(input)
	if err != nil {
		return nil, err
	}
	waitInput := &ec2.DescribeSnapshotsInput{
		SnapshotIds: []*string{snapshot.SnapshotId},
	}
	if err := c.core.CloudAccounts.ec2(volume.Kube.CloudAccount, volume.Kube.Config.Region).WaitUntilSnapshotCompleted(waitInput); err != nil {
		return nil, err // TODO destroy snapshot that failed to complete?
	}
	return snapshot, nil
}

func (c *Volumes) deleteSnapshot(volume *models.Volume, snapshot *ec2.Snapshot) error {
	input := &ec2.DeleteSnapshotInput{
		SnapshotId: snapshot.SnapshotId,
	}
	if _, err := c.core.CloudAccounts.ec2(volume.Kube.CloudAccount, volume.Kube.Config.Region).DeleteSnapshot(input); isErrAndNotAWSNotFound(err) {
		return err
	}
	return nil
}
