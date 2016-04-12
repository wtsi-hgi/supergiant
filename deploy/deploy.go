package deploy

import (
	"github.com/supergiant/supergiant/client"
)

func Deploy(appName *string, componentName *string) error {

	sg := client.New("http://localhost:8080/v0", "", "", true)

	app, err := sg.Apps().Get(appName)
	if err != nil {
		return err
	}

	component, err := app.Components().Get(componentName)
	if err != nil {
		return err
	}

	var currentRelease *client.ReleaseResource
	if component.CurrentReleaseTimestamp != nil {
		currentRelease, err = component.CurrentRelease()
		if err != nil {
			return err
		}
	}

	targetRelease, err := component.TargetRelease()
	if err != nil {
		return err
	}

	targetList, err := targetRelease.Instances().List()
	if err != nil {
		return err
	}
	targetInstances := targetList.Items

	if currentRelease == nil { // first release
		for _, instance := range targetInstances {
			if err = instance.Start(); err != nil {
				return err
			}
		}
		for _, instance := range targetInstances {
			if err = instance.WaitForStarted(); err != nil {
				return err
			}
		}
		return nil
	}

	currentList, err := currentRelease.Instances().List()
	if err != nil {
		return err
	}
	currentInstances := currentList.Items

	// remove instances
	if currentRelease.InstanceCount > targetRelease.InstanceCount {
		instancesRemoving := currentRelease.InstanceCount - targetRelease.InstanceCount
		for _, instance := range currentInstances[len(currentInstances)-instancesRemoving:] {
			if err := instance.Stop(); err != nil {
				return err
			}
		}
		// add new instances
	} else if currentRelease.InstanceCount < targetRelease.InstanceCount {
		instancesAdding := targetRelease.InstanceCount - currentRelease.InstanceCount
		newInstances := targetInstances[len(targetInstances)-instancesAdding:]
		for _, instance := range newInstances {
			if err := instance.Start(); err != nil {
				return err
			}
		}
		for _, instance := range newInstances {
			if err := instance.WaitForStarted(); err != nil {
				return err
			}
		}
	}

	// update instances

	if *currentRelease.InstanceGroup == *targetRelease.InstanceGroup {
		return nil // no need to update restart instances
	}

	// NOTE we only want to update the minimum of (target, current) instance
	// counts. When adding instances, we wouldn't want to use target instance
	// count because we would restart new instances. When removing instances, we
	// couldn't use current count without getting index out of range
	var instancesRestarting int
	if currentRelease.InstanceCount < targetRelease.InstanceCount {
		instancesRestarting = currentRelease.InstanceCount
	} else {
		instancesRestarting = targetRelease.InstanceCount
	}

	for i := 0; i < instancesRestarting; i++ {
		currentInstance := currentInstances[i]
		targetInstance := targetInstances[i]

		currentInstance.Stop()
		currentInstance.WaitForStopped()

		targetInstance.Start()
		targetInstance.WaitForStarted()
	}

	return nil
}
