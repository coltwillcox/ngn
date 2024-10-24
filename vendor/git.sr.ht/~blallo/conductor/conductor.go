package conductor

import (
	"context"
	"os"
)

// Sender is the return type of the [Send] function.
type Sender[T any] func(cmd T)

// Notifier is the return type of the [Notify] function.
type Notifier[T any] func(cmd T, signals ...os.Signal)

// Conductor is a type useful to convey commands (represented by the generic type T)
// across API and goroutine boundaries. It can be thought as a generalization of as
// [context.Context] and, as such, it also implements the [context.Context] interface.
type Conductor[T any] interface {
	// Cmd returns a channel where to listen to for commands. Is the analogous of
	// [context.Context.Done].
	Cmd() <-chan T
	// WithContext assigns the given [context.Context] to the [Conductor], replacing
	// the one currently assigned to it.
	WithContext(context.Context) Conductor[T]
	// WithContextPolicy attaches a [Policy] to the given [Conductor].
	WithContextPolicy(Policy[T]) Conductor[T]
	context.Context
}

// Send may be used on a [Conductor] to create a function to send a command to the
// interested listeners. It accepts a variadic amount of arguments to accommodate
// custom behavior, depending on the specific instance of a [Conductor] it acts on.
func Send[T any](conductor Conductor[T], args ...any) Sender[T] {
	switch c := any(conductor).(type) {
	case *simple[T]:
		return c.send
	case *tagged[T]:
		if len(args) == 0 {
			return c.broadcast
		}
		return func(cmd T) {
			c.send(cmd, args)
		}
	default:
		panic("conductor not supported")
	}
}

// Notify may be used on a [Conductor] to create a function to register it to an [os.Signal],
// in the same spirit as [os/signal.Notify]. The optional variadic args may be used to
// configure this mechanism, depending on the specific instance of the provided [Conductor].
func Notify[T any](conductor Conductor[T], args ...any) Notifier[T] {
	switch c := any(conductor).(type) {
	case *simple[T]:
		return c.notify
	case *tagged[T]:
		if len(args) == 0 {
			return c.notifyAll
		}
		return func(cmd T, signals ...os.Signal) {
			c.notifyTagged(cmd, args, signals)
		}
	default:
		panic("conductor not supported")
	}
}
