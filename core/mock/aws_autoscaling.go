package mock

import (
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/autoscaling"
)

func (f *FakeAwsAutoscaling) ReturnOnDescribeAutoScalingGroups(groups []*autoscaling.Group, err error) *FakeAwsAutoscaling {
	f.DescribeAutoScalingGroupsFn = func() (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
		return &autoscaling.DescribeAutoScalingGroupsOutput{
			AutoScalingGroups: groups,
		}, err
	}
	return f
}

func (f *FakeAwsAutoscaling) OnAttachLoadBalancers(clbk func(*autoscaling.AttachLoadBalancersInput) (*autoscaling.AttachLoadBalancersOutput, error)) *FakeAwsAutoscaling {
	f.AttachLoadBalancersFn = func(input *autoscaling.AttachLoadBalancersInput) (*autoscaling.AttachLoadBalancersOutput, error) {
		return clbk(input)
	}
	return f
}

func (f *FakeAwsAutoscaling) OnDetachLoadBalancers(clbk func(*autoscaling.DetachLoadBalancersInput) (*autoscaling.DetachLoadBalancersOutput, error)) *FakeAwsAutoscaling {
	f.DetachLoadBalancersFn = func(input *autoscaling.DetachLoadBalancersInput) (*autoscaling.DetachLoadBalancersOutput, error) {
		return clbk(input)
	}
	return f
}

type FakeAwsAutoscaling struct {
	DescribeAutoScalingGroupsFn func() (*autoscaling.DescribeAutoScalingGroupsOutput, error)
	AttachLoadBalancersFn       func(*autoscaling.AttachLoadBalancersInput) (*autoscaling.AttachLoadBalancersOutput, error)
	DetachLoadBalancersFn       func(*autoscaling.DetachLoadBalancersInput) (*autoscaling.DetachLoadBalancersOutput, error)
}

func (f *FakeAwsAutoscaling) AttachInstancesRequest(*autoscaling.AttachInstancesInput) (*request.Request, *autoscaling.AttachInstancesOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) AttachInstances(*autoscaling.AttachInstancesInput) (*autoscaling.AttachInstancesOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) AttachLoadBalancersRequest(*autoscaling.AttachLoadBalancersInput) (*request.Request, *autoscaling.AttachLoadBalancersOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) AttachLoadBalancers(input *autoscaling.AttachLoadBalancersInput) (*autoscaling.AttachLoadBalancersOutput, error) {
	return f.AttachLoadBalancersFn(input)
}

