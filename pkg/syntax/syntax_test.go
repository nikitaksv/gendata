package syntax

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		in      []byte
		wantErr bool
	}{
		{
			name:    "valid Properties",
			wantErr: false,
			in:      []byte(`{{Properties}}{{Name.CamelCase}}{{/Properties}}`),
		},
		{
			name:    "valid type.IsArray",
			wantErr: false,
			in:      []byte(`{{Type.IsArray}}{{Type}}[]{{/Type.IsArray}}`),
		},
		{
			name:    "invalid Properties not close",
			wantErr: true,
			in:      []byte(`{{Properties}}{{Name.CamelCase}}`),
		},
		{
			name:    "invalid Properties not open",
			wantErr: true,
			in:      []byte(`{{Name.CamelCase}}{{/Properties}}`),
		},
		{
			name:    "invalid Type.IsArray not close",
			wantErr: true,
			in:      []byte(`{{Type.IsArray}}{{Type}}[]`),
		},
		{
			name:    "invalid Type.IsArray not open",
			wantErr: true,
			in:      []byte(`{{Type}}[]{{Type.IsArray}}`),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := Validate(test.in); err != nil && !test.wantErr {
				t.Error(err)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tmpl := []byte(`
package main

{{ SPLIT }}

type {{ Name.PascalCase }} struct {
{{ Properties }}
	{{ Name.PascalCase }} {{ Type.Doc }}
{{ /Properties }}
}
{{ /SPLIT }}
`)
	outExpected := []byte(`
package main

{{ SPLIT }}

type {{ .Key.PascalCase }} struct {
{{- range .Properties }}
	{{ .Key.PascalCase }} {{ .Type.Doc }}
{{- end }}
}
{{ /SPLIT }}
`)
	out, err := Parse(tmpl)
	assert.NoError(t, err)
	assert.Equal(t, string(outExpected), string(out))
}

func TestParse_Error(t *testing.T) {
	tmpl := []byte(`
package main

type {{ Name.PascalCase }} struct {
{{ Properties }}
	{{ Name.PascalCase }} {{ Type.Doc }}
}
`)
	_, err := Parse(tmpl)
	assert.Error(t, err)
}

func TestParseWithSplit(t *testing.T) {
	tmpl := []byte(`
package main

{{ SPLIT }}

type {{ Name.PascalCase }} struct {
{{ Properties }}
	{{ Name.PascalCase }} {{ Type.Doc }}
{{ /Properties }}
}
{{ /SPLIT }}
`)
	outExpected := []byte(`
package main


type {{ .Key.PascalCase }} struct {
{{- range .Properties }}
	{{ .Key.PascalCase }} {{ .Type.Doc }}
{{- end }}
}
`)
	splittedExpected := []byte(`
type {{ .Key.PascalCase }} struct {
{{- range .Properties }}
	{{ .Key.PascalCase }} {{ .Type.Doc }}
{{- end }}
}
`)
	out, splitted, err := ParseWithSplit(tmpl)
	assert.NoError(t, err)
	assert.Equal(t, string(outExpected), string(out))
	assert.Equal(t, string(splittedExpected), string(splitted))
}

func TestParseWithSplit_Error(t *testing.T) {
	tmpl := []byte(`{{ SPLIT }}`)

	_, _, err := ParseWithSplit(tmpl)
	assert.Error(t, err)

	tmpl = []byte(`
{{ SPLIT }}
{{ SPLIT }}
{{ /SPLIT }}
{{ /SPLIT }}
`)

	_, _, err = ParseWithSplit(tmpl)
	assert.Error(t, err)

	tmpl = []byte(`
{{ /SPLIT }}
{{ /SPLIT }}
{{ SPLIT }}
{{ SPLIT }}
`)

	_, _, err = ParseWithSplit(tmpl)
	assert.Error(t, err)

	tmpl = []byte(`
{{ /SPLIT }}
{{ SPLIT }}
{{ SPLIT }}
`)

	_, _, err = ParseWithSplit(tmpl)
	assert.Error(t, err)

	tmpl = []byte(`
{{ SPLIT }}
{{ SPLIT }}
{{ /SPLIT }}
`)

	_, _, err = ParseWithSplit(tmpl)
	assert.Error(t, err)
}

func Test_printLex(t *testing.T) {
	m := map[int]string{
		0: "qwerty",
		1: "qwerty",
	}
	assert.Equal(t, "qwerty:1", printLex(m))
	m = map[int]string{}
	assert.Equal(t, "", printLex(m))
}
