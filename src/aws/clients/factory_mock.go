//go:build testing

package clients

type ClientPreBuilds struct {
	EC2Client         EC2Client
	AutoScalingClient AutoScalingClient
	MetadataClient    MetadataClient
}

func (fac ClientPreBuilds) EC2() EC2Client {
	return fac.EC2Client
}

func (fac ClientPreBuilds) AutoScaling() AutoScalingClient {
	return fac.AutoScalingClient
}

func (fac ClientPreBuilds) Metadata() MetadataClient {
	return fac.MetadataClient
}
