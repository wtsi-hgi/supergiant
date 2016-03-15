package job

import (
	"encoding/json"
	"fmt"
	"guber"
	"strings"
	"supergiant/core/model"
	"supergiant/core/storage"
)

type DeployComponentMessage struct {
	AppName       string
	ComponentName string
}

// DeployComponent implements job.Performable interface
type DeployComponent struct {
	db   *storage.Client
	kube *guber.Client
}

func (j DeployComponent) MaxAttempts() int {
	return 20
}

func (j DeployComponent) Perform(data string) error {
	var message *DeployComponentMessage
	if err := json.Unmarshal([]byte(data), message); err != nil {
		return err
	}

	component, err := j.db.ComponentStorage.Get(message.AppName, message.ComponentName)
	if err != nil {
		return err
	}

	// create namespace
	namespace, err := j.kube.Namespaces().Get(message.AppName)
	if err != nil {
		namespace = &guber.Namespace{
			Metadata: &guber.Metadata{
				Name: message.AppName,
			},
		}
		namespace, err = j.kube.Namespaces().Create(namespace)
		if err != nil {
			return err
		}
	}

	// get repo names from images
	var repoNames []string
	for _, container := range component.Containers {
		repoName := strings.Split(container.Image, "/")[0]
		repoNames = append(repoNames, repoName)
	}

	// create secrets for namespace
	// and create imagePullSecrets from repos
	var imagePullSecrets []*guber.ImagePullSecret
	for _, repoName := range repoNames {
		repo, err := j.db.ImageRepoStorage.Get(repoName)
		if err != nil {
			return err
		}

		secret, err := j.kube.Secrets(namespace.Metadata.Name).Get(repo.Name)
		if err != nil {
			secret = &guber.Secret{
				Metadata: &guber.Metadata{
					Name: repo.Name,
				},
				Type: "kubernetes.io/dockercfg",
				Data: map[string]string{
					".dockercfg": repo.Key,
				},
			}
			secret, err = j.kube.Secrets(namespace.Metadata.Name).Create(secret)
			if err != nil {
				return err
			}
		}

		imagePullSecret := &guber.ImagePullSecret{repo.Name}
		imagePullSecrets = append(imagePullSecrets, imagePullSecret)
	}

	// get all (uniq) ports from containers array
	// divide into private(ClusterIP) and public(NodePort) groups
	var externalPorts []*model.Port
	var internalPorts []*model.Port
	for _, container := range component.Containers {
		for _, port := range container.Ports {
			if port.Public == true {
				externalPorts = append(externalPorts, port)
			} else {
				internalPorts = append(internalPorts, port)
			}
		}
	}

	// create internal service
	if len(internalPorts) > 0 {
		internalServiceName := component.Name
		service, err := j.kube.Services(namespace.Metadata.Name).Get(internalServiceName)
		if err != nil {

			var servicePorts []*guber.ServicePort
			for _, port := range internalPorts {
				servicePort := &guber.ServicePort{
					Name:     string(port.Number),
					Protocol: port.Protocol,
					Port:     port.Number,
				}
				servicePorts = append(servicePorts, servicePort)
			}

			service = &guber.Service{
				Metadata: &guber.Metadata{
					Name: internalServiceName,
				},
				Spec: &guber.ServiceSpec{
					Selector: map[string]string{
						"deployment": component.ActiveDeploymentID,
					},
					Ports: servicePorts,
				},
			}
			service, err = j.kube.Services(namespace.Metadata.Name).Create(service)
			if err != nil {
				return err
			}
		}
	}

	// create external service
	if len(externalPorts) > 0 {
		externalServiceName := fmt.Sprintf("%s-public", component.Name)
		service, err := j.kube.Services(namespace.Metadata.Name).Get(externalServiceName)
		if err != nil {

			var servicePorts []*guber.ServicePort
			for _, port := range externalPorts {
				servicePort := &guber.ServicePort{
					Name:     string(port.Number),
					Protocol: port.Protocol,
					Port:     port.Number,
					// NodePort:
				}
				servicePorts = append(servicePorts, servicePort)
			}

			service = &guber.Service{
				Metadata: &guber.Metadata{
					Name: externalServiceName,
				},
				Spec: &guber.ServiceSpec{
					Type: "NodePort",
					Selector: map[string]string{
						"deployment": component.ActiveDeploymentID,
					},
					Ports: servicePorts,
				},
			}
			service, err = j.kube.Services(namespace.Metadata.Name).Create(service)
			if err != nil {
				return err
			}
		}
	}

	// for each instance,

	//    create RC, and interpolate data into Env vars

	return nil
}
