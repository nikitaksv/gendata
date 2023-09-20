package parser

import "github.com/nikitaksv/gendata/pkg/meta"

type Parser interface {
	Parse(data []byte, opts ...Option) (*meta.Meta, error)
}

type Option func(opts *options) error

type options struct{}
