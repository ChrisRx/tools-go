package alias

import (
	"go.chrisrx.dev/x/slices"
)

type options struct {
	Ignore  []string
	Include []string
}

func newOptions(opts []Option) *options {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func (o *options) ShouldInclude(name string) bool {
	if len(o.Include) > 0 {
		return slices.Contains(o.Include, name)
	}
	if len(o.Ignore) > 0 {
		return !slices.Contains(o.Ignore, name)
	}
	return true
}

type Option func(o *options)

func Ignore(ignore ...string) Option {
	return func(o *options) {
		o.Ignore = slices.DeleteFunc(ignore, func(s string) bool {
			return s == ""
		})
	}
}

func Include(include ...string) Option {
	return func(o *options) {
		o.Include = slices.DeleteFunc(include, func(s string) bool {
			return s == ""
		})
	}
}
