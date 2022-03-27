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
LexTypeDoc -> {{ type.doc }};{{Type.Doc}}
LexBeginTypeIsArray -> {{ type.isarray }};{{Type.IsArray}}
LexEndTypeIsArray -> {{ /type.isarray }};{{/Type.IsArray}}

LexBeginTypeIsObject -> {{ type.isobject }};{{Type.IsObject}}
LexEndTypeIsObject -> {{ /type.isobject }};{{/Type.IsObject}}
`)
	want := []byte(`
LexNameAsIsCase -> {{ .Key.String }};{{ .Key.String }}
LexNameCamelCase -> {{ .Key.CamelCase }};{{ .Key.CamelCase }}
LexNamePascalCase -> {{ .Key.PascalCase }};{{ .Key.PascalCase }}
LexNameSnakeCase -> {{ .Key.SnakeCase }};{{ .Key.SnakeCase }}
LexNameKebabCase -> {{ .Key.KebabCase }};{{ .Key.KebabCase }}
LexNameDotCase -> {{ .Key.DotCase }};{{ .Key.DotCase }}

LexBeginProps -> {{- range .Properties }};{{- range .Properties }}
LexEndProps -> {{- end }};{{- end }}
LexType -> {{ .Type.String }};{{ .Type.String }}
LexTypeDoc -> {{ .Type.Doc }};{{ .Type.Doc }}
LexBeginTypeIsArray -> {{- if .Type.IsArray }};{{- if .Type.IsArray }}
LexEndTypeIsArray -> {{- end }};{{- end }}

LexBeginTypeIsObject -> {{- if .Type.IsObject }};{{- if .Type.IsObject }}
LexEndTypeIsObject -> {{- end }};{{- end }}
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
{{ Type.IsObject }}
{{ /Type.IsObject }}
{{ SPLIT }}
{{ /SPLIT }}
`)

	for _, lexers := range StartEndLexers {
		var lexs []string
		for _, lex := range lexers {
			for _, token := range lex.Lex(in) {
				lexs = append(lexs, token)
			}
		}
		assert.Len(t, lexs, 2)
	}
}

func TestExtractSplit(t *testing.T) {
	in := []byte(`
package main

{{ SPLIT }}

type {{ Name.PascalCase }} struct {
	{{ Properties }}
	{{ Name.PascalCase }} {{ Type }}
	{{ /Properties }}
}
{{ /SPLIT }}
`)

	expected := []byte(`
type {{ Name.PascalCase }} struct {
	{{ Properties }}
	{{ Name.PascalCase }} {{ Type }}
	{{ /Properties }}
}
`)
	actual := ExtractSplit(in)
	assert.Equal(t, string(expected), string(actual))
}
