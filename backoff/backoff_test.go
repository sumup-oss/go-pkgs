package backoff_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	backoff "github.com/sumup-oss/go-pkgs/backoff"
	"testing"
	"time"
)

// RandomGeneratorMock...
type RandomGeneratorMock struct {
	mock.Mock
}

func (r *RandomGeneratorMock) Int63n(n int64) int64 {
	args := r.Called(n)
	return args.Get(0).(int64)
}

func getDurationForRetry(retry uint, duration time.Duration) int64 {
	return int64(duration * (1 << retry))
}

func TestNewBackOff(t *testing.T) {
	t.Run(
		"check backoff implementation with default config using Full Jitter",
		func(t *testing.T) {
			t.Parallel()
			randomGeneratorMock := &RandomGeneratorMock{}
			randomGeneratorMock.On("Int63n", getDurationForRetry(0, time.Second)).Return(int64(100))
			randomGeneratorMock.On("Int63n", getDurationForRetry(1, time.Second)).Return(int64(200))
			randomGeneratorMock.On("Int63n", getDurationForRetry(2, time.Second)).Return(int64(300))
			randomGeneratorMock.On("Int63n", getDurationForRetry(3, time.Second)).Return(int64(400))

			backoffDefault := backoff.NewBackoffWithRandomGen(randomGeneratorMock, backoff.DefaultConfig)
			duration := backoffDefault.Next()

			assert.Equal(t, int64(101), duration.Nanoseconds())

			duration = backoffDefault.Next()
			assert.Equal(t, int64(201), duration.Nanoseconds())

			duration = backoffDefault.Next()
			assert.Equal(t, int64(301), duration.Nanoseconds())

			duration = backoffDefault.Next()
			assert.Equal(t, int64(401), duration.Nanoseconds())
		},
	)

	t.Run(
		"check backoff implementation using custom config and using Equal Jitter",
		func(t *testing.T) {
			t.Parallel()
			randomGeneratorMock := &RandomGeneratorMock{}

			randomGeneratorMock.On("Int63n", getDurationForRetry(0, time.Minute)/2+1).Return(int64(100))
			randomGeneratorMock.On("Int63n", getDurationForRetry(1, time.Minute)/2+1).Return(int64(200))
			randomGeneratorMock.On("Int63n", getDurationForRetry(2, time.Minute)/2+1).Return(int64(300))
			randomGeneratorMock.On("Int63n", getDurationForRetry(3, time.Minute)/2+1).Return(int64(400))

			backoffEqJitter := backoff.NewBackoffWithRandomGen(
				randomGeneratorMock,
				&backoff.Config{
					Base:   time.Minute,
					Max:    time.Minute * 10,
					Jitter: backoff.EqualJitter,
				},
			)

			duration := backoffEqJitter.Next()
			assert.Equal(t, int64(30000000101), duration.Nanoseconds())

			duration = backoffEqJitter.Next()
			assert.Equal(t, int64(60000000201), duration.Nanoseconds())

			duration = backoffEqJitter.Next()
			assert.Equal(t, int64(120000000301), duration.Nanoseconds())

			duration = backoffEqJitter.Next()
			assert.Equal(t, int64(240000000401), duration.Nanoseconds())
		},
	)
}
