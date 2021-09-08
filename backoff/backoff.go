package backoff

import (
	"time"
)

// Config is used for the backoff constructor if you don't want to use the default config.
type Config struct {
	// Base is the duration used for the next duration calculation.
	Base time.Duration
	// Max is the maximum duration the backoff can return.
	Max time.Duration
	// Jitter is the Jitter used to randomize the next duration.
	Jitter Jitter
}

var DefaultConfig = &Config{
	Base:   time.Second,
	Max:    time.Second * 30,
	Jitter: FullJitter,
}

// Backoff is used to calculate the next backoff duration using a Jitter.
//
// Check here why to use Jitter - https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/
//
// The backoff is calculated by min(Max, rand(0, Base*(2^retries))).
//
// Example:
//	b := backoff.NewBackoff(backoff.DefaultConfig)
//
//  for {
//		err := someFunc()
//		if err == nil {
//			// all good
//			return nil
//		}
//
//		// get the next sleep duration
//		sleep := b.Next()
//
// 		select {
//		case <-time.After(sleep):
//			continue
//		case <-ctx.Done():
//			return errors.Wrap(err, "retry canceled")
//		}
//	}
type Backoff struct {
	config *Config

	randomGen  RandomGenerator
	retryCount uint
}

// NewBackoff uses the golang rand generator from the standard library.
//
// It is safe to be used in multiple go routines.
func NewBackoff(config *Config) *Backoff {
	return NewBackoffWithRandomGen(
		DefaultRandomGenerator(),
		config,
	)
}

// NewBackoffWithRandomGen is used when you want to pass a custom RandomGenerator.
//
// Example:
//  // Note that the backoff created bellow can be used safely only in a single go routine.
//	b := backoff.New(
//		rand.New(rand.NewSource(time.Now().UnixNano())),
//		backoff.DefaultConfig
//	)
func NewBackoffWithRandomGen(randomGen RandomGenerator, config *Config) *Backoff {
	if config.Base == 0 {
		config.Base = DefaultConfig.Base
	}

	if config.Max == 0 {
		config.Max = DefaultConfig.Max
	}

	if config.Jitter == nil {
		config.Jitter = DefaultConfig.Jitter
	}

	return &Backoff{
		config:    config,
		randomGen: randomGen,
	}
}

// Next returns the next duration for the retry.
func (b *Backoff) Next() time.Duration {
	d := b.config.Base * (1 << b.retryCount) // nolint: durationcheck
	if d > b.config.Max {
		d = b.config.Max
	} else {
		b.retryCount++
	}

	return b.config.Jitter(b.randomGen, int64(d))
}
