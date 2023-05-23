package metadata

import "context"

type Metadata map[string]interface{}

type ctxkey struct{}

// Len returns the number of items in m
func (m Metadata) Len() int {
	return len(m)
}

// Join joins any number of mds into a single Metadata
func Join(ms ...Metadata) Metadata {
	var out Metadata = make(map[string]interface{}, 0)
	for _, item := range ms {
		for key, val := range item {
			out[key] = val
		}
	}
	return out
}

// WithContext creates a new context with m attached
func (m Metadata) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxkey{}, m)
}

// FromContext .
func FromContext(ctx context.Context) (Metadata, bool) {
	m, ok := ctx.Value(ctxkey{}).(Metadata)
	return m, ok
}

// SetContext .
func SetContext(ctx context.Context, key string, val interface{}) context.Context {
	if m, ok := FromContext(ctx); ok {
		m[key] = val
		return m.WithContext(ctx)
	}
	m := &Metadata{key: val}
	return m.WithContext(ctx)
}

// Int64 get int64 value from metadata in context
func Int64(ctx context.Context, key string) int64 {
	if m, ok := FromContext(ctx); ok {
		val, _ := m[key].(int64)
		return val
	}
	return 0
}

// Bool get bool value from metadata in context
func Bool(ctx context.Context, key string) bool {
	if m, ok := FromContext(ctx); ok {
		val, _ := m[key].(bool)
		return val
	}
	return false
}

// String get string value from metadata in context
func String(ctx context.Context, key string) string {
	if m, ok := FromContext(ctx); ok {
		val, _ := m[key].(string)
		return val
	}
	return ""
}

// Value get value from metadata in context return nil if not found
func Value(ctx context.Context, key string) interface{} {
	if m, ok := FromContext(ctx); ok {
		return m[key]
	}
	return nil
}
