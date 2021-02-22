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

type GenerateRequest struct {
	Template Template `json:"template"`
	Schema   Schema   `json:"schema"`
}

type Template struct {
	Content     string `json:"content"`
	LangType    string `json:"langType"`
	SortProps   bool   `json:"sortProps"`
	Split       bool   `json:"split"`
	TypeAliases Types  `json:"typeAliases"`
	Class       Class  `json:"class"`
}

type Class struct {
	RootName   string `json:"rootName"`
	PrefixName string `json:"prefixName"`
}

type Schema struct {
	Content string   `json:"content"`
	Type    DataType `json:"type"`
}

type GenerateResponse struct {
	Files []*File `json:"files"`
	Error error   `json:"error"`
}

func (g GenerateResponse) Failed() error {
	return g.Error
}

type File struct {
	Name      string `json:"name"`
	Extension string `json:"extension"`
	Content   string `json:"content"`
}

type GetTypesRequest struct{}

type GetTypesResponse Types

type Types struct {
	Null        string `json:"null"`
	Int         string `json:"int"`
	String      string `json:"string"`
	Bool        string `json:"bool"`
	Float       string `json:"float"`
	Object      string `json:"object"`
	Array       string `json:"array"`
	ArrayObject string `json:"arrayObject"`
	ArrayInt    string `json:"arrayInt"`
	ArrayString string `json:"arrayString"`
	ArrayBool   string `json:"arrayBool"`
	ArrayFloat  string `json:"arrayFloat"`
}

func (t Types) ToMap() map[string]string {
	return map[string]string{
		"null":        t.Null,
		"int":         t.Int,
		"string":      t.String,
		"bool":        t.Bool,
		"float":       t.Float,
		"object":      t.Object,
		"array":       t.Array,
		"arrayObject": t.ArrayObject,
		"arrayInt":    t.ArrayInt,
		"arrayString": t.ArrayString,
		"arrayBool":   t.ArrayBool,
		"arrayFloat":  t.ArrayFloat,
	}
}
