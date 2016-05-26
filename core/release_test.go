package core

import (
	"regexp"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/supergiant/guber"
	"github.com/supergiant/supergiant/common"
	"github.com/supergiant/supergiant/core/mock"
)

var fakeRelease = &common.Release{
	InstanceCount: 1,
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
var fakeReleaseJSON = `{
  "instance_count": 1,
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

func TestReleaseList(t *testing.T) {
	Convey("Given a ReleaseCollection with 1 Release", t, func() {
		fakeEtcd := new(mock.FakeEtcd).ReturnValuesOnGet([]string{fakeReleaseJSON}, nil)
		core := newMockCore(fakeEtcd)

		// Kube Services are fetched on decorate()
		core.k8s = new(mock.FakeGuber).ReturnOnServiceGet(nil, new(guber.Error404)) // service does not exist

		core.AppsInterface = &AppCollection{core}
		core.EntrypointsInterface = &EntrypointCollection{core}
		core.ImageRegistriesInterface = &ImageRegistryCollection{core}
		core.dockerhub = core.ImageRegistries().New()
		// core.dockerhub.ImageReposInterface = new(FakeImageRepoCollection).ReturnOnGet(nil, &etcd.Error{Code: etcd.ErrorCodeKeyNotFound})

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
				So(list.Items[0].Release, ShouldResemble, fakeRelease)
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
		// core.dockerhub = core.ImageRegistries().New()

		app := core.Apps().New()
		app.Name = common.IDString("test")

		component := app.Components().New()
		component.Name = common.IDString("component-test")

		releases := component.Releases()

		fakeCopy := *fakeRelease
		release := releases.New()
		release.Release = &fakeCopy
		release.Meta = common.NewMeta() // because fakeCopy has no Meta

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
