package common

// App is the main top-level Resource within Supergiant, acting as a logical
// namespace for Components and all of their controlled assets -- along with
// provisioning an actual Kubernetes Namespace, it is used as a base name for
// app-specific cloud assets.
type App struct {
	Name ID `json:"name" validate:"nonzero,max=24,regexp=^[a-z]([-a-z0-9]*[a-z0-9])?$"`
	*Meta
}

type Component struct {
	Name ID `json:"name" validate:"nonzero,max=24,regexp=^[a-z]([-a-z0-9]*[a-z0-9])?$"`

	CustomDeployScript *CustomDeployScript `json:"custom_deploy_script"`

	CurrentReleaseTimestamp ID `json:"current_release_id" sg:"readonly"` // TODO this should be release_timestamp, not release_id
	TargetReleaseTimestamp  ID `json:"target_release_id" sg:"readonly"`

	Addresses *ComponentAddresses `json:"addresses,omitempty" sg:"readonly,nostore"`

	*Meta
}

// NOTE the word Blueprint is used for Volumes and Containers, since they are
// both "definitions" that create "instances" of the real thing
type Release struct {
	// NOTE Timestamp here does not use the Timestamp type.
	Timestamp ID `json:"timestamp"`

	// TODO
	// InstanceGroup is used as a labeling mechanism for instances. If nil,
	// InstanceGroup is set equal to the release's Timestamp. If a value is
	// supplied by the user, it MUST be the current (previous) Release's
	// Timestamp.
	//
	// The purpose of InstanceGroup is to prevent restarting between Releases.
	// I'm pretty sure the ONLY scenario in which this value makes sense is when
	// changing InstanceCount, and the value supplied in such a scenario must be
	// the previous Release's timestamp.
	//
	// It seems as though we need to break deploys up into:
	//   - config changes
	//	 - adding/removing instances
	//
	// It might still makes sense to have Release as a grouping mechanism, though,
	// because it could allow for grouping metrics recorded per-Release. However,
	// you may be able to separate the operations, but have every operation create
	// a new record that records the config / instance count at that time.
	InstanceGroup ID `json:"instance_group"`

	InstanceCount int `json:"instance_count" validate:"min=1" sg:"default=1"`

	// These attributes, when changed from last Release, indicate a restart is
	// needed (or just new instances through other means).
	Volumes                []*VolumeBlueprint    `json:"volumes"`
	Containers             []*ContainerBlueprint `json:"containers" validate:"min=1"`
	TerminationGracePeriod int                   `json:"termination_grace_period" validate:"min=0" sg:"default=10"`

	// Retired defines whether or not a Release still has active assets, like pods
	// or services. When retired is true, we skip attempting to delete assets.
	Retired bool `json:"retired" sg:"readonly"`

	// Committed defines whether or not a Release is being / has been deployed.
	Committed bool `json:"committed" sg:"readonly"`

	*Meta
}

// NOTE Instances are not stored in etcd, so the json tags here apply to HTTP
type Instance struct {
	ID ID `json:"id"` // actually just the number (starting w/ 1) of the instance order in the release

	// BaseName is the name of the instance without the Release ID appended. It is
	// used for naming volumes, which move between releases.
	BaseName string `json:"base_name"`
	Name     string `json:"name"`

	Status string `json:"status"`

	CPU *ResourceMetrics `json:"cpu"`
	RAM *ResourceMetrics `json:"ram"`
}

type Entrypoint struct {
	// NOTE eventually the plan for Domain is to use it for DNS application.
	// That's not hooked up yet, and thus, we aren't doing any type of validation
	// on hostnames.
	Domain ID `json:"domain" validate:"nonzero"` // e.g. test.example.com

	// the ELB address
	Address string `json:"address" sg:"readonly"`

	// NOTE we actually don't need this -- we can always attach the policy, and enable per port
	// IPWhitelistEnabled bool   `json:"ip_whitelist_enabled"`

	*Meta
}

type Task struct {
	ID         ID     `json:"id"`
	ActionData string `json:"action_data" validate:"nonzero"`

	MaxAttempts int `json:"max_attempts" validate:"min=1" sg:"default=10"`

	Status   string `json:"status" sg:"readonly"`
	Attempts int    `json:"attempts" sg:"readonly"`
	Error    string `json:"error" sg:"readonly"`

	*Meta
}

type ImageRegistry struct {
	Name ID `json:"name"`

	// more to come soon... it's just Dockerhub for now

	*Meta
}

// TODO this should really be renamed to Org probably
type ImageRepo struct {
	Name ID     `json:"name" validate:"nonzero"`
	Key  string `json:"key" validate:"nonzero" sg:"private"`

	*Meta
}

