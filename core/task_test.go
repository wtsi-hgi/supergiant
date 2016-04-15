package core

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	etcd "github.com/coreos/etcd/client"

	"github.com/supergiant/supergiant/common"
	"github.com/supergiant/supergiant/core/mock"
)

func TestTaskList(t *testing.T) {
	Convey("Given a TaskCollection with 1 Task", t, func() {
		fakeEtcd := new(mock.FakeEtcd).ReturnOnGet(
			&etcd.Response{
				Node: &etcd.Node{
					Nodes: etcd.Nodes{
						&etcd.Node{
							Key: "/supergiant/tasks/00000000000000004521",
							Value: `{
                "created": "Tue, 12 Apr 2016 03:54:56 UTC",
                "updated": null,
                "tags": {},
                "type": 0,
                "status": "QUEUED",
                "attempts": 0,
                "max_attempts": 20,
                "error": ""
      				}`,
						},
					},
				},
			}, nil,
		)
		core := newMockCore(fakeEtcd)
		tasks := core.Tasks()

		Convey("When List() is called", func() {
			list, err := tasks.List()

			Convey("The return value should be a TaskList with 1 Task", func() {
				expected := tasks.New()
				expected.ID = common.IDString("00000000000000004521")
				expected.Created = common.TimestampFromString("Tue, 12 Apr 2016 03:54:56 UTC")
				expected.Type = common.TaskType(0)
				expected.Status = "QUEUED"
				expected.MaxAttempts = 20

				So(err, ShouldBeNil)
				So(list.Items, ShouldHaveLength, 1)
				So(list.Items[0], ShouldResemble, expected)
			})
		})
	})
}

func TestTaskCreate(t *testing.T) {
	Convey("Given a TaskCollection and a new TaskResource", t, func() {

		fakeEtcd := new(mock.FakeEtcd).OnCreateInOrder(func(val string) (*etcd.Response, error) {
			return &etcd.Response{
				Node: &etcd.Node{
					Key:   "/supergiant/tasks/00000000000000004521",
					Value: val,
				},
			}, nil
		})
		core := newMockCore(fakeEtcd)

		tasks := core.Tasks()

		task := tasks.New()

		Convey("When Create() is called", func() {
			err := tasks.Create(task)

			Convey("The Task should be created in etcd with ID correctly parsed and a Created Timestamp", func() {
				So(*task.ID, ShouldEqual, "00000000000000004521")
				So(task.Created, ShouldHaveSameTypeAs, new(common.Timestamp))
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestTaskGet(t *testing.T) {
	Convey("Given a TaskCollection with a TaskResource", t, func() {
		fakeEtcd := new(mock.FakeEtcd).ReturnOnGet(
			&etcd.Response{
				Node: &etcd.Node{
					Key: "/supergiant/tasks/00000000000000004521",
					Value: `{
            "created": "Tue, 12 Apr 2016 03:54:56 UTC",
            "updated": null,
            "tags": {},
            "type": 0,
            "status": "QUEUED",
            "attempts": 0,
            "max_attempts": 20,
            "error": ""
  				}`,
				},
			}, nil,
		)
		core := newMockCore(fakeEtcd)
		tasks := core.Tasks()

		Convey("When Get() is called with the Task ID", func() {
			expected := tasks.New()
			expected.ID = common.IDString("00000000000000004521")
			expected.Created = common.TimestampFromString("Tue, 12 Apr 2016 03:54:56 UTC")
			expected.Type = common.TaskType(0)
			expected.Status = "QUEUED"
			expected.MaxAttempts = 20

			task, err := tasks.Get(expected.ID)

			Convey("The return value should be the TaskResource", func() {
				So(task, ShouldResemble, expected)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestTaskUpdate(t *testing.T) {
	Convey("Given a TaskCollection with a TaskResource", t, func() {
		etcdKeyUpdated := ""

		fakeEtcd := new(mock.FakeEtcd).OnUpdate(func(key string, val string) error {
			etcdKeyUpdated = key
			return nil
		})
		core := newMockCore(fakeEtcd)
		tasks := core.Tasks()

		task := tasks.New()
		task.ID = common.IDString("00000000000000004521")

		Convey("When Update() is called", func() {
			err := task.Update()

			Convey("The Task should be updated in etcd with an Updated Timestamp", func() {
				So(etcdKeyUpdated, ShouldEqual, "/supergiant/tasks/00000000000000004521")
				So(task.Updated, ShouldHaveSameTypeAs, new(common.Timestamp))
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestTaskDelete(t *testing.T) {
	Convey("Given a TaskCollection with a TaskResource", t, func() {
		etcdKeyDeleted := ""

		fakeEtcd := new(mock.FakeEtcd).OnDelete(func(key string) error {
			etcdKeyDeleted = key
			return nil
		})
		core := newMockCore(fakeEtcd)

		tasks := core.Tasks()
		task := tasks.New()
		task.ID = common.IDString("00000000000000004521")

		Convey("When Delete() is called", func() {
			err := task.Delete()

			Convey("The Task should be deleted in etcd", func() {
				So(err, ShouldBeNil)
				So(etcdKeyDeleted, ShouldEqual, "/supergiant/tasks/00000000000000004521")
			})
		})
	})
}
