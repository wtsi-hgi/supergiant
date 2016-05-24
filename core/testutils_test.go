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