type Node struct {
	ID   ID     `json:"id"`
	Name string `json:"name" sg:"readonly"`
	// This is the only input for Node
	Class      string `json:"class" validate:"nonzero"`
	ExternalIP string `json:"external_ip" sg:"readonly"`

	// LaunchTime time.Time `json:"-"`
	// ServerUptime int              `json:"server_uptime" db:"-"`

	ProviderCreationTimestamp *Timestamp `json:"provider_creation_timestamp" sg:"readonly"`

	OutOfDisk bool             `json:"out_of_disk" sg:"readonly,nostore"`
	Status    string           `json:"status" sg:"readonly,nostore"`
	CPU       *ResourceMetrics `json:"cpu" sg:"readonly,nostore"`
	RAM       *ResourceMetrics `json:"ram" sg:"readonly,nostore"`

	*Meta
}

//==============================================================================

// ID is defined as a string pointer in order to check for nil in the context of
// relations. NOTE this may not be best practice.
type ID *string

// Tags is a meta field for holding unstructued key/val info.
type Tags map[string]string

// Meta is a set of fields on all Resources for holding metadata.
type Meta struct {
	Created *Timestamp `json:"created" sg:"readonly"`
	Updated *Timestamp `json:"updated" sg:"readonly"`
	Tags    Tags       `json:"tags"`
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
	Image   string   `json:"image" validate:"nonzero,regexp=^[-\\w\\.\\/]+(:[-\\w\\.]+)?$"`
	Command []string `json:"command"` // TODO need validation here, I think we need to reqire command
	Timeout uint     `json:"timeout" sg:"default=1800"`
}

type VolumeBlueprint struct {
	Name ID     `json:"name" validate:"nonzero,regexp=^\\w[-\\w\\.]*$/"` // TODO max length
	Type string `json:"type" validate:"regexp=^(gp2)$" sg:"default=gp2"` // TODO support other vol types
	Size int    `json:"size" validate:"min=1"`
}

type ContainerBlueprint struct {
	Image   string         `json:"image" validate:"nonzero,regexp=^[-\\w\\.\\/]+(:[-\\w\\.]+)?$"`
	Name    string         `json:"name,omitempty" validate:"regexp=^[-\\w\\.\\/]+(:[-\\w\\.]+)?$"`
	Command []string       `json:"command,omitempty"`
	Ports   []*Port        `json:"ports,omitempty"`
	Env     []*EnvVar      `json:"env,omitempty"`
	CPU     *CpuAllocation `json:"cpu" validate:"nonzero"`
	RAM     *RamAllocation `json:"ram" validate:"nonzero"`
	Mounts  []*Mount       `json:"mounts,omitempty"`
}

type EnvVar struct {
	Name  string `json:"name" validate:"nonzero"`
	Value string `json:"value" validate:"nonzero"` // this may be templated, "something_{{ instance_id }}"
}

type Mount struct {
	Volume ID     `json:"volume" validate:"nonzero"` // TODO should be VolumeName
	Path   string `json:"path" validate:"nonzero"`
}

type Port struct {
	// TODO Kube only accepts TCP|UDP for protocol values, but we accept values
	// like HTTP, which are used to display component addresses. We should either
	// build a map defining the accepted application protocols on top of TCP|UDP,
	// or make a sep. field.
	Protocol string `json:"protocol" validate:"nonzero" sg:"default=TCP"`
	Number   int    `json:"number" validate:"nonzero,max=40000"`
	Public   bool   `json:"public"`

	// EntrypointDomain specifies which Entrypoint this Port is added to. Does not
	// apply when Public is false.
	EntrypointDomain ID `json:"entrypoint_domain,omitempty"`

	// ExternalNumber instructs the Entrypoint to set the actual Port number
	// specified as the external load balancer port. Does not apply when
	// EntrypointDomain is nil.
	//
	// TODO validation needed just like on Number, but it can't be nonzero since
	// the value provided can be 0.
	ExternalNumber int `json:"external_number"`
}

type CpuAllocation struct {
	Min *CoresValue `json:"min"` // NOTE validations removed since they are not numerical values anymore
	Max *CoresValue `json:"max"`
}

type RamAllocation struct {
	Min *BytesValue `json:"min"`
	Max *BytesValue `json:"max"`
}

const (
	InstanceStatusStopped = "STOPPED"
	InstanceStatusStarted = "STARTED"
)

// NOTE this is not to be confused with our concept of Resources like Apps and
// Components -- this is for CPU / RAM / disk.
type ResourceMetrics struct {
	Usage int `json:"usage"`
	Limit int `json:"limit"`
}
