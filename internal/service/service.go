package service

import (
	"context"
	"encoding/json"
	"html/template"
	"os"
	"strings"
	"time"

	"github.com/nikitaksv/gendata/internal/generator"
	"github.com/nikitaksv/gendata/internal/generator/meta"
	"github.com/nikitaksv/gendata/internal/generator/parser"
	"github.com/nikitaksv/gendata/internal/syntax"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type GenRequest struct {
	Tmpl   []byte  `json:"tmpl"`
	Data   []byte  `json:"data"`
	Config *Config `json:"config"`
}

type GenFileRequest struct {
	TmplFile   string  `json:"tmplFile"`
	DataFile   string  `json:"dataFile"`
	ConfigFile string  `json:"configFile"`
	Config     *Config `json:"config"`
}

type PredefinedLangSettingsListRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type PredefinedLangSettingsListResponse struct {
	Items []*LangSettings `json:"items"`
}

type GenResponse struct {
	RenderedFiles []*generator.RenderedFile `json:"renderedFiles"`
	RenderTime    time.Duration             `json:"renderTime"`
}

type Config struct {
	Lang         string        `json:"lang"`
	LangSettings *LangSettings `json:"langSettings,omitempty"`
	DataFormat   string        `json:"dataFormat"`
	// Root object name
	RootClassName   string `json:"rootClassName"`
	PrefixClassName string `json:"prefixClassName"`
	SuffixClassName string `json:"suffixClassName"`
	// Sort object properties
	SortProperties bool `json:"sortProperties"`
}

type Service interface {
	Gen(ctx context.Context, request *GenRequest) (*GenResponse, error)
	GenFile(ctx context.Context, request *GenFileRequest) (*GenResponse, error)
	PredefinedLangSettings(ctx context.Context, request *PredefinedLangSettingsListRequest) (*PredefinedLangSettingsListResponse, error)
}

func NewService(log *zap.Logger) Service {
	return &service{log: log}
}

type service struct {
	log *zap.Logger
}

func (s *service) PredefinedLangSettings(_ context.Context, request *PredefinedLangSettingsListRequest) (*PredefinedLangSettingsListResponse, error) {
	items := make([]*LangSettings, 0, len(PredefinedLangSettings))
	for _, langSettings := range PredefinedLangSettings {
		if request.Code == "" || request.Name == "" || langSettings.Code == request.Code || langSettings.Name == request.Name {
			if langSettings.ConfigMapping.TypeDocMapping == nil {
				langSettings.ConfigMapping.TypeDocMapping = langSettings.ConfigMapping.TypeMapping
			}
			items = append(items, langSettings)
		}
	}
	return &PredefinedLangSettingsListResponse{
		Items: items,
	}, nil
}

func (s *service) Gen(ctx context.Context, request *GenRequest) (*GenResponse, error) {
	beginTs := time.Now()
	if len(request.Tmpl) == 0 {
		return nil, errors.New("template is empty")
	}
	if len(request.Data) == 0 {
		return nil, errors.New("data is empty")
	}

	var langSettings LangSettings
	if request.Config.Lang != "" {
		prefLangSettings, ok := getPredefinedLangSettings(request.Config.Lang)
		if !ok {
			return nil, errors.Errorf("config.lang \"%s\" not supported", request.Config.Lang)
		}
		langSettings = prefLangSettings
	} else if request.Config.LangSettings != nil {
		if request.Config.LangSettings.ConfigMapping == nil {
			return nil, errors.New("config.langSettings.configMapping is empty")
		}
		if request.Config.LangSettings.ConfigMapping.TypeMapping == nil {
			return nil, errors.New("config.langSettings.configMapping.typeMapping is empty")
		}
		langSettings = *request.Config.LangSettings
	} else {
		return nil, errors.Errorf("config.lang and config.langSettings is empty")
	}
	if request.Config.DataFormat == "" {
		return nil, errors.New("config.dataFormat is empty, allowed 'json'")
	}
	if request.Config.RootClassName == "" {
		request.Config.RootClassName = "RootClass"
	}

	if langSettings.ConfigMapping.TypeDocMapping == nil {
		langSettings.ConfigMapping.TypeDocMapping = langSettings.ConfigMapping.TypeMapping
	}

	var parser_ parser.Parser
	if request.Config.DataFormat == "json" {
		var err error
		parser_, err = parser.NewParserJSON(
			parser.WithOptions(&parser.Options{
				RootClassName:   request.Config.RootClassName,
				PrefixClassName: request.Config.PrefixClassName,
				SuffixClassName: request.Config.SuffixClassName,
				SortProperties:  request.Config.SortProperties,
				TypeFormatters: &meta.TypeFormatters{
					Type: langSettings.ConfigMapping.TypeMapping.TypeFormatters(),
					Doc:  langSettings.ConfigMapping.TypeDocMapping.TypeFormatters(),
				},
				ClassNameFormatter: langSettings.ConfigMapping.ClassNameFormatter(),
			}))
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("config.data_content_type is unknown, only is `json` supported")
	}

	m, err := parser_.Parse(request.Data)
	if err != nil {
		return nil, err
	}

	gen := generator.NewGenerator()
	renderedFiles, err := gen.Generate(ctx, request.Tmpl, m, generator.WithOptions(&generator.Options{
		SplitObjectsByFiles: langSettings.SplitObjectByFiles,
		FileExtension:       langSettings.FileExtension,
		FileNameFormatter:   langSettings.ConfigMapping.FileNameFormatter(),
	}))
	if err != nil {
		return nil, err
	}
	return &GenResponse{
		RenderedFiles: renderedFiles,
		RenderTime:    time.Since(beginTs),
	}, nil
}

func (s *service) GenFile(ctx context.Context, request *GenFileRequest) (*GenResponse, error) {
	if request.TmplFile == "" {
		return nil, errors.New("template file is empty")
	}
	if request.DataFile == "" {
		return nil, errors.New("data file is empty")
	}
	if request.Config == nil && request.ConfigFile == "" {
		return nil, errors.New("config file and config params is empty")
	}

	tmplBs, err := os.ReadFile(request.TmplFile)
	if err != nil {
		return nil, err
	}
	dataBs, err := os.ReadFile(request.DataFile)
	if err != nil {
		return nil, err
	}
	if request.ConfigFile != "" {
		cfgBs, err := os.ReadFile(request.ConfigFile)
		if err != nil {
			return nil, err
		}
		request.Config = &Config{}
		if err := json.Unmarshal(cfgBs, request.Config); err != nil {
			return nil, errors.Wrap(err, "incorrect config file")
		}
	}

	return s.Gen(ctx, &GenRequest{
		Tmpl:   tmplBs,
		Data:   dataBs,
		Config: request.Config,
	})
}

type LangSettings struct {
	Code               string         `json:"code"`
	Name               string         `json:"name"`
	FileExtension      string         `json:"fileExtension"`
	SplitObjectByFiles bool           `json:"splitObjectByFiles"`
	ConfigMapping      *ConfigMapping `json:"configMapping"`
}

var PredefinedLangSettings = []*LangSettings{
	{
		Code:               "go",
		Name:               "GoLang",
		FileExtension:      ".go",
		SplitObjectByFiles: false,
		ConfigMapping: &ConfigMapping{
			TypeMapping: &TypeMapping{
				Array:       "[]interface{}",
				ArrayBool:   "[]bool",
				ArrayFloat:  "[]float64",
				ArrayInt:    "[]int",
				ArrayObject: "[]*{{ Name.PascalCase }}",
				ArrayString: "[]string",
				Bool:        "bool",
				Float:       "float64",
				Int:         "int",
				Null:        "interface{}",
				Object:      "*{{ Name.PascalCase}}",
				String:      "string",
			},
			TypeDocMapping:   nil,
			ClassNameMapping: "{{ Name.PascalCase }}",
			FileNameMapping:  "{{ Name.CamelCase }}",
		},
	},
	{
		Code:               "php",
		Name:               "PHP",
		FileExtension:      ".php",
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
				Float:       "float64",
				Int:         "int",
				Null:        "null",
				Object:      "{{ Name.PascalCase}}",
				String:      "string",
			},
			TypeDocMapping: &TypeMapping{
				Array:       "array",
				ArrayBool:   "bool[]",
				ArrayFloat:  "float[]",
				ArrayInt:    "int[]",
				ArrayObject: "{{ Name.PascalCase }}[]",
				ArrayString: "string[]",
				Bool:        "bool",
				Float:       "float64",
				Int:         "int",
				Null:        "null",
				Object:      "{{ Name.PascalCase}}",
				String:      "string",
			},
			ClassNameMapping: "{{ Name.PascalCase }}",
			FileNameMapping:  "{{ Name.PascalCase }}",
		},
	},
}

