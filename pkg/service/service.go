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
	GenerateTemplate(ctx context.Context, request *dto.GenerateTemplateRequest) (*dto.GenerateTemplateResponse, error)
}

type service struct {
	resource resource.Resource
}

func (s service) GenerateTemplate(ctx context.Context, request *dto.GenerateTemplateRequest) (*dto.GenerateTemplateResponse, error) {
	// Create Meta
	m := meta.Meta{
		Key:        meta.Key(request.Template.Class.RootName),
		Type:       meta.TypeString,
		Properties: nil,
	}
	if m.Key.String() == "" {
		m.Key = "RootClass"
	}

	// Parse String
	tmpl, err := template.New(request.Template.Class.RootName).Parse(request.Template.String)
	if err != nil {
		return nil, err
	}

	// Parse String
	err = json.Unmarshal([]byte(request.Schema.String), &m)
	if err != nil {
		return nil, err
	}

	// Generate String
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, m)
	if err != nil {
		return nil, err
	}

	return &dto.GenerateTemplateResponse{
		Files: []dto.File{
			{
				Name:      m.Key.CamelCase().String(),
				Extension: "php",
				Bytes:     buf.String(),
			},
		},
	}, nil
}

func NewService(r resource.Resource) *service {
	return &service{resource: r}
}
