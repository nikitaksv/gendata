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
	"time"

	"github.com/nikitaksv/strcase"

	"github.com/araddon/dateparse"
	"github.com/nikitaksv/dynjson"
)

const (
	TypeNull        = "null"
	TypeInt         = "int"
	TypeString      = "string"
	TypeBool        = "bool"
	TypeFloat       = "float"
	TypeObject      = "object"
	TypeDate        = "date"
	TypeTime        = "time"
	TypeDateTime    = TypeDate + TypeTime
	TypeDuration    = "duration"
	TypeArray       = "array"
	TypeArrayObject = "arrayObject"
	TypeArrayInt    = "arrayInt"
	TypeArrayString = "arrayString"
	TypeArrayBool   = "arrayBool"
	TypeArrayFloat  = "arrayFloat"
)

type TypeFormatter func(t Type) string

type TypeFormatters struct {
	Type TypeFormatter
	Doc  TypeFormatter
}

type Meta struct {
	Key        Key
	Type       Type
	Properties []*Property
}

func (m *Meta) Sort() {
	sort.Slice(m.Properties, func(i, j int) bool { return m.Properties[i].Key < m.Properties[j].Key })
}

func (m *Meta) Clone() *Meta {
	if m == nil {
		return nil
	}

	nm := &Meta{
		Key:        m.Key,
		Type:       m.Type,
		Properties: make([]*Property, len(m.Properties)),
	}

	for i, property := range m.Properties {
		nm.Properties[i] = &Property{
			Nest: property.Nest.Clone(),
			Key:  property.Key,
			Type: property.Type,
		}
	}

	return nm
}

type Property struct {
	Nest *Meta
	Key  Key
	Type Type
}

type Key string

func (k Key) String() string {
	return string(k)
}

// CamelCase ex. camelCase
func (k Key) CamelCase() string {
	return strcase.ToCamelCase(k.String())
}

// CamelCaseAcronym ex. camelCaseID
func (k Key) CamelCaseAcronym() string {
	return strcase.ToCamelCaseAcronym(k.String())
}

// PascalCase ex. PascalCase
func (k Key) PascalCase() string {
	return strcase.ToPascalCase(k.String())
}

// PascalCaseAcronym ex. PascalCaseID
func (k Key) PascalCaseAcronym() string {
	return strcase.ToPascalCaseAcronym(k.String())
}

// SnakeCase ex. snake_case
func (k Key) SnakeCase() string {
	return strcase.ToSnakeCase(k.String())
}

// SnakeCaseAcronym ex. snake_case_ID
func (k Key) SnakeCaseAcronym() string {
	return strcase.ToSnakeCaseAcronym(k.String())
}

// KebabCase ex. kebab-case
func (k Key) KebabCase() string {
	return strcase.ToKebabCase(k.String())
}

// KebabCaseAcronym ex. kebab-case-ID
func (k Key) KebabCaseAcronym() string {
	return strcase.ToKebabCaseAcronym(k.String())
}

// DotCase ex. dot.case
func (k Key) DotCase() string {
	return strcase.ToDotCase(k.String())
}

// DotCaseAcronym ex. dot.case.ID
func (k Key) DotCaseAcronym() string {
	return strcase.ToDotCaseAcronym(k.String())
}

// MergeCase ex. mergecase
func (k Key) MergeCase() string {
	return strcase.ToMergeCase(k.String())
}

// MergeCaseAcronym ex. mergecaseID
func (k Key) MergeCaseAcronym() string {
	return strcase.ToMergeCaseAcronym(k.String())
}

type Type struct {
	Formatters *TypeFormatters `json:"formatters"`

	Key   Key    `json:"key"`
	Value string `json:"value"`
}

func (t Type) String() string {
	return t.Formatters.Type(t)
}
func (t Type) Doc() string {
	return t.Formatters.Doc(t)
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
func (t Type) IsArrayObject() bool {
	return t.Value == TypeArrayObject
}
func (t Type) IsObject() bool {
	return t.Value == TypeObject
}
func (t Type) IsTime() bool {
	return t.Value == TypeTime
}
func (t Type) IsDate() bool {
	return t.Value == TypeDate
}
func (t Type) IsDateTime() bool {
	return t.Value == TypeDateTime
}
func (t Type) IsDuration() bool {
	return t.Value == TypeDuration
}

func TypeOf(key Key, v interface{}) Type {
	t := Type{Key: key}
	switch vType := v.(type) {
	case *dynjson.Object:
		t.Value = TypeObject
		return t
	case *dynjson.Array:
		return typeOfArray(key, vType.Elements)
	case []interface{}:
		return typeOfArray(key, vType)
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
		if vType == "" {
			t.Value = TypeString
		} else if _, err := time.ParseDuration(vType); err == nil {
			t.Value = TypeDuration
		} else if _, err := time.Parse(time.DateOnly, vType); err == nil {
			t.Value = TypeDate
		} else if _, err := time.Parse(time.TimeOnly, vType); err == nil {
			t.Value = TypeTime
		} else if tim, err := dateparse.ParseAny(vType); err == nil {
			h, m, s := tim.Clock()
			if h == 0 && m == 0 && s == 0 {
				t.Value = TypeDate
			} else {
				t.Value = TypeDateTime
			}
		} else {
			t.Value = TypeString
		}
		return t
	default:
		t.Value = TypeNull
		return t
	}
}

//nolint:gocyclo
func typeOfArray(key Key, arr []interface{}) Type {
	t := Type{Key: key}

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
			mx[typeOfArray(key, vType.Elements).Value]++
		case map[string]interface{}:
			mx[TypeArrayObject]++
		case []interface{}:
			mx[typeOfArray(key, vType).Value]++
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
