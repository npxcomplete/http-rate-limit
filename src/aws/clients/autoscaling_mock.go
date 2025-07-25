//go:build testing

package clients

import (
	"context"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/golang/mock/gomock"
)

type MockAutoScalingClient struct {
	Ctrl                                        *gomock.Controller
	DescribeAutoScalingInstancesWithContextFunc func(ctx context.Context, input *autoscaling.DescribeAutoScalingInstancesInput, opts ...request.Option) (*autoscaling.DescribeAutoScalingInstancesOutput, error)
	DescribeAutoScalingGroupsWithContextFunc    func(ctx context.Context, input *autoscaling.DescribeAutoScalingGroupsInput, opts ...request.Option) (*autoscaling.DescribeAutoScalingGroupsOutput, error)
}

func (m *MockAutoScalingClient) DescribeAutoScalingInstancesWithContext(ctx context.Context, input *autoscaling.DescribeAutoScalingInstancesInput, opts ...request.Option) (*autoscaling.DescribeAutoScalingInstancesOutput, error) {
	if m.DescribeAutoScalingInstancesWithContextFunc != nil {
		return m.DescribeAutoScalingInstancesWithContextFunc(ctx, input, opts...)
	}
	return nil, nil
}

func (m *MockAutoScalingClient) DescribeAutoScalingGroupsWithContext(ctx context.Context, input *autoscaling.DescribeAutoScalingGroupsInput, opts ...request.Option) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	if m.DescribeAutoScalingGroupsWithContextFunc != nil {
		return m.DescribeAutoScalingGroupsWithContextFunc(ctx, input, opts...)
	}
	return nil, nil
}
