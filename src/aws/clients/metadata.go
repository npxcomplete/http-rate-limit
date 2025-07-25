package clients

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
)

type MetadataClient interface {
	GetMetadataWithContext(ctx context.Context, path string) (string, error)
}

type metadataAPI struct{ svc *ec2metadata.EC2Metadata }

func (m metadataAPI) GetMetadataWithContext(ctx context.Context, path string) (string, error) {
	return m.svc.GetMetadataWithContext(ctx, path)
}

func NewMetadataClient(region string) (MetadataClient, error) {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		return nil, err
	}
	return metadataAPI{svc: ec2metadata.New(sess)}, nil
}
