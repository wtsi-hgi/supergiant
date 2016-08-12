package core

import "github.com/supergiant/supergiant/pkg/models"

type InstanceObserver struct {
	core *Core
}

func (s *InstanceObserver) Perform() error {
	var instances []*models.Instance
	if err := s.core.DB.Preload("Component.App.Kube.CloudAccount").Where("started = ?", true).Find(&instances); err != nil {
		return err
	}

	for _, instance := range instances {
		pod, err := s.core.Instances.pod(instance)
		if err != nil { // TODO should technically be checking error type here
			s.core.Log.Warnf("Could not load pod for Instance %d", instance.ID)
			continue
		}

		stats, err := pod.HeapsterStats()
		if err != nil {
			s.core.Log.Warnf("Could not load Heapster stats for pod %s", pod.Metadata.Name)
		} else {
			cpuUsage := stats.Stats["cpu-usage"]
			memUsage := stats.Stats["memory-usage"]

			// NOTE we took out the limit displayed by Heapster, and replaced it with
			// the actual limit value assigned to the pod. Heapster was returning limit
			// values less than the usage, which was causing errors when calculating
			// percentages.

			if cpuUsage != nil && memUsage != nil {
				instance.CPUUsage = cpuUsage.Minute.Average
				instance.CPULimit = int64(totalCpuLimit(pod).Millicores)

				instance.RAMUsage = memUsage.Minute.Average
				instance.RAMLimit = totalRamLimit(pod).Bytes

				// NOTE Heapster will return limits lower than usage values when there
				// are no limits set, so we lookup the host Node and supply that for limit.
				if instance.CPULimit < instance.CPUUsage || instance.RAMLimit < instance.RAMUsage {
					node := new(models.Node)
					if err := s.core.DB.First(node, "name = ?", pod.Spec.NodeName); err != nil {
						return err
					}
					instance.CPULimit = node.CPULimit
					instance.RAMLimit = node.RAMLimit
				}

				if err := s.core.DB.Save(instance); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
