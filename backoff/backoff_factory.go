package backoff

import (
	"math/rand"
	"time"
)

// BackoffFactory is used for creating backoff objects with a single shared RandomGenerator.
//
// BackoffFactory is useful if you need to repeatedly create backoff objects in multiple go
// routines.
//
// Sharing a random generator is good, since every time a generator is created it must be seeded,
// which is expensive operation.
//
// Example:
//  // the shared RandomGenerator seeding happens here
//	backoffFactory := backoff.NewBackoffFactory()
//
//	for i := 0; i < 10; i++ {
//		go func() {
//			// created backoff instances use a shared RandomGenerator
//			backoff := backoffFactory.CreateBackoff(backoff.DefaultConfig)
//
//			// use the backoff
//		}()
//	}
type BackoffFactory struct {
	randomGen RandomGenerator
}

// NewBackoffFactory creates a BackoffFactory instance using a standard random generator.
//
// The shared random generator is seeded with the current time in nanoseconds, and wrapped with
// SyncRandomGenerator for multi go routine safety.
func NewBackoffFactory() *BackoffFactory {
	return &BackoffFactory{
		randomGen: NewSyncRandomGenerator(rand.New(rand.NewSource(time.Now().UnixNano()))), // nolint: gosec
	}
}

// CreateBackoff returns a Backoff instance using a shared RandomGenerator.
func (bf *BackoffFactory) CreateBackoff(config *Config) *Backoff {
	return NewBackoffWithRandomGen(
		bf.randomGen,
		config,
	)
}
