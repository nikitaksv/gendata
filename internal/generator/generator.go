package generator

import (
	"bytes"
	"context"
	"github.com/nikitaksv/gendata/internal/generator/meta"
	"github.com/nikitaksv/gendata/internal/lexer"
	"github.com/nikitaksv/gendata/internal/syntax"
	"github.com/pkg/errors"
	"html/template"
	"io"
)

type Option func(opts *Options) error

func WithOptions(opts *Options) Option {
	return func(opts_ *Options) error {
		opts_.FileExtension = opts.FileExtension
		opts_.SplitObjectsByFiles = opts.SplitObjectsByFiles
		return nil
	}
}

type Options struct {
	// Split nested object by separate template file
	SplitObjectsByFiles bool   `json:"split_objects_by_files"`
	FileExtension       string `json:"-"`
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
	Content  io.ReadWriter
}

type generator struct {
}

func NewGenerator() Generator {
	return &generator{}
}

func (g *generator) Generate(_ context.Context, in []byte, m *meta.Nest, opts ...Option) ([]*RenderedFile, error) {
	for _, l := range lexer.Lexers {
		in = l.Replace(in)
	}
	if err := syntax.Validate(in); err != nil {
		return nil, err
	}

	options := &Options{}
	if err := options.apply(opts...); err != nil {
		return nil, errors.Wrap(err, "incorrect generator option")
	}

	nests := []*meta.Nest{m}
	if options.SplitObjectsByFiles == true {
		nests = nestSplit(m, options)
	}

	renderedFiles := make([]*RenderedFile, 0, len(nests))
	for _, nest := range nests {
		renderedFile := &RenderedFile{FileName: nest.Key.String() + options.FileExtension, Content: &bytes.Buffer{}}
		tmpl, err := template.New(renderedFile.FileName).Parse(string(in))
		if err != nil {
			return nil, errors.Wrap(err, "incorrect data template")
		}
		if err := tmpl.Execute(renderedFile.Content, nest); err != nil {
			return nil, errors.Wrap(err, "incorrect data template")
		}
		renderedFiles = append(renderedFiles, renderedFile)
	}

	return renderedFiles, nil
}

func nestSplit(m *meta.Nest, opt *Options) []*meta.Nest {
	nests := []*meta.Nest{
		{
			Key:        m.Key,
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
