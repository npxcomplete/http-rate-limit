package clients

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

// realEC2Client wraps the AWS SDK EC2 service client to satisfy EC2Client.
type ec2API struct{ svc *ec2.EC2 }

func (c ec2API) DescribeInstancesPagesWithContext(ctx context.Context, input *ec2.DescribeInstancesInput, fn func(*ec2.DescribeInstancesOutput, bool) bool) error {
	return c.svc.DescribeInstancesPagesWithContext(ctx, input, fn)
}

// NewEC2Client returns a client for the given region using the AWS SDK.
func NewEC2Client(region string) (EC2Client, error) {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		return nil, err
	}
	return ec2API{svc: ec2.New(sess)}, nil
}