func (f *FakeAwsAutoscaling) CompleteLifecycleActionRequest(*autoscaling.CompleteLifecycleActionInput) (*request.Request, *autoscaling.CompleteLifecycleActionOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) CompleteLifecycleAction(*autoscaling.CompleteLifecycleActionInput) (*autoscaling.CompleteLifecycleActionOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) CreateAutoScalingGroupRequest(*autoscaling.CreateAutoScalingGroupInput) (*request.Request, *autoscaling.CreateAutoScalingGroupOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) CreateAutoScalingGroup(*autoscaling.CreateAutoScalingGroupInput) (*autoscaling.CreateAutoScalingGroupOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) CreateLaunchConfigurationRequest(*autoscaling.CreateLaunchConfigurationInput) (*request.Request, *autoscaling.CreateLaunchConfigurationOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) CreateLaunchConfiguration(*autoscaling.CreateLaunchConfigurationInput) (*autoscaling.CreateLaunchConfigurationOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) CreateOrUpdateTagsRequest(*autoscaling.CreateOrUpdateTagsInput) (*request.Request, *autoscaling.CreateOrUpdateTagsOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) CreateOrUpdateTags(*autoscaling.CreateOrUpdateTagsInput) (*autoscaling.CreateOrUpdateTagsOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DeleteAutoScalingGroupRequest(*autoscaling.DeleteAutoScalingGroupInput) (*request.Request, *autoscaling.DeleteAutoScalingGroupOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DeleteAutoScalingGroup(*autoscaling.DeleteAutoScalingGroupInput) (*autoscaling.DeleteAutoScalingGroupOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DeleteLaunchConfigurationRequest(*autoscaling.DeleteLaunchConfigurationInput) (*request.Request, *autoscaling.DeleteLaunchConfigurationOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DeleteLaunchConfiguration(*autoscaling.DeleteLaunchConfigurationInput) (*autoscaling.DeleteLaunchConfigurationOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DeleteLifecycleHookRequest(*autoscaling.DeleteLifecycleHookInput) (*request.Request, *autoscaling.DeleteLifecycleHookOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DeleteLifecycleHook(*autoscaling.DeleteLifecycleHookInput) (*autoscaling.DeleteLifecycleHookOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DeleteNotificationConfigurationRequest(*autoscaling.DeleteNotificationConfigurationInput) (*request.Request, *autoscaling.DeleteNotificationConfigurationOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DeleteNotificationConfiguration(*autoscaling.DeleteNotificationConfigurationInput) (*autoscaling.DeleteNotificationConfigurationOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DeletePolicyRequest(*autoscaling.DeletePolicyInput) (*request.Request, *autoscaling.DeletePolicyOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DeletePolicy(*autoscaling.DeletePolicyInput) (*autoscaling.DeletePolicyOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DeleteScheduledActionRequest(*autoscaling.DeleteScheduledActionInput) (*request.Request, *autoscaling.DeleteScheduledActionOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DeleteScheduledAction(*autoscaling.DeleteScheduledActionInput) (*autoscaling.DeleteScheduledActionOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DeleteTagsRequest(*autoscaling.DeleteTagsInput) (*request.Request, *autoscaling.DeleteTagsOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DeleteTags(*autoscaling.DeleteTagsInput) (*autoscaling.DeleteTagsOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeAccountLimitsRequest(*autoscaling.DescribeAccountLimitsInput) (*request.Request, *autoscaling.DescribeAccountLimitsOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeAccountLimits(*autoscaling.DescribeAccountLimitsInput) (*autoscaling.DescribeAccountLimitsOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeAdjustmentTypesRequest(*autoscaling.DescribeAdjustmentTypesInput) (*request.Request, *autoscaling.DescribeAdjustmentTypesOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeAdjustmentTypes(*autoscaling.DescribeAdjustmentTypesInput) (*autoscaling.DescribeAdjustmentTypesOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeAutoScalingGroupsRequest(*autoscaling.DescribeAutoScalingGroupsInput) (*request.Request, *autoscaling.DescribeAutoScalingGroupsOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeAutoScalingGroups(*autoscaling.DescribeAutoScalingGroupsInput) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	return f.DescribeAutoScalingGroupsFn()
}

func (f *FakeAwsAutoscaling) DescribeAutoScalingGroupsPages(*autoscaling.DescribeAutoScalingGroupsInput, func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool) error {
	return nil
}

func (f *FakeAwsAutoscaling) DescribeAutoScalingInstancesRequest(*autoscaling.DescribeAutoScalingInstancesInput) (*request.Request, *autoscaling.DescribeAutoScalingInstancesOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeAutoScalingInstances(*autoscaling.DescribeAutoScalingInstancesInput) (*autoscaling.DescribeAutoScalingInstancesOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeAutoScalingInstancesPages(*autoscaling.DescribeAutoScalingInstancesInput, func(*autoscaling.DescribeAutoScalingInstancesOutput, bool) bool) error {
	return nil
}

func (f *FakeAwsAutoscaling) DescribeAutoScalingNotificationTypesRequest(*autoscaling.DescribeAutoScalingNotificationTypesInput) (*request.Request, *autoscaling.DescribeAutoScalingNotificationTypesOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeAutoScalingNotificationTypes(*autoscaling.DescribeAutoScalingNotificationTypesInput) (*autoscaling.DescribeAutoScalingNotificationTypesOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeLaunchConfigurationsRequest(*autoscaling.DescribeLaunchConfigurationsInput) (*request.Request, *autoscaling.DescribeLaunchConfigurationsOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeLaunchConfigurations(*autoscaling.DescribeLaunchConfigurationsInput) (*autoscaling.DescribeLaunchConfigurationsOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeLaunchConfigurationsPages(*autoscaling.DescribeLaunchConfigurationsInput, func(*autoscaling.DescribeLaunchConfigurationsOutput, bool) bool) error {
	return nil
}

func (f *FakeAwsAutoscaling) DescribeLifecycleHookTypesRequest(*autoscaling.DescribeLifecycleHookTypesInput) (*request.Request, *autoscaling.DescribeLifecycleHookTypesOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeLifecycleHookTypes(*autoscaling.DescribeLifecycleHookTypesInput) (*autoscaling.DescribeLifecycleHookTypesOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeLifecycleHooksRequest(*autoscaling.DescribeLifecycleHooksInput) (*request.Request, *autoscaling.DescribeLifecycleHooksOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeLifecycleHooks(*autoscaling.DescribeLifecycleHooksInput) (*autoscaling.DescribeLifecycleHooksOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeLoadBalancersRequest(*autoscaling.DescribeLoadBalancersInput) (*request.Request, *autoscaling.DescribeLoadBalancersOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeLoadBalancers(*autoscaling.DescribeLoadBalancersInput) (*autoscaling.DescribeLoadBalancersOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeMetricCollectionTypesRequest(*autoscaling.DescribeMetricCollectionTypesInput) (*request.Request, *autoscaling.DescribeMetricCollectionTypesOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeMetricCollectionTypes(*autoscaling.DescribeMetricCollectionTypesInput) (*autoscaling.DescribeMetricCollectionTypesOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeNotificationConfigurationsRequest(*autoscaling.DescribeNotificationConfigurationsInput) (*request.Request, *autoscaling.DescribeNotificationConfigurationsOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeNotificationConfigurations(*autoscaling.DescribeNotificationConfigurationsInput) (*autoscaling.DescribeNotificationConfigurationsOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeNotificationConfigurationsPages(*autoscaling.DescribeNotificationConfigurationsInput, func(*autoscaling.DescribeNotificationConfigurationsOutput, bool) bool) error {
	return nil
}

func (f *FakeAwsAutoscaling) DescribePoliciesRequest(*autoscaling.DescribePoliciesInput) (*request.Request, *autoscaling.DescribePoliciesOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribePolicies(*autoscaling.DescribePoliciesInput) (*autoscaling.DescribePoliciesOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribePoliciesPages(*autoscaling.DescribePoliciesInput, func(*autoscaling.DescribePoliciesOutput, bool) bool) error {
	return nil
}

func (f *FakeAwsAutoscaling) DescribeScalingActivitiesRequest(*autoscaling.DescribeScalingActivitiesInput) (*request.Request, *autoscaling.DescribeScalingActivitiesOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeScalingActivities(*autoscaling.DescribeScalingActivitiesInput) (*autoscaling.DescribeScalingActivitiesOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeScalingActivitiesPages(*autoscaling.DescribeScalingActivitiesInput, func(*autoscaling.DescribeScalingActivitiesOutput, bool) bool) error {
	return nil
}

func (f *FakeAwsAutoscaling) DescribeScalingProcessTypesRequest(*autoscaling.DescribeScalingProcessTypesInput) (*request.Request, *autoscaling.DescribeScalingProcessTypesOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeScalingProcessTypes(*autoscaling.DescribeScalingProcessTypesInput) (*autoscaling.DescribeScalingProcessTypesOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeScheduledActionsRequest(*autoscaling.DescribeScheduledActionsInput) (*request.Request, *autoscaling.DescribeScheduledActionsOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeScheduledActions(*autoscaling.DescribeScheduledActionsInput) (*autoscaling.DescribeScheduledActionsOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeScheduledActionsPages(*autoscaling.DescribeScheduledActionsInput, func(*autoscaling.DescribeScheduledActionsOutput, bool) bool) error {
	return nil
}

func (f *FakeAwsAutoscaling) DescribeTagsRequest(*autoscaling.DescribeTagsInput) (*request.Request, *autoscaling.DescribeTagsOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeTags(*autoscaling.DescribeTagsInput) (*autoscaling.DescribeTagsOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeTagsPages(*autoscaling.DescribeTagsInput, func(*autoscaling.DescribeTagsOutput, bool) bool) error {
	return nil
}

func (f *FakeAwsAutoscaling) DescribeTerminationPolicyTypesRequest(*autoscaling.DescribeTerminationPolicyTypesInput) (*request.Request, *autoscaling.DescribeTerminationPolicyTypesOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DescribeTerminationPolicyTypes(*autoscaling.DescribeTerminationPolicyTypesInput) (*autoscaling.DescribeTerminationPolicyTypesOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DetachInstancesRequest(*autoscaling.DetachInstancesInput) (*request.Request, *autoscaling.DetachInstancesOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DetachInstances(*autoscaling.DetachInstancesInput) (*autoscaling.DetachInstancesOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DetachLoadBalancersRequest(*autoscaling.DetachLoadBalancersInput) (*request.Request, *autoscaling.DetachLoadBalancersOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DetachLoadBalancers(input *autoscaling.DetachLoadBalancersInput) (*autoscaling.DetachLoadBalancersOutput, error) {
	return f.DetachLoadBalancersFn(input)
}

func (f *FakeAwsAutoscaling) DisableMetricsCollectionRequest(*autoscaling.DisableMetricsCollectionInput) (*request.Request, *autoscaling.DisableMetricsCollectionOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) DisableMetricsCollection(*autoscaling.DisableMetricsCollectionInput) (*autoscaling.DisableMetricsCollectionOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) EnableMetricsCollectionRequest(*autoscaling.EnableMetricsCollectionInput) (*request.Request, *autoscaling.EnableMetricsCollectionOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) EnableMetricsCollection(*autoscaling.EnableMetricsCollectionInput) (*autoscaling.EnableMetricsCollectionOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) EnterStandbyRequest(*autoscaling.EnterStandbyInput) (*request.Request, *autoscaling.EnterStandbyOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) EnterStandby(*autoscaling.EnterStandbyInput) (*autoscaling.EnterStandbyOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) ExecutePolicyRequest(*autoscaling.ExecutePolicyInput) (*request.Request, *autoscaling.ExecutePolicyOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) ExecutePolicy(*autoscaling.ExecutePolicyInput) (*autoscaling.ExecutePolicyOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) ExitStandbyRequest(*autoscaling.ExitStandbyInput) (*request.Request, *autoscaling.ExitStandbyOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) ExitStandby(*autoscaling.ExitStandbyInput) (*autoscaling.ExitStandbyOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) PutLifecycleHookRequest(*autoscaling.PutLifecycleHookInput) (*request.Request, *autoscaling.PutLifecycleHookOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) PutLifecycleHook(*autoscaling.PutLifecycleHookInput) (*autoscaling.PutLifecycleHookOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) PutNotificationConfigurationRequest(*autoscaling.PutNotificationConfigurationInput) (*request.Request, *autoscaling.PutNotificationConfigurationOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) PutNotificationConfiguration(*autoscaling.PutNotificationConfigurationInput) (*autoscaling.PutNotificationConfigurationOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) PutScalingPolicyRequest(*autoscaling.PutScalingPolicyInput) (*request.Request, *autoscaling.PutScalingPolicyOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) PutScalingPolicy(*autoscaling.PutScalingPolicyInput) (*autoscaling.PutScalingPolicyOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) PutScheduledUpdateGroupActionRequest(*autoscaling.PutScheduledUpdateGroupActionInput) (*request.Request, *autoscaling.PutScheduledUpdateGroupActionOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) PutScheduledUpdateGroupAction(*autoscaling.PutScheduledUpdateGroupActionInput) (*autoscaling.PutScheduledUpdateGroupActionOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) RecordLifecycleActionHeartbeatRequest(*autoscaling.RecordLifecycleActionHeartbeatInput) (*request.Request, *autoscaling.RecordLifecycleActionHeartbeatOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) RecordLifecycleActionHeartbeat(*autoscaling.RecordLifecycleActionHeartbeatInput) (*autoscaling.RecordLifecycleActionHeartbeatOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) ResumeProcessesRequest(*autoscaling.ScalingProcessQuery) (*request.Request, *autoscaling.ResumeProcessesOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) ResumeProcesses(*autoscaling.ScalingProcessQuery) (*autoscaling.ResumeProcessesOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) SetDesiredCapacityRequest(*autoscaling.SetDesiredCapacityInput) (*request.Request, *autoscaling.SetDesiredCapacityOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) SetDesiredCapacity(*autoscaling.SetDesiredCapacityInput) (*autoscaling.SetDesiredCapacityOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) SetInstanceHealthRequest(*autoscaling.SetInstanceHealthInput) (*request.Request, *autoscaling.SetInstanceHealthOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) SetInstanceHealth(*autoscaling.SetInstanceHealthInput) (*autoscaling.SetInstanceHealthOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) SetInstanceProtectionRequest(*autoscaling.SetInstanceProtectionInput) (*request.Request, *autoscaling.SetInstanceProtectionOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) SetInstanceProtection(*autoscaling.SetInstanceProtectionInput) (*autoscaling.SetInstanceProtectionOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) SuspendProcessesRequest(*autoscaling.ScalingProcessQuery) (*request.Request, *autoscaling.SuspendProcessesOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) SuspendProcesses(*autoscaling.ScalingProcessQuery) (*autoscaling.SuspendProcessesOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) TerminateInstanceInAutoScalingGroupRequest(*autoscaling.TerminateInstanceInAutoScalingGroupInput) (*request.Request, *autoscaling.TerminateInstanceInAutoScalingGroupOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) TerminateInstanceInAutoScalingGroup(*autoscaling.TerminateInstanceInAutoScalingGroupInput) (*autoscaling.TerminateInstanceInAutoScalingGroupOutput, error) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) UpdateAutoScalingGroupRequest(*autoscaling.UpdateAutoScalingGroupInput) (*request.Request, *autoscaling.UpdateAutoScalingGroupOutput) {
	return nil, nil
}

func (f *FakeAwsAutoscaling) UpdateAutoScalingGroup(*autoscaling.UpdateAutoScalingGroupInput) (*autoscaling.UpdateAutoScalingGroupOutput, error) {
	return nil, nil
}
