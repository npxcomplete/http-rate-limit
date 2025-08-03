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

	// GIVEN EC2 returns instances for the provided tag filter
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

	// WHEN retrieving instances for the role=web tag
	inst, err := InstancesForTag(context.Background(), factory, "role", "web")
	ips := IPsForInstances(inst)

	// THEN both instance IPs are returned
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"1.1.1.1", "2.2.2.2"}, ips)
}

func TestInstancesForTagError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// GIVEN describing instances fails
	mock := &clients.MockEC2Client{Ctrl: ctrl}
	mock.DescribeInstancesPagesWithContextFunc = func(ctx context.Context, input *ec2.DescribeInstancesInput, fn func(*ec2.DescribeInstancesOutput, bool) bool) error {
		return assert.AnError
	}

	factory := clients.ClientPreBuilds{EC2Client: mock}

	// WHEN retrieving instances for the tag
	_, err := InstancesForTag(context.Background(), factory, "role", "web")

	// THEN an error is returned
	assert.Error(t, err)
}

func TestIPsForTag(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// GIVEN EC2 returns instances for the provided tag filter
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

	// WHEN retrieving IPs for the role=web tag
	ips, err := IPsForTag(context.Background(), factory, "role", "web")

	// THEN both instance IPs are returned
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"1.1.1.1", "2.2.2.2"}, ips)
}

func TestIPsForTagError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// GIVEN describing instances fails
	mock := &clients.MockEC2Client{Ctrl: ctrl}
	mock.DescribeInstancesPagesWithContextFunc = func(ctx context.Context, input *ec2.DescribeInstancesInput, fn func(*ec2.DescribeInstancesOutput, bool) bool) error {
		return assert.AnError
	}

	factory := clients.ClientPreBuilds{EC2Client: mock}

	// WHEN retrieving IPs for the tag
	_, err := IPsForTag(context.Background(), factory, "role", "web")

	// THEN an error is returned
	assert.Error(t, err)
}
