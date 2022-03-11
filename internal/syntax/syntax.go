package syntax

import (
	"github.com/pkg/errors"

	"github.com/nikitaksv/gendata/internal/lexer"
)

var ErrSyntax = errors.New("syntax error")

func Validate(in []byte) error {
	for _, lexers := range lexer.StartEndLexers {
		startLex := lexers[0].Lex(in)
		endLex := lexers[1].Lex(in)
		if len(startLex) > len(endLex) {
			return errors.Wrapf(ErrSyntax, "need close '%s' tag", string(startLex[len(startLex)-1]))
		}
		if len(endLex) > len(startLex) {
			return errors.Wrapf(ErrSyntax, "need open '%s' tag", string(endLex[0]))
		}
	}

	return nil
}
