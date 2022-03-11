package parser

import "github.com/nikitaksv/gendata/internal/generator/meta"

type Option func(opts *Options) error

func WithRootName(rootName string) Option {
	return func(opts *Options) error {
		opts.RootName = rootName
		return nil
	}
}
func WithOptions(opts Options) Option {
	return func(opts_ *Options) error {
		opts_ = &opts
		return nil
	}
}

type Options struct {
	RootName string
}

func (o *Options) apply(opts ...Option) error {
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return err
		}
	}
	return nil
}

type Parser interface {
	Parse(data []byte, opts ...Option) (*meta.Nest, error)
}
