package backoff

import "time"

type Jitter func(randomGen RandomGenerator, factor int64) time.Duration

func FullJitter(randomGen RandomGenerator, factor int64) time.Duration {
	return time.Duration(1 + randomGen.Int63n(factor))
}

func EqualJitter(randomGen RandomGenerator, factor int64) time.Duration {
	half := 1 + factor/2
	return time.Duration(half + randomGen.Int63n(half))
}
