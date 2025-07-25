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

	ips, err := IPsForLocalASG(context.Background(), factory)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"1.1.1.1", "2.2.2.2"}, ips)
}

func TestInstancesForLocalASGError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	md := &clients.MockMetadataClient{Ctrl: ctrl}
	md.GetMetadataWithContextFunc = func(ctx context.Context, path string) (string, error) {
		return "i-self", nil
	}

	asg := &clients.MockAutoScalingClient{Ctrl: ctrl}
	asg.DescribeAutoScalingInstancesWithContextFunc = func(ctx context.Context, input *autoscaling.DescribeAutoScalingInstancesInput, opts ...request.Option) (*autoscaling.DescribeAutoScalingInstancesOutput, error) {
		return nil, assert.AnError
	}

	factory := clients.ClientPreBuilds{AutoScalingClient: asg, MetadataClient: md}
	_, err := InstancesForLocalASG(context.Background(), factory)
	assert.Error(t, err)
}
