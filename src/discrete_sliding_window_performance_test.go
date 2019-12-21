package ratelimit

import (
	"math/rand"
	"sync"
	"testing"

	"github.com/npxcomplete/random/src/strings"

	"github.com/npxcomplete/http-rate-limit/src/test_logger"
)

func Benchmark_access_attempts_on_single_user(b *testing.B) {
	logs := test_logger.NoopLogger{}
	var config = defaultConfig
	config.CapacityBound = 2

	limiter := NewSlidingWindowRateLimiter(config)
	limiter.clock = HardwareClock{}
	limiter.log = logs

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.AttemptAccess("Dave", 1)
	}
}

func Benchmark_cost_of_random_names(b *testing.B) {
	gen := random_strings.ByteStringGenerator{
		Alphabet:  random_strings.EnglishAlphabet,
		RandomGen: rand.New(rand.NewSource(0)),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.String(5)
	}
}

func Benchmark_access_attempts_on_many_users_with_high_eviction(b *testing.B) {
	logs := test_logger.NoopLogger{}
	var config = defaultConfig
	config.CapacityBound = 2

	limiter := NewSlidingWindowRateLimiter(config)
	limiter.clock = HardwareClock{}
	limiter.log = logs

	gen := random_strings.ByteStringGenerator{
		Alphabet:  random_strings.EnglishAlphabet,
		RandomGen: rand.New(rand.NewSource(0)),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.AttemptAccess(gen.String(5), 1)
	}
}

func Benchmark_access_attempts_on_many_users(b *testing.B) {
	logs := test_logger.NoopLogger{}
	var config = defaultConfig
	config.CapacityBound = 500

	limiter := NewSlidingWindowRateLimiter(config)
	limiter.clock = HardwareClock{}
	limiter.log = logs

	gen := random_strings.ByteStringGenerator{
		Alphabet:  random_strings.EnglishAlphabet,
		RandomGen: rand.New(rand.NewSource(0)),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.AttemptAccess(gen.String(5), 1)
	}
}

var thread_count = 4
var capacity = 5

func Benchmark_serial_access(b *testing.B) {
	logs := test_logger.NoopLogger{}
	var config = defaultConfig
	config.CapacityBound = capacity

	limiter := NewSlidingWindowRateLimiter(config)
	limiter.clock = HardwareClock{}
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
		limiter.AttemptAccess(keys[rand.Int31n(int32(len(keys)))], 1)
	}
}
func Benchmark_parallel_access(b *testing.B) {
	logs := test_logger.NoopLogger{}
	var config = defaultConfig
	config.CapacityBound = capacity

	limiter := NewSlidingWindowRateLimiter(config)
	limiter.clock = HardwareClock{}
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
				limiter.AttemptAccess(keys[rand.Int31n(int32(len(keys)))], 1)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
