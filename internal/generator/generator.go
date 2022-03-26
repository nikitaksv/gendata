package generator

import (
	"bytes"
	"context"
	"html/template"
	"io"

	"github.com/nikitaksv/gendata/internal/generator/meta"
	"github.com/nikitaksv/gendata/internal/syntax"
	"github.com/pkg/errors"
)

type FileNameFormatter func(key meta.Key) string

type Option func(opts *Options) error

func WithOptions(opts *Options) Option {
	return func(opts_ *Options) error {
		opts_.FileExtension = opts.FileExtension
		opts_.SplitObjectsByFiles = opts.SplitObjectsByFiles
		opts_.FileNameFormatter = opts.FileNameFormatter
		return nil
	}
}

type Options struct {
	// Split nested object by separate template file
	SplitObjectsByFiles bool
	FileExtension       string
	FileNameFormatter   FileNameFormatter
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
	Generate(ctx context.Context, tmpl []byte, m *meta.Meta, opts ...Option) ([]*RenderedFile, error)
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

func (g *generator) Generate(_ context.Context, in []byte, m *meta.Meta, opts ...Option) ([]*RenderedFile, error) {
	in, splitted, err := syntax.ParseWithSplit(in)
	if err != nil {
		return nil, err
	}

	options := &Options{}
	if err := options.apply(opts...); err != nil {
		return nil, errors.Wrap(err, "incorrect generator option")
	}
	if options.FileNameFormatter == nil {
		options.FileNameFormatter = func(key meta.Key) string {
			return key.String()
		}
	}

	nests := nestSplit(m, options)

	renderedFiles := make([]*RenderedFile, 0, len(nests))
	for i, nest := range nests {
		var renderedFile *RenderedFile
		if i > 0 && !options.SplitObjectsByFiles {
			renderedFile = renderedFiles[0]
		} else {
			renderedFile = &RenderedFile{
				FileName: options.FileNameFormatter(nest.Key) + options.FileExtension,
				Content:  &bytes.Buffer{},
			}
		}

		parseData := string(in)
		if i > 0 && len(splitted) > 0 {
			parseData = string(splitted)
		}

		tmpl, err := template.New(renderedFile.FileName).Parse(parseData)
		if err != nil {
			return nil, errors.Wrap(err, "incorrect data template")
		}
		if err := tmpl.Execute(renderedFile.Content, nest); err != nil {
			return nil, errors.Wrap(err, "incorrect data template")
		}

		if i == 0 || options.SplitObjectsByFiles {
			renderedFiles = append(renderedFiles, renderedFile)
		}
	}

	return renderedFiles, nil
}

func nestSplit(m *meta.Meta, opt *Options) []*meta.Meta {
	nests := []*meta.Meta{
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
