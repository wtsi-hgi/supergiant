package core

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func autoscalingGroups(c *Core) (groups []*autoscaling.Group, err error) {
	params := &autoscaling.DescribeAutoScalingGroupsInput{
		// VPCZoneIdentifier: aws.String(AwsSubnetID),
		// NOTE I think we have to just filter this client-side? Seems weird
		MaxRecords: aws.Int64(100),
	}
	resp, err := c.autoscaling.DescribeAutoScalingGroups(params)
	if err != nil {
		return nil, err
	}

	for _, group := range resp.AutoScalingGroups {
		if *group.VPCZoneIdentifier == c.AwsSubnetID {
			groups = append(groups, group)
		}
	}
	return groups, nil
}

func autoscalingGroup(c *Core, id *string) (*autoscaling.Group, error) {
	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{id},
	}
	resp, err := c.autoscaling.DescribeAutoScalingGroups(input)
	if err != nil {
		return nil, err
	}
	return resp.AutoScalingGroups[0], nil
}

func autoscalingGroupInstanceType(c *Core, group *autoscaling.Group) string {
	input := &autoscaling.DescribeLaunchConfigurationsInput{
		LaunchConfigurationNames: []*string{group.LaunchConfigurationName},
	}
	resp, err := c.autoscaling.DescribeLaunchConfigurations(input)
	if err != nil {
		panic(err)
	}
	return *resp.LaunchConfigurations[0].InstanceType
}

func autoscalingGroupByInstanceType(c *Core, instanceType string) (*autoscaling.Group, error) {
	groups, err := autoscalingGroups(c)
	if err != nil {
		return nil, err
	}
	for _, group := range groups {
		if autoscalingGroupInstanceType(c, group) == instanceType {
			return group, nil
		}
	}
	return nil, fmt.Errorf("Could not find AWS AutoScaling Group with instance type %s", instanceType)
}

func ec2InstancesFromAutoscalingGroups(c *Core) (instances []*ec2.Instance, err error) {
	var groupNames []*string
	groups, err := autoscalingGroups(c)
	if err != nil {
		return nil, err
	}
	for _, group := range groups {
		groupNames = append(groupNames, group.AutoScalingGroupName)
	}

	return findEc2Instances(c, []*ec2.Filter{
		&ec2.Filter{
			Name:   aws.String("tag:aws:autoscaling:groupName"),
			Values: groupNames,
		},
		&ec2.Filter{
			Name: aws.String("instance-state-name"),
			Values: []*string{
				aws.String("pending"),
				aws.String("running"),
			},
		},
	})
}

func ec2Instance(c *Core, id *string) (*ec2.Instance, error) {

	Log.Debugf("Trying to find EC2 instance by ID %s (using Filter)", *id)

	instances, err := findEc2Instances(c, []*ec2.Filter{
		&ec2.Filter{
			Name:   aws.String("instance-id"),
			Values: []*string{id},
		},
	})
	if err != nil {
		return nil, err
	}
	return instances[0], nil
}

func findEc2Instances(c *Core, filters []*ec2.Filter) (instances []*ec2.Instance, err error) {
	input := &ec2.DescribeInstancesInput{
		Filters: filters,
	}
	resp, err := c.ec2.DescribeInstances(input)
	if err != nil {
		return nil, err
	}
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			instances = append(instances, instance)
		}
	}
	return instances, nil
}
