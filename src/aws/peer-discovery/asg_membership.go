package peer_discovery

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/npxcomplete/http-rate-limit/src/aws/clients"
)

// InstancesForLocalASG returns all EC2 instances that are in the same
// Auto Scaling Group as the running instance.
func InstancesForLocalASG(ctx context.Context, factory clients.AWSClientFactory) ([]*ec2.Instance, error) {
	md := factory.Metadata()
	instanceID, err := md.GetMetadataWithContext(ctx, "instance-id")
	if err != nil {
		return nil, err
	}

	asg := factory.AutoScaling()
	instOut, err := asg.DescribeAutoScalingInstancesWithContext(ctx, &autoscaling.DescribeAutoScalingInstancesInput{
		InstanceIds: []*string{aws.String(instanceID)},
	})
	if err != nil {
		return nil, err
	}
	if len(instOut.AutoScalingInstances) == 0 {
		return []*ec2.Instance{}, nil
	}
	asgName := aws.StringValue(instOut.AutoScalingInstances[0].AutoScalingGroupName)

	groupOut, err := asg.DescribeAutoScalingGroupsWithContext(ctx, &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String(asgName)},
	})
	if err != nil {
		return nil, err
	}
	if len(groupOut.AutoScalingGroups) == 0 {
		return []*ec2.Instance{}, nil
	}

	var ids []*string
	for _, inst := range groupOut.AutoScalingGroups[0].Instances {
		if inst.InstanceId != nil {
			ids = append(ids, inst.InstanceId)
		}
	}

	ec2Client := factory.EC2()
	input := &ec2.DescribeInstancesInput{InstanceIds: ids}

	var instances []*ec2.Instance
	err = ec2Client.DescribeInstancesPagesWithContext(ctx, input, func(out *ec2.DescribeInstancesOutput, last bool) bool {
		for _, r := range out.Reservations {
			for _, inst := range r.Instances {
				instances = append(instances, inst)
			}
		}
		return !last
	})
	if err != nil {
		return nil, err
	}
	return instances, nil
}

// IPsForLocalASG returns the private IP addresses of all instances in the same
// Auto Scaling Group as the running instance.
func IPsForLocalASG(ctx context.Context, factory clients.AWSClientFactory) ([]string, error) {
	instances, err := InstancesForLocalASG(ctx, factory)
	if err != nil {
		return nil, err
	}
	return IPsForInstances(instances), nil
}
