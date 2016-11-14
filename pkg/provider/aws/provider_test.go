package aws_test

import (
	"errors"
	"testing"

	"github.com/Sirupsen/logrus"
	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/supergiant/supergiant/pkg/core"
	"github.com/supergiant/supergiant/pkg/kubernetes"
	"github.com/supergiant/supergiant/pkg/model"
	"github.com/supergiant/supergiant/pkg/provider/aws"
	"github.com/supergiant/supergiant/test/fake_aws_provider"
	"github.com/supergiant/supergiant/test/fake_core"

	. "github.com/smartystreets/goconvey/convey"
)

type mockS3Client struct {
	s3iface.S3API
}

func TestAWSProviderValidateAccount(t *testing.T) {
	Convey("AWS Provider ValidateAccount works correctly", t, func() {
		table := []struct {
			// Input
			cloudAccount *model.CloudAccount
			// Mocks
			// Expectations
			err error
		}{
			// A successful example
			{
				// Input
				cloudAccount: &model.CloudAccount{},
			},
		}

		for _, item := range table {

			c := &core.Core{
				DB:  new(fake_core.DB),
				Log: logrus.New(),
			}

			provider := &aws.Provider{
				Core: c,
				EC2: func(kube *model.Kube) ec2iface.EC2API {
					return &fake_aws_provider.EC2{
						// NOTE won't matter until error is mocked
						DescribeKeyPairsFn: func(input *ec2.DescribeKeyPairsInput) (*ec2.DescribeKeyPairsOutput, error) {
							output := &ec2.DescribeKeyPairsOutput{}
							return output, nil
						},
					}
				},
			}

			err := provider.ValidateAccount(item.cloudAccount)

			So(err, ShouldResemble, item.err)
		}
	})
}

//------------------------------------------------------------------------------

//------------------------------------------------------------------------------

func TestAWSProviderDeleteNode(t *testing.T) {
	Convey("AWS Provider DeleteNode works correctly", t, func() {
		table := []struct {
			// Input
			node *model.Node
			// Mocks
			// Expectations
			err error
		}{
			// A successful example
			{
				// Input
				node: &model.Node{
					Kube: &model.Kube{
						NodeSizes: []string{"m4.large"},
						AWSConfig: &model.AWSKubeConfig{},
						// Relations
						Entrypoints: []*model.Entrypoint{
							{
								Name: "my-entrypoint",
							},
						},
					},
				},
			},
		}

		for _, item := range table {

			c := &core.Core{
				DB:  new(fake_core.DB),
				Log: logrus.New(),
			}

			provider := &aws.Provider{
				Core: c,
				EC2: func(kube *model.Kube) ec2iface.EC2API {
					return &fake_aws_provider.EC2{
						TerminateInstancesFn: func(input *ec2.TerminateInstancesInput) (*ec2.TerminateInstancesOutput, error) {
							output := &ec2.TerminateInstancesOutput{}
							return output, nil
						},
					}
				},
				ELB: func(kube *model.Kube) elbiface.ELBAPI {
					return &fake_aws_provider.ELB{
						DeregisterInstancesFromLoadBalancerFn: func(input *elb.DeregisterInstancesFromLoadBalancerInput) (*elb.DeregisterInstancesFromLoadBalancerOutput, error) {
							output := &elb.DeregisterInstancesFromLoadBalancerOutput{}
							return output, nil
						},
					}
				},
			}

			action := &core.Action{Status: new(model.ActionStatus)}
			err := provider.DeleteNode(item.node, action)

			So(err, ShouldResemble, item.err)
		}
	})
}

//------------------------------------------------------------------------------

func TestAWSProviderCreateVolume(t *testing.T) {
	Convey("AWS Provider CreateVolume works correctly", t, func() {
		table := []struct {
			// Input
			volume *model.Volume
			// Mocks
			// Expectations
			err error
		}{
			// A successful example
			{
				// Input
				volume: &model.Volume{
					Kube: &model.Kube{
						AWSConfig: &model.AWSKubeConfig{},
					},
				},
			},
		}

		for _, item := range table {

			c := &core.Core{
				DB:  new(fake_core.DB),
				Log: logrus.New(),
			}

			provider := &aws.Provider{
				Core: c,
				EC2: func(kube *model.Kube) ec2iface.EC2API {
					return &fake_aws_provider.EC2{
						CreateVolumeFn: func(input *ec2.CreateVolumeInput) (*ec2.Volume, error) {
							output := &ec2.Volume{
								VolumeId: awssdk.String("VolumeId"),
								Size:     awssdk.Int64(10),
							}
							return output, nil
						},
					}
				},
			}

			action := &core.Action{Status: new(model.ActionStatus)}
			err := provider.CreateVolume(item.volume, action)

			So(err, ShouldResemble, item.err)
		}
	})
}

