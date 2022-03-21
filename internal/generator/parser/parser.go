package parser

import "github.com/nikitaksv/gendata/internal/generator/meta"

type ClassNameFormatter func(key meta.Key) string

type Option func(opts *Options) error

func WithOptions(opts *Options) Option {
	return func(opts_ *Options) error {
		opts_.RootClassName = opts.RootClassName
		opts_.PrefixClassName = opts.PrefixClassName
		opts_.SuffixClassName = opts.SuffixClassName
		opts_.SortProperties = opts.SortProperties
		opts_.TypeFormatters = opts.TypeFormatters
		opts_.ClassNameFormatter = opts.ClassNameFormatter
		return nil
	}
}

type Options struct {
	// Root object name
	RootClassName      string               `json:"root_class_name"`
	PrefixClassName    string               `json:"prefix_class_name"`
	SuffixClassName    string               `json:"suffix_class_name"`
	SortProperties     bool                 `json:"sort_properties"`
	TypeFormatters     *meta.TypeFormatters `json:"type_formatters"`
	ClassNameFormatter ClassNameFormatter   `json:"class_name_formatter"`
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
	Parse(data []byte) (*meta.Nest, error)
}
