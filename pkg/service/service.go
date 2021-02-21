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
	"bytes"
	"context"
	"encoding/json"
	"text/template"

	"github.com/nikitaksv/jgen/pkg/dto"
	"github.com/nikitaksv/jgen/pkg/resource"
	"github.com/nikitaksv/jgen/pkg/service/meta"
)

type Service interface {
	Generate(ctx context.Context, request *dto.GenerateRequest) (*dto.GenerateResponse, error)
	GetTypes(ctx context.Context, request *dto.GetTypesRequest) (*dto.GetTypesResponse, error)
}

type service struct {
	resource resource.Resource
}

func (s service) GetTypes(_ context.Context, _ *dto.GetTypesRequest) (*dto.GetTypesResponse, error) {
	return &dto.GetTypesResponse{
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
	}, nil
}

func (s service) Generate(_ context.Context, request *dto.GenerateRequest) (*dto.GenerateResponse, error) {
	// Create TypeAliases
	tA := meta.TypeAliases(request.Template.TypeAliases.ToMap())
	// Create Meta
	m := meta.Meta{
		Key:         meta.Key(request.Template.Class.RootName),
		Type:        meta.NewType(meta.TypeString, tA.Apply(meta.TypeString)),
		TypeAliases: tA,
		Properties:  nil,
	}
	if m.Key.String() == "" {
		m.Key = "RootClass"
	}

	// Parse Content
	tmpl, err := template.New(request.Template.Class.RootName).Parse(request.Template.Content)
	if err != nil {
		return nil, err
	}

	// Parse Content
	err = json.Unmarshal([]byte(request.Schema.Content), &m)
	if err != nil {
		return nil, err
	}

	// Generate Content
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, m)
	if err != nil {
		return nil, err
	}

	return &dto.GenerateResponse{
		Files: []dto.File{
			{
				Name:      m.Key.CamelCase().String(),
				Extension: request.Template.LangType,
				Content:   buf.String(),
			},
		},
	}, nil
}

func NewService(r resource.Resource) *service {
	return &service{resource: r}
}
