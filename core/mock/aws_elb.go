package mock

import (
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/elb"
)

func (f *FakeAwsELB) OnCreateLoadBalancer(clbk func(*elb.CreateLoadBalancerInput) (*elb.CreateLoadBalancerOutput, error)) *FakeAwsELB {
	f.CreateLoadBalancerFn = func(input *elb.CreateLoadBalancerInput) (*elb.CreateLoadBalancerOutput, error) {
		return clbk(input)
	}
	return f
}

func (f *FakeAwsELB) OnDeleteLoadBalancer(clbk func(*elb.DeleteLoadBalancerInput) (*elb.DeleteLoadBalancerOutput, error)) *FakeAwsELB {
	f.DeleteLoadBalancerFn = func(input *elb.DeleteLoadBalancerInput) (*elb.DeleteLoadBalancerOutput, error) {
		return clbk(input)
	}
	return f
}

func (f *FakeAwsELB) OnConfigureHealthCheck(clbk func(*elb.ConfigureHealthCheckInput) (*elb.ConfigureHealthCheckOutput, error)) *FakeAwsELB {
	f.ConfigureHealthCheckFn = func(input *elb.ConfigureHealthCheckInput) (*elb.ConfigureHealthCheckOutput, error) {
		return clbk(input)
	}
	return f
}

func (f *FakeAwsELB) OnDeleteLoadBalancerListeners(clbk func(*elb.DeleteLoadBalancerListenersInput) (*elb.DeleteLoadBalancerListenersOutput, error)) *FakeAwsELB {
	f.DeleteLoadBalancerListenersFn = func(input *elb.DeleteLoadBalancerListenersInput) (*elb.DeleteLoadBalancerListenersOutput, error) {
		return clbk(input)
	}
	return f
}

type FakeAwsELB struct {
	CreateLoadBalancerFn          func(*elb.CreateLoadBalancerInput) (*elb.CreateLoadBalancerOutput, error)
	DeleteLoadBalancerFn          func(*elb.DeleteLoadBalancerInput) (*elb.DeleteLoadBalancerOutput, error)
	ConfigureHealthCheckFn        func(*elb.ConfigureHealthCheckInput) (*elb.ConfigureHealthCheckOutput, error)
	DeleteLoadBalancerListenersFn func(*elb.DeleteLoadBalancerListenersInput) (*elb.DeleteLoadBalancerListenersOutput, error)
}

func (f *FakeAwsELB) AddTagsRequest(*elb.AddTagsInput) (*request.Request, *elb.AddTagsOutput) {
	return nil, nil
}

func (f *FakeAwsELB) AddTags(*elb.AddTagsInput) (*elb.AddTagsOutput, error) {
	return nil, nil
}

func (f *FakeAwsELB) ApplySecurityGroupsToLoadBalancerRequest(*elb.ApplySecurityGroupsToLoadBalancerInput) (*request.Request, *elb.ApplySecurityGroupsToLoadBalancerOutput) {
	return nil, nil
}

func (f *FakeAwsELB) ApplySecurityGroupsToLoadBalancer(*elb.ApplySecurityGroupsToLoadBalancerInput) (*elb.ApplySecurityGroupsToLoadBalancerOutput, error) {
	return nil, nil
}

func (f *FakeAwsELB) AttachLoadBalancerToSubnetsRequest(*elb.AttachLoadBalancerToSubnetsInput) (*request.Request, *elb.AttachLoadBalancerToSubnetsOutput) {
	return nil, nil
}

func (f *FakeAwsELB) AttachLoadBalancerToSubnets(*elb.AttachLoadBalancerToSubnetsInput) (*elb.AttachLoadBalancerToSubnetsOutput, error) {
	return nil, nil
}

func (f *FakeAwsELB) ConfigureHealthCheckRequest(*elb.ConfigureHealthCheckInput) (*request.Request, *elb.ConfigureHealthCheckOutput) {
	return nil, nil
}

