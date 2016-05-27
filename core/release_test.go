package core

import (
	"regexp"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/supergiant/guber"
	"github.com/supergiant/supergiant/common"
	"github.com/supergiant/supergiant/core/mock"
)

func TestReleaseList(t *testing.T) {
	Convey("Given a ReleaseCollection with 1 Release", t, func() {
		fakeEtcd := new(mock.FakeEtcd).ReturnValuesOnGet([]string{fakeReleaseJSON}, nil)
		core := newMockCore(fakeEtcd)

		// Kube Services are fetched on decorate()
		core.k8s = new(mock.FakeGuber).ReturnOnServiceGet(nil, new(guber.Error404)) // service does not exist

		core.AppsInterface = &AppCollection{core}
		core.EntrypointsInterface = &EntrypointCollection{core}
		core.ImageRegistriesInterface = &ImageRegistryCollection{core}

		app := core.Apps().New()
		app.Name = common.IDString("test")

		component := app.Components().New()
		component.Name = common.IDString("test")

		releases := component.Releases()

		Convey("When List() is called", func() {
			list, err := releases.List()

			Convey("The return value should be a ReleaseList with 1 Release", func() {
				So(err, ShouldBeNil)
				So(list.Items, ShouldHaveLength, 1)
				So(list.Items[0].Release, ShouldResemble, fakeRelease())
			})
		})
	})
}

func TestReleaseCreate(t *testing.T) {
	Convey("Given a ReleaseCollection and a new ReleaseResource", t, func() {
		etcdKeyCreated := ""
		updatedTargetReleaseTimestamp := ""

		fakeEtcd := new(mock.FakeEtcd)

		fakeEtcd.OnCreate(func(key string, val string) error {
			etcdKeyCreated = key
			return nil
		})

		fakeEtcd.OnUpdate(func(key string, val string) error {
			re := regexp.MustCompile(`"target_release_id":"([0-9]+)"`)
			updatedTargetReleaseTimestamp = re.FindStringSubmatch(val)[1]
			return nil
		})

		core := newMockCore(fakeEtcd)

		// Kube Services are fetched on decorate()
		core.k8s = new(mock.FakeGuber).ReturnOnServiceGet(nil, new(guber.Error404)) // service does not exist

		core.AppsInterface = &AppCollection{core}
		core.EntrypointsInterface = &EntrypointCollection{core}
		core.ImageRegistriesInterface = &ImageRegistryCollection{core}

		app := core.Apps().New()
		app.Name = common.IDString("test")

		component := app.Components().New()
		component.Name = common.IDString("component-test")

		releases := component.Releases()

		release := releases.New()
		release.Release = fakeRelease()

		Convey("When Create() is called", func() {
			err := releases.Create(release)

			Convey("The Release should be created in etcd", func() {
				So(err, ShouldBeNil)
				So(etcdKeyCreated, ShouldStartWith, "/supergiant/releases/test/component-test/201")
				So(release.Created, ShouldHaveSameTypeAs, new(common.Timestamp))
			})

			Convey("The Release should have a new Release Timestamp", func() {
				So(*release.Timestamp, ShouldStartWith, "201") // e.g. 20160519152252
			})

			Convey("Component should be updated with new TargetReleaseTimestamp", func() {
				So(updatedTargetReleaseTimestamp, ShouldEqual, *release.Timestamp)
			})
		})

		Convey("When Component already has a TargetReleaseTimestamp, and Create() is called", func() {
			component.TargetReleaseTimestamp = common.IDString("something")
			err := releases.Create(release)

			Convey("An error is returned", func() {
				So(err.Error(), ShouldEqual, "Component already has a target Release")
			})
		})
	})
}

func TestReleaseMergeCreate(t *testing.T) {
	Convey("Given a ReleaseCollection and a new ReleaseResource", t, func() {
		etcdKeyCreated := ""
		updatedTargetReleaseTimestamp := ""

		fakeEtcd := new(mock.FakeEtcd)

		fakeEtcd.OnCreate(func(key string, val string) error {
			etcdKeyCreated = key
			return nil
		})

		fakeEtcd.OnUpdate(func(key string, val string) error {
			re := regexp.MustCompile(`"target_release_id":"([0-9]+)"`)
			updatedTargetReleaseTimestamp = re.FindStringSubmatch(val)[1]
			return nil
		})

		core := newMockCore(fakeEtcd)

		// Kube Services are fetched on decorate()
		core.k8s = new(mock.FakeGuber).ReturnOnServiceGet(nil, new(guber.Error404)) // service does not exist

		core.AppsInterface = &AppCollection{core}
		core.EntrypointsInterface = &EntrypointCollection{core}
		core.ImageRegistriesInterface = &ImageRegistryCollection{core}

		app := core.Apps().New()
		app.Name = common.IDString("test")

		component := app.Components().New()
		component.Name = common.IDString("component-test")

		fakeCurrentRelease := fakeRelease()

		fakeCurrentRelease.Timestamp = common.IDString("20160519152252")
		fakeCurrentRelease.Meta.Created = common.TimestampFromString("Tue, 12 Apr 2016 03:54:56 UTC")
		fakeCurrentRelease.Committed = true

		component.ReleasesInterface = new(FakeReleaseCollection).ReturnOnGet(fakeCurrentRelease, nil)
		component.CurrentReleaseTimestamp = fakeCurrentRelease.Timestamp

		// we do this and not component.Releases() because we mock it above
		releases := &ReleaseCollection{core, component}

		release := releases.New()
		release.TerminationGracePeriod = 420

		Convey("When MergeCreate() is called", func() {
			err := releases.MergeCreate(release)

			Convey("The Release should be created in etcd", func() {
				So(err, ShouldBeNil)
				So(etcdKeyCreated, ShouldStartWith, "/supergiant/releases/test/component-test/201")
			})

			Convey("The Release should have certain different Timestamp, Meta.Created, and Committed values", func() {
				So(*release.Timestamp, ShouldStartWith, "201") // e.g. 20160519152252
				So(*release.Timestamp, ShouldNotEqual, *fakeCurrentRelease.Timestamp)

				So(release.Created, ShouldHaveSameTypeAs, new(common.Timestamp))
				So(*release.Created, ShouldNotEqual, *fakeCurrentRelease.Created)

				So(release.Committed, ShouldBeFalse)
			})

			// TODO this really tests that release.collection.Create is called
			Convey("Component should be updated with new TargetReleaseTimestamp", func() {
				So(updatedTargetReleaseTimestamp, ShouldEqual, *release.Timestamp)
			})

			Convey("The Release should have values merged with CurrentRelease", func() {
				So(release.Containers[0].Image, ShouldEqual, "mysql") // old
				So(release.TerminationGracePeriod, ShouldEqual, 420)  // new
			})
		})

		Convey("When Component has no CurrentReleaseTimestamp, and MergeCreate() is called", func() {
			component.CurrentReleaseTimestamp = nil
			err := releases.MergeCreate(release)

			Convey("An error is returned", func() {
				So(err.Error(), ShouldEqual, "Attempting MergeCreate with no current Release")
			})
		})
	})
}

