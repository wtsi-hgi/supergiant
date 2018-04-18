package model

type KubeList struct {
	BaseList
	Items []*Kube `json:"items"`
}

// Kube objects contains global info about kubernetes ckusters.
type Kube struct {
	BaseModel

	// belongs_to CloudAccount
	CloudAccount     *CloudAccount `json:"cloud_account,omitempty" gorm:"ForeignKey:CloudAccountName;AssociationForeignKey:Name"`
	CloudAccountName string        `json:"cloud_account_name" validate:"nonzero" gorm:"not null;index" sg:"immutable"`

	// has_many Nodes
	Nodes     []*Node `json:"nodes,omitempty" gorm:"ForeignKey:KubeName;AssociationForeignKey:Name" sg:"store_as_json_in=NodesJSON"`
	NodesJSON []byte  `json:"-"`
	// has_many LoadBalancers
	LoadBalancers     []*LoadBalancer `json:"load_balancers,omitempty" gorm:"ForeignKey:KubeName;AssociationForeignKey:Name" sg:"store_as_json_in=LoadBalancersJSON"`
	LoadBalancersJSON []byte          `json:"-"`
	// has_many KubeResources
	KubeResources     []*KubeResource `json:"kube_resources,omitempty" gorm:"ForeignKey:KubeName;AssociationForeignKey:Name" sg:"store_as_json_in=KubeResourcesJSON"`
	KubeResourcesJSON []byte          `json:"-"`
	// has_many HelmReleases
	HelmReleases     []*HelmRelease `json:"helm_releases,omitempty" gorm:"ForeignKey:KubeName;AssociationForeignKey:Name" sg:"store_as_json_in=HelmReleasesJSON"`
	HelmReleasesJSON []byte         `json:"-"`

	Name string `json:"name" validate:"nonzero,max=12,regexp=^[a-z]([-a-z0-9]*[a-z0-9])?$" gorm:"not null;unique_index" sg:"immutable"`
	// Kubernetes
	KubernetesVersion string `json:"kubernetes_version" validate:"nonzero" sg:"default=1.8.7"`
	SSHPubKey         string `json:"ssh_pub_key"`
	ETCDDiscoveryURL  string `json:"etcd_discovery_url" sg:"readonly"`

	// Kubernetes Master
	MasterNodeSize     string   `json:"master_node_size" validate:"nonzero" sg:"immutable"`
	MasterID           string   `json:"master_id" sg:"readonly"`
	MasterPrivateIP    string   `json:"master_private_ip" sg:"readonly"`
	KubeMasterCount    int      `json:"kube_master_count"`
	MasterNodes        []string `json:"master_nodes" gorm:"-" sg:"store_as_json_in=MasterNodesJSON"`
	MasterNodesJSON    []byte   `json:"-"`
	MasterName         string   `json:"master_name" sg:"readonly"`
	CustomFiles        string   `json:"custom_files" sg:"readonly"`
	ProviderString     string   `json:"provider_string" sg:"readonly"`
	KubeProviderString string   `json:"Kube_provider_string" sg:"readonly"`
	ServiceString      string   `json:"service_string" sg:"readonly"`

	NodeSizes     []string `json:"node_sizes" gorm:"-" validate:"min=1" sg:"store_as_json_in=NodeSizesJSON"`
	NodeSizesJSON []byte   `json:"-" gorm:"not null"`

	Username string `json:"username" validate:"nonzero" sg:"immutable"`
	Password string `json:"password" validate:"nonzero" sg:"immutable"`

	RBACEnabled bool `json:"rbac_enabled"`

	HeapsterVersion          string `json:"heapster_version" validate:"nonzero" sg:"default=v1.4.0,immutable"`
	HeapsterMetricResolution string `json:"heapster_metric_resolution" validate:"regexp=^([0-9]+[smhd])+$" sg:"default=20s,immutable"`

	// NOTE due to how we marshal this as JSON, it's difficult to have this stored
	// as an interface, because unmarshalling causes us to lose the underlying
	// type. So, this is kindof like a whacky form of single-table inheritance.
	AWSConfig     *AWSKubeConfig `json:"aws_config,omitempty" gorm:"-" sg:"store_as_json_in=AWSConfigJSON,immutable"`
	AWSConfigJSON []byte         `json:"-"`

	DigitalOceanConfig     *DOKubeConfig `json:"digitalocean_config,omitempty" gorm:"-" sg:"store_as_json_in=DigitalOceanConfigJSON,immutable"`
	DigitalOceanConfigJSON []byte        `json:"-"`

	OpenStackConfig     *OSKubeConfig `json:"openstack_config,omitempty" gorm:"-" sg:"store_as_json_in=OpenStackConfigJSON,immutable"`
	OpenStackConfigJSON []byte        `json:"-"`

	GCEConfig     *GCEKubeConfig `json:"gce_config,omitempty" gorm:"-" sg:"store_as_json_in=GCEConfigJSON,immutable"`
	GCEConfigJSON []byte         `json:"-"`

	PACKConfig     *PACKKubeConfig `json:"packet_config,omitempty" gorm:"-" sg:"store_as_json_in=PACKConfigJSON,immutable"`
	PACKConfigJSON []byte          `json:"-"`

	MasterPublicIP string `json:"master_public_ip" sg:"readonly"`

	Ready bool `json:"ready" sg:"readonly" gorm:"index"`
	// This is used to store unstructured data such as metrics from Heapster.
	ExtraData     map[string]interface{} `json:"extra_data" gorm:"-" sg:"store_as_json_in=ExtraDataJSON,readonly"`
	ExtraDataJSON []byte                 `json:"-"`
}

