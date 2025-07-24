package peer_discovery

import (
	"context"
	"github.com/npxcomplete/http-rate-limit/src/aws/clients"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)


func TestInstancesForTagReturnsInstances(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := &clients.MockEC2Client{Ctrl: ctrl}
	mock.DescribeInstancesPagesWithContextFunc = func(ctx context.Context, input *ec2.DescribeInstancesInput, fn func(*ec2.DescribeInstancesOutput, bool) bool) error {
		out := &ec2.DescribeInstancesOutput{Reservations: []*ec2.Reservation{
			{Instances: []*ec2.Instance{{PrivateIpAddress: aws.String("1.1.1.1")}}},
			{Instances: []*ec2.Instance{{PrivateIpAddress: aws.String("2.2.2.2")}}},
		}}
		fn(out, true)
		return nil
	}

	factory := clients.ClientPreBuilds{EC2Client: mock}
	inst, err := InstancesForTag(context.Background(), factory, "role", "web")
	assert.NoError(t, err)
	ips := IPsForInstances(inst)
	assert.ElementsMatch(t, []string{"1.1.1.1", "2.2.2.2"}, ips)
}

func TestInstancesForTagError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := &clients.MockEC2Client{Ctrl: ctrl}
	mock.DescribeInstancesPagesWithContextFunc = func(ctx context.Context, input *ec2.DescribeInstancesInput, fn func(*ec2.DescribeInstancesOutput, bool) bool) error {
		return assert.AnError
	}

	factory := clients.ClientPreBuilds{EC2Client: mock}
	_, err := InstancesForTag(context.Background(), factory, "role", "web")
	assert.Error(t, err)
}
