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

	"github.com/nikitaksv/dynjson"
	"github.com/nikitaksv/strcase"
)

const (
	TypeNull        Type = "null"
	TypeInt              = "int"
	TypeString           = "string"
	TypeBool             = "bool"
	TypeFloat            = "float"
	TypeObject           = "object"
	TypeArray            = "array"
	TypeArrayObject      = "arrayObject"
	TypeArrayInt         = "arrayInt"
	TypeArrayString      = "arrayString"
	TypeArrayBool        = "arrayBool"
	TypeArrayFloat       = "arrayFloat"
)

type Meta struct {
	index      int
	Key        Key         `json:"key"`
	Type       Type        `json:"type"`
	Properties []*Property `json:"properties"`
}

func (m *Meta) UnmarshalJSON(data []byte) error {
	j := &dynjson.Json{}
	err := j.UnmarshalJSON(data)
	if err != nil {
		return err
	}
	parseMap(m, j.Value.(*dynjson.Object))

	return nil
}

func (m *Meta) getProperty(key Key) (*Property, bool) {
	for _, p := range m.Properties {
		if p.Key == key {
			return p, true
		}
	}
	return nil, false
}

func (m Meta) Sort() {
	sortProperties(m.Properties)
}

type Property struct {
	// for origin order sorting
	index int

	Key  Key   `json:"key"`
	Type Type  `json:"type"`
	Nest *Meta `json:"nest"`
}

func sortProperties(props []*Property) {
	sort.Slice(props, func(i, j int) bool { return props[i].Key < props[j].Key })
}

type Key string

func (k Key) String() string {
	return string(k)
}

// CamelCase ex. camelCase
func (k Key) CamelCase() Key {
	return Key(strcase.ToCamelCase(k.String()))
}

// PascalCase ex. PascalCase
func (k Key) PascalCase() Key {
	return Key(strcase.ToPascalCase(k.String()))
}

// SnakeCase ex. snake_case
func (k Key) SnakeCase() Key {
	return Key(strcase.ToSnakeCase(k.String()))
}

// KebabCase ex. kebab-case
func (k Key) KebabCase() Key {
	return Key(strcase.ToKebabCase(k.String()))
}

// DotCase ex. dot.case
func (k Key) DotCase() Key {
	return Key(strcase.ToDotCase(k.String()))
}

type Type string

func (t Type) String() string {
	return string(t)
}
func (t Type) Long() Type {
	return t
}
func (t Type) Short() Type {
	return t
}
func (t Type) IsNull() bool {
	return t == TypeNull
}
func (t Type) IsInt() bool {
	return t == TypeInt
}
func (t Type) IsBool() bool {
	return t == TypeBool
}
func (t Type) IsFloat() bool {
	return t == TypeFloat
}
func (t Type) IsNumber() bool {
	return t.IsFloat() || t.IsInt()
}
func (t Type) IsString() bool {
	return t == TypeString
}
func (t Type) IsArray() bool {
	return t == TypeArray ||
		t == TypeArrayBool ||
		t == TypeArrayFloat ||
		t == TypeArrayObject ||
		t == TypeArrayString
}
func (t Type) IsObject() bool {
	return t == TypeObject
}

// Returning meta-type data
func TypeOf(v interface{}) Type {
	switch v.(type) {
	case []interface{}:
		return typeOfArray(v.([]interface{}))
	case map[string]interface{}:
		return TypeObject
	case bool:
		return TypeBool
	case float32, float64:
		vFloat64 := v.(float64)
		if vFloat64 == math.Trunc(vFloat64) {
			return TypeInt
		}

		return TypeFloat
	case int, int8, int16, int32, int64:
		return TypeInt
	case string:
		return TypeString
	default:
		return TypeNull
	}
}

// If json/xml array have mixed type data. This function detect most superior data type.
func typeOfArray(arr []interface{}) Type {
	var t Type

	mx := map[Type]int{
		TypeArrayBool:   0,
		TypeArrayFloat:  0,
		TypeArrayInt:    0,
		TypeArrayString: 0,
		TypeArrayObject: 0,
		TypeArray:       0,
	}

	for _, v := range arr {
		switch v.(type) {
		case map[string]interface{}:
			mx[TypeArrayObject]++
		case []interface{}:
			mx[typeOfArray(v.([]interface{}))]++
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
			vS := v.(string)
			if vFloat64, err := strconv.ParseFloat(vS, 64); err == nil {
				if vFloat64 == math.Trunc(vFloat64) {
					mx[TypeArrayInt]++
				} else {
					mx[TypeArrayInt] = 0
					mx[TypeArrayFloat]++
				}
			} else if _, err := strconv.ParseBool(vS); err == nil {
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
		return TypeArray
	}

	max := 0
	for k, v := range mx {
		if v > max {
			max = v
			t = k
		}
	}

	return t
}