func (f *FakeAwsELB) ConfigureHealthCheck(input *elb.ConfigureHealthCheckInput) (*elb.ConfigureHealthCheckOutput, error) {
	return f.ConfigureHealthCheckFn(input)
}

func (f *FakeAwsELB) CreateAppCookieStickinessPolicyRequest(*elb.CreateAppCookieStickinessPolicyInput) (*request.Request, *elb.CreateAppCookieStickinessPolicyOutput) {
	return nil, nil
}

func (f *FakeAwsELB) CreateAppCookieStickinessPolicy(*elb.CreateAppCookieStickinessPolicyInput) (*elb.CreateAppCookieStickinessPolicyOutput, error) {
	return nil, nil
}

func (f *FakeAwsELB) CreateLBCookieStickinessPolicyRequest(*elb.CreateLBCookieStickinessPolicyInput) (*request.Request, *elb.CreateLBCookieStickinessPolicyOutput) {
	return nil, nil
}

func (f *FakeAwsELB) CreateLBCookieStickinessPolicy(*elb.CreateLBCookieStickinessPolicyInput) (*elb.CreateLBCookieStickinessPolicyOutput, error) {
	return nil, nil
}

func (f *FakeAwsELB) CreateLoadBalancerRequest(*elb.CreateLoadBalancerInput) (*request.Request, *elb.CreateLoadBalancerOutput) {
	return nil, nil
}

func (f *FakeAwsELB) CreateLoadBalancer(input *elb.CreateLoadBalancerInput) (*elb.CreateLoadBalancerOutput, error) {
	return f.CreateLoadBalancerFn(input)
}

func (f *FakeAwsELB) CreateLoadBalancerListenersRequest(*elb.CreateLoadBalancerListenersInput) (*request.Request, *elb.CreateLoadBalancerListenersOutput) {
	return nil, nil
}

func (f *FakeAwsELB) CreateLoadBalancerListeners(*elb.CreateLoadBalancerListenersInput) (*elb.CreateLoadBalancerListenersOutput, error) {
	return nil, nil
}

func (f *FakeAwsELB) CreateLoadBalancerPolicyRequest(*elb.CreateLoadBalancerPolicyInput) (*request.Request, *elb.CreateLoadBalancerPolicyOutput) {
	return nil, nil
}

func (f *FakeAwsELB) CreateLoadBalancerPolicy(*elb.CreateLoadBalancerPolicyInput) (*elb.CreateLoadBalancerPolicyOutput, error) {
	return nil, nil
}

func (f *FakeAwsELB) DeleteLoadBalancerRequest(*elb.DeleteLoadBalancerInput) (*request.Request, *elb.DeleteLoadBalancerOutput) {
	return nil, nil
}

func (f *FakeAwsELB) DeleteLoadBalancer(input *elb.DeleteLoadBalancerInput) (*elb.DeleteLoadBalancerOutput, error) {
	return f.DeleteLoadBalancerFn(input)
}

func (f *FakeAwsELB) DeleteLoadBalancerListenersRequest(*elb.DeleteLoadBalancerListenersInput) (*request.Request, *elb.DeleteLoadBalancerListenersOutput) {
	return nil, nil
}

func (f *FakeAwsELB) DeleteLoadBalancerListeners(input *elb.DeleteLoadBalancerListenersInput) (*elb.DeleteLoadBalancerListenersOutput, error) {
	return f.DeleteLoadBalancerListenersFn(input)
}

func (f *FakeAwsELB) DeleteLoadBalancerPolicyRequest(*elb.DeleteLoadBalancerPolicyInput) (*request.Request, *elb.DeleteLoadBalancerPolicyOutput) {
	return nil, nil
}

func (f *FakeAwsELB) DeleteLoadBalancerPolicy(*elb.DeleteLoadBalancerPolicyInput) (*elb.DeleteLoadBalancerPolicyOutput, error) {
	return nil, nil
}

