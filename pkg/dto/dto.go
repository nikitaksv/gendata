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

package dto

const (
	DataTypeJson DataType = "json"
	DataTypeXml           = "xml"
	DataTypeYaml          = "yaml"
	DataTypeToml          = "toml"
)

type DataType string

type GenerateTemplateRequest struct {
	Template Template `json:"template"`
	Schema   Schema   `json:"schema"`
}

type Template struct {
	String    string            `json:"string"`
	LangType  string            `json:"langType"`
	SortProps bool              `json:"sortProps"`
	TypeMap   map[string]string `json:"typeMap"`
	Class     Class             `json:"class"`
}

type Class struct {
	RootName   string `json:"rootName"`
	PrefixName string `json:"prefixName"`
}

type Schema struct {
	String string   `json:"string"`
	Type   DataType `json:"type"`
}

type GenerateTemplateResponse struct {
	Files []File `json:"files"`
	Error error  `json:"error"`
}

func (g GenerateTemplateResponse) Failed() error {
	return g.Error
}

type File struct {
	Name      string `json:"name"`
	Extension string `json:"extension"`
	Bytes     string `json:"bytes"`
}
