package logz

// WithAttrs wraps a given logger and ensures that
func WithAttrs(logger Logger, attrs map[string]any) Logger {
	// Ensure that we do not wrap twice: unwrap and patch
	// the alredy existing attributes.
	switch l := logger.(type) {
	case *withAttrs:
		// NOTE: we want the attrs given in this call have precedence
		// over those already present in the given withAttrs Logger.
		newAttrs := make(map[string]any)
		for k, v := range l.attrs {
			newAttrs[k] = v
		}
		for k, v := range attrs {
			newAttrs[k] = v
		}
		// NOTE: we do not patch the existing withAttrs Logger, but
		// return a new one, not to override the one passed in.
		return &withAttrs{
			wrapped: l.wrapped,
			attrs:   newAttrs,
		}
	}

	return &withAttrs{
		wrapped: logger,
		attrs:   attrs,
	}
}

type withAttrs struct {
	wrapped Logger
	attrs   map[string]any
}

func (l *withAttrs) patch(data map[string]any) map[string]any {
	out := make(map[string]any)
	for k, v := range data {
		out[k] = v
	}

	for k, v := range l.attrs {
		out[k] = v
	}

	return out
}

func (l *withAttrs) Log(level LogLevel, data map[string]interface{}) {
	l.wrapped.Log(level, l.patch(data))
}

func (l *withAttrs) Trace(data map[string]interface{}) {
	l.wrapped.Trace(l.patch(data))
}

func (l *withAttrs) Debug(data map[string]interface{}) {
	l.wrapped.Debug(l.patch(data))
}

func (l *withAttrs) Info(data map[string]interface{}) {
	l.wrapped.Info(l.patch(data))
}

func (l *withAttrs) Warn(data map[string]interface{}) {
	l.wrapped.Warn(l.patch(data))
}

func (l *withAttrs) Err(data map[string]interface{}) {
	l.wrapped.Err(l.patch(data))
}

func (l *withAttrs) SetLevel(level LogLevel) Logger {
	l.wrapped.SetLevel(level)
	return l
}
