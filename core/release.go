package core

import (
	"errors"
	"fmt"
	"time"

	"github.com/imdario/mergo"
	"github.com/supergiant/guber"
	"github.com/supergiant/supergiant/common"
)

type ReleasesInterface interface {
	Component() *ComponentResource

	List() (*ReleaseList, error)
	New() *ReleaseResource
	Create(*ReleaseResource) error
	MergeCreate(*ReleaseResource) error
	Get(common.ID) (*ReleaseResource, error)
	Update(common.ID, *ReleaseResource) error
	Patch(common.ID, *ReleaseResource) error
	Delete(*ReleaseResource) error
}

type ReleaseCollection struct {
	core      *Core
	component *ComponentResource
}

type ReleaseResource struct {
	core       *Core
	collection ReleasesInterface
	*common.Release

	// Relations
	InstancesInterface InstancesInterface `json:"-"`

	serviceSet *ServiceSet

	imageRepos  []*ImageRepoResource
	entrypoints map[string]*EntrypointResource // a map for quick lookup
}

type ReleaseList struct {
	Items []*ReleaseResource `json:"items"`
}

// initializeResource implements the Collection interface.
func (c *ReleaseCollection) initializeResource(in Resource) {
	r := in.(*ReleaseResource)
	r.collection = c
	r.core = c.core
	if r.InstancesInterface == nil { // don't want to reset for testing purposes
		r.InstancesInterface = &InstanceCollection{
			core:    c.core,
			release: r,
		}
	}
}

func (c *ReleaseCollection) Component() *ComponentResource {
	return c.component
}

// List returns an ReleaseList.
func (c *ReleaseCollection) List() (*ReleaseList, error) {
	list := new(ReleaseList)
	err := c.core.db.list(c, list)
	return list, err
}

// New initializes an Release with a pointer to the Collection.
func (c *ReleaseCollection) New() *ReleaseResource {
	r := &ReleaseResource{
		Release: &common.Release{
			Meta: common.NewMeta(),
		},
	}
	c.initializeResource(r)
	return r
}

// Create takes an Release and creates it in etcd.
func (c *ReleaseCollection) Create(r *ReleaseResource) error {
	if c.Component().TargetReleaseTimestamp != nil {
		return errors.New("Component already has a target Release")
	}

	r.Timestamp = newReleaseTimestamp()
	if r.InstanceGroup == nil {
		r.InstanceGroup = r.Timestamp
	} else if *r.InstanceGroup != *r.Timestamp && *r.InstanceGroup != *c.Component().CurrentReleaseTimestamp {
		return errors.New("Release InstanceGroup field can only be set to either the current or target Release's Timestamp value.")
	}

	if err := c.core.db.create(c, r.Timestamp, r); err != nil {
		return err
	}

	c.Component().TargetReleaseTimestamp = r.Timestamp
	return c.Component().Update()
}

// MergeCreate creates a Release by taking a new Release and merging it with the
// Component's current Release.
func (c *ReleaseCollection) MergeCreate(r *ReleaseResource) error {
	if c.Component().CurrentReleaseTimestamp == nil {
		return errors.New("Attempting MergeCreate with no current Release")
	}

	current, err := c.Component().CurrentRelease()
	if err != nil {
		return err
	}

	if err := mergo.Merge(r, *current); err != nil {
		return err
	}

	// TODO
	r.InstanceGroup = nil
	r.Committed = false
	r.Created = nil
	r.Updated = nil

	return c.Create(r)
}

// Get takes an id and returns an ReleaseResource if it exists.
func (c *ReleaseCollection) Get(timestamp common.ID) (*ReleaseResource, error) {
	r := c.New()
	if err := c.core.db.get(c, timestamp, r); err != nil {
		return nil, err
	}
	return r, nil
}

// Update saves the Release in etcd through an update.
func (c *ReleaseCollection) Update(timestamp common.ID, r *ReleaseResource) error {
	return c.core.db.update(c, timestamp, r)
}

// Patch partially updates the App in etcd.
func (c *ReleaseCollection) Patch(name common.ID, r *ReleaseResource) error {
	return c.core.db.patch(c, name, r)
}

