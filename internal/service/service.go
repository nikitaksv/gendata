package service

import (
	"context"
	"github.com/nikitaksv/gendata/internal/generator"
	"github.com/nikitaksv/gendata/internal/generator/meta"
	"github.com/nikitaksv/gendata/internal/generator/parser"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"os"
	"time"
)

type GenRequest struct {
	Tmpl   []byte  `json:"tmpl"`
	Data   []byte  `json:"data"`
	Config *Config `json:"config"`
}

type GenFileRequest struct {
	TmplFile string  `json:"tmpl_file"`
	DataFile string  `json:"data_file"`
	Config   *Config `json:"config"`
}

type GenResponse struct {
	RenderedFiles []*generator.RenderedFile `json:"rendered_files"`
	RenderTime    time.Duration             `json:"render_time"`
}

type Config struct {
	Lang       string `json:"lang"`
	DataFormat string `json:"data_format"`
	// Root object name
	RootClassName   string `json:"root_class_name"`
	PrefixClassName string `json:"prefix_class_name"`
	SuffixClassName string `json:"suffix_class_name"`
	// Sort object properties
	SortProperties bool `json:"sort_properties"`
}

type Service interface {
	Gen(ctx context.Context, request *GenRequest) (*GenResponse, error)
	GenFile(ctx context.Context, request *GenFileRequest) (*GenResponse, error)
}

type service struct {
	log *zap.Logger
}

func NewService(log *zap.Logger) Service {
	return &service{log: log}
}

func (s *service) Gen(ctx context.Context, request *GenRequest) (*GenResponse, error) {
	beginTs := time.Now()
	if len(request.Tmpl) == 0 {
		return nil, errors.New("template is empty")
	}
	if len(request.Data) == 0 {
		return nil, errors.New("data is empty")
	}
	if request.Config.Lang == "" {
		return nil, errors.Errorf("config.lang is empty")
	}
	if _, ok := langSettings[request.Config.Lang]; !ok {
		return nil, errors.Errorf("config.lang \"%s\" not supported", request.Config.Lang)
	}
	if request.Config.DataFormat == "" {
		return nil, errors.New("config.data_format is empty")
	}
	if request.Config.RootClassName == "" {
		request.Config.RootClassName = "RootClass"
	}

	var parser_ parser.Parser
	if request.Config.DataFormat == "json" {
		var err error
		parser_, err = parser.NewParserJSON(
			parser.WithOptions(&parser.Options{
				RootClassName:      request.Config.RootClassName,
				PrefixClassName:    request.Config.PrefixClassName,
				SuffixClassName:    request.Config.SuffixClassName,
				SortProperties:     request.Config.SortProperties,
				TypeFormatters:     langSettings[request.Config.Lang].Formatters,
				ClassNameFormatter: langSettings[request.Config.Lang].ClassNameFormatter,
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
		SplitObjectsByFiles: langSettings[request.Config.Lang].SplitObjectByFiles,
		FileExtension:       langSettings[request.Config.Lang].FileExtension,
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
	tmplBs, err := os.ReadFile(request.TmplFile)
	if err != nil {
		return nil, err
	}
	dataBs, err := os.ReadFile(request.DataFile)
	if err != nil {
		return nil, err
	}

	return s.Gen(ctx, &GenRequest{
		Tmpl:   tmplBs,
		Data:   dataBs,
		Config: request.Config,
	})
}

type LangSettings struct {
	Code               string                    `json:"code"`
	Name               string                    `json:"name"`
	FileExtension      string                    `json:"file_extension"`
	SplitObjectByFiles bool                      `json:"split_object_by_files"`
	Formatters         *meta.TypeFormatters      `json:"type_formatters"`
	ClassNameFormatter parser.ClassNameFormatter `json:"class_name_formatter"`
}

var langSettings = map[string]*LangSettings{
	"php": {
		Code:               "php",
		Name:               "PHP",
		FileExtension:      ".php",
		SplitObjectByFiles: true,
		ClassNameFormatter: func(key meta.Key) string {
			return key.PascalCase()
		},
		Formatters: &meta.TypeFormatters{
			Type: func(t meta.Type) string {
				if t.IsArray() {
					return "array"
				} else if t.IsObject() {
					return t.Key.PascalCase()
				} else if t.IsNull() {
					return "null"
				} else if t.IsBool() {
					return "bool"
				} else if t.IsInt() {
					return "int"
				} else if t.IsFloat() {
					return "float"
				} else if t.IsString() {
					return "string"
				}
				return "mixed"
			},
			Doc: func(t meta.Type) string {
				if t.Value == meta.TypeArrayString {
					return "string[]"
				} else if t.Value == meta.TypeArrayInt {
					return "int[]"
				} else if t.Value == meta.TypeArrayFloat {
					return "float[]"
				} else if t.Value == meta.TypeArrayBool {
					return "bool[]"
				} else if t.Value == meta.TypeArrayObject {
					return t.Key.PascalCase() + "[]"
				} else if t.Value == meta.TypeArray {
					return "array"
				} else if t.IsObject() {
					return t.Key.PascalCase()
				} else if t.IsNull() {
					return "null"
				} else if t.IsBool() {
					return "bool"
				} else if t.IsInt() {
					return "int"
				} else if t.IsFloat() {
					return "float"
				} else if t.IsString() {
					return "string"
				}
				return "mixed"
			},
		},
	},
	"go": {
		Code:          "go",
		Name:          "Go",
		FileExtension: ".go",
		ClassNameFormatter: func(key meta.Key) string {
			return key.CamelCase()
		},
		Formatters: &meta.TypeFormatters{
			Type: func(t meta.Type) string {
				if t.Value == meta.TypeArrayString {
					return "[]string"
				} else if t.Value == meta.TypeArrayInt {
					return "[]int"
				} else if t.Value == meta.TypeArrayFloat {
					return "[]float64"
				} else if t.Value == meta.TypeArrayBool {
					return "[]bool"
				} else if t.Value == meta.TypeArrayObject {
					return "[]*" + t.Key.PascalCase()
				} else if t.Value == meta.TypeArray {
					return "[]interface{}"
				} else if t.IsObject() {
					return "*" + t.Key.PascalCase()
				} else if t.IsNull() {
					return "nil"
				} else if t.IsBool() {
					return "bool"
				} else if t.IsInt() {
					return "int"
				} else if t.IsFloat() {
					return "float64"
				} else if t.IsString() {
					return "string"
				}
				return "interface{}"
			},
			Doc: func(t meta.Type) string {
				return t.Formatters.Type(t)
			},
		},
	},
}
