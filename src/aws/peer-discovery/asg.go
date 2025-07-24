package peer_discovery

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
)

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
func DiscoverPeerIPs(sess *session.Session) ([]string, error) {
	meta := ec2metadata.New(sess)
	instanceID, err := meta.GetMetadata("instance-id")
	if err != nil {
		return nil, err
	}

	asgSvc := autoscaling.New(sess)
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

	ec2Svc := ec2.New(sess)
	descInst, err := ec2Svc.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: instanceIDs,
	})
	if err != nil {
		return nil, err
	}

	ips := []string{}
	for _, res := range descInst.Reservations {
		for _, inst := range res.Instances {
			if inst.PrivateIpAddress != nil {
				ips = append(ips, aws.StringValue(inst.PrivateIpAddress))
			}
		}
	}
	return ips, nil
}
