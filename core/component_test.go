package core

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/supergiant/supergiant/common"
	"github.com/supergiant/supergiant/core/mock"
)

func TestComponentList(t *testing.T) {
	Convey("Given an ComponentCollection with 1 Component", t, func() {
		fakeEtcd := new(mock.FakeEtcd).ReturnValuesOnGet(
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

		core.AppsInterface = &AppCollection{core}
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

		core.AppsInterface = &AppCollection{core}
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
		fakeEtcd := new(mock.FakeEtcd).ReturnValueOnGet(
			`{
				"name": "component-test",
				"created": "Tue, 12 Apr 2016 03:54:56 UTC",
				"updated": null,
				"tags": {}
			}`,
			nil,
		)
		core := newMockCore(fakeEtcd)

		core.AppsInterface = &AppCollection{core}
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

		core.AppsInterface = &AppCollection{core}
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

		core.AppsInterface = &AppCollection{core}
		app := core.Apps().New()
		app.Name = common.IDString("test")

		component := app.Components().New()
		component.Name = common.IDString("component-test")

		fakeReleases := &FakeReleaseCollection{
			component: component,
			core:      core,
		}

		fakeReleases.ReturnValuesOnList([]*common.Release{
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
