package peer_discovery

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// EC2Client abstracts the EC2 API so that it can be mocked in tests.
type EC2Client interface {
	DescribeInstancesPagesWithContext(ctx context.Context, input *ec2.DescribeInstancesInput, fn func(*ec2.DescribeInstancesOutput, bool) bool) error
}

// EC2ClientFactory creates a new EC2 client instance. It allows the caller to
// inject different implementations (for example, mocks) when testing.
type EC2ClientFactory func() (EC2Client, error)

// realEC2Client wraps the AWS SDK EC2 service client to satisfy EC2Client.
type realEC2Client struct{ svc *ec2.EC2 }

func (c realEC2Client) DescribeInstancesPagesWithContext(ctx context.Context, input *ec2.DescribeInstancesInput, fn func(*ec2.DescribeInstancesOutput, bool) bool) error {
	return c.svc.DescribeInstancesPagesWithContext(ctx, input, fn)
}

// NewEC2Client returns a client for the given region using the AWS SDK.
func NewEC2Client(region string) (EC2Client, error) {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		return nil, err
	}
	return realEC2Client{svc: ec2.New(sess)}, nil
}

// InstancesForTag returns all EC2 instances from the given client that
// have the provided tag key and value.
func InstancesForTag(ctx context.Context, factory EC2ClientFactory, key, value string) ([]*ec2.Instance, error) {
	client, err := factory()
	if err != nil {
		return nil, err
	}
	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{{
			Name:   aws.String("tag:" + key),
			Values: []*string{aws.String(value)},
		}},
	}

	var instances []*ec2.Instance
	err = client.DescribeInstancesPagesWithContext(ctx, input, func(output *ec2.DescribeInstancesOutput, lastPage bool) bool {
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
func IPsForTag(ctx context.Context, factory EC2ClientFactory, key, value string) ([]string, error) {
	instances, err := InstancesForTag(ctx, factory, key, value)
	if err != nil {
		return nil, err
	}
	return IPsForInstances(instances), nil
}
