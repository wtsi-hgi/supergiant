package core

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/supergiant/guber"
	"github.com/supergiant/supergiant/common"
)

func RunCustomDeployment(core *Core, component *ComponentResource) error {
	cd := component.CustomDeployScript
	name := fmt.Sprintf("supergiant-custom-deploy-%s-%s-%s", *component.App().Name, *component.Name, *component.TargetReleaseTimestamp)
	podDef := &guber.Pod{
		Metadata: &guber.Metadata{
			Name: name,
		},
		Spec: &guber.PodSpec{
			Containers: []*guber.Container{
				&guber.Container{
					Name:    "container",
					Image:   cd.Image,
					Command: cd.Command,
				},
			},
			ImagePullPolicy: "Always",
		},
	}

	log.Printf("Creating pod %s", name)

	pod, err := core.K8S.Pods(*component.App().Name).Create(podDef)
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
		} else if pod == nil {
			return false, fmt.Errorf("pod %s does not exist", name)
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
			return false, err
		} else if pod == nil {
			return true, nil // done
		}

		if !pod.IsReady() {
			dumpContainerStatuses(pod)
			return false, fmt.Errorf("pod %s failed during deploy", name)
		}

		if latestLog, _ := pod.Log("container"); latestLog != "" {
			log = latestLog
		}

		return false, nil // pod still exists, keep going
	})

	fmt.Println(log)

	if err != nil {
		return err
	}

	// // Now we need to check to see if there were reported errors about the pod
	// query := &guber.QueryParams{
	// 	FieldSelector: "involvedObject.kind=Pod,involvedObject.name=" + name,
	// }
	// events, err := core.K8S.Events(*component.App().Name).Query(query)
	// if err != nil {
	// 	return err
	// }
	//
	// for _, event := range events.Items {
	// 	fmt.Println("EVENT: ", fmt.Sprintf("%#v", event))
	// }

	return nil
}

func dumpContainerStatuses(pod *guber.Pod) error {
	dump, err := json.Marshal(pod.Status.ContainerStatuses)
	if err != nil {
		return err
	}
	log.Println(string(dump))
	return nil
}
