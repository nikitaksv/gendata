package lexer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLexer_Replace(t *testing.T) {
	in := []byte(`
LexNameAsIsCase -> {{ name }};{{Name}}
LexNameCamelCase -> {{ name.camelcase }};{{Name.CamelCase}}
LexNamePascalCase -> {{ name.pascalcase }};{{Name.PascalCase}}
LexNameSnakeCase -> {{ name.snakecase }};{{Name.SnakeCase}}
LexNameKebabCase -> {{ name.kebabcase }};{{Name.KebabCase}}
LexNameDotCase -> {{ name.dotcase }};{{Name.DotCase}}

LexBeginProps -> {{ properties }};{{Properties}}
LexEndProps -> {{ /properties }};{{/Properties}}
LexType -> {{ type }};{{Type}}
LexTypeShort -> {{ type.short }};{{Type.Short}}
LexBeginTypeIsArray -> {{ type.isarray }};{{Type.IsArray}}
LexEndTypeIsArray -> {{ /type.isarray }};{{/Type.IsArray}}
`)
	want := []byte(`
LexNameAsIsCase -> {{ .Key.String }};{{ .Key.String }}
LexNameCamelCase -> {{ .Key.CamelCase }};{{ .Key.CamelCase }}
LexNamePascalCase -> {{ .Key.PascalCase }};{{ .Key.PascalCase }}
LexNameSnakeCase -> {{ .Key.SnakeCase }};{{ .Key.SnakeCase }}
LexNameKebabCase -> {{ .Key.CamelCase }};{{ .Key.CamelCase }}
LexNameDotCase -> {{ .Key.DotCase }};{{ .Key.DotCase }}

LexBeginProps -> {{- range .Properties }};{{- range .Properties }}
LexEndProps -> {{- end }};{{- end }}
LexType -> {{ .Type.String }};{{ .Type.String }}
LexTypeShort -> {{ .Type.Short }};{{ .Type.Short }}
LexBeginTypeIsArray -> {{- if .Type.IsArray }};{{- if .Type.IsArray }}
LexEndTypeIsArray -> {{- end }};{{- end }}
`)

	var actual []byte
	for _, lexer := range Lexers {
		if len(actual) == 0 {
			actual = lexer.Replace(in)
		} else {
			actual = lexer.Replace(actual)
		}
	}

	fmt.Println(string(actual))
	assert.Equal(t, string(want), string(actual))
}

func TestLexer_Lex(t *testing.T) {
	in := []byte(`
{{ Properties }}
{{ /Properties }}
{{ Type.IsArray }}
{{ /Type.IsArray }}
`)

	for _, lexers := range StartEndLexers {
		var lexs []string
		for _, lex := range lexers {
			for _, bs := range lex.Lex(in) {
				lexs = append(lexs, string(bs))
			}
		}
		assert.Len(t, lexs, 2)
	}
}
