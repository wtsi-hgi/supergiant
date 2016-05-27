package core

import (
	"github.com/supergiant/supergiant/common"
	"github.com/supergiant/supergiant/core/mock"
)

func newMockCore(fakeEtcd *mock.FakeEtcd) *Core {
	return &Core{
		db: &db{
			&etcdClient{fakeEtcd},
		},
	}
}

// Components
//==============================================================================

func (f *FakeComponentCollection) ReturnValuesOnList(components []*common.Component) *FakeComponentCollection {
	var items []*ComponentResource
	for _, component := range components {
		items = append(items, &ComponentResource{
			core:       f.core,
			collection: f,
			Component:  component,
		})
	}
	f.ListFn = func() (*ComponentList, error) {
		return &ComponentList{Items: items}, nil
	}
	return f
}

func (f *FakeComponentCollection) OnDelete(clbk func(Resource) error) *FakeComponentCollection {
	f.DeleteFn = func(r Resource) error {
		return clbk(r)
	}
	return f
}

type FakeComponentCollection struct {
	core     *Core
	app      *AppResource
	ListFn   func() (*ComponentList, error)
	NewFn    func() *ComponentResource
	CreateFn func() error
	GetFn    func() (*ComponentResource, error)
	UpdateFn func() error
	PatchFn  func() error
	DeleteFn func(Resource) error
	DeployFn func(Resource) error
}

func (f *FakeComponentCollection) App() *AppResource {
	return f.app
}

func (f *FakeComponentCollection) List() (*ComponentList, error) {
	return f.ListFn()
}

func (f *FakeComponentCollection) New() *ComponentResource {
	return f.NewFn()
}

func (f *FakeComponentCollection) Create(*ComponentResource) error {
	return f.CreateFn()
}

func (f *FakeComponentCollection) Get(common.ID) (*ComponentResource, error) {
	return f.GetFn()
}

func (f *FakeComponentCollection) Update(common.ID, *ComponentResource) error {
	return f.UpdateFn()
}

func (f *FakeComponentCollection) Patch(common.ID, *ComponentResource) error {
	return f.PatchFn()
}

func (f *FakeComponentCollection) Delete(r Resource) error {
	return f.DeleteFn(r)
}

func (f *FakeComponentCollection) Deploy(r Resource) error {
	return f.DeployFn(r)
}

// ImageRepos
//==============================================================================

func (f *FakeImageRepoCollection) ReturnOnGet(t *common.ImageRepo, err error) *FakeImageRepoCollection {
	f.GetFn = func() (*ImageRepoResource, error) {
		if err != nil {
			return nil, err
		}
		return &ImageRepoResource{ImageRepo: t}, nil
	}
	return f
}

func (f *FakeImageRepoCollection) OnDelete(clbk func(*ImageRepoResource) error) *FakeImageRepoCollection {
	f.DeleteFn = func(r *ImageRepoResource) error {
		return clbk(r)
	}
	return f
}

type FakeImageRepoCollection struct {
	core     *Core
	ListFn   func() (*ImageRepoList, error)
	NewFn    func() *ImageRepoResource
	CreateFn func() error
	GetFn    func() (*ImageRepoResource, error)
	UpdateFn func() error
	PatchFn  func() error
	DeleteFn func(*ImageRepoResource) error
}

func (f *FakeImageRepoCollection) List() (*ImageRepoList, error) {
	return f.ListFn()
}

func (f *FakeImageRepoCollection) New() *ImageRepoResource {
	return f.NewFn()
}

func (f *FakeImageRepoCollection) Create(*ImageRepoResource) error {
	return f.CreateFn()
}

func (f *FakeImageRepoCollection) Get(common.ID) (*ImageRepoResource, error) {
	return f.GetFn()
}

func (f *FakeImageRepoCollection) Update(common.ID, *ImageRepoResource) error {
	return f.UpdateFn()
}

func (f *FakeImageRepoCollection) Patch(common.ID, *ImageRepoResource) error {
	return f.PatchFn()
}

func (f *FakeImageRepoCollection) Delete(r *ImageRepoResource) error {
	return f.DeleteFn(r)
}

// Releases
//==============================================================================

func (f *FakeReleaseCollection) ReturnValuesOnList(ts []*common.Release) *FakeReleaseCollection {
	var items []*ReleaseResource
	for _, t := range ts {
		items = append(items, &ReleaseResource{
			core:       f.core,
			collection: f,
			Release:    t,
		})
	}
	f.ListFn = func() (*ReleaseList, error) {
		return &ReleaseList{Items: items}, nil
	}
	return f
}

func (f *FakeReleaseCollection) ReturnOnGet(t *common.Release, err error) *FakeReleaseCollection {
	f.GetFn = func() (*ReleaseResource, error) {
		if err != nil {
			return nil, err
		}
		return &ReleaseResource{
			core:       f.core,
			collection: f,
			Release:    t,
		}, nil
	}
	return f
}

func (f *FakeReleaseCollection) OnDelete(clbk func(*ReleaseResource) error) *FakeReleaseCollection {
	f.DeleteFn = func(r *ReleaseResource) error {
		return clbk(r)
	}
	return f
}