// Delete removes all assets (volumes, pods, etc.) and deletes the Release in
// etcd.
func (c *ReleaseCollection) Delete(r *ReleaseResource) error {

	// NOTE we attempt deletes on all Release instances just in case we have
	// lingering pods. This wouldn't be an issue if we deleted namespace first,
	// but there's issues with that.
	ch := make(chan error)
	for _, instance := range r.Instances().List().Items {
		go func(instance *InstanceResource) {
			ch <- instance.Delete()
		}(instance)
	}
	for i := 0; i < r.InstanceCount; i++ {
		if err := <-ch; err != nil {
			return err
		}
	}

	// NOTE this part of the delete, as opposed to abovewith Instances, deals
	// with resources shared between Releases.
	if r.Committed && !r.Retired {
		if err := r.serviceSet.delete(); err != nil {
			return err
		}

		for _, instance := range r.Instances().List().Items {
			if err := r.Instances().DeleteVolumes(instance); err != nil {
				return err
			}
		}
	}

	// TODO sloppy
	targetTimestamp := r.Component().TargetReleaseTimestamp
	currentTimestamp := r.Component().CurrentReleaseTimestamp
	if targetTimestamp != nil && *r.Timestamp == *targetTimestamp {
		r.Component().TargetReleaseTimestamp = nil
		r.Component().Update()
	} else if currentTimestamp != nil && *r.Timestamp == *currentTimestamp {
		r.Component().CurrentReleaseTimestamp = nil
		r.Component().Update()
	}

	return r.core.db.delete(c, r.Timestamp)
}

//------------------------------------------------------------------------------

// Key implements the Locatable interface.
func (c *ReleaseCollection) locationKey() string {
	return "releases"
}

// Parent implements the Locatable interface.
func (c *ReleaseCollection) parent() Locatable {
	return c.component
}

// Child implements the Locatable interface.
func (c *ReleaseCollection) child(key string) Locatable {
	r, err := c.Get(common.IDString(key))
	if err != nil {
		panic(fmt.Errorf("No child with key %s for %T", key, c))
	}
	return r
}

// Key implements the Locatable interface.
func (r *ReleaseResource) locationKey() string {
	return common.StringID(r.Timestamp)
}

// Parent implements the Locatable interface.
func (r *ReleaseResource) parent() Locatable {
	return r.collection.(Locatable)
}

// Child implements the Locatable interface.
func (r *ReleaseResource) child(key string) (l Locatable) {
	switch key {
	case "instances":
		l = r.Instances().(Locatable)
	default:
		panic(fmt.Errorf("No child with key %s for %T", key, r))
	}
	return
}

//------------------------------------------------------------------------------

// decorate implements the Resource interface
func (r *ReleaseResource) decorate() error {
	repos, err := r.getImageRepos()
	if err != nil {
		return err
	}
	r.imageRepos = repos

	var oldServiceSet *ServiceSet
	if r.Component().TargetReleaseTimestamp != nil && r.Component().CurrentReleaseTimestamp != nil && *r.Component().TargetReleaseTimestamp == *r.Timestamp {
		currentRelease, err := r.Component().CurrentRelease()
		if err != nil {
			return err
		}
		oldServiceSet = currentRelease.serviceSet
	}

	baseName := common.StringID(r.Component().Name)
	r.serviceSet = &ServiceSet{
		core:          r.core,
		release:       r,
		namespace:     common.StringID(r.App().Name),
		baseName:      baseName,
		labelSelector: map[string]string{"service": baseName},
		previous:      oldServiceSet,
	}

	r.entrypoints, err = r.getEntrypoints()
	if err != nil {
		return err
	}

	return nil
}

// Update is a proxy method to ReleaseCollection's Update.
func (r *ReleaseResource) Update() error {
	return r.collection.Update(r.Timestamp, r)
}

// Patch is a proxy method to collection Patch.
func (r *ReleaseResource) Patch() error {
	return r.collection.Patch(r.Timestamp, r)
}

// Delete is a proxy method to ReleaseCollection's Delete.
func (r *ReleaseResource) Delete() error {
	return r.collection.Delete(r)
}

func newReleaseTimestamp() common.ID {
	stamp := time.Now().Format("20060102150405")
	return &stamp
}

func (r *ReleaseResource) App() *AppResource {
	return r.Component().App()
}

func (r *ReleaseResource) Component() *ComponentResource {
	return r.collection.Component()
}

func (r *ReleaseResource) Instances() InstancesInterface {
	return r.InstancesInterface
}

