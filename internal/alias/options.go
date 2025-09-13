package alias

import (
	"go.chrisrx.dev/x/slices"
)

type options struct {
	Docs           string
	Ignore         []string
	Include        []string
	GoBuildVersion string
}

func newOptions(opts []Option) *options {
	o := &options{
		Docs: "none",
	}
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

func (o *options) IncludeDocs(docs []string, opts ...string) []string {
	if slices.Contains(opts, o.Docs) {
		return docs
	}
	return nil
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

func Docs(docs string) Option {
	return func(o *options) {
		o.Docs = docs
	}
}
func GoBuildVersion(v string) Option {
	return func(o *options) {
		o.GoBuildVersion = v
	}
}
