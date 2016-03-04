package model

type Volume struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Size int    `json:"size"`
}

type Container struct {
	Image        string              `json:"image"`
	Ports        []*Port             `json:"ports"`
	Env          []*EnvVar           `json:"env"`
	CPU          *ResourceAllocation `json:"cpu"`
	RAM          *ResourceAllocation `json:"ram"`
	VolumeMounts []*VolumeMount      `json:"volume_mounts"`
}

type Port struct {
	Protocol string `json:"protocol"`
	Public   bool   `json:"public"`
	Number   int    `json:"number"`
}

type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ResourceAllocation struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type VolumeMount struct {
	Volume string `json:"volume"`
	Path   string `json:"path"`
}

// Deployment is the top-level resource within an account
type Deployment struct {
	Name       string       `json:"name"`
	Instances  int          `json:"instances"`
	Volumes    []*Volume    `json:"volumes"`
	Containers []*Container `json:"containers"`
}
