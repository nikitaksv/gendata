package gen

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/nikitaksv/gendata/pkg/formatter"
	"github.com/nikitaksv/gendata/pkg/meta"
	parser2 "github.com/nikitaksv/gendata/pkg/parser"
	"github.com/pkg/errors"
)

type Gen interface {
	Gen(ctx context.Context, params *Params) (*RenderResult, error)
}

type Params struct {
	LangSettings []*LangSettings `json:"langSettings,omitempty" xml:"LangSettings" yaml:"langSettings"`
	// Root object name
	RootClassName   string `json:"rootClassName" xml:"RootClassName" yaml:"rootClassName"`
	PrefixClassName string `json:"prefixClassName" xml:"PrefixClassName" yaml:"prefixClassName"`
	SuffixClassName string `json:"suffixClassName" xml:"SuffixClassName" yaml:"suffixClassName"`
	// Sort object properties
	SortProperties bool    `json:"sortProperties" xml:"SortProperties" yaml:"sortProperties"`
	Templates      []*File `json:"templates"`
	Data           *File   `json:"data"`
}

type File struct {
	Name string        `json:"name"`
	Body io.ReadWriter `json:"body"`
}

type RenderResult struct {
	RenderedFiles []*File       `json:"renderedFiles"`
	RenderTime    time.Duration `json:"renderTime"`
}

func NewGen() Gen {
	return &gen{}
}

type gen struct {
}

