package peer_discovery

import (
	"context"
	"github.com/npxcomplete/http-rate-limit/src/aws/clients"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// InstancesForTag returns all EC2 instances from the given client that
// have the provided tag key and value.
func InstancesForTag(ctx context.Context, factory clients.AWSClientFactory, key, value string) ([]*ec2.Instance, error) {
	ec2Client := factory.EC2()
	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{{
			Name:   aws.String("tag:" + key),
			Values: []*string{aws.String(value)},
		}},
	}

	var instances []*ec2.Instance
	err := ec2Client.DescribeInstancesPagesWithContext(ctx, input, func(output *ec2.DescribeInstancesOutput, lastPage bool) bool {
		for _, r := range output.Reservations {
			for _, inst := range r.Instances {
				instances = append(instances, inst)
			}
		}
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return instances, nil
}

// IPsForInstances returns the private IP addresses of the provided EC2 instances.
func IPsForInstances(instances []*ec2.Instance) []string {
	var ips []string
	for _, inst := range instances {
		if inst.PrivateIpAddress != nil {
			ips = append(ips, aws.StringValue(inst.PrivateIpAddress))
		}
	}
	return ips
}

// IPsForTag is a convenience wrapper that returns the private IP addresses of
// all instances matching the given tag.
func IPsForTag(ctx context.Context, factory clients.AWSClientFactory, key, value string) ([]string, error) {
	instances, err := InstancesForTag(ctx, factory, key, value)
	if err != nil {
		return nil, err
	}
	return IPsForInstances(instances), nil
}
