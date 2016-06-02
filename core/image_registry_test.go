package core

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/supergiant/supergiant/common"
	"github.com/supergiant/supergiant/core/mock"
)

func TestImageRegistryList(t *testing.T) {
	Convey("Given an ImageRegistryCollection with 1 ImageRegistry", t, func() {
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
		core.ImageRegistriesInterface = &ImageRegistryCollection{core}
		registries := core.ImageRegistries()

		Convey("When List() is called", func() {
			list, err := registries.List()

			Convey("The return value should be an ImageRegistryList with 1 ImageRegistry", func() {
				expected := registries.New()
				expected.Name = common.IDString("test")
				expected.Created = common.TimestampFromString("Tue, 12 Apr 2016 03:54:56 UTC")

				So(list.Items, ShouldHaveLength, 1)
				So(list.Items[0], ShouldResemble, expected)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestImageRegistryCreate(t *testing.T) {
	Convey("Given an ImageRegistryCollection and a new ImageRegistryResource", t, func() {
		etcdKeyCreated := ""

		fakeEtcd := new(mock.FakeEtcd).OnCreate(func(key string, val string) error {
			etcdKeyCreated = key
			return nil
		})
		core := newMockCore(fakeEtcd)
		core.ImageRegistriesInterface = &ImageRegistryCollection{core}

		registries := core.ImageRegistries()

		registry := registries.New()
		registry.Name = common.IDString("test")

		Convey("When Create() is called", func() {
			err := registries.Create(registry)

			Convey("The ImageRegistry should be created in etcd with a Created Timestamp", func() {
				So(etcdKeyCreated, ShouldEqual, "/supergiant/registries/test")
				So(registry.Created, ShouldHaveSameTypeAs, new(common.Timestamp))
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestImageRegistryGet(t *testing.T) {
	Convey("Given an ImageRegistryCollection with an ImageRegistryResource", t, func() {
		fakeEtcd := new(mock.FakeEtcd).ReturnValueOnGet(
			`{
				"name": "test",
				"key": "key",
				"created": "Tue, 12 Apr 2016 03:54:56 UTC",
				"updated": null,
				"tags": {}
			}`,
			nil,
		)
		core := newMockCore(fakeEtcd)
		core.ImageRegistriesInterface = &ImageRegistryCollection{core}
		registries := core.ImageRegistries()

		Convey("When Get() is called with the ImageRegistry name", func() {
			expected := registries.New()
			expected.Name = common.IDString("test")
			expected.Created = common.TimestampFromString("Tue, 12 Apr 2016 03:54:56 UTC")

			registry, err := registries.Get(expected.Name)

			Convey("The return value should be the ImageRegistryResource", func() {
				So(registry, ShouldResemble, expected)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestImageRegistryUpdate(t *testing.T) {
	Convey("Given an ImageRegistryCollection with an ImageRegistryResource", t, func() {
		etcdKeyUpdated := ""

		fakeEtcd := new(mock.FakeEtcd).OnUpdate(func(key string, val string) error {
			etcdKeyUpdated = key
			return nil
		})

		core := newMockCore(fakeEtcd)
		core.ImageRegistriesInterface = &ImageRegistryCollection{core}
		registries := core.ImageRegistries()

		registry := registries.New()
		registry.Name = common.IDString("test")

		Convey("When Update() is called", func() {
			err := registry.Update()

			Convey("The ImageRegistry should be updated in etcd with an Updated Timestamp", func() {
				So(err, ShouldBeNil)
				So(etcdKeyUpdated, ShouldEqual, "/supergiant/registries/test")
				So(registry.Updated, ShouldHaveSameTypeAs, new(common.Timestamp))
			})
		})
	})
}

func TestImageRegistryDelete(t *testing.T) {
	Convey("Given an ImageRegistryCollection with an ImageRegistryResource", t, func() {
		etcdKeyDeleted := ""

		fakeEtcd := new(mock.FakeEtcd).OnDelete(func(key string) error {
			etcdKeyDeleted = key
			return nil
		})
		core := newMockCore(fakeEtcd)
		core.ImageRegistriesInterface = &ImageRegistryCollection{core}

		registries := core.ImageRegistries()
		registry := registries.New()
		registry.Name = common.IDString("test")

		Convey("When Delete() is called", func() {
			err := registry.Delete()

			Convey("The ImageRegistry should be deleted in etcd", func() {
				So(err, ShouldBeNil)
				So(etcdKeyDeleted, ShouldEqual, "/supergiant/registries/test")
			})
		})
	})
}
