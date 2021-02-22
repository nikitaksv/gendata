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

package endpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/nikitaksv/jgen/pkg/dto"
	"github.com/nikitaksv/jgen/pkg/service"
	"github.com/nikitaksv/jgen/pkg/validation"
)

type Endpoints struct {
	Generate endpoint.Endpoint
	GetTypes endpoint.Endpoint
}

func New(s service.Service) Endpoints {
	return Endpoints{
		Generate: MakeGenerateEndpoint(s),
		GetTypes: MakeGetTypesEndpoint(s),
	}
}

func MakeGenerateEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*dto.GenerateRequest)

		err := validation.ValidateGenerateRequest(*req)
		if err != nil {
			return nil, err
		}

		res, err := s.Generate(ctx, req)
		if err != nil {
			return nil, err
		}
		return res, err
	}
}

func MakeGetTypesEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return s.GetTypes(ctx, request.(*dto.GetTypesRequest))
	}
}
