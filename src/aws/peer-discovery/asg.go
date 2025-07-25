package peer_discovery

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// metadataAPI describes the subset of the EC2 metadata client used by
// DiscoverPeerIPs. Defined as an interface for easy mocking in tests.
type metadataAPI interface {
	GetMetadata(path string) (string, error)
}

// autoscalingAPI describes the calls to the AutoScaling service used by
// DiscoverPeerIPs.
type autoscalingAPI interface {
	DescribeAutoScalingInstances(*autoscaling.DescribeAutoScalingInstancesInput) (*autoscaling.DescribeAutoScalingInstancesOutput, error)
	DescribeAutoScalingGroups(*autoscaling.DescribeAutoScalingGroupsInput) (*autoscaling.DescribeAutoScalingGroupsOutput, error)
}

// ec2API describes the EC2 calls used by DiscoverPeerIPs.
type ec2API interface {
	DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
}

// ClientFactory provides AWS service clients required for peer discovery.
type ClientFactory interface {
	Metadata() metadataAPI
	AutoScaling() autoscalingAPI
	EC2() ec2API
}

// DefaultFactory creates real AWS service clients using the provided session.
type DefaultFactory struct{ Sess *session.Session }

// Metadata returns an EC2 metadata client.
func (f DefaultFactory) Metadata() metadataAPI { return ec2metadata.New(f.Sess) }

// AutoScaling returns an AutoScaling client.
func (f DefaultFactory) AutoScaling() autoscalingAPI { return autoscaling.New(f.Sess) }

// EC2 returns an EC2 client.
func (f DefaultFactory) EC2() ec2API { return ec2.New(f.Sess) }

// DiscoverPeerIPs returns the private IP addresses of all hosts
// that are part of the same autoscaling group as the instance
// associated with the provided session.
//
// Example IAM policy required for this call:
//
// ```json
//
//	{
//	  "Version": "2012-10-17",
//	  "Statement": [{
//	    "Effect": "Allow",
//	    "Action": [
//	      "autoscaling:DescribeAutoScalingInstances",
//	      "autoscaling:DescribeAutoScalingGroups",
//	      "ec2:DescribeInstances"
//	    ],
//	    "Resource": "*"
//	  }]
//	}
//
// ```
// DiscoverPeerIPs returns the private IP addresses of all hosts that are part
// of the same autoscaling group as the instance returned by the metadata
// service provided by the factory.
func DiscoverPeerIPs(factory ClientFactory) ([]string, error) {
	meta := factory.Metadata()
	instanceID, err := meta.GetMetadata("instance-id")
	if err != nil {
		return nil, err
	}

	asgSvc := factory.AutoScaling()
	descInstOut, err := asgSvc.DescribeAutoScalingInstances(&autoscaling.DescribeAutoScalingInstancesInput{
		InstanceIds: []*string{aws.String(instanceID)},
	})
	if err != nil {
		return nil, err
	}
	if len(descInstOut.AutoScalingInstances) == 0 {
		return nil, fmt.Errorf("instance %s not part of an autoscaling group", instanceID)
	}

	asgName := aws.StringValue(descInstOut.AutoScalingInstances[0].AutoScalingGroupName)

	descASGOut, err := asgSvc.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String(asgName)},
	})
	if err != nil {
		return nil, err
	}
	if len(descASGOut.AutoScalingGroups) == 0 {
		return nil, fmt.Errorf("autoscaling group %s not found", asgName)
	}

	instanceIDs := make([]*string, 0, len(descASGOut.AutoScalingGroups[0].Instances))
	for _, inst := range descASGOut.AutoScalingGroups[0].Instances {
		instanceIDs = append(instanceIDs, inst.InstanceId)
	}

	ec2Svc := factory.EC2()
	descInst, err := ec2Svc.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: instanceIDs,
	})
	if err != nil {
		return nil, err
	}

	var ips []string
	for _, res := range descInst.Reservations {
		for _, inst := range res.Instances {
			if inst.PrivateIpAddress != nil {
				ips = append(ips, aws.StringValue(inst.PrivateIpAddress))
			}
		}
	}
	return ips, nil
}
