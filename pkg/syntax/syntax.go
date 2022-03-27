package syntax

import (
	"fmt"

	"github.com/nikitaksv/gendata/pkg/lexer"
	"github.com/pkg/errors"
)

var ErrSyntax = errors.New("syntax error")

func Validate(in []byte) error {
	for _, lexers := range lexer.StartEndLexers {
		startLex := lexers[0].Lex(in)
		endLex := lexers[1].Lex(in)
		if len(startLex) > len(endLex) {
			return errors.Wrapf(ErrSyntax, "need close '%s' tag", printLex(startLex))
		}
		if len(endLex) > len(startLex) {
			return errors.Wrapf(ErrSyntax, "need open '%s' tag", printLex(endLex))
		}
	}

	startSplitLex := lexer.LexBeginSplit.Lex(in)
	if len(startSplitLex) > 1 {
		return errors.Wrapf(ErrSyntax, "have only one '%s' tag", printLex(startSplitLex))
	}
	endSplitLex := lexer.LexEndSplit.Lex(in)
	if len(endSplitLex) > 1 {
		return errors.Wrapf(ErrSyntax, "have only one '%s' tag", printLex(endSplitLex))
	}

	return nil
}

func Parse(in []byte) ([]byte, error) {
	for _, l := range lexer.Lexers {
		in = l.Replace(in)
	}
	if err := Validate(in); err != nil {
		return nil, err
	}

	return in, nil
}

func ParseWithSplit(in []byte) ([]byte, []byte, error) {
	in, err := Parse(in)
	if err != nil {
		return nil, nil, err
	}
	splitted := lexer.ExtractSplit(in)
	in = lexer.LexBeginSplit.Token.ReplaceAll(in, []byte{})
	in = lexer.LexEndSplit.Token.ReplaceAll(in, []byte{})
	return in, splitted, nil
}

func printLex(lex map[string]int) string {
	token := ""
	maxIdx := 0
	for t, i := range lex {
		if maxIdx == 0 || i >= maxIdx {
			token = t
			maxIdx = i
		}
	}

	return fmt.Sprintf("%s:%d", token, maxIdx)
}