//------------------------------------------------------------------------------

func TestAWSProviderKubernetesVolumeDefinition(t *testing.T) {
	Convey("AWS Provider KubernetesVolumeDefinition works correctly", t, func() {
		table := []struct {
			// Input
			volume *model.Volume
			// Mocks
			// Expectations
			kubeVol *kubernetes.Volume
		}{
			// A successful example
			{
				// Input
				volume: &model.Volume{
					Name:       "test",
					ProviderID: "provider-ID",
					Kube: &model.Kube{
						AWSConfig: &model.AWSKubeConfig{},
					},
				},
				// Expectations
				kubeVol: &kubernetes.Volume{
					Name: "test",
					AwsElasticBlockStore: &kubernetes.AwsElasticBlockStore{
						VolumeID: "provider-ID",
						FSType:   "ext4",
					},
				},
			},
		}

		for _, item := range table {

			c := &core.Core{
				DB:  new(fake_core.DB),
				Log: logrus.New(),
			}

			provider := &aws.Provider{
				Core: c,
			}

			kubeVol := provider.KubernetesVolumeDefinition(item.volume)

			So(kubeVol, ShouldResemble, item.kubeVol)
		}
	})
}

//------------------------------------------------------------------------------

func TestAWSProviderResizeVolume(t *testing.T) {
	Convey("AWS Provider ResizeVolume works correctly", t, func() {
		table := []struct {
			// Input
			volume *model.Volume
			// Mocks
			// Expectations
			err error
		}{
			// A successful example
			{
				// Input
				volume: &model.Volume{
					ProviderID: "ProviderID",
					Kube: &model.Kube{
						AWSConfig: &model.AWSKubeConfig{},
					},
				},
			},
		}

		for _, item := range table {

			c := &core.Core{
				DB:  new(fake_core.DB),
				Log: logrus.New(),
			}

			provider := &aws.Provider{
				Core: c,
				EC2: func(kube *model.Kube) ec2iface.EC2API {
					return &fake_aws_provider.EC2{
						CreateSnapshotFn: func(input *ec2.CreateSnapshotInput) (*ec2.Snapshot, error) {
							output := &ec2.Snapshot{
								SnapshotId: awssdk.String("SnapshotId"),
							}
							return output, nil
						},
						DescribeSnapshotsFn: func(input *ec2.DescribeSnapshotsInput) (*ec2.DescribeSnapshotsOutput, error) {
							output := &ec2.DescribeSnapshotsOutput{
								Snapshots: []*ec2.Snapshot{
									{
										SnapshotId: awssdk.String("SnapshotId"),
										State:      awssdk.String("completed"),
									},
								},
							}
							return output, nil
						},
						CreateVolumeFn: func(input *ec2.CreateVolumeInput) (*ec2.Volume, error) {
							output := &ec2.Volume{
								VolumeId: awssdk.String("VolumeId"),
								Size:     awssdk.Int64(10),
							}
							return output, nil
						},
						DeleteVolumeFn: func(input *ec2.DeleteVolumeInput) (*ec2.DeleteVolumeOutput, error) {
							output := &ec2.DeleteVolumeOutput{}
							return output, nil
						},
						DescribeVolumesFn: func(input *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error) {
							output := &ec2.DescribeVolumesOutput{
								Volumes: []*ec2.Volume{
									{
										State: awssdk.String("available"),
									},
								},
							}
							return output, nil
						},
					}
				},
			}

			action := &core.Action{Status: new(model.ActionStatus)}
			err := provider.ResizeVolume(item.volume, action)

			So(err, ShouldResemble, item.err)
		}
	})
}

//------------------------------------------------------------------------------

func TestAWSProviderWaitForVolumeAvailable(t *testing.T) {
	Convey("AWS Provider WaitForVolumeAvailable works correctly", t, func() {
		table := []struct {
			// Input
			volume *model.Volume
			// Mocks
			// Expectations
			err error
		}{
			// A successful example
			{
				// Input
				volume: &model.Volume{
					Kube: &model.Kube{
						AWSConfig: &model.AWSKubeConfig{},
					},
				},
			},
		}

		for _, item := range table {

			c := &core.Core{
				DB:  new(fake_core.DB),
				Log: logrus.New(),
			}

			provider := &aws.Provider{
				Core: c,
				EC2: func(kube *model.Kube) ec2iface.EC2API {
					return &fake_aws_provider.EC2{
						DescribeVolumesFn: func(input *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error) {
							output := &ec2.DescribeVolumesOutput{
								Volumes: []*ec2.Volume{
									{
										State: awssdk.String("available"),
									},
								},
							}
							return output, nil
						},
					}
				},
			}

			action := &core.Action{Status: new(model.ActionStatus)}
			err := provider.WaitForVolumeAvailable(item.volume, action)

			So(err, ShouldResemble, item.err)
		}
	})
}

