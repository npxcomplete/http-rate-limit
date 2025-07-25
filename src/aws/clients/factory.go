package clients

import (
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// ClientFactory provides AWS service clients required for peer discovery.
type AWSClientFactory interface {
	EC2() EC2Client
	AutoScaling() AutoScalingClient
	Metadata() MetadataClient
}

// DefaultFactory creates real AWS service clients using the provided session.
type DefaultAWSFactory struct{ Sess *session.Session }

func (f DefaultAWSFactory) EC2() EC2Client { return ec2API{ec2.New(f.Sess)} }
func (f DefaultAWSFactory) AutoScaling() AutoScalingClient {
	return autoScalingAPI{autoscaling.New(f.Sess)}
}
func (f DefaultAWSFactory) Metadata() MetadataClient { return metadataAPI{ec2metadata.New(f.Sess)} }