func (f *FakeAwsELB) DeregisterInstancesFromLoadBalancerRequest(*elb.DeregisterInstancesFromLoadBalancerInput) (*request.Request, *elb.DeregisterInstancesFromLoadBalancerOutput) {
	return nil, nil
}

func (f *FakeAwsELB) DeregisterInstancesFromLoadBalancer(*elb.DeregisterInstancesFromLoadBalancerInput) (*elb.DeregisterInstancesFromLoadBalancerOutput, error) {
	return nil, nil
}

func (f *FakeAwsELB) DescribeInstanceHealthRequest(*elb.DescribeInstanceHealthInput) (*request.Request, *elb.DescribeInstanceHealthOutput) {
	return nil, nil
}

func (f *FakeAwsELB) DescribeInstanceHealth(*elb.DescribeInstanceHealthInput) (*elb.DescribeInstanceHealthOutput, error) {
	return nil, nil
}

func (f *FakeAwsELB) DescribeLoadBalancerAttributesRequest(*elb.DescribeLoadBalancerAttributesInput) (*request.Request, *elb.DescribeLoadBalancerAttributesOutput) {
	return nil, nil
}

func (f *FakeAwsELB) DescribeLoadBalancerAttributes(*elb.DescribeLoadBalancerAttributesInput) (*elb.DescribeLoadBalancerAttributesOutput, error) {
	return nil, nil
}

func (f *FakeAwsELB) DescribeLoadBalancerPoliciesRequest(*elb.DescribeLoadBalancerPoliciesInput) (*request.Request, *elb.DescribeLoadBalancerPoliciesOutput) {
	return nil, nil
}

func (f *FakeAwsELB) DescribeLoadBalancerPolicies(*elb.DescribeLoadBalancerPoliciesInput) (*elb.DescribeLoadBalancerPoliciesOutput, error) {
	return nil, nil
}

func (f *FakeAwsELB) DescribeLoadBalancerPolicyTypesRequest(*elb.DescribeLoadBalancerPolicyTypesInput) (*request.Request, *elb.DescribeLoadBalancerPolicyTypesOutput) {
	return nil, nil
}

func (f *FakeAwsELB) DescribeLoadBalancerPolicyTypes(*elb.DescribeLoadBalancerPolicyTypesInput) (*elb.DescribeLoadBalancerPolicyTypesOutput, error) {
	return nil, nil
}

func (f *FakeAwsELB) DescribeLoadBalancersRequest(*elb.DescribeLoadBalancersInput) (*request.Request, *elb.DescribeLoadBalancersOutput) {
	return nil, nil
}

func (f *FakeAwsELB) DescribeLoadBalancers(*elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error) {
	return nil, nil
}

func (f *FakeAwsELB) DescribeLoadBalancersPages(*elb.DescribeLoadBalancersInput, func(*elb.DescribeLoadBalancersOutput, bool) bool) error {
	return nil
}

func (f *FakeAwsELB) DescribeTagsRequest(*elb.DescribeTagsInput) (*request.Request, *elb.DescribeTagsOutput) {
	return nil, nil
}

func (f *FakeAwsELB) DescribeTags(*elb.DescribeTagsInput) (*elb.DescribeTagsOutput, error) {
	return nil, nil
}

func (f *FakeAwsELB) DetachLoadBalancerFromSubnetsRequest(*elb.DetachLoadBalancerFromSubnetsInput) (*request.Request, *elb.DetachLoadBalancerFromSubnetsOutput) {
	return nil, nil
}

func (f *FakeAwsELB) DetachLoadBalancerFromSubnets(*elb.DetachLoadBalancerFromSubnetsInput) (*elb.DetachLoadBalancerFromSubnetsOutput, error) {
	return nil, nil
}

