package formatter

import (
	"github.com/nikitaksv/gendata/pkg/meta"
	"github.com/pkg/errors"
)

type ClassNameFormatter func(key meta.Key) (string, error)

type Option func(opts *options) error

func WithTypeNameFormatter(formatter *meta.TypeFormatters) Option {
	return func(opts *options) error {
		opts.typeFormatters = formatter
		return nil
	}
}

func WithClassNameFormatter(formatter ClassNameFormatter) Option {
	return func(opts *options) error {
		opts.classNameFormatter = formatter
		return nil
	}
}

func WithRootClassName(name string) Option {
	return func(opts *options) error {
		opts.rootClassName = name
		return nil
	}
}

func WithPrefixClassName(name string) Option {
	return func(opts *options) error {
		opts.prefixClassName = name
		return nil
	}
}

func WithSuffixClassName(name string) Option {
	return func(opts *options) error {
		opts.suffixClassName = name
		return nil
	}
}

func WithSortProperties(sort bool) Option {
	return func(opts *options) error {
		opts.sortProperties = sort
		return nil
	}
}

type options struct {
	typeFormatters     *meta.TypeFormatters
	classNameFormatter ClassNameFormatter
	rootClassName      string
	prefixClassName    string
	suffixClassName    string
	sortProperties     bool
}

func (o *options) apply(opts ...Option) error {
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

func (f *formatter) Format(m *meta.Meta, opts ...Option) (*meta.Meta, error) {
	options := &options{}
	if err := options.apply(opts...); err != nil {
		return nil, err
	}

	if options.sortProperties {
		m.Sort()
	}

	if m.Key.String() == "" {
		m.Key = meta.Key(options.rootClassName)
	}

	key, err := f.className(m.Key, options)
	if err != nil {
		return nil, errors.WithMessagef(err, "can't format name on \"%s\" meta key", m.Key.String())
	}
	m.Key = meta.Key(key)
	m.Type.Key = m.Key
	m.Type.Formatters = options.typeFormatters
	for _, property := range m.Properties {
		property.Type.Formatters = options.typeFormatters

		if property.Type.IsObject() || property.Type.Value == meta.TypeArrayObject {
			key, err := f.className(property.Key, options)
			if err != nil {
				return nil, errors.WithMessagef(err, "can't format name on \"%s\" property", property.Key.String())
			}
			property.Key = meta.Key(key)
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

func (f *formatter) className(name meta.Key, options *options) (string, error) {
	className := ""
	if options.prefixClassName != "" {
		className = options.prefixClassName + name.String() + options.suffixClassName
	} else {
		className = name.String() + options.suffixClassName
	}
	if options.classNameFormatter != nil {
		var err error
		className, err = options.classNameFormatter(meta.Key(className))
		if err != nil {
			return "", err
		}
	}

	return className, nil
}