//nolint:gocyclo
func (_ *gen) Gen(_ context.Context, params *Params) (*RenderResult, error) {
	beginTs := time.Now()

	if len(params.Templates) == 0 {
		return nil, errors.New("templates is empty")
	}

	if params.Data.Body == nil {
		return nil, errors.Errorf("data \"%s\" is empty", params.Data.Name)
	}

	var parser_ parser2.Parser
	if filepath.Ext(params.Data.Name) == ".json" {
		var err error
		parser_, err = parser2.NewParserJSON()
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("dataFormat is unknown, only is `json` supported")
	}

	dataBody := &bytes.Buffer{}
	if _, err := io.Copy(dataBody, params.Data.Body); err != nil {
		return nil, errors.Errorf("can't read data body \"%s\"", params.Data.Name)
	}
	dataBodyBs := dataBody.Bytes()
	if len(dataBodyBs) == 0 {
		return nil, errors.Errorf("data \"%s\" is empty", params.Data.Name)
	}

	_meta, err := parser_.Parse(dataBodyBs)
	if err != nil {
		return nil, errors.WithMessagef(err, "error parsing data file \"%s\"", params.Data.Name)
	}

	langSettings := append([]*LangSettings{}, PredefinedLangSettings...)
	for _, setting := range params.LangSettings {
		if setting.Code == "" {
			return nil, errors.New("langSetting.code is required")
		}
		if len(setting.FileExtensions) == 0 {
			return nil, errors.New("langSetting.fileExtension is required")
		}
		if setting.ConfigMapping == nil {
			return nil, errors.New("langSetting.configMapping is required")
		}
		if setting.ConfigMapping.TypeMapping == nil {
			return nil, errors.New("langSetting.configMapping.typeMapping is required")
		}
		langSettings = append(langSettings, setting)
	}

	const tmplExt = ".tmpl"

	// lang index => template indexes
	templateLang := make(map[int][]int)
LOOP:
	for tmplIdx, file := range params.Templates {
		parts := strings.Split(file.Name, ".")
		if len(parts) == 1 {
			return nil, errors.Errorf("template name \"%s\" is not have file extension", file.Name)
		}

		for _, part := range parts[len(parts)-2:] {
			for langIdx, setting := range langSettings {
				for _, ext := range setting.FileExtensions {
					if part == ext {
						if idxs, ok := templateLang[langIdx]; ok {
							templateLang[langIdx] = append(idxs, tmplIdx)
						} else {
							templateLang[langIdx] = []int{tmplIdx}
						}
						continue LOOP
					}
				}
			}
		}
		// common language
		if idxs, ok := templateLang[0]; ok {
			templateLang[0] = append(idxs, tmplIdx)
		} else {
			templateLang[0] = []int{tmplIdx}
		}
	}

	if params.RootClassName == "" {
		name := params.Data.Name
		if name != "" {
			params.RootClassName = name[:len(name)-len(filepath.Ext(name))]
		}
	}
	if params.RootClassName == "" {
		params.RootClassName = "Root"
	}

	renderedFiles := make([]*File, 0, len(params.Templates))
	_formatter := formatter.NewFormatter()
	for langIdx, tmplIdxs := range templateLang {
		lang := langSettings[langIdx]

		formattedMeta, err := _formatter.Format(_meta.Clone(),
			formatter.WithPrefixClassName(params.PrefixClassName),
			formatter.WithSuffixClassName(params.SuffixClassName),
			formatter.WithRootClassName(params.RootClassName),
			formatter.WithSortProperties(params.SortProperties),

			formatter.WithClassNameFormatter(lang.ConfigMapping.ClassNameFormatter()),
			formatter.WithTypeNameFormatter(&meta.TypeFormatters{
				Type: lang.ConfigMapping.TypeMapping.TypeFormatters(),
				Doc:  lang.ConfigMapping.TypeDocMapping.TypeFormatters(),
			}),
		)
		if err != nil {
			return nil, errors.WithMessagef(err, "formatter error in lang template \"%s\"", lang.Name)
		}

		for _, idx := range tmplIdxs {
			tmpl := params.Templates[idx]
			tmplBody := &bytes.Buffer{}
			if _, err := io.Copy(tmplBody, tmpl.Body); err != nil {
				return nil, errors.Errorf("can't read template body \"%s\"", tmpl.Name)
			}
			tmplBodyBs := tmplBody.Bytes()
			if len(tmplBodyBs) == 0 {
				return nil, errors.Errorf("template body \"%s\" is empty", tmpl.Name)
			}

			outName := strings.TrimRight(tmpl.Name, tmplExt)
			if strings.Contains(outName, "{{") {
				body := bytes.NewBuffer(nil)
				t, err := template.New("").Parse(outName)
				if err != nil {
					return nil, errors.WithMessagef(err, "incorrect template name \"%s\"", tmpl.Name)
				}
				if err := t.Execute(body, formattedMeta); err != nil {
					return nil, errors.WithMessagef(err, "incorrect template name \"%s\"", tmpl.Name)
				}
				outName = body.String()
			}

			body := bytes.NewBuffer(nil)
			t, err := template.New(tmpl.Name).Parse(string(tmplBodyBs))
			if err != nil {
				return nil, errors.WithMessagef(err, "incorrect template \"%s\"", tmpl.Name)
			}
			if err := t.Execute(body, formattedMeta); err != nil {
				return nil, errors.WithMessagef(err, "incorrect template \"%s\"", tmpl.Name)
			}

			renderedFiles = append(renderedFiles, &File{
				Name: outName,
				Body: body,
			})
		}
	}

	return &RenderResult{
		RenderedFiles: renderedFiles,
		RenderTime:    time.Since(beginTs),
	}, nil
}

type LangSettings struct {
	ConfigMapping      *ConfigMapping `json:"configMapping"  yaml:"configMapping" xml:"ConfigMapping"`
	Code               string         `json:"code" yaml:"code" xml:"Code"`
	Name               string         `json:"name" yaml:"name" xml:"Name"`
	FileExtensions     []string       `json:"fileExtensions" yaml:"fileExtensions" xml:"FileExtensions"`
	SplitObjectByFiles bool           `json:"splitObjectByFiles" yaml:"splitObjectByFiles" xml:"SplitObjectByFiles"`
}