func (f *FakeAwsELB) DisableAvailabilityZonesForLoadBalancerRequest(*elb.DisableAvailabilityZonesForLoadBalancerInput) (*request.Request, *elb.DisableAvailabilityZonesForLoadBalancerOutput) {
	return nil, nil
}

func (f *FakeAwsELB) DisableAvailabilityZonesForLoadBalancer(*elb.DisableAvailabilityZonesForLoadBalancerInput) (*elb.DisableAvailabilityZonesForLoadBalancerOutput, error) {
	return nil, nil
}

func (f *FakeAwsELB) EnableAvailabilityZonesForLoadBalancerRequest(*elb.EnableAvailabilityZonesForLoadBalancerInput) (*request.Request, *elb.EnableAvailabilityZonesForLoadBalancerOutput) {
	return nil, nil
}

func (f *FakeAwsELB) EnableAvailabilityZonesForLoadBalancer(*elb.EnableAvailabilityZonesForLoadBalancerInput) (*elb.EnableAvailabilityZonesForLoadBalancerOutput, error) {
	return nil, nil
}

func (f *FakeAwsELB) ModifyLoadBalancerAttributesRequest(*elb.ModifyLoadBalancerAttributesInput) (*request.Request, *elb.ModifyLoadBalancerAttributesOutput) {
	return nil, nil
}

func (f *FakeAwsELB) ModifyLoadBalancerAttributes(*elb.ModifyLoadBalancerAttributesInput) (*elb.ModifyLoadBalancerAttributesOutput, error) {
	return nil, nil
}

func (f *FakeAwsELB) RegisterInstancesWithLoadBalancerRequest(*elb.RegisterInstancesWithLoadBalancerInput) (*request.Request, *elb.RegisterInstancesWithLoadBalancerOutput) {
	return nil, nil
}

func (f *FakeAwsELB) RegisterInstancesWithLoadBalancer(*elb.RegisterInstancesWithLoadBalancerInput) (*elb.RegisterInstancesWithLoadBalancerOutput, error) {
	return nil, nil
}

func (f *FakeAwsELB) RemoveTagsRequest(*elb.RemoveTagsInput) (*request.Request, *elb.RemoveTagsOutput) {
	return nil, nil
}

func (f *FakeAwsELB) RemoveTags(*elb.RemoveTagsInput) (*elb.RemoveTagsOutput, error) {
	return nil, nil
}

func (f *FakeAwsELB) SetLoadBalancerListenerSSLCertificateRequest(*elb.SetLoadBalancerListenerSSLCertificateInput) (*request.Request, *elb.SetLoadBalancerListenerSSLCertificateOutput) {
	return nil, nil
}

func (f *FakeAwsELB) SetLoadBalancerListenerSSLCertificate(*elb.SetLoadBalancerListenerSSLCertificateInput) (*elb.SetLoadBalancerListenerSSLCertificateOutput, error) {
	return nil, nil
}

func (f *FakeAwsELB) SetLoadBalancerPoliciesForBackendServerRequest(*elb.SetLoadBalancerPoliciesForBackendServerInput) (*request.Request, *elb.SetLoadBalancerPoliciesForBackendServerOutput) {
	return nil, nil
}

func (f *FakeAwsELB) SetLoadBalancerPoliciesForBackendServer(*elb.SetLoadBalancerPoliciesForBackendServerInput) (*elb.SetLoadBalancerPoliciesForBackendServerOutput, error) {
	return nil, nil
}

func (f *FakeAwsELB) SetLoadBalancerPoliciesOfListenerRequest(*elb.SetLoadBalancerPoliciesOfListenerInput) (*request.Request, *elb.SetLoadBalancerPoliciesOfListenerOutput) {
	return nil, nil
}

func (f *FakeAwsELB) SetLoadBalancerPoliciesOfListener(*elb.SetLoadBalancerPoliciesOfListenerInput) (*elb.SetLoadBalancerPoliciesOfListenerOutput, error) {
	return nil, nil
}