//------------------------------------------------------------------------------

func TestAWSProviderDeleteVolume(t *testing.T) {
	Convey("AWS Provider DeleteVolume works correctly", t, func() {
		table := []struct {
			// Input
			volume *model.Volume
			// Mocks
			// Expectations
			err error
		}{
			// A successful example
			{
				// Input
				volume: &model.Volume{
					ProviderID: "ProviderID",
					Kube: &model.Kube{
						AWSConfig: &model.AWSKubeConfig{},
					},
				},
			},
		}

		for _, item := range table {

			c := &core.Core{
				DB:  new(fake_core.DB),
				Log: logrus.New(),
			}

			provider := &aws.Provider{
				Core: c,
				EC2: func(kube *model.Kube) ec2iface.EC2API {
					return &fake_aws_provider.EC2{
						DescribeVolumesFn: func(input *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error) {
							output := &ec2.DescribeVolumesOutput{
								Volumes: []*ec2.Volume{
									{
										State: awssdk.String("available"),
									},
								},
							}
							return output, nil
						},
						DeleteVolumeFn: func(input *ec2.DeleteVolumeInput) (*ec2.DeleteVolumeOutput, error) {
							output := &ec2.DeleteVolumeOutput{}
							return output, nil
						},
					}
				},
			}

			action := &core.Action{Status: new(model.ActionStatus)}
			err := provider.DeleteVolume(item.volume, action)

			So(err, ShouldResemble, item.err)
		}
	})
}

//------------------------------------------------------------------------------

func TestAWSProviderCreateEntrypoint(t *testing.T) {
	Convey("AWS Provider CreateEntrypoint works correctly", t, func() {
		table := []struct {
			// Input
			entrypoint *model.Entrypoint
			// Mocks
			// Expectations
			err error
		}{
			// A successful example
			{
				// Input
				entrypoint: &model.Entrypoint{
					Kube: &model.Kube{
						AWSConfig: &model.AWSKubeConfig{},
						// Relations
						Nodes: []*model.Node{
							{
								Name: "a-node.host",
							},
						},
					},
				},
			},
		}

		for _, item := range table {

			c := &core.Core{
				DB:  new(fake_core.DB),
				Log: logrus.New(),
			}

			provider := &aws.Provider{
				Core: c,
				ELB: func(kube *model.Kube) elbiface.ELBAPI {
					return &fake_aws_provider.ELB{
						CreateLoadBalancerFn: func(input *elb.CreateLoadBalancerInput) (*elb.CreateLoadBalancerOutput, error) {
							output := &elb.CreateLoadBalancerOutput{
								DNSName: awssdk.String("elb.dns.blah.blah"),
							}
							return output, nil
						},
						// NOTE for both of the following, it doesn't really matter due to
						// how we mock (we don't care about the return, just the error
						// here). It will only matter once we mock an error.
						RegisterInstancesWithLoadBalancerFn: func(input *elb.RegisterInstancesWithLoadBalancerInput) (*elb.RegisterInstancesWithLoadBalancerOutput, error) {
							output := &elb.RegisterInstancesWithLoadBalancerOutput{}
							return output, nil
						},
						ConfigureHealthCheckFn: func(input *elb.ConfigureHealthCheckInput) (*elb.ConfigureHealthCheckOutput, error) {
							output := &elb.ConfigureHealthCheckOutput{}
							return output, nil
						},
					}
				},
			}

			action := &core.Action{Status: new(model.ActionStatus)}
			err := provider.CreateEntrypoint(item.entrypoint, action)

			So(err, ShouldResemble, item.err)
		}
	})
}

//------------------------------------------------------------------------------

