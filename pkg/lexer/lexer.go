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

import (
	"regexp"
)

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
		replace: []byte(`{{ .Key.KebabCase }}`),
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
	LexTypeDoc = &Lexer{
		Token:   regexp.MustCompile(`(?i){{([\s]+)?Type.Doc([\s]+)?}}`),
		replace: []byte(`{{ .Type.Doc }}`),
	}
	LexBeginTypeIsArray = &Lexer{
		Token:   regexp.MustCompile(`(?i){{([\s]+)?Type.IsArray([\s]+)?}}`),
		replace: []byte(`{{- if .Type.IsArray }}`),
	}
	LexEndTypeIsArray = &Lexer{
		Token:   regexp.MustCompile(`(?i){{([\s]+)?/Type.IsArray([\s]+)?}}`),
		replace: []byte(`{{- end }}`),
	}
	LexBeginTypeIsObject = &Lexer{
		Token:   regexp.MustCompile(`(?i){{([\s]+)?Type.IsObject([\s]+)?}}`),
		replace: []byte(`{{- if .Type.IsObject }}`),
	}
	LexEndTypeIsObject = &Lexer{
		Token:   regexp.MustCompile(`(?i){{([\s]+)?/Type.IsObject([\s]+)?}}`),
		replace: []byte(`{{- end }}`),
	}
	LexBeginSplit = &Lexer{
		Token:   regexp.MustCompile(`(?i){{([\s]+)?SPLIT([\s]+)?}}(\n)?`),
		replace: []byte(""),
	}
	LexEndSplit = &Lexer{
		Token:   regexp.MustCompile(`(?i){{([\s]+)?/SPLIT([\s]+)?}}(\n)?`),
		replace: []byte(""),
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
		LexTypeDoc,
		LexBeginTypeIsArray,
		LexEndTypeIsArray,
		LexBeginTypeIsObject,
		LexEndTypeIsObject,
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
		{
			LexBeginTypeIsObject,
			LexEndTypeIsObject,
		},
		{
			LexBeginSplit,
			LexEndSplit,
		},
	}
)

type Lexer struct {
	Token   *regexp.Regexp
	replace []byte
}

func (l Lexer) Replace(in []byte) []byte {
	if l.replace != nil {
		return l.Token.ReplaceAllLiteral(in, l.replace)
	}
	return in
}

func (l Lexer) Lex(in []byte) map[string]int {
	if loc := l.Token.FindIndex(in); loc != nil {
		return map[string]int{
			string(in[loc[0]:loc[1]]): loc[1],
		}
	}
	return map[string]int{}
}

func ExtractSplit(in []byte) []byte {
	if len(LexBeginSplit.Lex(in)) == 0 {
		return in
	}
	if len(LexEndSplit.Lex(in)) == 0 {
		return in
	}
	beginSplitIdxs := LexBeginSplit.Token.FindIndex(in)
	endSplitIdxs := LexEndSplit.Token.FindIndex(in)
	if len(beginSplitIdxs) != 2 && len(endSplitIdxs) != 2 {
		return in
	}

	return in[beginSplitIdxs[1]:endSplitIdxs[0]]
}
