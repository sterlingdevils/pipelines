package retrypipe

import (
	"time"
)

type RetryThing[K comparable, T Retryable[K]] struct {
	thing      T
	key        K
	created    time.Time
	NextRetry  time.Time
	ExpireTime time.Time
}

func (p RetryThing[K, _]) Key() K {
	return p.key
}

func (p RetryThing[_, T]) Thing() T {
	return p.thing
}

func (p RetryThing[_, _]) Created() time.Time {
	return p.created
}

func (RetryThing[K, T]) New(k K, t T) *RetryThing[K, T] {
	return &RetryThing[K, T]{thing: t, key: k, created: time.Now()}
}
