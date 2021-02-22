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

package http

import (
	"context"
	"encoding/json"
	"net/http"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/nikitaksv/jgen/pkg/dto"
	"github.com/nikitaksv/jgen/pkg/endpoint"
)

func MakeHTTPHandler(e endpoint.Endpoints) http.Handler {
	options := []kithttp.ServerOption{}

	r := mux.NewRouter()

	r.Methods("POST", "OPTIONS").Path("/generate").Handler(
		kithttp.NewServer(
			e.Generate,
			decodeGenerateTemplateRequest,
			encodeResponse,
			options...,
		),
	)

	r.Methods("GET", "OPTIONS").Path("/types").Handler(
		kithttp.NewServer(
			e.GetTypes,
			decodeGetTypesRequest,
			encodeResponse,
			options...,
		),
	)

	hs := handlers.CORS(
		handlers.AllowedHeaders([]string{"X-Trace-Id"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"}),
	)(r)
	hs = handlers.ContentTypeHandler(hs, "application/json")
	hs = handlers.CompressHandler(hs)
	hs = handlers.RecoveryHandler()(hs)
	return hs
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	return encode(ctx, w, response, http.StatusOK)
}

func encode(_ context.Context, w http.ResponseWriter, response interface{}, code int) error {
	w.Header().Set("Content-Language", "en")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(response)
}

func decodeGenerateTemplateRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	req := dto.GenerateRequest{}
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}

	return &req, nil
}
func decodeGetTypesRequest(_ context.Context, _ *http.Request) (request interface{}, err error) {
	req := dto.GetTypesRequest{}
	return &req, nil
}
