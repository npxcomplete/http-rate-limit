//go:build testing
package clients

import (
	"context"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/mock/gomock"
)

type MockEC2Client struct {
	Ctrl     *gomock.Controller
	DescribeInstancesPagesWithContextFunc func(ctx context.Context, input *ec2.DescribeInstancesInput, fn func(*ec2.DescribeInstancesOutput, bool) bool) error
}


func (m *MockEC2Client) DescribeInstancesPagesWithContext(ctx context.Context, input *ec2.DescribeInstancesInput, fn func(*ec2.DescribeInstancesOutput, bool) bool) error {
	if m.DescribeInstancesPagesWithContextFunc != nil {
		return m.DescribeInstancesPagesWithContextFunc(ctx, input, fn)
	}
	return nil
}
