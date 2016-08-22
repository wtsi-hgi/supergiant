package core

import "github.com/supergiant/supergiant/pkg/models"

type NodeSynchronizer struct {
	core *Core
}

func (s *NodeSynchronizer) Perform() error {
	var kubes []*models.Kube
	if err := s.core.DB.Where("ready = ?", true).Preload("CloudAccount").Preload("Nodes", "provider_id <> ?", "").Find(&kubes); err != nil {
		return err
	}

	for _, kube := range kubes {
		servers, err := s.core.Kubes.servers(kube)
		if err != nil {
			return err
		}

		serverIDs := make(map[string]struct{})

		// Create new Nodes from newly discovered servers
		for _, server := range servers {
			exists := false
			for _, node := range kube.Nodes {
				if node.ProviderID == *server.InstanceId {
					serverIDs[node.ProviderID] = struct{}{}
					exists = true
					break
				}
			}
			if !exists {
				node := &models.Node{
					KubeID: kube.ID,
				}
				s.core.Nodes.setAttrsFromServer(node, server)
				if err := s.core.Nodes.Collection.Create(node); err != nil {
					return err
				}
			}
		}

		// Delete any Nodes which no longer exist
		for _, node := range kube.Nodes {
			if _, exists := serverIDs[node.ProviderID]; exists {
				continue
			}
			s.core.Log.Warnf("Deleting node with ID %s because it no longer exists in AWS", node.ProviderID)
			if err := s.core.Nodes.Delete(node.ID, node).Now(); err != nil {
				return err
			}
		}
	}

	return nil
}
