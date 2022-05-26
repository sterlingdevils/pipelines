package retrypipe

import (
	"context"
)

type RetryThing[K comparable, T any] struct {
	Thing T
	key   K
	ctx   context.Context
	can   context.CancelFunc
}

func (p *RetryThing[K, _]) Key() K {
	return p.key
}

func (p *RetryThing[K, _]) Context() context.Context {
	return p.ctx
}

func NewRetryThing[K comparable, T any](k K, t T) *RetryThing[K, T] {
	c, cancel := context.WithCancel(context.Background())
	p := &RetryThing[K, T]{Thing: t, key: k, ctx: c, can: cancel}

	return p
}
