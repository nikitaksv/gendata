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

package meta

import (
	"math"
	"sort"
	"strconv"

	"github.com/nikitaksv/strcase"

	"github.com/nikitaksv/dynjson"
)

const (
	TypeNull        string = "null"
	TypeInt                = "int"
	TypeString             = "string"
	TypeBool               = "bool"
	TypeFloat              = "float"
	TypeObject             = "object"
	TypeArray              = "array"
	TypeArrayObject        = "arrayObject"
	TypeArrayInt           = "arrayInt"
	TypeArrayString        = "arrayString"
	TypeArrayBool          = "arrayBool"
	TypeArrayFloat         = "arrayFloat"
)

type TypeFormatter func(t Type) string

type TypeFormatters struct {
	Type TypeFormatter `json:"-"`
	Doc  TypeFormatter `json:"-"`
}

type Meta struct {
	Key        Key         `json:"key"`
	Type       Type        `json:"type"`
	Properties []*Property `json:"properties"`
}

func (n *Meta) Sort() {
	sort.Slice(n.Properties, func(i, j int) bool { return n.Properties[i].Key < n.Properties[j].Key })
}

type Property struct {
	Key  Key   `json:"key"`
	Type Type  `json:"type"`
	Nest *Meta `json:"nest"`
}

type Key string

func (k Key) String() string {
	return string(k)
}

// CamelCase ex. camelCase
func (k Key) CamelCase() string {
	return strcase.ToCamelCase(k.String())
}

// PascalCase ex. PascalCase
func (k Key) PascalCase() string {
	return strcase.ToPascalCase(k.String())
}

// SnakeCase ex. snake_case
func (k Key) SnakeCase() string {
	return strcase.ToSnakeCase(k.String())
}

// KebabCase ex. kebab-case
func (k Key) KebabCase() string {
	return strcase.ToKebabCase(k.String())
}

// DotCase ex. dot.case
func (k Key) DotCase() string {
	return strcase.ToDotCase(k.String())
}

type Type struct {
	Key   Key    `json:"key"`
	Value string `json:"value"`

	formatters *TypeFormatters
}

func (t Type) String() string {
	return t.formatters.Type(t)
}
func (t Type) Doc() string {
	return t.formatters.Doc(t)
}
func (t Type) IsNull() bool {
	return t.Value == TypeNull
}
func (t Type) IsInt() bool {
	return t.Value == TypeInt
}
func (t Type) IsBool() bool {
	return t.Value == TypeBool
}
func (t Type) IsFloat() bool {
	return t.Value == TypeFloat
}
func (t Type) IsString() bool {
	return t.Value == TypeString
}
func (t Type) IsArray() bool {
	return t.Value == TypeArray || t.Value == TypeArrayObject || t.Value == TypeArrayFloat ||
		t.Value == TypeArrayBool || t.Value == TypeArrayString || t.Value == TypeArrayInt
}
func (t Type) IsObject() bool {
	return t.Value == TypeObject
}

func TypeOf(key Key, v interface{}, f *TypeFormatters) Type {
	t := Type{
		Key:        key,
		formatters: f,
	}
	switch vType := v.(type) {
	case *dynjson.Object:
		t.Value = TypeObject
		return t
	case *dynjson.Array:
		return typeOfArray(key, vType.Elements, f)
	case []interface{}:
		return typeOfArray(key, vType, f)
	case map[string]interface{}:
		t.Value = TypeObject
		return t
	case bool:
		t.Value = TypeBool
		return t
	case float32, float64:
		t.Value = TypeFloat
		if vFloat64, ok := v.(float64); ok && vFloat64 == math.Trunc(vFloat64) {
			t.Value = TypeInt
		}
		return t
	case int, int8, int16, int32, int64:
		t.Value = TypeInt
		return t
	case string:
		t.Value = TypeString
		return t
	default:
		t.Value = TypeNull
		return t
	}
}

func typeOfArray(key Key, arr []interface{}, f *TypeFormatters) Type {
	t := Type{
		Key:        key,
		formatters: f,
	}

	mx := map[string]int{
		TypeArrayBool:   0,
		TypeArrayFloat:  0,
		TypeArrayInt:    0,
		TypeArrayString: 0,
		TypeArrayObject: 0,
		TypeArray:       0,
	}

	for _, v := range arr {
		switch vType := v.(type) {
		case *dynjson.Object:
			mx[TypeArrayObject]++
		case *dynjson.Array:
			mx[typeOfArray(key, vType.Elements, f).Value]++
		case map[string]interface{}:
			mx[TypeArrayObject]++
		case []interface{}:
			mx[typeOfArray(key, vType, f).Value]++
		case int, int8, int16, int32, int64:
			mx[TypeArrayInt]++
		case float32, float64:
			mx[TypeArrayInt] = 0
			mx[TypeArrayFloat]++
		case bool:
			mx[TypeArrayInt] = 0
			mx[TypeArrayFloat] = 0
			mx[TypeArrayBool]++
		case string:
			if vFloat64, err := strconv.ParseFloat(vType, 64); err == nil {
				if vFloat64 == math.Trunc(vFloat64) {
					mx[TypeArrayInt]++
				} else {
					mx[TypeArrayInt] = 0
					mx[TypeArrayFloat]++
				}
			} else if _, err := strconv.ParseBool(vType); err == nil {
				mx[TypeArrayInt] = 0
				mx[TypeArrayFloat] = 0
				mx[TypeArrayBool]++
			} else {
				if mx[TypeArrayInt] > 0 || mx[TypeArrayFloat] > 0 || mx[TypeArrayBool] > 0 || mx[TypeArrayObject] > 0 {
					// Then array have a mixed types
					mx[TypeArray]++
				} else {
					mx[TypeArrayString]++
				}
			}
		default:
			mx[TypeArray]++
		}
	}

	if mx[TypeArray] > 0 {
		t.Value = TypeArray
		return t
	}

	max := 0
	for k, v := range mx {
		if v > max {
			max = v
			t.Value = k
		}
	}

	return t
}
