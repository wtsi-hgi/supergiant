package model

type Volume struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Size int    `json:"size"`
}

type Container struct {
	Image  string              `json:"image"`
	Ports  []*Port             `json:"ports"`
	Env    []*EnvVar           `json:"env"`
	CPU    *ResourceAllocation `json:"cpu"`
	RAM    *ResourceAllocation `json:"ram"`
	Mounts []*Mount            `json:"mounts"`
}

type Port struct {
	Protocol string `json:"protocol"`
	Number   int    `json:"number"`
	Public   bool   `json:"public"`
}

type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ResourceAllocation struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type Mount struct {
	Volume string `json:"volume"`
	Path   string `json:"path"`
}

type CustomDeployScript struct {
	Image   string `json:"image"`
	Command string `json:"command"`
}

type Component struct {
	Name string `json:"name"`

	TerminationGracePeriod int `json:"termination_grace_period"`

	Instances  int          `json:"instances"`
	Volumes    []*Volume    `json:"volumes"`
	Containers []*Container `json:"containers"`

	// TODO kinda weird,
	// you choose a container that has the deploy file, and then reference it as a command
	CustomDeployScript *CustomDeployScript `json:"custom_deploy_script"`

	CurrentReleaseID    int    `json:"current_release_id"`
	ActiveDeploymentID  string `json:"active_deployment_id"`
	StandbyDeploymentID string `json:"standby_deployment_id"`
}
