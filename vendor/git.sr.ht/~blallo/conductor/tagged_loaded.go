package conductor

import (
	"context"
	"time"
)

type loaded[T any] struct {
	wrapped       *tagged[T]
	tag           string
	discriminator []any
}

/* Implement context.Context */

var _ context.Context = &loaded[struct{}]{}

func (l *loaded[T]) Deadline() (time.Time, bool) {
	return l.wrapped.Deadline()
}

func (l *loaded[T]) Done() <-chan struct{} {
	return l.wrapped.Done()
}

func (l *loaded[T]) Err() error {
	return l.wrapped.Err()
}

func (l *loaded[T]) Value(key any) any {
	return l.wrapped.Value(key)
}

/* Implement Conductor[T] */

func (l *loaded[T]) Cmd() <-chan T {
	return l.wrapped.cmd(l.tag, l.discriminator...)
}

func (l *loaded[T]) WithContext(ctx context.Context) Conductor[T] {
	return l.wrapped.WithContext(ctx)
}

func (l *loaded[T]) WithContextPolicy(policy Policy[T]) Conductor[T] {
	return l.wrapped.WithContextPolicy(policy)
}
