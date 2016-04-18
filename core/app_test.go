package core

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/supergiant/guber"
	"github.com/supergiant/supergiant/common"
	"github.com/supergiant/supergiant/core/mock"
)

func TestAppList(t *testing.T) {
	Convey("Given an AppCollection with 1 App", t, func() {
		fakeEtcd := new(mock.FakeEtcd).ReturnValuesOnGet(
			[]string{
				`{
					"name": "test",
					"created": "Tue, 12 Apr 2016 03:54:56 UTC",
					"updated": null,
					"tags": {}
				}`,
			},
			nil,
		)
		core := newMockCore(fakeEtcd)
		apps := core.Apps()

		Convey("When List() is called", func() {
			list, err := apps.List()

			Convey("The return value should be an AppList with 1 App", func() {
				expected := apps.New()
				expected.Name = common.IDString("test")
				expected.Created = common.TimestampFromString("Tue, 12 Apr 2016 03:54:56 UTC")

				So(list.Items, ShouldHaveLength, 1)
				So(list.Items[0], ShouldResemble, expected)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestAppCreate(t *testing.T) {
	Convey("Given an AppCollection and a new AppResource", t, func() {
		etcdKeyCreated := ""
		namespaceCreated := ""

		fakeEtcd := new(mock.FakeEtcd).OnCreate(func(key string, val string) error {
			etcdKeyCreated = key
			return nil
		})

		core := newMockCore(fakeEtcd)
		apps := core.Apps()

		core.k8s = new(mock.FakeGuber).OnNamespaceCreate(func(namespace *guber.Namespace) error {
			namespaceCreated = namespace.Metadata.Name
			return nil
		})

		app := apps.New()
		app.Name = common.IDString("test")

		Convey("When Create() is called", func() {
			err := apps.Create(app)

			Convey("The App should be created in etcd with a Created Timestamp", func() {
				So(etcdKeyCreated, ShouldEqual, "/supergiant/apps/test")
				So(app.Created, ShouldHaveSameTypeAs, new(common.Timestamp))
				So(err, ShouldBeNil)
			})

			Convey("A Kubernetes Namespace should be created", func() {
				So(namespaceCreated, ShouldEqual, "test")
			})
		})
	})
}

func TestAppGet(t *testing.T) {
	Convey("Given an AppCollection with an AppResource", t, func() {
		fakeEtcd := new(mock.FakeEtcd).ReturnValueOnGet(
			`{
				"name": "test",
				"created": "Tue, 12 Apr 2016 03:54:56 UTC",
				"updated": null,
				"tags": {}
			}`,
			nil,
		)
		core := newMockCore(fakeEtcd)
		apps := core.Apps()

		Convey("When Get() is called with the App name", func() {
			expected := apps.New()
			expected.Name = common.IDString("test")
			expected.Created = common.TimestampFromString("Tue, 12 Apr 2016 03:54:56 UTC")

			app, err := apps.Get(expected.Name)

			Convey("The return value should be the AppResource", func() {
				So(app, ShouldResemble, expected)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestAppUpdate(t *testing.T) {
	Convey("Given an AppCollection with an AppResource", t, func() {
		etcdKeyUpdated := ""

		fakeEtcd := new(mock.FakeEtcd).OnUpdate(func(key string, val string) error {
			etcdKeyUpdated = key
			return nil
		})
		core := newMockCore(fakeEtcd)
		apps := core.Apps()

		app := apps.New()
		app.Name = common.IDString("test")

		Convey("When Update() is called", func() {
			err := app.Update()

			Convey("The App should be updated in etcd with an Updated Timestamp", func() {
				So(etcdKeyUpdated, ShouldEqual, "/supergiant/apps/test")
				So(app.Updated, ShouldHaveSameTypeAs, new(common.Timestamp))
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestAppDelete(t *testing.T) {
	Convey("Given an AppCollection with an AppResource", t, func() {
		etcdKeyDeleted := ""
		namespaceDeleted := ""
		componentDeleted := ""

		fakeEtcd := new(mock.FakeEtcd).OnDelete(func(key string) error {
			etcdKeyDeleted = key
			return nil
		})
		core := newMockCore(fakeEtcd)

		core.k8s = new(mock.FakeGuber).OnNamespaceDelete(func(name string) error {
			namespaceDeleted = name
			return nil
		})

		apps := core.Apps()
		app := apps.New()
		app.Name = common.IDString("test")

		fakeComponents := &FakeComponentCollection{
			app:  app,
			core: core,
		}
		fakeComponents.ReturnValuesOnGet([]*common.Component{
			&common.Component{
				Name: common.IDString("component-test"),
			},
		})
		fakeComponents.OnDelete(func(r Resource) error {
			componentDeleted = *(r.(*ComponentResource).Name)
			return nil
		})
		app.ComponentsInterface = fakeComponents

		Convey("When Delete() is called", func() {
			err := app.Delete()

			Convey("The App should be deleted in etcd", func() {
				So(err, ShouldBeNil)
				So(etcdKeyDeleted, ShouldEqual, "/supergiant/apps/test")
			})

			Convey("The Kubernetes Namespace should be deleted", func() {
				So(namespaceDeleted, ShouldEqual, "test")
			})

			Convey("The App's Components should be deleted", func() {
				So(componentDeleted, ShouldEqual, "component-test")
			})
		})
	})
}

// TODO move to shared folder
func newMockCore(fakeEtcd *mock.FakeEtcd) *Core {
	return &Core{
		db: &database{
			&etcdClient{fakeEtcd},
		},
	}
}

func (f *FakeComponentCollection) ReturnValuesOnGet(components []*common.Component) *FakeComponentCollection {
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

func (f *FakeComponentCollection) Delete(r Resource) error {
	return f.DeleteFn(r)
}

func (f *FakeComponentCollection) Deploy(r Resource) error {
	return f.DeployFn(r)
}
