package core

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/supergiant/guber"
	"github.com/supergiant/supergiant/common"
)

func RunCustomDeployment(core *Core, component *ComponentResource) error {
	cd := component.CustomDeployScript
	name := fmt.Sprintf("supergiant-custom-deploy-%s-%s-%s", common.StringID(component.App().Name), common.StringID(component.Name), *component.TargetReleaseTimestamp)
	podDef := &guber.Pod{
		Metadata: &guber.Metadata{
			Name: name,
		},
		Spec: &guber.PodSpec{
			Containers: []*guber.Container{
				&guber.Container{
					Name:            "container",
					Image:           cd.Image,
					Command:         cd.Command,
					ImagePullPolicy: "Always",
				},
			},
			RestartPolicy: "OnFailure",
		},
	}

	Log.Infof("Creating pod %s", name)

	pod, err := core.k8s.Pods(common.StringID(component.App().Name)).Create(podDef)
	if err != nil {
		return err
	}

	defer func() {
		if pod != nil {
			pod.Delete()
		}
	}()

	// Wait for pod to start
	msg := fmt.Sprintf("%s (pod start)", name)
	err = common.WaitFor(msg, time.Minute*2, time.Second*5, func() (bool, error) {
		pod, err = pod.Reload()
		if err != nil {
			return false, err
		}
		return pod.IsReady(), nil
	})

	if err != nil {
		dumpContainerStatuses(pod) // not doing anything with error here
		return err
	}

	var timeout time.Duration
	if cd.Timeout == 0 {
		timeout = 30 * time.Minute
	} else {
		timeout = time.Duration(cd.Timeout) * time.Second
	}

	var log string

	err = common.WaitFor(name, timeout, time.Second*5, func() (bool, error) {
		pod, err = pod.Reload()
		if err != nil {
			if isKubeNotFoundErr(err) {
				// This or the Phase == "Succeeded" line may fire, but this one is much
				// less likely. The pod seems to linger for a while as we capture the Status
				return true, nil // done
			} else {
				return false, err
			}
		}

		if !pod.IsReady() {
			if pod.Status.Phase == "Succeeded" {
				return true, nil
			} else {
				dumpContainerStatuses(pod)
				return false, fmt.Errorf("pod %s failed during deploy", name)
			}
		}

		if latestLog, _ := pod.Log("container"); latestLog != "" {
			log = latestLog
		}

		return false, nil // pod still exists, keep going
	})

	Log.Info(log)

	if err != nil {
		return err
	}

	return nil
}

func dumpContainerStatuses(pod *guber.Pod) error {
	dump, err := json.Marshal(pod.Status.ContainerStatuses)
	if err != nil {
		return err
	}
	Log.Error(string(dump))
	return nil
}
