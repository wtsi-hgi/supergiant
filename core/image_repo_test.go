package core

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/supergiant/supergiant/common"
	"github.com/supergiant/supergiant/core/mock"
)

func TestImageRepoList(t *testing.T) {
	Convey("Given an ImageRepoCollection with 1 ImageRepo", t, func() {
		fakeEtcd := new(mock.FakeEtcd).ReturnValuesOnGet(
			[]string{
				`{
					"name": "test",
					"key": "key",
					"created": "Tue, 12 Apr 2016 03:54:56 UTC",
					"updated": null,
					"tags": {}
				}`,
			},
			nil,
		)
		core := newMockCore(fakeEtcd)
		repos := core.ImageRepos()

		Convey("When List() is called", func() {
			list, err := repos.List()

			Convey("The return value should be an ImageRepoList with 1 ImageRepo", func() {
				expected := repos.New()
				expected.Name = common.IDString("test")
				expected.Key = "key"
				expected.Created = common.TimestampFromString("Tue, 12 Apr 2016 03:54:56 UTC")

				So(list.Items, ShouldHaveLength, 1)
				So(list.Items[0], ShouldResemble, expected)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestImageRepoCreate(t *testing.T) {
	Convey("Given an ImageRepoCollection and a new ImageRepoResource", t, func() {
		etcdKeyCreated := ""

		fakeEtcd := new(mock.FakeEtcd).OnCreate(func(key string, val string) error {
			etcdKeyCreated = key
			return nil
		})
		core := newMockCore(fakeEtcd)

		repos := core.ImageRepos()

		repo := repos.New()
		repo.Name = common.IDString("test")
		repo.Key = "key"

		Convey("When Create() is called", func() {
			err := repos.Create(repo)

			Convey("The ImageRepo should be created in etcd with a Created Timestamp", func() {
				So(etcdKeyCreated, ShouldEqual, "/supergiant/repos/dockerhub/test")
				So(repo.Created, ShouldHaveSameTypeAs, new(common.Timestamp))
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestImageRepoGet(t *testing.T) {
	Convey("Given an ImageRepoCollection with an ImageRepoResource", t, func() {
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
		repos := core.ImageRepos()

		Convey("When Get() is called with the ImageRepo name", func() {
			expected := repos.New()
			expected.Name = common.IDString("test")
			expected.Key = "key"
			expected.Created = common.TimestampFromString("Tue, 12 Apr 2016 03:54:56 UTC")

			repo, err := repos.Get(expected.Name)

			Convey("The return value should be the ImageRepoResource", func() {
				So(repo, ShouldResemble, expected)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestImageRepoUpdate(t *testing.T) {
	Convey("Given an ImageRepoCollection with an ImageRepoResource", t, func() {
		etcdKeyUpdated := ""

		fakeEtcd := new(mock.FakeEtcd).OnUpdate(func(key string, val string) error {
			etcdKeyUpdated = key
			return nil
		})

		core := newMockCore(fakeEtcd)
		repos := core.ImageRepos()

		repo := repos.New()
		repo.Name = common.IDString("test")
		repo.Key = "key"

		Convey("When Update() is called", func() {
			err := repo.Update()

			Convey("The ImageRepo should be updated in etcd with an Updated Timestamp", func() {
				So(err, ShouldBeNil)
				So(etcdKeyUpdated, ShouldEqual, "/supergiant/repos/dockerhub/test")
				So(repo.Updated, ShouldHaveSameTypeAs, new(common.Timestamp))
			})
		})
	})
}

func TestImageRepoDelete(t *testing.T) {
	Convey("Given an ImageRepoCollection with an ImageRepoResource", t, func() {
		etcdKeyDeleted := ""

		fakeEtcd := new(mock.FakeEtcd).OnDelete(func(key string) error {
			etcdKeyDeleted = key
			return nil
		})
		core := newMockCore(fakeEtcd)

		repos := core.ImageRepos()
		repo := repos.New()
		repo.Name = common.IDString("test")

		Convey("When Delete() is called", func() {
			err := repo.Delete()

			Convey("The ImageRepo should be deleted in etcd", func() {
				So(err, ShouldBeNil)
				So(etcdKeyDeleted, ShouldEqual, "/supergiant/repos/dockerhub/test")
			})
		})
	})
}
