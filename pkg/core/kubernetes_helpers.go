package core

import (
	"strconv"
	"strings"

	"github.com/supergiant/guber"
	"github.com/supergiant/supergiant/pkg/model"
)

// kube_helpers.go is a collection of helper methods that convert a Supergiant
// resource definition into a Kubernetes resource defition.
// (and some other assorted things that should maybe be moved out...)

func isKubeNotFoundErr(err error) bool {
	_, yes := err.(*guber.Error404)
	return yes
}

func isKubeAlreadyExistsErr(err error) bool {
	_, yes := err.(*guber.Error409)
	return yes
}

func kubeVolumeMounts(m *model.ContainerBlueprint) (volMounts []*guber.VolumeMount) {
	for _, mount := range m.Mounts {
		volMounts = append(volMounts, &guber.VolumeMount{
			Name:      mount.Volume,
			MountPath: mount.Path,
		})
	}
	return
}

func kubeContainerPorts(m *model.ContainerBlueprint) (cPorts []*guber.ContainerPort) {
	for _, port := range m.Ports {
		cPorts = append(cPorts, &guber.ContainerPort{ContainerPort: port.Number})
	}
	return
}

func interpolatedEnvVars(m *model.ContainerBlueprint, instance *model.Instance) (envVars []*guber.EnvVar) {
	for _, envVar := range m.Env {
		envVars = append(envVars, asKubeEnvVar(envVar, instance))
	}
	return envVars
}

func asKubeContainer(m *model.ContainerBlueprint, instance *model.Instance) *guber.Container { // NOTE how instance must be passed here
	// TODO
	resources := &guber.Resources{
		Requests: new(guber.ResourceValues),
		Limits:   new(guber.ResourceValues),
	}

	if m.RAMRequest == nil {
		m.RAMRequest = new(model.BytesValue)
	}
	if m.CPURequest == nil {
		m.CPURequest = new(model.CoresValue)
	}
	resources.Requests.Memory = m.RAMRequest.ToKubeMebibytes()
	resources.Requests.CPU = m.CPURequest.ToKubeMillicores()

	if m.RAMLimit != nil {
		resources.Limits.Memory = m.RAMLimit.ToKubeMebibytes()
	}
	if m.CPULimit != nil {
		resources.Limits.CPU = m.CPULimit.ToKubeMillicores()
	}

	container := &guber.Container{
		Name:         m.NameOrDefault(),
		Image:        m.Image,
		Env:          interpolatedEnvVars(m, instance),
		Resources:    resources,
		VolumeMounts: kubeVolumeMounts(m),
		Ports:        kubeContainerPorts(m),

		// TODO this should be an option, enabled by default with volumes
		SecurityContext: &guber.SecurityContext{
			Privileged: true,
		},

		// TODO option
		ImagePullPolicy: "Always",
	}

	if m.Command != nil {
		container.Command = m.Command
	}

	return container
}

// EnvVar
//==============================================================================
func interpolatedValue(m *model.EnvVar, instance *model.Instance) string {
	r := strings.NewReplacer(
		"{{ instance_id }}", strconv.Itoa(instance.Num),
		"{{ other_stuff }}", "TODO")
	return r.Replace(m.Value)
}

func asKubeEnvVar(m *model.EnvVar, instance *model.Instance) *guber.EnvVar {
	return &guber.EnvVar{
		Name:  m.Name,
		Value: interpolatedValue(m, instance),
	}
}

// Port
//==============================================================================
func asKubeServicePort(m *model.Port) *guber.ServicePort {
	return &guber.ServicePort{
		Name:     strconv.Itoa(m.Number),
		Port:     m.Number,
		Protocol: "TCP", // this is default; only other option is UDP
	}
}

// ImageRepo
//==============================================================================
func provisionSecret(core *Core, app *model.App, key *model.PrivateImageKey) error {
	secret := &guber.Secret{
		Metadata: &guber.Metadata{
			Name: key.Username,
		},
		Type: "kubernetes.io/dockerconfigjson",
		Data: map[string]string{
			".dockerconfigjson": key.Key,
		},
	}
	if _, err := core.K8S(app.Kube).Secrets(app.Name).Create(secret); err != nil && !isKubeAlreadyExistsErr(err) {
		return err
	}
	return nil
}

// Misc
func totalCpuLimit(pod *guber.Pod) *model.CoresValue {
	cores := new(model.CoresValue)
	for _, container := range pod.Spec.Containers {
		if container.Resources != nil && container.Resources.Limits != nil {
			cores.Millicores += model.CoresFromString(container.Resources.Limits.CPU).Millicores
		}
	}
	return cores
}

func totalRamLimit(pod *guber.Pod) *model.BytesValue {
	bytes := new(model.BytesValue)
	for _, container := range pod.Spec.Containers {
		if container.Resources != nil && container.Resources.Limits != nil {
			bytes.Bytes += model.BytesFromString(container.Resources.Limits.Memory).Bytes
		}
	}
	return bytes
}
