package backoff

import (
	"time"
)

type RandomGenerator interface {
	Int63n(n int64) int64
}

type Config struct {
	Base                 time.Duration
	Max                  time.Duration
	BackOffResetDuration time.Duration
	Jitter               Jitter
}

var DefaultConfig = &Config{
	Base:                 time.Second,
	Max:                  time.Second * 30,
	BackOffResetDuration: time.Second * 60,
	Jitter:               FullJitter,
}

// Backoff The backoff is calculated by min(Max, rand(0, Base*(2^retries))).
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

	return &Backoff{
		config:    config,
		randomGen: randomGen,
	}
}

func (b *Backoff) Next() time.Duration {
	now := time.Now()
	elapsed := float64(now.Sub(b.tn)) / float64(time.Second)

	if elapsed > float64(b.config.BackOffResetDuration) {
		b.retryCount = 0
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
