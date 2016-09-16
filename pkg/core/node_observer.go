package core

import "github.com/supergiant/supergiant/pkg/model"

type NodeObserver struct {
	core *Core
}

func (s *NodeObserver) Perform() error {
	var kubes []*model.Kube
	if err := s.core.DB.Where("ready = ?", true).Preload("CloudAccount").Preload("Nodes", "provider_id <> ?", "").Find(&kubes); err != nil {
		return err
	}

	for _, kube := range kubes {
		for _, node := range kube.Nodes {
			k8sNode, err := s.core.K8S(kube).Nodes().Get(node.Name)
			if err != nil {
				if isKubeNotFoundErr(err) {
					// node.Ready = false    TODO we should probably be setting ready in this observer size
					s.core.Log.Warn(err.Error())
					continue
				} else {
					return err
				}
			}

			// node.Ready = true
			if ip := k8sNode.ExternalIP(); ip != "" {
				node.ExternalIP = ip
			}
			node.OutOfDisk = k8sNode.IsOutOfDisk()

			stats, err := k8sNode.HeapsterStats()
			if err != nil {
				s.core.Log.Warnf("Could not load Heapster stats for node %s", node.Name)
			} else {
				cpuUsage := stats.Stats["cpu-usage"]
				memUsage := stats.Stats["memory-usage"]

				if cpuUsage != nil && memUsage != nil {
					node.CPUUsage = cpuUsage.Minute.Average
					node.CPULimit = stats.Stats["cpu-limit"].Minute.Average

					node.RAMUsage = memUsage.Minute.Average
					node.RAMLimit = stats.Stats["memory-limit"].Minute.Average
				}
			}

			if err := s.core.DB.Save(node); err != nil {
				return err
			}
		}
	}

	return nil
}
