package core

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/elb"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/supergiant/supergiant/common"
	"github.com/supergiant/supergiant/core/mock"
)

func TestEntrypointList(t *testing.T) {
	Convey("Given an EntrypointCollection with 1 Entrypoint", t, func() {
		fakeEtcd := new(mock.FakeEtcd).ReturnValuesOnGet(
			[]string{
				`{
					"domain": "example.com",
					"created": "Tue, 12 Apr 2016 03:54:56 UTC",
					"updated": null,
					"tags": {}
				}`,
			},
			nil,
		)
		core := newMockCore(fakeEtcd)
		entrypoints := core.Entrypoints()

		Convey("When List() is called", func() {
			list, err := entrypoints.List()

			Convey("The return value should be an EntrypointList with 1 Entrypoint", func() {
				expected := entrypoints.New()
				expected.Domain = common.IDString("example.com")
				expected.Created = common.TimestampFromString("Tue, 12 Apr 2016 03:54:56 UTC")

				So(list.Items, ShouldHaveLength, 1)
				So(list.Items[0], ShouldResemble, expected)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestEntrypointCreate(t *testing.T) {
	Convey("Given an EntrypointCollection and a new EntrypointResource", t, func() {
		etcdKeyCreated := ""
		elbCreated := ""
		elbHealthCheckConfiguredOn := ""
		elbAttachedTo := ""

		fakeEtcd := new(mock.FakeEtcd)

		fakeEtcd.OnCreate(func(key string, val string) error {
			etcdKeyCreated = key
			return nil
		})

		fakeEtcd.OnUpdate(func(_ string, _ string) error {
			// just mocking update since we use it in Entrypoint create (but we assert outcome below)
			return nil
		})

		fakeELB := new(mock.FakeAwsELB)

		fakeELB.OnCreateLoadBalancer(func(input *elb.CreateLoadBalancerInput) (*elb.CreateLoadBalancerOutput, error) {
			elbCreated = *input.LoadBalancerName
			return &elb.CreateLoadBalancerOutput{
				DNSName: aws.String("scatman.com"),
			}, nil
		})

		fakeELB.OnConfigureHealthCheck(func(input *elb.ConfigureHealthCheckInput) (*elb.ConfigureHealthCheckOutput, error) {
			elbHealthCheckConfiguredOn = *input.LoadBalancerName
			return nil, nil
		})

		fakeAutoScaling := new(mock.FakeAwsAutoscaling)

		fakeAutoScaling.ReturnOnDescribeAutoScalingGroups(
			[]*autoscaling.Group{
				&autoscaling.Group{
					AutoScalingGroupName: aws.String("autoscaling-test"),
					VPCZoneIdentifier:    aws.String("subnet-69420666"),
				},
			}, nil,
		)

		fakeAutoScaling.OnAttachLoadBalancers(func(input *autoscaling.AttachLoadBalancersInput) (*autoscaling.AttachLoadBalancersOutput, error) {
			elbAttachedTo = *input.AutoScalingGroupName
			return nil, nil
		})

		core := newMockCore(fakeEtcd)
		core.elb = fakeELB
		core.autoscaling = fakeAutoScaling

		core.AwsSubnetID = "subnet-69420666"

		entrypoints := core.Entrypoints()

		entrypoint := entrypoints.New()
		entrypoint.Domain = common.IDString("example.com")

		Convey("When Create() is called", func() {
			err := entrypoints.Create(entrypoint)

			Convey("The Entrypoint should be created in etcd with a Created Timestamp", func() {
				So(etcdKeyCreated, ShouldEqual, "/supergiant/entrypoints/example.com")
				So(entrypoint.Created, ShouldHaveSameTypeAs, new(common.Timestamp))
				So(err, ShouldBeNil)
			})

			Convey("An ELB should be created with a name in the format of supergiant-:escaped_domain", func() {
				So(elbCreated, ShouldEqual, "supergiant-example-com")
			})

			Convey("The ELB should have a health check configured", func() {
				So(elbHealthCheckConfiguredOn, ShouldEqual, "supergiant-example-com")
			})

			Convey("The ELB should be attached to AWS autoscaling groups in the same subnet", func() {
				So(elbAttachedTo, ShouldEqual, "autoscaling-test")
			})

			Convey("The ELB address should be saved (updated) on the Entrypoint", func() {
				So(entrypoint.Address, ShouldEqual, "scatman.com")
			})
		})
	})
}

func TestEntrypointGet(t *testing.T) {
	Convey("Given an EntrypointCollection with an EntrypointResource", t, func() {
		fakeEtcd := new(mock.FakeEtcd).ReturnValueOnGet(
			`{
				"domain": "example.com",
				"created": "Tue, 12 Apr 2016 03:54:56 UTC",
				"updated": null,
				"tags": {}
			}`,
			nil,
		)
		core := newMockCore(fakeEtcd)
		entrypoints := core.Entrypoints()

		Convey("When Get() is called with the Entrypoint name", func() {
			expected := entrypoints.New()
			expected.Domain = common.IDString("example.com")
			expected.Created = common.TimestampFromString("Tue, 12 Apr 2016 03:54:56 UTC")

			entrypoint, err := entrypoints.Get(expected.Domain)

			Convey("The return value should be the EntrypointResource", func() {
				So(entrypoint, ShouldResemble, expected)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestEntrypointUpdate(t *testing.T) {
	Convey("Given an EntrypointCollection with an EntrypointResource", t, func() {
		etcdKeyUpdated := ""

		fakeEtcd := new(mock.FakeEtcd).OnUpdate(func(key string, val string) error {
			etcdKeyUpdated = key
			return nil
		})

		core := newMockCore(fakeEtcd)
		entrypoints := core.Entrypoints()

		entrypoint := entrypoints.New()
		entrypoint.Domain = common.IDString("example.com")

		Convey("When Update() is called", func() {
			err := entrypoint.Update()

			Convey("The Entrypoint should be updated in etcd with an Updated Timestamp", func() {
				So(err, ShouldBeNil)
				So(etcdKeyUpdated, ShouldEqual, "/supergiant/entrypoints/example.com")
				So(entrypoint.Updated, ShouldHaveSameTypeAs, new(common.Timestamp))
			})
		})
	})
}

func TestEntrypointDelete(t *testing.T) {
	Convey("Given an EntrypointCollection and an EntrypointResource", t, func() {
		etcdKeyDeleted := ""
		elbDeleted := ""
		elbDetachedFrom := ""

		fakeEtcd := new(mock.FakeEtcd)

		fakeEtcd.OnDelete(func(key string) error {
			etcdKeyDeleted = key
			return nil
		})

		fakeELB := new(mock.FakeAwsELB)

		fakeELB.OnDeleteLoadBalancer(func(input *elb.DeleteLoadBalancerInput) (*elb.DeleteLoadBalancerOutput, error) {
			elbDeleted = *input.LoadBalancerName
			return nil, nil
		})

		fakeAutoScaling := new(mock.FakeAwsAutoscaling)

		fakeAutoScaling.ReturnOnDescribeAutoScalingGroups(
			[]*autoscaling.Group{
				&autoscaling.Group{
					AutoScalingGroupName: aws.String("autoscaling-test"),
					VPCZoneIdentifier:    aws.String("subnet-69420666"),
				},
			}, nil,
		)

		fakeAutoScaling.OnDetachLoadBalancers(func(input *autoscaling.DetachLoadBalancersInput) (*autoscaling.DetachLoadBalancersOutput, error) {
			elbDetachedFrom = *input.AutoScalingGroupName
			return nil, nil
		})

		core := newMockCore(fakeEtcd)
		core.elb = fakeELB
		core.autoscaling = fakeAutoScaling

		core.AwsSubnetID = "subnet-69420666"

		entrypoints := core.Entrypoints()

		entrypoint := entrypoints.New()
		entrypoint.Domain = common.IDString("example.com")

		Convey("When Delete() is called", func() {
			err := entrypoint.Delete()

			Convey("The Entrypoint should be deleted in etcd", func() {
				So(etcdKeyDeleted, ShouldEqual, "/supergiant/entrypoints/example.com")
				So(entrypoint.Created, ShouldHaveSameTypeAs, new(common.Timestamp))
				So(err, ShouldBeNil)
			})

			Convey("The ELB should be detached from AWS autoscaling groups", func() {
				So(elbDetachedFrom, ShouldEqual, "autoscaling-test")
			})

			Convey("The ELB should be deleted", func() {
				So(elbDeleted, ShouldEqual, "supergiant-example-com")
			})
		})
	})
}