func (r *ReleaseResource) IsStarted() bool {
	for _, instance := range r.Instances().List().Items {
		if !instance.IsStarted() {
			return false
		}
	}
	return true
}

func (r *ReleaseResource) IsStopped() bool {
	for _, instance := range r.Instances().List().Items {
		if !instance.IsStopped() {
			return false
		}
	}
	return true
}

func (r *ReleaseResource) imageRepoNames() (repoNames []string) { // TODO convert Image into Value object w/ repo, image, version
	for _, container := range r.Containers {
		if repoName := imageRepoName(container); repoName != "" {
			repoNames = append(repoNames, repoName)
		}
	}
	return uniqStrs(repoNames)
}

func (r *ReleaseResource) getEntrypoints() (map[string]*EntrypointResource, error) { // TODO convert Image into Value object w/ repo, image, version
	entrypoints := make(map[string]*EntrypointResource)
	for _, port := range r.serviceSet.externalPortDefs() {
		if port.EntrypointDomain == nil {
			continue
		}
		entrypoint, err := r.core.Entrypoints().Get(port.EntrypointDomain)
		if err != nil {

			// TODO
			if isEtcdNotFoundErr(err) {
				Log.Errorf("Entrypoint %s does not exist", *port.EntrypointDomain)
				continue
			}

			return nil, err
		}
		entrypoints[*port.EntrypointDomain] = entrypoint
	}
	return entrypoints, nil
}

// Operations-------------------------------------------------------------------

// NOTE it seems weird here, but "Provision" == "CreateUnlessExists"
func (r *ReleaseResource) provisionSecrets() error {
	for _, repo := range r.imageRepos {
		if err := r.provisionSecret(repo); err != nil {
			return err
		}
	}
	return nil
}

func (r *ReleaseResource) provisionSecret(repo *ImageRepoResource) error {
	// TODO why don't we just do a call directly to Create, and then check the error?
	_, err := r.core.k8s.Secrets(common.StringID(r.App().Name)).Get(common.StringID(repo.Name))
	if err == nil { // Secret already exists
		return nil
	} else if !isKubeNotFoundErr(err) { // to continue, it has to be a not found error
		return err
	}

	_, err = r.core.k8s.Secrets(common.StringID(r.App().Name)).Create(asKubeSecret(repo))
	return err
}

// Provision creates needed assets for all instances. It does not actually
// start instances -- that is handled by deploy.go.
func (r *ReleaseResource) Provision() error {
	if err := r.provisionSecrets(); err != nil {
		return err
	}

	if err := r.serviceSet.provision(); err != nil {
		return err
	}

	// Concurrently provision volumes
	// which is not actually concurrent... just sends all requests, and then
	// loops waiting, which prevents concurrently polling while waiting.
	var newVols []*AwsVolume
	for _, vol := range r.volumes() {
		if ok, err := vol.Exists(); err != nil {
			return err
		} else if !ok {
			if err = vol.Create(); err != nil {
				return err
			}
			newVols = append(newVols, vol)
		}
	}
	for _, vol := range newVols {
		if err := vol.waitForAvailable(); err != nil {
			return err
		}
	}

	return nil
}

func (r *ReleaseResource) volumes() (vols []*AwsVolume) {
	for _, instance := range r.Instances().List().Items {
		vols = append(vols, instance.Volumes()...)
	}
	return vols
}

func (r *ReleaseResource) getImageRepos() (repos []*ImageRepoResource, err error) { // Not returning ImageRepoResource, since they are defined before hand
	for _, repoName := range r.imageRepoNames() {
		repo, err := r.core.ImageRepos().Get(&repoName)
		if err != nil {

			if isEtcdNotFoundErr(err) {
				// if there is no repo, we can assume this is a public repo (though it
				// may not be) -- this represents a TODO on how to report errors from
				// Kubernetes
				continue
			}

			return nil, err
		}
		repos = append(repos, repo)
	}
	return repos, nil
}

// TODO naming inconsistencies for kube definitions of resources
// ImagePullSecrets returns repo names defined for Kube pods
func (r *ReleaseResource) ImagePullSecrets() (pullSecrets []*guber.ImagePullSecret, err error) { // TODO don't need to return error here it seems
	for _, repo := range r.imageRepos {
		pullSecrets = append(pullSecrets, asKubeImagePullSecret(repo))
	}
	return pullSecrets, nil
}
