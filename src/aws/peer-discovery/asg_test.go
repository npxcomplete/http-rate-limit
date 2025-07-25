package peer_discovery

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDiscoverPeerIPs_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	factory := NewMockClientFactory(ctrl)
	meta := NewMockMetadataAPI(ctrl)
	asg := NewMockAutoScalingAPI(ctrl)
	ec2c := NewMockEC2API(ctrl)

	factory.EXPECT().Metadata().Return(meta)
	factory.EXPECT().AutoScaling().Return(asg)
	factory.EXPECT().EC2().Return(ec2c)

	meta.EXPECT().GetMetadata("instance-id").Return("i-123", nil)

	asg.EXPECT().DescribeAutoScalingInstances(&autoscaling.DescribeAutoScalingInstancesInput{
		InstanceIds: []*string{aws.String("i-123")},
	}).Return(&autoscaling.DescribeAutoScalingInstancesOutput{
		AutoScalingInstances: []*autoscaling.Instance{
			{InstanceId: aws.String("i-123"), AutoScalingGroupName: aws.String("asg")},
		},
	}, nil)

	asg.EXPECT().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String("asg")},
	}).Return(&autoscaling.DescribeAutoScalingGroupsOutput{
		AutoScalingGroups: []*autoscaling.Group{
			{Instances: []*autoscaling.Instance{
				{InstanceId: aws.String("i-123")},
				{InstanceId: aws.String("i-456")},
			}},
		},
	}, nil)

	ec2c.EXPECT().DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String("i-123"), aws.String("i-456")},
	}).Return(&ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{{
			Instances: []*ec2.Instance{{PrivateIpAddress: aws.String("10.0.0.1")}, {PrivateIpAddress: aws.String("10.0.0.2")}},
		}},
	}, nil)

	ips, err := DiscoverPeerIPs(factory)
	assert.NoError(t, err)
	assert.Equal(t, []string{"10.0.0.1", "10.0.0.2"}, ips)
}

func TestDiscoverPeerIPs_MetadataError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	factory := NewMockClientFactory(ctrl)
	meta := NewMockMetadataAPI(ctrl)

	factory.EXPECT().Metadata().Return(meta)
	meta.EXPECT().GetMetadata("instance-id").Return("", assert.AnError)

	ips, err := DiscoverPeerIPs(factory)
	assert.Error(t, err)
	assert.Nil(t, ips)
}