func TestAWSProviderDeleteEntrypoint(t *testing.T) {
	Convey("AWS Provider DeleteEntrypoint works correctly", t, func() {
		table := []struct {
			// Input
			entrypoint *model.Entrypoint
			// Mocks
			mockDeleteLoadBalancerError error
			// Expectations
			err error
		}{
			// A successful example
			{
				// Input
				entrypoint: &model.Entrypoint{
					Kube: &model.Kube{
						AWSConfig: &model.AWSKubeConfig{},
						// Relations
						Nodes: []*model.Node{
							{
								Name: "a-node.host",
							},
						},
					},
				},
				err: nil,
			},

			// On DeleteLoadBalancer error
			{
				// Input
				entrypoint: &model.Entrypoint{
					Kube: &model.Kube{
						AWSConfig: &model.AWSKubeConfig{},
						// Relations
						Nodes: []*model.Node{
							{
								Name: "a-node.host",
							},
						},
					},
				},
				mockDeleteLoadBalancerError: errors.New("ERROR"),
				err: errors.New("ERROR"),
			},
		}

		for _, item := range table {

			c := &core.Core{
				DB:  new(fake_core.DB),
				Log: logrus.New(),
			}

			provider := &aws.Provider{
				Core: c,
				ELB: func(kube *model.Kube) elbiface.ELBAPI {
					return &fake_aws_provider.ELB{
						DeleteLoadBalancerFn: func(input *elb.DeleteLoadBalancerInput) (*elb.DeleteLoadBalancerOutput, error) {
							output := &elb.DeleteLoadBalancerOutput{}
							return output, item.mockDeleteLoadBalancerError
						},
					}
				},
			}

			action := &core.Action{Status: new(model.ActionStatus)}
			err := provider.DeleteEntrypoint(item.entrypoint, action)

			So(err, ShouldResemble, item.err)
		}
	})
}

//------------------------------------------------------------------------------

func TestAWSProviderCreateEntrypointListener(t *testing.T) {
	Convey("AWS Provider CreateEntrypointListener works correctly", t, func() {
		table := []struct {
			// Input
			entrypointListener *model.EntrypointListener
			// Mocks
			// Expectations
			err error
		}{
			// A successful example
			{
				// Input
				entrypointListener: &model.EntrypointListener{
					Entrypoint: &model.Entrypoint{
						Kube: &model.Kube{
							AWSConfig: &model.AWSKubeConfig{},
						},
					},
				},
			},
		}

		for _, item := range table {

			c := &core.Core{
				DB:  new(fake_core.DB),
				Log: logrus.New(),
			}

			provider := &aws.Provider{
				Core: c,
				ELB: func(kube *model.Kube) elbiface.ELBAPI {
					return &fake_aws_provider.ELB{
						// NOTE won't matter until error is mocked
						CreateLoadBalancerListenersFn: func(input *elb.CreateLoadBalancerListenersInput) (*elb.CreateLoadBalancerListenersOutput, error) {
							output := &elb.CreateLoadBalancerListenersOutput{}
							return output, nil
						},
					}
				},
			}

			action := &core.Action{Status: new(model.ActionStatus)}
			err := provider.CreateEntrypointListener(item.entrypointListener, action)

			So(err, ShouldResemble, item.err)
		}
	})
}

//------------------------------------------------------------------------------

func TestAWSProviderDeleteEntrypointListener(t *testing.T) {
	Convey("AWS Provider DeleteEntrypointListener works correctly", t, func() {
		table := []struct {
			// Input
			entrypointListener *model.EntrypointListener
			// Mocks
			mockDeleteEntrypointListenerError error
			// Expectations
			err error
		}{
			// A successful example
			{
				// Input
				entrypointListener: &model.EntrypointListener{
					Entrypoint: &model.Entrypoint{
						Kube: &model.Kube{
							AWSConfig: &model.AWSKubeConfig{},
						},
					},
				},
			},

			// On non-404 error
			{
				// Input
				entrypointListener: &model.EntrypointListener{
					Entrypoint: &model.Entrypoint{
						Kube: &model.Kube{
							AWSConfig: &model.AWSKubeConfig{},
						},
					},
				},
				mockDeleteEntrypointListenerError: errors.New("ERROR"),
				err: errors.New("ERROR"),
			},
		}

		for _, item := range table {

			c := &core.Core{
				DB:  new(fake_core.DB),
				Log: logrus.New(),
			}

			provider := &aws.Provider{
				Core: c,
				ELB: func(kube *model.Kube) elbiface.ELBAPI {
					return &fake_aws_provider.ELB{
						// NOTE won't matter until error is mocked
						DeleteLoadBalancerListenersFn: func(input *elb.DeleteLoadBalancerListenersInput) (*elb.DeleteLoadBalancerListenersOutput, error) {
							output := &elb.DeleteLoadBalancerListenersOutput{}
							return output, item.mockDeleteEntrypointListenerError
						},
					}
				},
			}

			action := &core.Action{Status: new(model.ActionStatus)}
			err := provider.DeleteEntrypointListener(item.entrypointListener, action)

			So(err, ShouldResemble, item.err)
		}
	})
}
