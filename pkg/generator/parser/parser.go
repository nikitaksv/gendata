package parser

import (
	"github.com/nikitaksv/gendata/pkg/generator/meta"
)

type Option func(opts *Options) error

type Options struct{}

type Parser interface {
	Parse(data []byte, opts ...Option) (*meta.Meta, error)
}
