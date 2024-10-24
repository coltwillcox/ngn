package conductor

// Policy is invoked when the [Conductor] gets cancelled, in the sense of [context] (i.e.
// being it a [context.Context], it can be canceled in the usual way). When this happens,
// a Policy is a way to decide what to do with all the listeners loaded in a [Conductor].
type Policy[T any] interface {
	// Decide gets invoked when the [Conductor] is canceled.
	Decide(args ...any) (T, bool)
}

type constantPolicy[T any] struct {
	cmd T
}

func (p *constantPolicy[T]) Decide(...any) (T, bool) {
	return p.cmd, true
}

// ConstantPolicy creates a [Policy] that fires the given command to all the listeners.
func ConstantPolicy[T any](cmd T) Policy[T] {
	return &constantPolicy[T]{
		cmd: cmd,
	}
}

type setPolicy[T any] struct {
	mapping map[any]T
}

func (p *setPolicy[T]) Decide(tags ...any) (zero T, falsy bool) {
	for _, tag := range tags {
		if cmd, ok := p.mapping[tag]; ok {
			return cmd, true
		}
	}

	return zero, false
}

// SetPolicy creates a [Policy] that uses the provided map to conditionally fire
// the command found at the associated identifier in the map. It is to be used only
// with a [Tagged] [Conductor]. If attached to a [Simple], it is a noop.
func SetPolicy[T any](mapping map[any]T) Policy[T] {
	return &setPolicy[T]{
		mapping: mapping,
	}
}