// AWSKubeConfig holds aws specific information about AWS based KUbernetes clusters.
type AWSKubeConfig struct {
	Region           string `json:"region" validate:"nonzero,regexp=^[a-z]{2}-[a-z]+-[0-9]$"`
	AvailabilityZone string `json:"availability_zone"`
	VPCIPRange       string `json:"vpc_ip_range" validate:"nonzero" sg:"default=172.20.0.0/16"`
	// TODO this should be a slice of objects instead of maps, since we have a rigid key structure
	PublicSubnetIPRange []map[string]string `json:"public_subnet_ip_range"`
	MultiAZ             bool                `json:"multi_az"`
	BucketName          string              `json:"bucket_name,omitempty" sg:"readonly"`
	NodeVolumeSize      int                 `json:"node_volume_size" sg:"default=100"`
	MasterVolumeSize    int                 `json:"master_volume_size" sg:"default=100"`

	MasterRoleName                string            `json:"master_role"`
	NodeRoleName                  string            `json:"node_role"`
	Tags                          map[string]string `json:"tags"`
	LastSelectedAZ                string            `json:"last_selected_az" sg:"readonly"` // if using multiAZ this is the last az the node build used.
	PrivateKey                    string            `json:"private_key,omitempty" sg:"readonly"`
	VPCID                         string            `json:"vpc_id"`
	VPCMANAGED                    bool              `json:"vpc_managed"`
	InternetGatewayID             string            `json:"internet_gateway_id"`
	RouteTableID                  string            `json:"route_table_id"`
	RouteTableSubnetAssociationID []string          `json:"route_table_subnet_association_id" sg:"readonly"`
	PrivateNetwork                bool              `json:"private_network"`
	ELBSecurityGroupID            string            `json:"elb_security_group_id" sg:"readonly"`
	NodeSecurityGroupID           string            `json:"node_security_group_id" sg:"readonly"`
	ElasticFileSystemID           string            `json:"elastic_filesystem_id"`
	ElasticFileSystemTargets      []string          `json:"elastic_filesystem_targets" sg:"readonly"`
	BuildElasticFileSystem        bool              `json:"build_elastic_filesystem"`
}

// DOKubeConfig holds do specific information about DO based KUbernetes clusters.
type DOKubeConfig struct {
	Region            string   `json:"region" validate:"nonzero"`
	SSHKeyFingerprint []string `json:"ssh_key_fingerprint" validate:"nonzero"`
}

// OSKubeConfig holds do specific information about Open Stack based KUbernetes clusters.
type OSKubeConfig struct {
	Region             string `json:"region" validate:"nonzero"`
	PrivateSubnetRange string `json:"private_subnet_ip_range" validate:"nonzero" sg:"default=172.20.0.0/24"`
	PublicGatwayID     string `json:"public_gateway_id" validate:"nonzero" sg:"default=disabled"`

	NetworkID             string `json:"network_id" sg:"readonly"`
	SubnetID              string `json:"subnet_id" sg:"readonly"`
	RouterID              string `json:"router_id" sg:"readonly"`
	FloatingIPID          string `json:"floating_ip_id" sg:"readonly"`
	ImageName             string `json:"image_name" validate:"nonzero"`
	MasterSecurityGroupID string `json:"master_security_group_id" sg:"readonly"`
	NodeSecurityGroupID   string `json:"node_security_group_id" sg:"readonly"`
	KeyPair               string `json:"key_pair" sg:"readonly"`
	SSHPubKey             string `json:"ssh_pub_key" validate:"nonzero"`
}

// GCEKubeConfig holds do specific information about DO based KUbernetes clusters.
type GCEKubeConfig struct {
	Zone                string   `json:"zone" validate:"nonzero"`
	MasterInstanceGroup string   `json:"master_instance_group" sg:"readonly"`
	MinionInstanceGroup string   `json:"minion_instance_group" sg:"readonly"`
	MasterNodes         []string `json:"master_nodes" sg:"readonly"`
	MasterName          string   `json:"master_name" sg:"readonly"`
	KubeMasterCount     int      `json:"kube_master_count"`

	// Template vars
	SSHPubKey         string `json:"ssh_pub_key" validate:"nonzero"`
	KubernetesVersion string `json:"kubernetes_version" validate:"nonzero" sg:"default=1.5.1"`
	ETCDDiscoveryURL  string `json:"etcd_discovery_url" sg:"readonly"`
	MasterPrivateIP   string `json:"master_private_ip" sg:"readonly"`
}

type PACKKubeConfig struct {
	Project   string `json:"project" validate:"nonzero"`
	ProjectID string `json:"project_id" sg:"readonly"`
	Facility  string `json:"facility" validate:"nonzero"`
}
