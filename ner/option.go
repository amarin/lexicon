package ner

// Option modifies one Extract call.
type Option func(*options)

// Explain fills Span.Evidence.
func Explain() Option { return func(o *options) { o.explain = true } }
