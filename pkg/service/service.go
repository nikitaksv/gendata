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
	"github.com/nikitaksv/jgen/pkg/dto"
	"github.com/nikitaksv/jgen/pkg/resources"
	"github.com/nikitaksv/jgen/pkg/service/validation"
)

type Service interface {
	GenerateTemplate(request *dto.GenerateTemplateRequest) (*dto.GenerateTemplateResponse, error)
}

type service struct {
	resources resources.Resources
}

func (s service) GenerateTemplate(request *dto.GenerateTemplateRequest) (*dto.GenerateTemplateResponse, error) {
	err := validation.ValidateGenerateTemplateRequest(request)
	if err != nil {
		return nil, err
	}

	return &dto.GenerateTemplateResponse{Data: nil}, nil
}

func NewService(r resources.Resources) *service {
	return &service{resources: r}
}