func TestReleaseGet(t *testing.T) {
	Convey("Given a ReleaseCollection with 1 Release", t, func() {
		fakeEtcd := new(mock.FakeEtcd).ReturnValueOnGet(fakeReleaseJSON, nil)
		core := newMockCore(fakeEtcd)

		// Kube Services are fetched on decorate()
		core.k8s = new(mock.FakeGuber).ReturnOnServiceGet(nil, new(guber.Error404)) // service does not exist

		core.AppsInterface = &AppCollection{core}
		core.EntrypointsInterface = &EntrypointCollection{core}
		core.ImageRegistriesInterface = &ImageRegistryCollection{core}

		app := core.Apps().New()
		app.Name = common.IDString("test")

		component := app.Components().New()
		component.Name = common.IDString("test")

		releases := component.Releases()

		Convey("When Get() is called", func() {
			release, err := releases.Get(common.IDString("buttman"))

			Convey("The return value should be the Release", func() {
				So(err, ShouldBeNil)
				So(release.Release, ShouldResemble, fakeRelease())
			})
		})
	})
}

func TestReleaseUpdate(t *testing.T) {
	Convey("Given an ReleaseCollection with an ReleaseResource", t, func() {
		etcdKeyUpdated := ""

		fakeEtcd := new(mock.FakeEtcd).OnUpdate(func(key string, val string) error {
			etcdKeyUpdated = key
			return nil
		})

		core := newMockCore(fakeEtcd)

		core.k8s = new(mock.FakeGuber).ReturnOnServiceGet(nil, new(guber.Error404))

		core.AppsInterface = &AppCollection{core}
		core.EntrypointsInterface = &EntrypointCollection{core}
		core.ImageRegistriesInterface = &ImageRegistryCollection{core}

		app := core.Apps().New()
		app.Name = common.IDString("test")

		component := app.Components().New()
		component.Name = common.IDString("component-test")

		release := component.Releases().New()
		release.Release = fakeRelease()
		release.Timestamp = common.IDString("20160519152252")

		Convey("When Update() is called", func() {
			err := release.Update()

			Convey("The Release should be updated in etcd with an Updated Timestamp", func() {
				So(err, ShouldBeNil)
				So(etcdKeyUpdated, ShouldEqual, "/supergiant/releases/test/component-test/"+*release.Timestamp)
				So(component.Updated, ShouldHaveSameTypeAs, new(common.Timestamp))
			})
		})
	})
}

//------------------------------------------------------------------------------
var fakeReleaseJSON = `{
	"created": null,
	"tags": {
		"test": "tag"
	},
	"instance_count": 1,
	"termination_grace_period": 666,
	"volumes": [
		{
			"name": "data",
			"type": "gp2",
			"size": 20
		}
	],
	"containers": [
		{
			"image": "mysql",
			"ports": [
				{
					"protocol": "TCP",
					"number": 3306
				}
			],
			"mounts": [
				{
					"volume": "data",
					"path": "/var/lib/mysql"
				}
			]
		}
	]
}`

// method so we don't have to worry about deep copying
func fakeRelease() *common.Release {
	return &common.Release{
		Meta: &common.Meta{
			Tags: map[string]string{
				"test": "tag",
			},
		},
		InstanceCount:          1,
		TerminationGracePeriod: 666,
		Volumes: []*common.VolumeBlueprint{
			&common.VolumeBlueprint{
				Name: common.IDString("data"),
				Type: "gp2",
				Size: 20,
			},
		},
		Containers: []*common.ContainerBlueprint{
			&common.ContainerBlueprint{
				Image: "mysql",
				Ports: []*common.Port{
					&common.Port{
						Protocol: "TCP",
						Number:   3306,
					},
				},
				Mounts: []*common.Mount{
					&common.Mount{
						Volume: common.IDString("data"),
						Path:   "/var/lib/mysql",
					},
				},
			},
		},
	}
}
