package leakybucket

import (
	"github.com/npxcomplete/http-rate-limit/src/test_logger"
	random_strings "github.com/npxcomplete/random/src/strings"
	"math/rand"
	"sync"
	"testing"
)

var uniformLimit = TenantLimit{
	Rate:  100,
	Burst: 100,
}

func uniformLimits(tenant string) *TenantLimit {
	return &uniformLimit
}

func Benchmark_leaky_bucket_access_attempts_on_single_user(b *testing.B) {
	logs := test_logger.NoopLogger{}

	limiter := NewRateLimiter(Config{
		Tenancy: uniformLimits,
		TenantCapacity: 10,
	})

	limiter.log = logs

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.AttemptAccess(1, "Dave")
	}
}

func Benchmark_leaky_bucket_cost_of_random_names(b *testing.B) {
	gen := random_strings.ByteStringGenerator{
		Alphabet:  random_strings.EnglishAlphabet,
		RandomGen: rand.New(rand.NewSource(0)),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.String(5)
	}
}

func Benchmark_leaky_bucket_access_attempts_on_many_users_with_high_eviction(b *testing.B) {
	logs := test_logger.NoopLogger{}

	limiter := NewRateLimiter(Config{
		Tenancy: uniformLimits,
		TenantCapacity: 10,
	})
	limiter.log = logs

	gen := random_strings.ByteStringGenerator{
		Alphabet:  random_strings.EnglishAlphabet,
		RandomGen: rand.New(rand.NewSource(0)),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.AttemptAccess(1, gen.String(5))
	}
}

func Benchmark_leaky_bucket_access_attempts_on_many_users(b *testing.B) {
	logs := test_logger.NoopLogger{}
	limiter := NewRateLimiter(Config{
		Tenancy: uniformLimits,
		TenantCapacity: 10,
	})
	limiter.log = logs

	gen := random_strings.ByteStringGenerator{
		Alphabet:  random_strings.EnglishAlphabet,
		RandomGen: rand.New(rand.NewSource(0)),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.AttemptAccess( 1, gen.String(5))
	}
}

var thread_count = 4
var capacity = 5

func Benchmark_leaky_bucket_serial_access(b *testing.B) {
	logs := test_logger.NoopLogger{}
	limiter := NewRateLimiter(Config{
		Tenancy: uniformLimits,
		TenantCapacity: 10,
	})
	limiter.log = logs

	gen := random_strings.ByteStringGenerator{
		Alphabet:  random_strings.EnglishAlphabet,
		RandomGen: rand.New(rand.NewSource(0)),
	}

	keys := make([]string, capacity*2)
	for i := 0; i < len(keys); i++ {
		keys[i] = gen.String(12)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.AttemptAccess( 1, keys[rand.Int31n(int32(len(keys)))])
	}
}
func Benchmark_leaky_bucket_parallel_access(b *testing.B) {
	logs := test_logger.NoopLogger{}
	limiter := NewRateLimiter(Config{
		Tenancy: uniformLimits,
		TenantCapacity: 10,
	})
	limiter.log = logs

	gen := random_strings.ByteStringGenerator{
		Alphabet:  random_strings.EnglishAlphabet,
		RandomGen: rand.New(rand.NewSource(0)),
	}

	keys := make([]string, capacity*2)
	for i := 0; i < len(keys); i++ {
		keys[i] = gen.String(12)
	}

	var wg sync.WaitGroup
	b.ResetTimer()
	for t := 0; t < thread_count; t++ {
		wg.Add(1)
		go func() {
			for i := 0; i < b.N; i++ {
				limiter.AttemptAccess(1, keys[rand.Int31n(int32(len(keys)))])
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func Benchmark_leaky_bucket_churn_without_gen(b *testing.B) {
	limiter := NewRateLimiter(Config{
		Tenancy: uniformLimits,
		TenantCapacity: 10,
	})
	limiter.log = test_logger.NoopLogger{}

	keys := []string{
		"aaaa",
		"bbbb",
		"cccc",
		"dddd",
	}

	for i := 0; i < b.N; i++ {
		limiter.AttemptAccess(1, keys[i%len(keys)])
	}
}
