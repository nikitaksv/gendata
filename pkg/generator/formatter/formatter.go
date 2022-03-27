package formatter

import (
	"github.com/nikitaksv/gendata/pkg/generator/meta"
)

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
	RootClassName      string               `json:"rootClassName"`
	PrefixClassName    string               `json:"prefixClassName"`
	SuffixClassName    string               `json:"suffixClassName"`
	SortProperties     bool                 `json:"sortProperties"`
	TypeFormatters     *meta.TypeFormatters `json:"typeFormatters"`
	ClassNameFormatter ClassNameFormatter   `json:"classNameFormatter"`
}

func (o *Options) apply(opts ...Option) error {
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return err
		}
	}
	return nil
}

type Formatter interface {
	Format(m *meta.Meta, opts ...Option) (*meta.Meta, error)
}

func NewFormatter() Formatter {
	return &formatter{}
}

type formatter struct{}

func (f formatter) Format(m *meta.Meta, opts ...Option) (*meta.Meta, error) {
	options := &Options{}
	if err := options.apply(opts...); err != nil {
		return nil, err
	}

	if options.SortProperties {
		m.Sort()
	}

	if m.Key.String() == "" {
		m.Key = meta.Key(options.RootClassName)
	}

	m.Key = meta.Key(f.className(m.Key, options))
	m.Type.Key = m.Key
	m.Type.Formatters = options.TypeFormatters
	for _, property := range m.Properties {
		property.Type.Formatters = options.TypeFormatters

		if property.Type.IsObject() || property.Type.Value == meta.TypeArrayObject {
			property.Key = meta.Key(f.className(property.Key, options))
			property.Type.Key = property.Key
		}
		if property.Nest != nil {
			var err error
			property.Nest, err = f.Format(property.Nest, opts...)
			if err != nil {
				return nil, err
			}
		}
	}

	return m, nil
}

func (f formatter) className(name meta.Key, options *Options) string {
	className := ""
	if options.PrefixClassName != "" {
		className = options.PrefixClassName + name.PascalCase() + options.SuffixClassName
	} else {
		className = name.String() + options.SuffixClassName
	}
	if options.ClassNameFormatter != nil {
		className = options.ClassNameFormatter(meta.Key(className))
	}

	return className
}
