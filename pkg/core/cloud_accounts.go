package core

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/supergiant/supergiant/pkg/models"
)

// TODO this and the similar concept in Kubes should be moved to core, not global vars
var globalAWSSession = session.New()

type CloudAccounts struct {
	Collection
}

func (c *CloudAccounts) Create(m *models.CloudAccount) error {
	// Validate the credentials
	if _, err := c.ec2(m, "us-east-1").DescribeKeyPairs(new(ec2.DescribeKeyPairsInput)); err != nil {
		return err
	}
	return c.Collection.Create(m)
}

func (c *CloudAccounts) Delete(id *int64, m *models.CloudAccount) error {
	if err := c.core.DB.Find(&m.Kubes, "cloud_account_id = ?", id); err != nil {
		return err
	}
	if len(m.Kubes) > 0 {
		return errors.New("Cannot delete CloudAccount that has active Kubes")
	}
	return c.Collection.Delete(id, m)
}

////////////////////////////////////////////////////////////////////////////////
// Private methods                                                            //
////////////////////////////////////////////////////////////////////////////////

func (c *CloudAccounts) awsConfig(m *models.CloudAccount, region string) *aws.Config {
	creds := credentials.NewStaticCredentials(m.Credentials["access_key"], m.Credentials["secret_key"], "")
	creds.Get()
	return aws.NewConfig().WithRegion(region).WithCredentials(creds)
}

func (c *CloudAccounts) ec2(m *models.CloudAccount, region string) *ec2.EC2 {
	return ec2.New(globalAWSSession, c.awsConfig(m, region))
}

func (c *CloudAccounts) iam(m *models.CloudAccount, region string) *iam.IAM {
	return iam.New(globalAWSSession, c.awsConfig(m, region))
}

func (c *CloudAccounts) elb(m *models.CloudAccount, region string) *elb.ELB {
	return elb.New(globalAWSSession, c.awsConfig(m, region))
}

func (c *CloudAccounts) autoscaling(m *models.CloudAccount, region string) *autoscaling.AutoScaling {
	return autoscaling.New(globalAWSSession, c.awsConfig(m, region))
}
