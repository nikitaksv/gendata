/*
 * Copyright (c) 2022 Nikita Krasnikov
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package lexer

import "regexp"

var (
	LexNameAsIsCase = &Lexer{
		Token:   regexp.MustCompile(`(?i){{([\s]+)?Name([\s]+)?}}`),
		replace: []byte(`{{ .Key.String }}`),
	}
	LexNameCamelCase = &Lexer{
		Token:   regexp.MustCompile(`(?i){{([\s]+)?Name.CamelCase([\s]+)?}}`),
		replace: []byte(`{{ .Key.CamelCase }}`),
	}
	LexNamePascalCase = &Lexer{
		Token:   regexp.MustCompile(`(?i){{([\s]+)?Name.PascalCase([\s]+)?}}`),
		replace: []byte(`{{ .Key.PascalCase }}`),
	}
	LexNameSnakeCase = &Lexer{
		Token:   regexp.MustCompile(`(?i){{([\s]+)?Name.SnakeCase([\s]+)?}}`),
		replace: []byte(`{{ .Key.SnakeCase }}`),
	}
	LexNameKebabCase = &Lexer{
		Token:   regexp.MustCompile(`(?i){{([\s]+)?Name.KebabCase([\s]+)?}}`),
		replace: []byte(`{{ .Key.CamelCase }}`),
	}
	LexNameDotCase = &Lexer{
		Token:   regexp.MustCompile(`(?i){{([\s]+)?Name.DotCase([\s]+)?}}`),
		replace: []byte(`{{ .Key.DotCase }}`),
	}
	LexBeginProps = &Lexer{
		Token:   regexp.MustCompile(`(?i){{([\s]+)?Properties([\s]+)?}}`),
		replace: []byte(`{{- range .Properties }}`),
	}
	LexEndProps = &Lexer{
		Token:   regexp.MustCompile(`(?i){{([\s]+)?/Properties([\s]+)?}}`),
		replace: []byte(`{{- end }}`),
	}
	LexType = &Lexer{
		Token:   regexp.MustCompile(`(?i){{([\s]+)?Type([\s]+)?}}`),
		replace: []byte(`{{ .Type.String }}`),
	}
	LexTypeShort = &Lexer{
		Token:   regexp.MustCompile(`(?i){{([\s]+)?Type.Short([\s]+)?}}`),
		replace: []byte(`{{ .Type.Short }}`),
	}
	LexBeginTypeIsArray = &Lexer{
		Token:   regexp.MustCompile(`(?i){{([\s]+)?Type.IsArray([\s]+)?}}`),
		replace: []byte(`{{- if .Type.IsArray }}`),
	}
	LexEndTypeIsArray = &Lexer{
		Token:   regexp.MustCompile(`(?i){{([\s]+)?/Type.IsArray([\s]+)?}}`),
		replace: []byte(`{{- end }}`),
	}

	Lexers = []*Lexer{
		LexNameAsIsCase,
		LexNameCamelCase,
		LexNamePascalCase,
		LexNameSnakeCase,
		LexNameKebabCase,
		LexNameDotCase,
		LexBeginProps,
		LexEndProps,
		LexType,
		LexTypeShort,
		LexBeginTypeIsArray,
		LexEndTypeIsArray,
	}
	StartEndLexers = [][]*Lexer{
		{
			LexBeginProps,
			LexEndProps,
		},
		{
			LexBeginTypeIsArray,
			LexEndTypeIsArray,
		},
	}
)

type Lexer struct {
	Token   *regexp.Regexp
	replace []byte
}

func (l Lexer) Replace(in []byte) []byte {
	return l.Token.ReplaceAllLiteral(in, l.replace)
}
func (l Lexer) Lex(in []byte) [][]byte {
	var lex [][]byte
	for _, bs := range l.Token.FindSubmatch(in) {
		if len(bs) != 0 && string(bs) != " " {
			lex = append(lex, bs)
		}
	}
	return lex
}
