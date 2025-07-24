package clients

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// ClientFactory provides AWS service clients required for peer discovery.
type AWSClientFactory interface {
	EC2() EC2Client
}

// DefaultFactory creates real AWS service clients using the provided session.
type DefaultAWSFactory struct{ Sess *session.Session }

func (f DefaultAWSFactory) EC2() EC2Client { return ec2API { ec2.New(f.Sess) } }