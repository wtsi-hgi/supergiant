package core

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	etcd "github.com/coreos/etcd/client"

	"github.com/supergiant/guber"
	"github.com/supergiant/supergiant/common"
	"github.com/supergiant/supergiant/mock"
)

func TestAppList(t *testing.T) {
	Convey("Given an AppCollection with 1 App", t, func() {
		fakeEtcd := &mock.FakeKeysAPI{
			GetFn: func() (*etcd.Response, error) {
				return &etcd.Response{
					Node: &etcd.Node{
						Nodes: etcd.Nodes{
							&etcd.Node{
								Value: `{
                  "name": "test",
                  "created": "Tue, 12 Apr 2016 03:54:56 UTC",
                  "updated": null,
                  "tags": {}
                }`,
							},
						},
					},
				}, nil
			},
		}
		core := newMockCore(fakeEtcd)
		apps := &AppCollection{core}

		app := apps.New()
		app.Name = common.IDString("test")
		app.Meta.Created = new(common.Timestamp)
		app.Meta.Created.UnmarshalJSON([]byte(`"Tue, 12 Apr 2016 03:54:56 UTC"`))

		Convey("When List() is called", func() {
			list, err := apps.List()

			Convey("The return value should be an AppList with 1 App", func() {
				So(list.Items, ShouldHaveLength, 1)
				So(list.Items[0], ShouldResemble, app)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestAppCreate(t *testing.T) {
	Convey("Given an AppCollection and a new AppResource", t, func() {
		createCalled := false
		namespaceCreated := ""

		fakeEtcd := &mock.FakeKeysAPI{
			CreateFn: func(val string) (*etcd.Response, error) {
				createCalled = true

				return &etcd.Response{
					Node: &etcd.Node{
						Value: val,
					},
				}, nil
			},
		}
		core := newMockCore(fakeEtcd)
		apps := &AppCollection{core}

		fakeNamespaces := &mock.FakeGuberNamespaces{
			CreateFn: func(namespace *guber.Namespace) (*guber.Namespace, error) {
				namespaceCreated = namespace.Metadata.Name

				return namespace, nil
			},
		}

		core.k8s = &mock.FakeGuberClient{
			NamespacesFn: func() guber.NamespaceCollection {
				return fakeNamespaces
			},
		}

		app := apps.New()
		app.Name = common.IDString("test")

		Convey("When Create() is called", func() {
			err := apps.Create(app)

			Convey("The App should be created in etcd with a Created Timestamp", func() {
				So(createCalled, ShouldBeTrue)
				So(app.Created, ShouldHaveSameTypeAs, new(common.Timestamp))
				So(err, ShouldBeNil)
			})

			Convey("A Kubernetes Namespace should be created", func() {
				So(namespaceCreated, ShouldEqual, "test")
			})
		})
	})
}

func newMockCore(fakeEtcd *mock.FakeKeysAPI) *Core {
	return &Core{
		db: &database{
			&etcdClient{fakeEtcd},
		},
	}
}

func TestAppGet(t *testing.T) {
	Convey("Given an AppCollection with an AppResource", t, func() {
		fakeEtcd := &mock.FakeKeysAPI{
			GetFn: func() (*etcd.Response, error) {
				return &etcd.Response{
					Node: &etcd.Node{
						Value: `{
              "name": "test",
              "created": "Tue, 12 Apr 2016 03:54:56 UTC",
              "updated": null,
              "tags": {}
            }`,
					},
				}, nil
			},
		}
		core := newMockCore(fakeEtcd)
		apps := &AppCollection{core}

		newApp := apps.New()
		newApp.Name = common.IDString("test")
		newApp.Meta.Created = new(common.Timestamp)
		newApp.Meta.Created.UnmarshalJSON([]byte(`"Tue, 12 Apr 2016 03:54:56 UTC"`))

		Convey("When Get() is called with the App name", func() {
			app, err := apps.Get(newApp.Name)

			Convey("The return value should be the AppResource", func() {
				So(app, ShouldResemble, newApp)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestAppUpdate(t *testing.T) {
	Convey("Given an AppCollection with an AppResource", t, func() {
		updateCalled := false

		fakeEtcd := &mock.FakeKeysAPI{
			GetFn: func() (*etcd.Response, error) {
				return &etcd.Response{
					Node: &etcd.Node{
						Value: `{
							"name": "test",
							"created": "Tue, 12 Apr 2016 03:54:56 UTC",
							"updated": null,
							"tags": {}
						}`,
					},
				}, nil
			},
			UpdateFn: func(val string) (*etcd.Response, error) {
				updateCalled = true

				return &etcd.Response{
					Node: &etcd.Node{
						Value: val,
					},
				}, nil
			},
		}
		core := newMockCore(fakeEtcd)
		apps := &AppCollection{core}

		app, _ := apps.Get(common.IDString("test"))

		Convey("When Update() is called", func() {
			app.Tags["foo"] = "bar"
			err := app.Update()

			Convey("The App should be updated in etcd with an Updated Timestamp", func() {
				So(updateCalled, ShouldBeTrue)
				So(app.Tags["foo"], ShouldEqual, "bar")
				So(app.Updated, ShouldHaveSameTypeAs, new(common.Timestamp))
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestAppDelete(t *testing.T) {
	Convey("Given an AppCollection with an AppResource", t, func() {
		deleteCalled := false
		namespaceDeleted := ""
		componentDeleted := ""

		fakeEtcd := &mock.FakeKeysAPI{
			DeleteFn: func() (*etcd.Response, error) {
				deleteCalled = true

				return new(etcd.Response), nil
			},
		}

		core := newMockCore(fakeEtcd)

		fakeNamespaces := &mock.FakeGuberNamespaces{
			DeleteFn: func(name string) (bool, error) {
				namespaceDeleted = name

				return true, nil
			},
		}
		core.k8s = &mock.FakeGuberClient{
			NamespacesFn: func() guber.NamespaceCollection {
				return fakeNamespaces
			},
		}

		apps := &AppCollection{core}
		app := apps.New()
		app.Name = common.IDString("test")

		fakeComponents := &FakeComponentCollection{
			app: app,
			DeleteFn: func(r *ComponentResource) error {
				componentDeleted = *r.Name

				return nil
			},
		}
		fakeComponents.ListFn = func() (*ComponentList, error) {
			return &ComponentList{
				Items: []*ComponentResource{
					&ComponentResource{
						collection: fakeComponents,
						Component: &common.Component{
							Name: common.IDString("component-test"),
						},
					},
				},
			}, nil
		}

		app.ComponentsInterface = fakeComponents

		Convey("When Delete() is called", func() {
			err := app.Delete()

			Convey("The App should be deleted in etcd", func() {
				So(deleteCalled, ShouldBeTrue)
				So(err, ShouldBeNil)
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

type FakeComponentCollection struct {
	app      *AppResource
	ListFn   func() (*ComponentList, error)
	NewFn    func() *ComponentResource
	CreateFn func() error
	GetFn    func() (*ComponentResource, error)
	UpdateFn func() error
	DeleteFn func(*ComponentResource) error
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

func (f *FakeComponentCollection) Delete(r *ComponentResource) error {
	return f.DeleteFn(r)
}
