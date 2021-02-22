/*
 * Copyright (c) 2021 Nikita Krasnikov
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

package service

import (
	"context"
	"reflect"
	"testing"

	"github.com/nikitaksv/jgen/pkg/dto"
	"github.com/nikitaksv/jgen/pkg/resource"
	"github.com/nikitaksv/jgen/pkg/service/meta"
	"go.uber.org/zap"
)

func newRes() resource.Resource {
	logger, _ := zap.NewDevelopment()
	return resource.NewResource(logger)
}

func TestNewService(t *testing.T) {
	res := newRes()
	type args struct {
		r resource.Resource
	}
	tests := []struct {
		name string
		args args
		want *service
	}{
		{name: "success", args: args{r: res}, want: &service{resource: res}},
		{name: "empty", args: args{r: nil}, want: &service{resource: nil}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewService(tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_service_Generate(t *testing.T) {
	res := newRes()
	type fields struct {
		resource resource.Resource
	}
	type args struct {
		in0     context.Context
		request *dto.GenerateRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *dto.GenerateResponse
		wantErr bool
	}{
		{
			name:   "Check Key",
			fields: fields{resource: res},
			args: args{
				in0: context.Background(),
				request: &dto.GenerateRequest{
					Template: dto.Template{
						Content:  "{{.Key}}",
						LangType: "php",
						Class: dto.Class{
							RootName:   "HELLO",
							PrefixName: "TEST_",
						},
					},
					Schema: dto.Schema{
						Content: "{\"a1\": 123}",
						Type:    "json",
					},
				},
			},
			want: &dto.GenerateResponse{
				Files: []*dto.File{
					{
						Name:      "TestHello",
						Extension: "php",
						Content:   "TEST_HELLO",
					},
				},
				Error: nil,
			},
			wantErr: false,
		},
		{
			name:   "Check Single Props",
			fields: fields{resource: res},
			args: args{
				in0: context.Background(),
				request: &dto.GenerateRequest{
					Template: dto.Template{
						Content:  "{{ range .Properties }}{{.Key.PascalCase}}{{end}}",
						LangType: "php",
						Class: dto.Class{
							RootName:   "HELLO",
							PrefixName: "TEST_",
						},
					},
					Schema: dto.Schema{
						Content: "{\"a1\": 123}",
						Type:    "json",
					},
				},
			},
			want: &dto.GenerateResponse{
				Files: []*dto.File{
					{
						Name:      "TestHello",
						Extension: "php",
						Content:   "A1",
					},
				},
				Error: nil,
			},
			wantErr: false,
		},
		{
			name:   "Check Multi Props",
			fields: fields{resource: res},
			args: args{
				in0: context.Background(),
				request: &dto.GenerateRequest{
					Template: dto.Template{
						Content:  "{{ range .Properties }}{{.Key.PascalCase}}{{end}}",
						LangType: "php",
						Class: dto.Class{
							RootName:   "HELLO",
							PrefixName: "TEST_",
						},
					},
					Schema: dto.Schema{
						Content: "{\"a1\": 123, \"baseUrl\":\"\", \"DNS\": []}",
						Type:    "json",
					},
				},
			},
			want: &dto.GenerateResponse{
				Files: []*dto.File{
					{
						Name:      "TestHello",
						Extension: "php",
						Content:   "A1BaseUrlDns",
					},
				},
				Error: nil,
			},
			wantErr: false,
		},
		{
			name:   "Check Array Objects",
			fields: fields{resource: res},
			args: args{
				in0: context.Background(),
				request: &dto.GenerateRequest{
					Template: dto.Template{
						Content:  "{{ range .Properties }}{{.Key.PascalCase}}{{end}}",
						LangType: "php",
						Class: dto.Class{
							RootName:   "HELLO",
							PrefixName: "TEST_",
						},
					},
					Schema: dto.Schema{
						Content: "[{\"a1\": 123, \"baseUrl\":\"\", \"DNS\": []}]",
						Type:    "json",
					},
				},
			},
			want: &dto.GenerateResponse{
				Files: []*dto.File{
					{
						Name:      "TestHello",
						Extension: "php",
						Content:   "A1BaseUrlDns",
					},
				},
				Error: nil,
			},
			wantErr: false,
		},
		{
			name:   "Check Multi Files",
			fields: fields{resource: res},
			args: args{
				in0: context.Background(),
				request: &dto.GenerateRequest{
					Template: dto.Template{
						Split:    true,
						Content:  "{{ range .Properties }}{{.Key.SnakeCase}};{{end}}",
						LangType: "php",
						Class: dto.Class{
							RootName:   "HELLO",
							PrefixName: "TEST_",
						},
					},
					Schema: dto.Schema{
						Content: "[{\"a1\": 123, \"baseUrl\":\"\", \"DNS\":\"\", \"user\": {\"first_name\":\"\",\"last-name\": \"\"}}]",
						Type:    "json",
					},
				},
			},
			want: &dto.GenerateResponse{
				Files: []*dto.File{
					{
						Name:      "TestHello",
						Extension: "php",
						Content:   "a1;base_url;dns;user;",
					},
					{
						Name:      "TestUser",
						Extension: "php",
						Content:   "first_name;last_name;",
					},
				},
				Error: nil,
			},
			wantErr: false,
		},
		{
			name:   "Sorting Props",
			fields: fields{resource: res},
			args: args{
				in0: context.Background(),
				request: &dto.GenerateRequest{
					Template: dto.Template{
						Content:   "{{ range .Properties }}{{.Key.SnakeCase}};{{end}}",
						LangType:  "php",
						SortProps: true,
						Class: dto.Class{
							RootName:   "HELLO",
							PrefixName: "TEST_",
						},
					},
					Schema: dto.Schema{
						Content: "{\"a2\":2,\"a1\": 1}",
						Type:    "json",
					},
				},
			},
			want: &dto.GenerateResponse{
				Files: []*dto.File{
					{
						Name:      "TestHello",
						Extension: "php",
						Content:   "a1;a2;",
					},
				},
				Error: nil,
			},
			wantErr: false,
		},
		{
			name:   "Error Schema",
			fields: fields{resource: res},
			args: args{
				in0: context.Background(),
				request: &dto.GenerateRequest{
					Template: dto.Template{
						Content:   "{{ range .Properties }}{{.Key.SnakeCase}};{{end}}",
						LangType:  "php",
						SortProps: true,
						Class: dto.Class{
							RootName:   "HELLO",
							PrefixName: "TEST_",
						},
					},
					Schema: dto.Schema{
						Content: "}",
						Type:    "json",
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "Error Template",
			fields: fields{resource: res},
			args: args{
				in0: context.Background(),
				request: &dto.GenerateRequest{
					Template: dto.Template{
						Content:   "{{ range .Proper ties }}{{.Key.SnakeCase}};{{end}}",
						LangType:  "php",
						SortProps: true,
						Class: dto.Class{
							RootName:   "HELLO",
							PrefixName: "TEST_",
						},
					},
					Schema: dto.Schema{
						Content: "{}",
						Type:    "json",
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "Empty Schema",
			fields: fields{resource: res},
			args: args{
				in0: context.Background(),
				request: &dto.GenerateRequest{
					Template: dto.Template{
						Content:   "{{ range .Properties }}{{.Key.SnakeCase}};{{end}}",
						LangType:  "php",
						SortProps: true,
						Class: dto.Class{
							RootName:   "HELLO",
							PrefixName: "TEST_",
						},
					},
					Schema: dto.Schema{
						Content: "",
						Type:    "json",
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "Check TypeAliases",
			fields: fields{resource: res},
			args: args{
				in0: context.Background(),
				request: &dto.GenerateRequest{
					Template: dto.Template{
						Content:  "{{ range .Properties }}{{.Type}};{{end}}",
						LangType: "php",
						TypeAliases: dto.Types{
							String: "alias",
						},
					},
					Schema: dto.Schema{
						Content: "{\"a1\":\"\"}",
						Type:    "json",
					},
				},
			},
			want: &dto.GenerateResponse{
				Files: []*dto.File{
					{
						Name:      "RootClass",
						Extension: "php",
						Content:   "alias;",
					},
				},
				Error: nil,
			},
			wantErr: false,
		},
		{
			name:   "Check Empty TypeAliases",
			fields: fields{resource: res},
			args: args{
				in0: context.Background(),
				request: &dto.GenerateRequest{
					Template: dto.Template{
						Content:  "{{ range .Properties }}{{.Type}};{{end}}",
						LangType: "php",
						TypeAliases: dto.Types{
							String: "",
						},
					},
					Schema: dto.Schema{
						Content: "{\"a1\":\"\"}",
						Type:    "json",
					},
				},
			},
			want: &dto.GenerateResponse{
				Files: []*dto.File{
					{
						Name:      "RootClass",
						Extension: "php",
						Content:   "string;",
					},
				},
				Error: nil,
			},
			wantErr: false,
		},
		{
			name:   "Check ArrayObject",
			fields: fields{resource: res},
			args: args{
				in0: context.Background(),
				request: &dto.GenerateRequest{
					Template: dto.Template{
						Split:    true,
						Content:  "{{ range .Properties }}{{ if .Type.IsArrayObject}}{{.Key}}[]{{end}}{{end}}",
						LangType: "php",
					},
					Schema: dto.Schema{
						Content: "{\"a1\":[{\"b2\": false}]}",
						Type:    "json",
					},
				},
			},
			want: &dto.GenerateResponse{
				Files: []*dto.File{
					{
						Name:      "RootClass",
						Extension: "php",
						Content:   "a1[]",
					},
					{
						Name:      "A1",
						Extension: "php",
						Content:   "",
					},
				},
				Error: nil,
			},
			wantErr: false,
		},
		{
			name:   "Check All Array types",
			fields: fields{resource: res},
			args: args{
				in0: context.Background(),
				request: &dto.GenerateRequest{
					Template: dto.Template{
						Split:    true,
						Content:  "{{ range .Properties }}{{ if .Type.IsArrayObject}}{{.Key}}[]{{else}}{{.Type}};{{end}}{{end}}",
						LangType: "php",
					},
					Schema: dto.Schema{
						Content: "{\"a1\":[{\"b2\": \"\"}],\"a2\":[true],\"a3\":[1.1],\"a4\":[1],\"a5\":[\"\"]}",
						Type:    "json",
					},
				},
			},
			want: &dto.GenerateResponse{
				Files: []*dto.File{
					{
						Name:      "RootClass",
						Extension: "php",
						Content:   "a1[]arrayBool;arrayFloat;arrayInt;arrayString;",
					},
					{
						Name:      "A1",
						Extension: "php",
						Content:   "string;",
					},
				},
				Error: nil,
			},
			wantErr: false,
		},
		{
			name:   "Check No split",
			fields: fields{resource: res},
			args: args{
				in0: context.Background(),
				request: &dto.GenerateRequest{
					Template: dto.Template{
						Split:    false,
						Content:  "{{ range .Properties }}{{ if .Type.IsArrayObject}}{{.Key}}[]{{else}}{{.Type}};{{end}}{{end}}",
						LangType: "php",
					},
					Schema: dto.Schema{
						Content: "{\"a1\":[{\"b2\": \"\"}],\"a2\":[true],\"a3\":[1.1],\"a4\":[1],\"a5\":[\"\"]}",
						Type:    "json",
					},
				},
			},
			want: &dto.GenerateResponse{
				Files: []*dto.File{
					{
						Name:      "RootClass",
						Extension: "php",
						Content:   "a1[]arrayBool;arrayFloat;arrayInt;arrayString;",
					},
				},
				Error: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := service{
				resource: tt.fields.resource,
			}
			got, err := s.Generate(tt.args.in0, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Generate() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_service_GetTypes(t *testing.T) {
	res := newRes()
	type fields struct {
		resource resource.Resource
	}
	type args struct {
		in0 context.Context
		in1 *dto.GetTypesRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *dto.GetTypesResponse
		wantErr bool
	}{
		{
			name:   "success",
			fields: fields{resource: res},
			args: args{
				in0: context.Background(),
				in1: nil,
			},
			want: &dto.GetTypesResponse{
				Null:        meta.TypeNull,
				Int:         meta.TypeInt,
				String:      meta.TypeString,
				Bool:        meta.TypeBool,
				Float:       meta.TypeFloat,
				Object:      meta.TypeObject,
				Array:       meta.TypeArray,
				ArrayObject: meta.TypeArrayObject,
				ArrayInt:    meta.TypeArrayInt,
				ArrayString: meta.TypeArrayString,
				ArrayBool:   meta.TypeArrayBool,
				ArrayFloat:  meta.TypeArrayFloat,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := service{
				resource: tt.fields.resource,
			}
			got, err := s.GetTypes(tt.args.in0, tt.args.in1)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTypes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTypes() got = %v, want %v", got, tt.want)
			}
		})
	}
}