var PredefinedLangSettings = []*LangSettings{
	{
		Code:               "common",
		Name:               "Common",
		FileExtensions:     []string{"*"},
		SplitObjectByFiles: false,
		ConfigMapping: &ConfigMapping{
			TypeMapping: &TypeMapping{
				Array:       "[]",
				ArrayBool:   "[]",
				ArrayFloat:  "[]",
				ArrayInt:    "[]",
				ArrayObject: "[]",
				ArrayString: "[]",
				Bool:        "bool",
				Float:       "float",
				Int:         "int",
				Null:        "null",
				Object:      "{{ .Key.CamelCase}}",
				String:      "string",
				Time:        "time",
				Date:        "date",
				DateTime:    "datetime",
				Duration:    "duration",
			},
			TypeDocMapping:   nil,
			ClassNameMapping: "{{ .Key.PascalCase }}",
		},
	},
	{
		Code:               "go",
		Name:               "GoLang",
		FileExtensions:     []string{"go"},
		SplitObjectByFiles: false,
		ConfigMapping: &ConfigMapping{
			TypeMapping: &TypeMapping{
				Array:       "[]interface{}",
				ArrayBool:   "[]bool",
				ArrayFloat:  "[]float64",
				ArrayInt:    "[]int",
				ArrayObject: "[]*{{ .Key.PascalCase }}",
				ArrayString: "[]string",
				Bool:        "bool",
				Float:       "float64",
				Int:         "int",
				Null:        "interface{}",
				Object:      "{{ .Key.PascalCase}}",
				String:      "string",
				Time:        "time.Time",
				Date:        "time.Time",
				DateTime:    "time.Time",
				Duration:    "time.Duration",
			},
			TypeDocMapping:   nil,
			ClassNameMapping: "{{ .Key.PascalCase }}",
		},
	},
	{
		Code:               "go1.20",
		Name:               "GoLang > v1.20",
		FileExtensions:     []string{"go"},
		SplitObjectByFiles: false,
		ConfigMapping: &ConfigMapping{
			TypeMapping: &TypeMapping{
				Array:       "[]any",
				ArrayBool:   "[]bool",
				ArrayFloat:  "[]float64",
				ArrayInt:    "[]int",
				ArrayObject: "[]*{{ .Key.PascalCase }}",
				ArrayString: "[]string",
				Bool:        "bool",
				Float:       "float64",
				Int:         "int",
				Null:        "any",
				Object:      "{{ .Key.PascalCase}}",
				String:      "string",
				Time:        "time.Time",
				Date:        "time.Time",
				DateTime:    "time.Time",
				Duration:    "time.Duration",
			},
			TypeDocMapping:   nil,
			ClassNameMapping: "{{ .Key.PascalCase }}",
		},
	},
	{
		Code:               "php",
		Name:               "PHP",
		FileExtensions:     []string{"php"},
		SplitObjectByFiles: true,
		ConfigMapping: &ConfigMapping{
			TypeMapping: &TypeMapping{
				Array:       "array",
				ArrayBool:   "array",
				ArrayFloat:  "array",
				ArrayInt:    "array",
				ArrayObject: "array",
				ArrayString: "array",
				Bool:        "bool",
				Float:       "float",
				Int:         "int",
				Null:        "null",
				Object:      "{{ .Key.PascalCase}}",
				String:      "string",
				Time:        "\\DateTime",
				Date:        "\\DateTime",
				DateTime:    "\\DateTime",
				Duration:    "\\DateInterval",
			},
			TypeDocMapping: &TypeMapping{
				Array:       "array",
				ArrayBool:   "bool[]",
				ArrayFloat:  "float[]",
				ArrayInt:    "int[]",
				ArrayObject: "{{ .Key.PascalCase }}[]",
				ArrayString: "string[]",
				Bool:        "bool",
				Float:       "float",
				Int:         "int",
				Null:        "null",
				Object:      "{{ .Key.PascalCase}}",
				String:      "string",
				Time:        "\\DateTime",
				Date:        "\\DateTime",
				DateTime:    "\\DateTime",
				Duration:    "\\DateInterval",
			},
			ClassNameMapping: "{{ .Key.PascalCase }}",
		},
	},
}

