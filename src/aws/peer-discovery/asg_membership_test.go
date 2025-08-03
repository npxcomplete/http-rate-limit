package peer_discovery

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/mock/gomock"
	"github.com/npxcomplete/http-rate-limit/src/aws/clients"
	"github.com/stretchr/testify/assert"
)

func TestIPsForLocalASG(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// GIVEN metadata reveals the instance ID and peers exist in the same Auto Scaling group
	md := &clients.MockMetadataClient{Ctrl: ctrl}
	md.GetMetadataWithContextFunc = func(ctx context.Context, path string) (string, error) {
		return "i-self", nil
	}

	asg := &clients.MockAutoScalingClient{Ctrl: ctrl}
	asg.DescribeAutoScalingInstancesWithContextFunc = func(ctx context.Context, input *autoscaling.DescribeAutoScalingInstancesInput, opts ...request.Option) (*autoscaling.DescribeAutoScalingInstancesOutput, error) {
		return &autoscaling.DescribeAutoScalingInstancesOutput{AutoScalingInstances: []*autoscaling.InstanceDetails{{AutoScalingGroupName: aws.String("asg")}}}, nil
	}
	asg.DescribeAutoScalingGroupsWithContextFunc = func(ctx context.Context, input *autoscaling.DescribeAutoScalingGroupsInput, opts ...request.Option) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
		return &autoscaling.DescribeAutoScalingGroupsOutput{AutoScalingGroups: []*autoscaling.Group{{Instances: []*autoscaling.Instance{{InstanceId: aws.String("i-self")}, {InstanceId: aws.String("i-peer")}}}}}, nil
	}

	ec2Mock := &clients.MockEC2Client{Ctrl: ctrl}
	ec2Mock.DescribeInstancesPagesWithContextFunc = func(ctx context.Context, input *ec2.DescribeInstancesInput, fn func(*ec2.DescribeInstancesOutput, bool) bool) error {
		out := &ec2.DescribeInstancesOutput{Reservations: []*ec2.Reservation{
			{Instances: []*ec2.Instance{{PrivateIpAddress: aws.String("1.1.1.1")}}},
			{Instances: []*ec2.Instance{{PrivateIpAddress: aws.String("2.2.2.2")}}},
		}}
		fn(out, true)
		return nil
	}

	factory := clients.ClientPreBuilds{EC2Client: ec2Mock, AutoScalingClient: asg, MetadataClient: md}

	// WHEN retrieving IPs for peers in the local Auto Scaling group
	ips, err := IPsForLocalASG(context.Background(), factory)

	// THEN all peer IPs are returned
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"1.1.1.1", "2.2.2.2"}, ips)
}

func TestInstancesForLocalASGError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// GIVEN Auto Scaling instance lookup fails
	md := &clients.MockMetadataClient{Ctrl: ctrl}
	md.GetMetadataWithContextFunc = func(ctx context.Context, path string) (string, error) {
		return "i-self", nil
	}

	asg := &clients.MockAutoScalingClient{Ctrl: ctrl}
	asg.DescribeAutoScalingInstancesWithContextFunc = func(ctx context.Context, input *autoscaling.DescribeAutoScalingInstancesInput, opts ...request.Option) (*autoscaling.DescribeAutoScalingInstancesOutput, error) {
		return nil, assert.AnError
	}

	factory := clients.ClientPreBuilds{AutoScalingClient: asg, MetadataClient: md}

	// WHEN retrieving instances for the local Auto Scaling group
	_, err := InstancesForLocalASG(context.Background(), factory)

	// THEN an error is returned
	assert.Error(t, err)
}

func TestInstancesForLocalASGMetadataError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// GIVEN the instance metadata service fails
	md := &clients.MockMetadataClient{Ctrl: ctrl}
	md.GetMetadataWithContextFunc = func(ctx context.Context, path string) (string, error) {
		return "", assert.AnError
	}

	factory := clients.ClientPreBuilds{MetadataClient: md}

	// WHEN retrieving instances for the local Auto Scaling group
	_, err := InstancesForLocalASG(context.Background(), factory)

	// THEN an error is returned
	assert.Error(t, err)
}

func TestInstancesForLocalASGNoInstances(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// GIVEN the Auto Scaling group has no instances
	md := &clients.MockMetadataClient{Ctrl: ctrl}
	md.GetMetadataWithContextFunc = func(ctx context.Context, path string) (string, error) {
		return "i-self", nil
	}

	asg := &clients.MockAutoScalingClient{Ctrl: ctrl}
	asg.DescribeAutoScalingInstancesWithContextFunc = func(ctx context.Context, input *autoscaling.DescribeAutoScalingInstancesInput, opts ...request.Option) (*autoscaling.DescribeAutoScalingInstancesOutput, error) {
		return &autoscaling.DescribeAutoScalingInstancesOutput{}, nil
	}

	factory := clients.ClientPreBuilds{AutoScalingClient: asg, MetadataClient: md}

	// WHEN retrieving instances for the local Auto Scaling group
	inst, err := InstancesForLocalASG(context.Background(), factory)

	// THEN an empty slice is returned without error
	assert.NoError(t, err)
	assert.Empty(t, inst)
}

func TestInstancesForLocalASGDescribeGroupsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// GIVEN retrieving Auto Scaling groups fails
	md := &clients.MockMetadataClient{Ctrl: ctrl}
	md.GetMetadataWithContextFunc = func(ctx context.Context, path string) (string, error) {
		return "i-self", nil
	}

	asg := &clients.MockAutoScalingClient{Ctrl: ctrl}
	asg.DescribeAutoScalingInstancesWithContextFunc = func(ctx context.Context, input *autoscaling.DescribeAutoScalingInstancesInput, opts ...request.Option) (*autoscaling.DescribeAutoScalingInstancesOutput, error) {
		return &autoscaling.DescribeAutoScalingInstancesOutput{AutoScalingInstances: []*autoscaling.InstanceDetails{{AutoScalingGroupName: aws.String("asg")}}}, nil
	}
	asg.DescribeAutoScalingGroupsWithContextFunc = func(ctx context.Context, input *autoscaling.DescribeAutoScalingGroupsInput, opts ...request.Option) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
		return nil, assert.AnError
	}

	factory := clients.ClientPreBuilds{AutoScalingClient: asg, MetadataClient: md}

	// WHEN retrieving instances for the local Auto Scaling group
	_, err := InstancesForLocalASG(context.Background(), factory)

	// THEN an error is returned
	assert.Error(t, err)
}

func TestInstancesForLocalASGNoGroups(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// GIVEN no Auto Scaling groups are found
	md := &clients.MockMetadataClient{Ctrl: ctrl}
	md.GetMetadataWithContextFunc = func(ctx context.Context, path string) (string, error) {
		return "i-self", nil
	}

	asg := &clients.MockAutoScalingClient{Ctrl: ctrl}
	asg.DescribeAutoScalingInstancesWithContextFunc = func(ctx context.Context, input *autoscaling.DescribeAutoScalingInstancesInput, opts ...request.Option) (*autoscaling.DescribeAutoScalingInstancesOutput, error) {
		return &autoscaling.DescribeAutoScalingInstancesOutput{AutoScalingInstances: []*autoscaling.InstanceDetails{{AutoScalingGroupName: aws.String("asg")}}}, nil
	}
	asg.DescribeAutoScalingGroupsWithContextFunc = func(ctx context.Context, input *autoscaling.DescribeAutoScalingGroupsInput, opts ...request.Option) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
		return &autoscaling.DescribeAutoScalingGroupsOutput{}, nil
	}

	factory := clients.ClientPreBuilds{AutoScalingClient: asg, MetadataClient: md}

	// WHEN retrieving instances for the local Auto Scaling group
	inst, err := InstancesForLocalASG(context.Background(), factory)

	// THEN an empty slice is returned without error
	assert.NoError(t, err)
	assert.Empty(t, inst)
}

func TestInstancesForLocalASGDescribeInstancesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// GIVEN EC2 instance description fails
	md := &clients.MockMetadataClient{Ctrl: ctrl}
	md.GetMetadataWithContextFunc = func(ctx context.Context, path string) (string, error) {
		return "i-self", nil
	}

	asg := &clients.MockAutoScalingClient{Ctrl: ctrl}
	asg.DescribeAutoScalingInstancesWithContextFunc = func(ctx context.Context, input *autoscaling.DescribeAutoScalingInstancesInput, opts ...request.Option) (*autoscaling.DescribeAutoScalingInstancesOutput, error) {
		return &autoscaling.DescribeAutoScalingInstancesOutput{AutoScalingInstances: []*autoscaling.InstanceDetails{{AutoScalingGroupName: aws.String("asg")}}}, nil
	}
	asg.DescribeAutoScalingGroupsWithContextFunc = func(ctx context.Context, input *autoscaling.DescribeAutoScalingGroupsInput, opts ...request.Option) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
		return &autoscaling.DescribeAutoScalingGroupsOutput{AutoScalingGroups: []*autoscaling.Group{{Instances: []*autoscaling.Instance{{InstanceId: aws.String("i-self")}}}}}, nil
	}

	ec2Mock := &clients.MockEC2Client{Ctrl: ctrl}
	ec2Mock.DescribeInstancesPagesWithContextFunc = func(ctx context.Context, input *ec2.DescribeInstancesInput, fn func(*ec2.DescribeInstancesOutput, bool) bool) error {
		return assert.AnError
	}

	factory := clients.ClientPreBuilds{EC2Client: ec2Mock, AutoScalingClient: asg, MetadataClient: md}

	// WHEN retrieving instances for the local Auto Scaling group
	_, err := InstancesForLocalASG(context.Background(), factory)

	// THEN an error is returned
	assert.Error(t, err)
}

func TestIPsForLocalASGError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// GIVEN Auto Scaling instance lookup fails
	md := &clients.MockMetadataClient{Ctrl: ctrl}
	md.GetMetadataWithContextFunc = func(ctx context.Context, path string) (string, error) {
		return "i-self", nil
	}

	asg := &clients.MockAutoScalingClient{Ctrl: ctrl}
	asg.DescribeAutoScalingInstancesWithContextFunc = func(ctx context.Context, input *autoscaling.DescribeAutoScalingInstancesInput, opts ...request.Option) (*autoscaling.DescribeAutoScalingInstancesOutput, error) {
		return nil, assert.AnError
	}

	factory := clients.ClientPreBuilds{AutoScalingClient: asg, MetadataClient: md}

	// WHEN retrieving peer IPs for the local Auto Scaling group
	_, err := IPsForLocalASG(context.Background(), factory)

	// THEN an error is returned
	assert.Error(t, err)
}
