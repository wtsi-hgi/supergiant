package core

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	etcd "github.com/coreos/etcd/client"
	"github.com/supergiant/guber"
	"github.com/supergiant/supergiant/common"
	"github.com/supergiant/supergiant/core/mock"
)

var fakeRelease = &common.Release{
	Meta: &common.Meta{
		Created: common.TimestampFromString("Tue, 12 Apr 2016 03:54:56 UTC"),
	},
	Timestamp:     common.IDString("20160412035456"),
	InstanceGroup: common.IDString("20160412035456"),
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
  "created": "Tue, 12 Apr 2016 03:54:56 UTC",
  "timestamp": "20160412035456",
  "instance_group": "20160412035456",
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
		core.dockerhub.ImageReposInterface = new(FakeImageRepoCollection).ReturnOnGet(nil, &etcd.Error{Code: etcd.ErrorCodeKeyNotFound})

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
