package deploy

import (
	"github.com/supergiant/supergiant/pkg/client"
	"github.com/supergiant/supergiant/pkg/model"
)

func Deploy(sg *client.Client, componentID *int64) error {
	includes := []string{
		"App.Kube.CloudAccount",
		"PrivateImageKeys",
		"CurrentRelease",
		"TargetRelease",
		"Instances.Volumes",
	}
	component := new(model.Component)
	if err := sg.Components.GetWithIncludes(componentID, component, includes); err != nil {
		return err
	}

	// If first deploy, concurrently start Instances and return
	if component.CurrentRelease == nil {
		for _, instance := range component.Instances {
			if instance.Started {
				continue
			}
			if err := sg.Instances.Start(instance); err != nil {
				return err
			}
		}
		for _, instance := range component.Instances {
			if err := sg.Instances.WaitForStarted(instance); err != nil {
				return err
			}
		}
		return nil
	}

	for _, instance := range component.Instances {
		// Remove Instances
		if (instance.Num + 1) > component.TargetRelease.InstanceCount {
			if err := sg.Instances.Delete(instance.ID, instance); err != nil {
				return err
			}
			if err := sg.Instances.WaitForDeleted(instance); err != nil {
				return err
			}
			continue
		}

		// Stop Instance if not using new Release
		if *instance.ReleaseID != *component.TargetReleaseID {
			if err := sg.Instances.Stop(instance); err != nil {
				return err
			}
			if err := sg.Instances.WaitForStopped(instance); err != nil {
				return err
			}
		}

		// Start Instance
		if instance.Started {
			continue
		}
		if err := sg.Instances.Start(instance); err != nil {
			return err
		}
		if err := sg.Instances.WaitForStarted(instance); err != nil {
			return err
		}
	}
	return nil
}
