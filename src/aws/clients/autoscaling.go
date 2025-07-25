package clients

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
)

type AutoScalingClient interface {
	DescribeAutoScalingInstancesWithContext(ctx context.Context, input *autoscaling.DescribeAutoScalingInstancesInput, opts ...request.Option) (*autoscaling.DescribeAutoScalingInstancesOutput, error)
	DescribeAutoScalingGroupsWithContext(ctx context.Context, input *autoscaling.DescribeAutoScalingGroupsInput, opts ...request.Option) (*autoscaling.DescribeAutoScalingGroupsOutput, error)
}

type autoScalingAPI struct{ svc *autoscaling.AutoScaling }

func (a autoScalingAPI) DescribeAutoScalingInstancesWithContext(ctx context.Context, input *autoscaling.DescribeAutoScalingInstancesInput, opts ...request.Option) (*autoscaling.DescribeAutoScalingInstancesOutput, error) {
	return a.svc.DescribeAutoScalingInstancesWithContext(ctx, input, opts...)
}

func (a autoScalingAPI) DescribeAutoScalingGroupsWithContext(ctx context.Context, input *autoscaling.DescribeAutoScalingGroupsInput, opts ...request.Option) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	return a.svc.DescribeAutoScalingGroupsWithContext(ctx, input, opts...)
}

func NewAutoScalingClient(region string) (AutoScalingClient, error) {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		return nil, err
	}
	return autoScalingAPI{svc: autoscaling.New(sess)}, nil
}