type ConfigMapping struct {
	TypeMapping      *TypeMapping `json:"typeMapping"`
	TypeDocMapping   *TypeMapping `json:"typeDocMapping"`
	ClassNameMapping string       `json:"classNameMapping"`
	FileNameMapping  string       `json:"fileNameMapping"`
}

func (m ConfigMapping) ClassNameFormatter() parser.ClassNameFormatter {
	return func(key meta.Key) string {
		bs, err := syntax.Parse([]byte(m.ClassNameMapping))
		if err != nil {
			return errors.WithMessage(err, "ClassNameFormatter syntax error").Error()
		}
		tmpl, err := template.New("").Parse(string(bs))
		if err != nil {
			return errors.WithMessage(err, "ClassNameFormatter template parse error").Error()
		}
		b := &strings.Builder{}
		if err := tmpl.Execute(b, struct {
			Key meta.Key
		}{key}); err != nil {
			return errors.WithMessage(err, "ClassNameFormatter template execute error").Error()
		}
		return b.String()
	}
}

func (m ConfigMapping) FileNameFormatter() generator.FileNameFormatter {
	return func(key meta.Key) string {
		bs, err := syntax.Parse([]byte(m.FileNameMapping))
		if err != nil {
			return errors.WithMessage(err, "ClassNameFormatter syntax error").Error()
		}
		tmpl, err := template.New("").Parse(string(bs))
		if err != nil {
			return errors.WithMessage(err, "ClassNameFormatter template parse error").Error()
		}
		b := &strings.Builder{}
		if err := tmpl.Execute(b, struct {
			Key meta.Key
		}{key}); err != nil {
			return errors.WithMessage(err, "ClassNameFormatter template execute error").Error()
		}
		return b.String()
	}
}

type TypeMapping struct {
	Array       string `json:"array"`
	ArrayBool   string `json:"arrayBool"`
	ArrayFloat  string `json:"arrayFloat"`
	ArrayInt    string `json:"arrayInt"`
	ArrayObject string `json:"arrayObject"`
	ArrayString string `json:"arrayString"`
	Bool        string `json:"bool"`
	Float       string `json:"float"`
	Int         string `json:"int"`
	Null        string `json:"null"`
	Object      string `json:"object"`
	String      string `json:"string"`
}

func (m TypeMapping) GetType(key string) (string, error) {
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
		bs, err := syntax.Parse([]byte(typ))
		if err != nil {
			return errors.WithMessage(err, "TypeFormatter syntax error").Error()
		}
		tmpl, err := template.New("").Parse(string(bs))
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

func getPredefinedLangSettings(lang string) (LangSettings, bool) {
	for _, langSettings := range PredefinedLangSettings {
		if langSettings.Code == lang || langSettings.Name == lang {
			return *langSettings, true
		}
	}
	return LangSettings{}, false
}
