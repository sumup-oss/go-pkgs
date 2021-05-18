package backoff

import "time"

type Jitter func(randomGen RandomGenerator, n int64) time.Duration

func FullJitter(randomGen RandomGenerator, n int64) time.Duration {
	return time.Duration(1 + randomGen.Int63n(n))
}

func EqualJitter(randomGen RandomGenerator, n int64) time.Duration {
	half := 1 + n/2
	return time.Duration(half + randomGen.Int63n(half))
}
