package leakybucket

import (
	ratelimit "github.com/npxcomplete/http-rate-limit/src"
	"io"
	"net/http"
)

func StdMiddleware(
	limiter RateLimiter,
) func(servlet http.Handler) http.Handler {
	ipAddr := func(req *http.Request) []string { return []string{ratelimit.IPAddressOf(req)} }
	return Middleware(limiter, ipAddr, ratelimit.FixedRequestCost)
}

func Middleware(
	limiter RateLimiter,
// For a given request return any identifiers
//   that may be linked to a throttle rule.
//
// These are typically IPAddresses, API AccessKeys, or
//   Cloud Identities such as an an AWS account number.
	tenantIdentifiers func(r *http.Request) []string,
// For many APIs where different types of requests
//   have disproportionate compute costs.
	costOfRequest func(req *http.Request) uint32,
) func(servlet http.Handler) http.Handler {
	return func(servlet http.Handler) http.Handler {
		twirpHandler := NewLeakyBucketServer(limiter)
		// You can use any mux you like - NewHelloWorldServer gives you an http.Handler.
		mux := http.NewServeMux()
		// The generated code includes a method, PathPrefix(), which
		// can be used to mount your service on a mux.
		mux.Handle(twirpHandler.PathPrefix(), twirpHandler)

		delegate := func(resp http.ResponseWriter, req *http.Request) {
			if limiter.AttemptAccess(costOfRequest(req), tenantIdentifiers(req)...) {
				servlet.ServeHTTP(resp, req)
				return
			} // else access attempt failed

			resp.WriteHeader(http.StatusTooManyRequests)
			io.WriteString(resp, "Throttle limit exceeded.")
			return
		}
		mux.Handle("/", http.HandlerFunc(delegate))

		return mux
	}
}