type FakeReleaseCollection struct {
	core          *Core
	component     *ComponentResource
	ListFn        func() (*ReleaseList, error)
	NewFn         func() *ReleaseResource
	CreateFn      func() error
	MergeCreateFn func() error
	GetFn         func() (*ReleaseResource, error)
	UpdateFn      func() error
	PatchFn       func() error
	DeleteFn      func(*ReleaseResource) error
}

func (f *FakeReleaseCollection) Component() *ComponentResource {
	return f.component
}

func (f *FakeReleaseCollection) List() (*ReleaseList, error) {
	return f.ListFn()
}

func (f *FakeReleaseCollection) New() *ReleaseResource {
	return f.NewFn()
}

func (f *FakeReleaseCollection) Create(*ReleaseResource) error {
	return f.CreateFn()
}

func (f *FakeReleaseCollection) MergeCreate(*ReleaseResource) error {
	return f.MergeCreateFn()
}

func (f *FakeReleaseCollection) Get(common.ID) (*ReleaseResource, error) {
	return f.GetFn()
}

func (f *FakeReleaseCollection) Update(common.ID, *ReleaseResource) error {
	return f.UpdateFn()
}

func (f *FakeReleaseCollection) Patch(common.ID, *ReleaseResource) error {
	return f.PatchFn()
}

func (f *FakeReleaseCollection) Delete(r *ReleaseResource) error {
	return f.DeleteFn(r)
}

// Entrypoints
//==============================================================================

func (f *FakeEntrypointCollection) ReturnOnGet(t *common.Entrypoint, err error) *FakeEntrypointCollection {
	f.GetFn = func() (*EntrypointResource, error) {
		if err != nil {
			return nil, err
		}
		return &EntrypointResource{
			core:       f.core,
			collection: f,
			Entrypoint: t,
		}, nil
	}
	return f
}

type FakeEntrypointCollection struct {
	core     *Core
	ListFn   func() (*EntrypointList, error)
	NewFn    func() *EntrypointResource
	CreateFn func() error
	GetFn    func() (*EntrypointResource, error)
	UpdateFn func() error
	PatchFn  func() error
	DeleteFn func(*EntrypointResource) error
}

func (f *FakeEntrypointCollection) List() (*EntrypointList, error) {
	return f.ListFn()
}

func (f *FakeEntrypointCollection) New() *EntrypointResource {
	return f.NewFn()
}

func (f *FakeEntrypointCollection) Create(*EntrypointResource) error {
	return f.CreateFn()
}

func (f *FakeEntrypointCollection) Get(common.ID) (*EntrypointResource, error) {
	return f.GetFn()
}

func (f *FakeEntrypointCollection) Update(common.ID, *EntrypointResource) error {
	return f.UpdateFn()
}

func (f *FakeEntrypointCollection) Patch(common.ID, *EntrypointResource) error {
	return f.PatchFn()
}

func (f *FakeEntrypointCollection) Delete(r *EntrypointResource) error {
	return f.DeleteFn(r)
}

// Instances
//==============================================================================

func (f *FakeInstanceCollection) ReturnValuesOnList(ts []*common.Instance) *FakeInstanceCollection {
	var items []*InstanceResource
	for _, t := range ts {
		items = append(items, &InstanceResource{
			core:       f.core,
			collection: f,
			Instance:   t,
		})
	}
	f.ListFn = func() *InstanceList {
		return &InstanceList{Items: items}
	}
	return f
}

func (f *FakeInstanceCollection) OnDelete(clbk func(*InstanceResource) error) *FakeInstanceCollection {
	f.DeleteFn = func(r *InstanceResource) error {
		return clbk(r)
	}
	return f
}

type FakeInstanceCollection struct {
	core     *Core
	release  *ReleaseResource
	ListFn   func() *InstanceList
	NewFn    func(common.ID) *InstanceResource
	GetFn    func(common.ID) (*InstanceResource, error)
	StartFn  func(Resource) error
	StopFn   func(Resource) error
	DeleteFn func(*InstanceResource) error
}

func (f *FakeInstanceCollection) App() *AppResource {
	return f.release.Component().App()
}

func (f *FakeInstanceCollection) Component() *ComponentResource {
	return f.release.Component()
}

func (f *FakeInstanceCollection) Release() *ReleaseResource {
	return f.release
}

func (f *FakeInstanceCollection) List() *InstanceList {
	return f.ListFn()
}

func (f *FakeInstanceCollection) New(id common.ID) *InstanceResource {
	return f.NewFn(id)
}

func (f *FakeInstanceCollection) Get(id common.ID) (*InstanceResource, error) {
	return f.GetFn(id)
}

func (f *FakeInstanceCollection) Start(ri Resource) error {
	return f.StartFn(ri)
}

func (f *FakeInstanceCollection) Stop(ri Resource) error {
	return f.StopFn(ri)
}

func (f *FakeInstanceCollection) Delete(r *InstanceResource) error {
	return f.DeleteFn(r)
}
