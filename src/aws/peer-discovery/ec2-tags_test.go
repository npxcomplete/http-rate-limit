package peer_discovery

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// mockEC2Client is a hand-written mock implementing EC2Client.
type mockEC2Client struct {
	ctrl     *gomock.Controller
	callFunc func(ctx context.Context, input *ec2.DescribeInstancesInput, fn func(*ec2.DescribeInstancesOutput, bool) bool) error
}

func (m *mockEC2Client) DescribeInstancesPagesWithContext(ctx context.Context, input *ec2.DescribeInstancesInput, fn func(*ec2.DescribeInstancesOutput, bool) bool) error {
	if m.callFunc != nil {
		return m.callFunc(ctx, input, fn)
	}
	return nil
}

func TestInstancesForTagReturnsInstances(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := &mockEC2Client{ctrl: ctrl}
	mock.callFunc = func(ctx context.Context, input *ec2.DescribeInstancesInput, fn func(*ec2.DescribeInstancesOutput, bool) bool) error {
		out := &ec2.DescribeInstancesOutput{Reservations: []*ec2.Reservation{
			{Instances: []*ec2.Instance{{PrivateIpAddress: aws.String("1.1.1.1")}}},
			{Instances: []*ec2.Instance{{PrivateIpAddress: aws.String("2.2.2.2")}}},
		}}
		fn(out, true)
		return nil
	}

	factory := func() (EC2Client, error) { return mock, nil }
	inst, err := InstancesForTag(context.Background(), factory, "role", "web")
	assert.NoError(t, err)
	ips := IPsForInstances(inst)
	assert.ElementsMatch(t, []string{"1.1.1.1", "2.2.2.2"}, ips)
}

func TestInstancesForTagError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := &mockEC2Client{ctrl: ctrl}
	mock.callFunc = func(ctx context.Context, input *ec2.DescribeInstancesInput, fn func(*ec2.DescribeInstancesOutput, bool) bool) error {
		return assert.AnError
	}

	factory := func() (EC2Client, error) { return mock, nil }
	_, err := InstancesForTag(context.Background(), factory, "role", "web")
	assert.Error(t, err)
}
