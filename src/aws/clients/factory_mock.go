//go:build testing
package clients

type ClientPreBuilds struct {
	EC2Client EC2Client
}

func (fac ClientPreBuilds) EC2() EC2Client {
	return fac.EC2Client
}