type ConfigMapping struct {
	TypeMapping      *TypeMapping `json:"typeMapping" xml:"TypeMapping" yaml:"typeMapping"`
	TypeDocMapping   *TypeMapping `json:"typeDocMapping" xml:"TypeDocMapping" yaml:"typeDocMapping"`
	ClassNameMapping string       `json:"classNameMapping" xml:"ClassNameMapping" yaml:"classNameMapping"`
	FileNameMapping  string       `json:"fileNameMapping" xml:"FileNameMapping" yaml:"fileNameMapping"`
}

func (m ConfigMapping) ClassNameFormatter() formatter.ClassNameFormatter {
	return func(key meta.Key) (string, error) {
		tmpl, err := template.New("").Parse(m.ClassNameMapping)
		if err != nil {
			return "", errors.WithMessage(err, "ClassNameFormatter template parse error")
		}
		b := &strings.Builder{}
		if err := tmpl.Execute(b, struct {
			Key meta.Key
		}{key}); err != nil {
			return "", errors.WithMessage(err, "ClassNameFormatter template execute error")
		}
		return b.String(), nil
	}
}

type TypeMapping struct {
	Array       string `json:"array" yaml:"array" xml:"Array"`
	ArrayBool   string `json:"arrayBool" yaml:"arrayBool" xml:"ArrayBool"`
	ArrayFloat  string `json:"arrayFloat" yaml:"arrayFloat" xml:"ArrayFloat"`
	ArrayInt    string `json:"arrayInt" yaml:"arrayInt" xml:"ArrayInt"`
	ArrayObject string `json:"arrayObject" yaml:"arrayObject" xml:"ArrayObject"`
	ArrayString string `json:"arrayString" yaml:"arrayString" xml:"ArrayString"`
	Bool        string `json:"bool" yaml:"bool" xml:"Bool"`
	Float       string `json:"float" yaml:"float" xml:"Float"`
	Int         string `json:"int" yaml:"int" xml:"Int"`
	Null        string `json:"null" yaml:"null" xml:"Null"`
	Object      string `json:"object" yaml:"object" xml:"Object"`
	String      string `json:"string" yaml:"string" xml:"String"`
	Time        string `json:"time" yaml:"time" xml:"Time"`
	Date        string `json:"date" yaml:"date" xml:"Date"`
	DateTime    string `json:"dateTime" yaml:"dateTime" xml:"DateTime"`
	Duration    string `json:"duration" yaml:"duration" xml:"Duration"`
}

func (m *TypeMapping) GetType(key string) (string, error) {
	switch key {
	case meta.TypeArray:
		return m.Array, nil
	case meta.TypeArrayBool:
		return m.ArrayBool, nil
	case meta.TypeArrayFloat:
		return m.ArrayFloat, nil
	case meta.TypeArrayInt:
		return m.ArrayInt, nil
	case meta.TypeArrayObject:
		return m.ArrayObject, nil
	case meta.TypeArrayString:
		return m.ArrayString, nil
	case meta.TypeBool:
		return m.Bool, nil
	case meta.TypeFloat:
		return m.Float, nil
	case meta.TypeInt:
		return m.Int, nil
	case meta.TypeNull:
		return m.Null, nil
	case meta.TypeObject:
		return m.Object, nil
	case meta.TypeString:
		return m.String, nil
	case meta.TypeTime:
		return m.Time, nil
	case meta.TypeDate:
		return m.Date, nil
	case meta.TypeDateTime:
		return m.DateTime, nil
	case meta.TypeDuration:
		return m.Duration, nil
	}
	return "", errors.Errorf("invalid TypeMapping key %s", key)
}

func (m *TypeMapping) TypeFormatters() meta.TypeFormatter {
	if m == nil {
		return nil
	}
	return func(t meta.Type) string {
		typ, err := m.GetType(t.Value)
		if err != nil {
			return errors.WithMessage(err, "TypeFormatter error").Error()
		}
		tmpl, err := template.New("").Parse(typ)
		if err != nil {
			return errors.WithMessage(err, "TypeFormatter template parse error").Error()
		}
		b := &strings.Builder{}
		if err := tmpl.Execute(b, t); err != nil {
			return errors.WithMessage(err, "TypeFormatter template execute error").Error()
		}
		return b.String()
	}
}
