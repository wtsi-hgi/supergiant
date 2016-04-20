package common

// Tags is a meta field for holding unstructued key/val info.
type Tags map[string]string

// Meta is a set of fields on all Resources for holding metadata.
type Meta struct {
	Created *Timestamp `json:"created"`
	Updated *Timestamp `json:"updated"`
	Tags    Tags       `json:"tags"`
}

// TODO weird placement
func NewMeta() *Meta {
	return &Meta{
		Tags: make(Tags),
	}
}

// App is the main top-level Resource within Supergiant, acting as a logical
// namespace for Components and all of their controlled assets -- along with
// provisioning an actual Kubernetes Namespace, it is used as a base name for
// app-specific cloud assets.
type App struct {
	Name ID `json:"name"`
	*Meta
}

type Component struct {
	Name ID `json:"name"` //  validate:"regexp=^[a-z]([-a-z0-9]*[a-z0-9])?$"

	CustomDeployScript *CustomDeployScript `json:"custom_deploy_script"`

	CurrentReleaseTimestamp ID `json:"current_release_id"`
	TargetReleaseTimestamp  ID `json:"target_release_id"`

	*Meta

	Addresses *ComponentAddresses `json:"addresses,omitempty" db:"-"`
}

type PortAddress struct {
	Port    string `json:"port"` // TODO really this should be the name of the port, which currently is the string of the number
	Address string `json:"address"`
}

type ComponentAddresses struct {
	External []*PortAddress `json:"external"`
	Internal []*PortAddress `json:"internal"`
}

type CustomDeployScript struct {
	Image   string   `json:"image"`
	Command []string `json:"command"`
	Timeout uint     `json:"timeout"`
}

// Volume
//==============================================================================
type VolumeBlueprint struct {
	Name ID     `json:"name"`
	Type string `json:"type"`
	Size int    `json:"size"`
}

// Container
//==============================================================================
type ContainerBlueprint struct {
	Image   string              `json:"image"`
	Name    string              `json:"name,omitempty"`
	Command []string            `json:"command,omitempty"`
	Ports   []*Port             `json:"ports"`
	Env     []*EnvVar           `json:"env"`
	CPU     *ResourceAllocation `json:"cpu"`
	RAM     *ResourceAllocation `json:"ram"`
	Mounts  []*Mount            `json:"mounts,omitempty"`
}

// EnvVar
//==============================================================================
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"` // this may be templated, "something_{{ instance_id }}"
}

// Mount
//==============================================================================
type Mount struct {
	Volume ID     `json:"volume"` // TODO should be VolumeName
	Path   string `json:"path"`
}

// Port
//==============================================================================
type Port struct {
	Protocol string `json:"protocol"`
	Number   int    `json:"number"`
	Public   bool   `json:"public"`

	// EntrypointDomain specifies which Entrypoint this Port is added to. Does not
	// apply when Public is false.
	EntrypointDomain ID `json:"entrypoint_domain,omitempty"`

	// ExternalNumber instructs the Entrypoint to set the actual Port number
	// specified as the external load balancer port. Does not apply when
	// EntrypointDomain is nil.
	ExternalNumber int `json:"external_number"`
}

// ResourceAllocation
//==============================================================================
type ResourceAllocation struct {
	Min uint `json:"min"`
	Max uint `json:"max"`
}

// Release
//==============================================================================
// NOTE the word Blueprint is used for Volumes and Containers, since they are
// both "definitions" that create "instances" of the real thing
type Release struct {
	// NOTE Timestamp here does not use the Timestamp type.
	Timestamp ID `json:"timestamp"`

	// InstanceGroup is used as a labeling mechanism for instances. If nil,
	// InstanceGroup is set equal to the release's Timestamp. If a value is
	// supplied by the user, it MUST be the current (previous) Release's
	// Timestamp.
	InstanceGroup ID `json:"instance_group"`

	InstanceCount int `json:"instance_count"`

	// These attributes, when changed from last Release, indicate a restart is
	// needed (or just new instances through other means).
	Volumes                []*VolumeBlueprint    `json:"volumes"`
	Containers             []*ContainerBlueprint `json:"containers"`
	TerminationGracePeriod int                   `json:"termination_grace_period"`

	// Retired defines whether or not a Release still has active assets, like pods
	// or services. When retired is true, we skip attempting to delete assets.
	Retired bool `json:"retired"`

	// Committed defines whether or not a Release is being / has been deployed.
	Committed bool `json:"committed"`

	*Meta
}

// Instance
//==============================================================================
type InstanceStatus string

const (
	InstanceStatusStopped InstanceStatus = "STOPPED"
	InstanceStatusStarted InstanceStatus = "STARTED"
)

// NOTE Instances are not stored in etcd, so the json tags here apply to HTTP
type Instance struct {
	ID ID `json:"id"` // actually just the number (starting w/ 1) of the instance order in the release

	// BaseName is the name of the instance without the Release ID appended. It is
	// used for naming volumes, which move between releases.
	BaseName string `json:"base_name"`
	Name     string `json:"name"`

	Status InstanceStatus `json:"status"`
}

// Entrypoint
//==============================================================================
type Entrypoint struct {
	Domain  ID     `json:"domain"`  // e.g. test.example.com
	Address string `json:"address"` // the ELB address

	// NOTE we actually don't need this -- we can always attach the policy, and enable per port
	// IPWhitelistEnabled bool   `json:"ip_whitelist_enabled"`

	*Meta
}

// Task
//==============================================================================
type Task struct {
	ID         ID     `json:"id"`
	ActionData string `json:"action_data"`

	Status      string `json:"status"`
	Attempts    int    `json:"attempts"`
	MaxAttempts int    `json:"max_attempts"` // this is static; config-level
	Error       string `json:"error"`

	*Meta
}

// ImageRegistry
//==============================================================================
type ImageRegistry struct {
	Name ID `json:"name"`

	// more to come soon... it's just Dockerhub for now

	*Meta
}

// ImageRepo
//==============================================================================
type ImageRepo struct {
	Name ID     `json:"name"`
	Key  string `json:"key,omitempty"`

	*Meta
}

// Node
//==============================================================================
// NOTE this is not to be confused with our concept of Resources like Apps and
// Components -- this is for CPU / RAM / disk.
type ResourceMetrics struct {
	Usage int `json:"usage"`
	Limit int `json:"limit"`
}

type Node struct {
	ID         ID     `json:"id"`
	Name       string `json:"name"`
	Class      string `json:"class"`
	ExternalIP string `json:"external_ip" db:"-"`

	// LaunchTime time.Time `json:"-"`
	// ServerUptime int              `json:"server_uptime" db:"-"`

	ProviderCreationTimestamp *Timestamp `json:"provider_creation_timestamp"`

	OutOfDisk bool             `json:"out_of_disk" db:"-"`
	Status    string           `json:"status" db:"-"`
	CPU       *ResourceMetrics `json:"cpu" db:"-"`
	RAM       *ResourceMetrics `json:"ram" db:"-"`

	*Meta
}
