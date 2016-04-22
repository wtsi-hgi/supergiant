package core

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/supergiant/guber"
	"github.com/supergiant/supergiant/common"
)

// kube_helpers.go is a collection of helper methods that convert a Supergiant
// resource definition into a Kubernetes resource defition.
// (and some other assorted things that should maybe be moved out...)

func isKubeNotFoundErr(err error) bool {
	_, yes := err.(*guber.Error404)
	return yes
}

func kubeVolumeMounts(m *common.ContainerBlueprint) (volMounts []*guber.VolumeMount) {
	for _, mount := range m.Mounts {
		volMounts = append(volMounts, asKubeVolumeMount(mount))
	}
	return volMounts
}

func kubeContainerPorts(m *common.ContainerBlueprint) (cPorts []*guber.ContainerPort) {
	for _, port := range m.Ports {
		cPorts = append(cPorts, asKubeContainerPort(port))
	}
	return cPorts
}

func interpolatedEnvVars(m *common.ContainerBlueprint, instance *InstanceResource) (envVars []*guber.EnvVar) {
	for _, envVar := range m.Env {
		envVars = append(envVars, asKubeEnvVar(envVar, instance))
	}
	return envVars
}

func ImageRepoName(m *common.ContainerBlueprint) string {
	return strings.Split(m.Image, "/")[0]
}

func asKubeContainer(m *common.ContainerBlueprint, instance *InstanceResource) *guber.Container { // NOTE how instance must be passed here
	// TODO
	resources := &guber.Resources{
		Requests: new(guber.ResourceValues),
		Limits:   new(guber.ResourceValues),
	}
	if m.RAM != nil {
		if m.RAM.Min != nil {
			resources.Requests.Memory = m.RAM.Min.ToKubeMebibytes()
		}
		if m.RAM.Max != nil {
			resources.Limits.Memory = m.RAM.Max.ToKubeMebibytes()
		}
	}
	if m.CPU != nil {
		if m.CPU.Min != nil {
			resources.Requests.CPU = m.CPU.Min.ToKubeMillicores()
		}
		if m.CPU.Max != nil {
			resources.Limits.CPU = m.CPU.Max.ToKubeMillicores()
		}
	}

	// TODO
	containerName := m.Name
	if m.Name == "" {
		rxp := regexp.MustCompile("[^A-Za-z0-9]")
		containerName = rxp.ReplaceAllString(m.Image, "-")
	}

	container := &guber.Container{
		Name:         containerName,
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
func interpolatedValue(m *common.EnvVar, instance *InstanceResource) string {
	r := strings.NewReplacer(
		"{{ instance_id }}", common.StringID(instance.ID),
		"{{ other_stuff }}", "TODO")
	return r.Replace(m.Value)
}

func asKubeEnvVar(m *common.EnvVar, instance *InstanceResource) *guber.EnvVar {
	return &guber.EnvVar{
		Name:  m.Name,
		Value: interpolatedValue(m, instance),
	}
}

// Volume
//==============================================================================
func asKubeVolume(m *AwsVolume) (*guber.Volume, error) {
	vol, err := m.awsVolume()
	if err != nil {
		return nil, err
	}

	return &guber.Volume{
		Name: common.StringID(m.Blueprint.Name), // NOTE this is not the physical volume name
		AwsElasticBlockStore: &guber.AwsElasticBlockStore{
			VolumeID: *vol.VolumeId,
			FSType:   "ext4",
		},
	}, nil
}

// Mount
//==============================================================================
func asKubeVolumeMount(m *common.Mount) *guber.VolumeMount {
	return &guber.VolumeMount{
		Name:      *m.Volume,
		MountPath: m.Path,
	}
}

// Port
//==============================================================================
func portName(m *common.Port) string {
	return strconv.Itoa(m.Number)
}

func asKubeContainerPort(m *common.Port) *guber.ContainerPort {
	return &guber.ContainerPort{
		ContainerPort: m.Number,
	}
}

func asKubeServicePort(m *common.Port) *guber.ServicePort {
	return &guber.ServicePort{
		Name:     portName(m),
		Port:     m.Number,
		Protocol: "TCP", // this is default; only other option is UDP
	}
}

// ImageRepo
//==============================================================================
func asKubeImagePullSecret(m *ImageRepoResource) *guber.ImagePullSecret {
	return &guber.ImagePullSecret{
		Name: common.StringID(m.Name),
	}
}

func asKubeSecret(m *ImageRepoResource) *guber.Secret {
	return &guber.Secret{
		Metadata: &guber.Metadata{
			Name: common.StringID(m.Name),
		},
		Type: "kubernetes.io/dockercfg",
		Data: map[string]string{
			".dockercfg": m.Key,
		},
	}
}
