//go:generate protoc  --go_out=. --twirp_out=. source_relative:src/protocol/leakybucket.proto
package leakybucket

import (
	"fmt"
	"net/http"
)

type ClientFactory interface {
	NewClients() []LeakyBucket
}

type IPLookupClientFactory struct {
	peerLookup func() []string
	port       uint16
}

func NewClientFactory(lookup func() []string) ClientFactory {
	return IPLookupClientFactory{peerLookup: lookup}
}

func (fac IPLookupClientFactory) NewClients() []LeakyBucket {
	clients := make([]LeakyBucket, 0, 8)
	for _, peer := range fac.peerLookup() {
		httpClient := &http.Client{}
		limiterTwirpClient := NewLeakyBucketJSONClient(
			fmt.Sprintf("%s:%d/twirp", peer, fac.port),
			httpClient,
		)
		clients = append(clients, limiterTwirpClient)
	}
	return clients
}
