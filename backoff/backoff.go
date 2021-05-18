package backoff

import (
	"time"
)

// RandomGenerator interface for the random generator.
// Standard library can be used as follows: rand.NewSource(time.Now().UnixNano()).
type RandomGenerator interface {
	Int63n(n int64) int64
}

// Config is used for the backoff constructor if you don't want to use the default config.
type Config struct {
	// Base is the duration used for the next duration calculation.
	Base time.Duration
	// Max is the maximum duration the backoff can return.
	Max time.Duration
	// BackoffResetDuration is the duration between two Next() calls that the Backoff will wait until he resets the counter.
	BackoffResetDuration time.Duration
	// Jitter is the Jitter used to randomize the next duration.
	Jitter Jitter
}

var DefaultConfig = &Config{
	Base:                 time.Second,
	Max:                  time.Second * 30,
	BackoffResetDuration: time.Second * 60,
	Jitter:               FullJitter,
}

// Backoff is used to calculate the next backoff duration using a Jitter.
// Check here why to use Jitter - https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/
// The backoff is calculated by min(Max, rand(0, Base*(2^retries))).
// example: backoffWithDefaultConfig := backoff.New(rand.New(rand.NewSource(time.Now().UnixNano())), backoff.DefaultConfig)
type Backoff struct {
	config *Config

	randomGen  RandomGenerator
	retryCount uint
	tn         time.Time
}

func New(randomGen RandomGenerator, config *Config) *Backoff {
	if config.Base == 0 {
		config.Base = DefaultConfig.Base
	}

	if config.Max == 0 {
		config.Max = DefaultConfig.Max
	}

	if config.Jitter == nil {
		config.Jitter = DefaultConfig.Jitter
	}

	if config.BackoffResetDuration == 0 {
		config.BackoffResetDuration = DefaultConfig.BackoffResetDuration
	}

	return &Backoff{
		config:    config,
		randomGen: randomGen,
	}
}

// Next returns the next duration for the retry.
func (b *Backoff) Next() time.Duration {
	now := time.Now()
	elapsed := now.Sub(b.tn)

	if b.config.BackoffResetDuration != 0 {
		if elapsed > b.config.BackoffResetDuration {
			b.retryCount = 0
		}
	}

	d := b.config.Base * (1 << b.retryCount)
	if d > b.config.Max {
		d = b.config.Max
	} else {
		b.retryCount++
	}

	b.tn = now

	return b.config.Jitter(b.randomGen, int64(d))
}

// Retried returns the number of retries
func (b *Backoff) Retried() uint {
	return b.retryCount
}
