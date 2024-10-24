package conductor

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"
)

var defaultTag string = "CONDUCTOR_INTERNAL_DEFAULT_TAG"

type tagged[T any] struct {
	tagged map[any]*simple[T]
	mu     sync.RWMutex
	ctx    context.Context
}

/* Implement context.Context */

var _ context.Context = &tagged[struct{}]{}

func (t *tagged[T]) Deadline() (time.Time, bool) {
	return t.ctx.Deadline()
}

func (t *tagged[T]) Done() <-chan struct{} {
	return t.ctx.Done()
}

func (t *tagged[T]) Err() error {
	return t.ctx.Err()
}

func (t *tagged[T]) Value(key any) any {
	return t.ctx.Value(key)
}

/* Implement Conductor[T] */

func (t *tagged[T]) Cmd() <-chan T {
	return t.cmd(defaultTag)
}

func (t *tagged[T]) WithContext(ctx context.Context) Conductor[T] {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ctx = ctx
	for _, tagged := range t.tagged {
		tagged.WithContext(ctx)
	}
	return t
}

func (c *tagged[T]) WithContextPolicy(policy Policy[T]) Conductor[T] {
	go func() {
		<-c.ctx.Done()
		c.mu.Lock()
		defer c.mu.Unlock()
		for tag, lis := range c.tagged {
			if cmd, ok := policy.Decide(tag); ok {
				lis.send(cmd)
			}
		}
	}()

	return c
}

/* Internal functions */

func (t *tagged[T]) cmd(tag string, discriminator ...any) <-chan T {
	t.mu.Lock()
	defer t.mu.Unlock()

	discriminator = append([]any{any(tag)}, discriminator...)
	if c, ok := t.tagged[tag]; ok {
		return c.cmd(3, discriminator...)
	}
	c := SimpleFromContext[T](t.ctx)
	t.tagged[tag] = c.(*simple[T])
	return c.(*simple[T]).cmd(2, discriminator...)
}

func (t *tagged[T]) send(cmd T, tags []any) {
	t.mu.RLock()
	fmt.Fprintf(logFile, "Sending %s to %s listener\n", fmtCmd(cmd), tags)
	for _, tag := range append(tags, defaultTag) {
		if c, ok := t.tagged[tag]; ok {
			c.send(cmd)
		}
	}
	t.mu.RUnlock()
}

func (t *tagged[T]) broadcast(cmd T) {
	t.mu.RLock()
	fmt.Fprintf(logFile, "Sending %s to all listener\n", fmtCmd(cmd))
	for _, c := range t.tagged {
		c.send(cmd)
	}
	t.mu.RUnlock()
}

func (t *tagged[T]) notifyAll(cmd T, signals ...os.Signal) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)
	for {
		select {
		case <-t.Done():
			return
		case <-ch:
			t.broadcast(cmd)
		}
	}
}

func (t *tagged[T]) notifyTagged(cmd T, tags []any, signals []os.Signal) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)
	for {
		select {
		case <-t.Done():
			return
		case <-ch:
			t.send(cmd, tags)
		}
	}
}

/* Public functions */

// Tagged creates a [Conductor] that supports tagged listeners.
func Tagged[T any]() Conductor[T] {
	return &tagged[T]{
		tagged: make(map[any]*simple[T]),
		ctx:    context.TODO(),
	}
}

// WithTag loads a tagged listener in a Tagged [Conductor].
func WithTag[T any](conductor Conductor[T], tag string, discriminator ...any) Conductor[T] {
	c, ok := any(conductor).(*tagged[T])
	if !ok {
		panic("not a conductor.Tagged")
	}
	return &loaded[T]{
		wrapped:       c,
		tag:           tag,
		discriminator: discriminator,
	}
}

// TaggedFromContext creates a Tagged [Conductor] from a given [context.Context].
func TaggedFromContext[T any](parent context.Context) Conductor[T] {
	t := Tagged[T]()
	t.(*tagged[T]).ctx = parent

	return t
}

// TaggedFromSimple transforms a Simple [Conductor] in a Tagged one.
func TaggedFromSimple[T any](s Conductor[T]) Conductor[T] {
	c, ok := any(s).(*simple[T])
	if !ok {
		panic("not a conductor.Simple")
	}
	return &tagged[T]{
		tagged: map[any]*simple[T]{
			defaultTag: c,
		},
		ctx: c.ctx,
	}
}
