package ratelimit

import (
	"math/rand"
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

	for i := 0; i < b.N; i++ {
		limiter.AttemptAccess("Dave", 1)
	}
}

func Benchmark_cost_of_random_names(b *testing.B) {
	gen := random_strings.ByteStringGenerator{
		Alphabet:  random_strings.EnglishAlphabet,
		RandomGen: rand.New(rand.NewSource(0)),
	}

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

	for i := 0; i < b.N; i++ {
		limiter.AttemptAccess(gen.String(5), 1)
	}
}

func Benchmark_churn_without_gen(b *testing.B) {
	var config = defaultConfig
	config.CapacityBound = 4

	limiter := NewSlidingWindowRateLimiter(config)
	limiter.clock = HardwareClock{}
	limiter.log = test_logger.NoopLogger{}

	keys := []string{
		"aaaa",
		"bbbb",
		"cccc",
		"dddd",
	}

	for i := 0; i < b.N; i++ {
		limiter.AttemptAccess(keys[i % len(keys)], 1)
	}
}
