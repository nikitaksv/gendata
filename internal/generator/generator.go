package generator

import (
	"context"
	"html/template"
	"io"
	"os"

	"github.com/nikitaksv/gendata/internal/generator/meta"
	"github.com/nikitaksv/gendata/internal/lexer"
	"github.com/nikitaksv/gendata/internal/syntax"
	"github.com/pkg/errors"
)

type Option func(opts *Options) error

func WithRootClassName(rootClassName string) Option {
	return func(opts *Options) error {
		opts.RootClassName = rootClassName
		return nil
	}
}
func WithPrefixClassName(prefix string) Option {
	return func(opts *Options) error {
		opts.PrefixClassName = prefix
		return nil
	}
}
func WithSuffixClassName(suffix string) Option {
	return func(opts *Options) error {
		opts.SuffixClassName = suffix
		return nil
	}
}
func WithSortProperties(sort bool) Option {
	return func(opts *Options) error {
		opts.SortProperties = sort
		return nil
	}
}
func WithSplitObjectsByFiles(split bool) Option {
	return func(opts *Options) error {
		opts.SplitObjectsByFiles = split
		return nil
	}
}

type Options struct {
	// Root object name
	RootClassName   string
	PrefixClassName string
	SuffixClassName string
	// Sort object properties
	SortProperties bool
	// Split nested object by separate template file
	SplitObjectsByFiles bool
}

func (o *Options) apply(opts ...Option) error {
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return err
		}
	}
	return nil
}

type Generator interface {
	Generate(ctx context.Context, tmpl []byte, m *meta.Nest, opts ...Option) ([]*RenderedFile, error)
}

type RenderedFile struct {
	FileName string
	Content  io.ReadWriteSeeker
}

type generator struct {
}

func NewGenerator() Generator {
	return &generator{}
}

func (g *generator) Generate(ctx context.Context, in []byte, m *meta.Nest, opts ...Option) ([]*RenderedFile, error) {
	for _, l := range lexer.Lexers {
		in = l.Replace(in)
	}
	if err := syntax.Validate(in); err != nil {
		return nil, err
	}

	options := &Options{
		RootClassName:       "RootClass",
		SortProperties:      false,
		SplitObjectsByFiles: true,
	}
	if err := options.apply(opts...); err != nil {
		return nil, errors.Wrap(err, "incorrect generator option")
	}

	nests := nestSplit(m, options)
	nests[0].Key = meta.Key(options.PrefixClassName + options.RootClassName + options.SuffixClassName)

	renderedFiles := make([]*RenderedFile, 0, len(nests))
	for _, nest := range nests {
		tmpl, err := template.New(nest.Key.String()).Parse(string(in))
		if err != nil {
			return nil, errors.Wrap(err, "incorrect data template")
		}
		if err := tmpl.Execute(os.Stdout, nest); err != nil {
			return nil, errors.Wrap(err, "incorrect data template")
		}
	}

	return renderedFiles, nil
}

func nestSplit(m *meta.Nest, opt *Options) []*meta.Nest {
	nests := []*meta.Nest{
		{
			Key:        meta.Key(opt.PrefixClassName + m.Key.String() + opt.SuffixClassName),
			Type:       m.Type,
			Properties: make([]*meta.Property, 0, len(m.Properties)),
		},
	}
	for _, property := range m.Properties {
		nests[0].Properties = append(nests[0].Properties, property)
		if property.Nest != nil {
			nests = append(nests, nestSplit(property.Nest, opt)...)
		}
	}
	return nests
}
