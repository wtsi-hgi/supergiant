package core

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/supergiant/supergiant/common"
	"github.com/supergiant/supergiant/core/mock"
)

func TestComponentList(t *testing.T) {
	Convey("Given an ComponentCollection with 1 Component", t, func() {
		fakeEtcd := new(mock.FakeEtcd).ReturnOnList(
			[]string{
				`{
					"name": "component-test",
					"created": "Tue, 12 Apr 2016 03:54:56 UTC",
					"updated": null,
					"tags": {}
				}`,
			},
			nil,
		)
		core := newMockCore(fakeEtcd)

		app := core.Apps().New()
		app.Name = common.IDString("test")

		components := app.Components()

		Convey("When List() is called", func() {
			list, err := components.List()

			Convey("The return value should be an ComponentList with 1 Component", func() {
				expected := components.New()
				expected.Name = common.IDString("component-test")
				expected.Created = common.TimestampFromString("Tue, 12 Apr 2016 03:54:56 UTC")

				So(list.Items, ShouldHaveLength, 1)
				So(list.Items[0], ShouldResemble, expected)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestComponentCreate(t *testing.T) {
	Convey("Given an ComponentCollection and a new ComponentResource", t, func() {
		etcdKeyCreated := ""

		fakeEtcd := new(mock.FakeEtcd).OnCreate(func(key string, val string) error {
			etcdKeyCreated = key
			return nil
		})

		core := newMockCore(fakeEtcd)

		app := core.Apps().New()
		app.Name = common.IDString("test")

		components := app.Components()

		component := components.New()
		component.Name = common.IDString("component-test")

		Convey("When Create() is called", func() {
			err := components.Create(component)

			Convey("The Component should be created in etcd with a Created Timestamp", func() {
				So(etcdKeyCreated, ShouldEqual, "/supergiant/components/test/component-test")
				So(component.Created, ShouldHaveSameTypeAs, new(common.Timestamp))
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestComponentGet(t *testing.T) {
	Convey("Given an ComponentCollection with an ComponentResource", t, func() {
		fakeEtcd := new(mock.FakeEtcd).ReturnOnGet(
			`{
				"name": "component-test",
				"created": "Tue, 12 Apr 2016 03:54:56 UTC",
				"updated": null,
				"tags": {}
			}`,
			nil,
		)
		core := newMockCore(fakeEtcd)

		app := core.Apps().New()
		app.Name = common.IDString("test")

		components := app.Components()

		Convey("When Get() is called with the Component name", func() {
			expected := components.New()
			expected.Name = common.IDString("component-test")
			expected.Created = common.TimestampFromString("Tue, 12 Apr 2016 03:54:56 UTC")

			component, err := components.Get(expected.Name)

			Convey("The return value should be the ComponentResource", func() {
				So(component, ShouldResemble, expected)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestComponentUpdate(t *testing.T) {
	Convey("Given an ComponentCollection with an ComponentResource", t, func() {
		etcdKeyUpdated := ""

		fakeEtcd := new(mock.FakeEtcd).OnUpdate(func(key string, val string) error {
			etcdKeyUpdated = key
			return nil
		})
		core := newMockCore(fakeEtcd)

		app := core.Apps().New()
		app.Name = common.IDString("test")

		component := app.Components().New()
		component.Name = common.IDString("component-test")

		Convey("When Update() is called", func() {
			err := component.Update()

			Convey("The Component should be updated in etcd with an Updated Timestamp", func() {
				So(etcdKeyUpdated, ShouldEqual, "/supergiant/components/test/component-test")
				So(component.Updated, ShouldHaveSameTypeAs, new(common.Timestamp))
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestComponentDelete(t *testing.T) {
	Convey("Given an ComponentCollection with an ComponentResource", t, func() {
		etcdKeyDeleted := ""
		releaseDeleted := ""

		fakeEtcd := new(mock.FakeEtcd).OnDelete(func(key string) error {
			etcdKeyDeleted = key
			return nil
		})
		core := newMockCore(fakeEtcd)

		app := core.Apps().New()
		app.Name = common.IDString("test")

		component := app.Components().New()
		component.Name = common.IDString("component-test")

		fakeReleases := &FakeReleaseCollection{
			component: component,
			core:      core,
		}

		fakeReleases.ReturnOnList([]*common.Release{
			&common.Release{
				Timestamp: common.IDString("20160412035456"),
			},
		})
		fakeReleases.OnDelete(func(r *ReleaseResource) error {
			releaseDeleted = *r.Timestamp
			return nil
		})
		component.ReleasesInterface = fakeReleases

		Convey("When Delete() is called", func() {
			err := component.Delete()

			Convey("The Component should be deleted in etcd", func() {
				So(err, ShouldBeNil)
				So(etcdKeyDeleted, ShouldEqual, "/supergiant/components/test/component-test")
			})

			Convey("The Component's Releases should be deleted", func() {
				So(releaseDeleted, ShouldEqual, "20160412035456")
			})
		})
	})
}

// Mock

func (f *FakeReleaseCollection) ReturnOnList(components []*common.Release) *FakeReleaseCollection {
	var items []*ReleaseResource
	for _, component := range components {
		items = append(items, &ReleaseResource{
			core:       f.core,
			collection: f,
			Release:    component,
		})
	}
	f.ListFn = func() (*ReleaseList, error) {
		return &ReleaseList{Items: items}, nil
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

func (f *FakeReleaseCollection) Delete(r *ReleaseResource) error {
	return f.DeleteFn(r)
}
