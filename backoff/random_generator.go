package backoff

import (
	"math/rand"
	"sync"
	"time"
)

// RandomGenerator interface for the random generator.
//
// Standard library can be used as follows: rand.NewSource(time.Now().UnixNano()).
//
// Note that the standard generator returned by rand.NewSource is multi go routine unsafe.
// For multi go routine safety use NewSyncRandomGenerator(rand.NewSource(time.Now().UnixNano())).
// Or just use the DefaultRandomGenerator function, that returns a go routine safe random generator.
type RandomGenerator interface {
	Int63n(n int64) int64
}

// DefaultRandomGenerator is the default generator used by NewBackoff function.
func DefaultRandomGenerator() *SyncRandomGenerator {
	return NewSyncRandomGenerator(rand.New(rand.NewSource(time.Now().UnixNano()))) // nolint: gosec
}

// SyncRandomGenerator is a wrapper that makes a RandomGenerator multi go routine safe.
type SyncRandomGenerator struct {
	RandomGenerator
	mu sync.Mutex
}

// NewSyncRandomGenerator creates SyncRandomGenerator instance.
func NewSyncRandomGenerator(generator RandomGenerator) *SyncRandomGenerator {
	return &SyncRandomGenerator{
		RandomGenerator: generator,
	}
}

// Int63n returns a int64 random integer in the half open interval [0, n).
func (g *SyncRandomGenerator) Int63n(n int64) int64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.RandomGenerator.Int63n(n)
}
