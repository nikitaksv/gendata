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
	"github.com/nikitaksv/dynjson"
)

func parseMap(obj *Meta, aMap *dynjson.Object) {
	for k, v := range aMap.Properties {
		prop := &Property{
			index: k,
			Key:   Key(v.Key),
			Type:  TypeOf(v.Value),
			Nest:  nil,
		}

		switch v.Value.(type) {
		case *dynjson.Object:
			prop.Type = TypeObject
			nestObj := v.Value.(*dynjson.Object)
			newObj := &Meta{
				Key:        prop.Key,
				Type:       TypeObject,
				Properties: make([]*Property, 0, len(nestObj.Properties)),
			}
			parseMap(newObj, nestObj)
			prop.Nest = newObj
		case *dynjson.Array:
			prop.Type = TypeArray
			nestedObj := &Meta{
				Key:        prop.Key,
				Type:       TypeArray,
				Properties: nil,
			}
			if valObj, ok := mergeArray(v.Value.(*dynjson.Array)).Elements[0].(*dynjson.Object); ok {
				parseMap(nestedObj, valObj)
				prop.Nest = nestedObj
			}
		}

		obj.Properties = append(obj.Properties, prop)
	}
}

func mergeArray(arr *dynjson.Array) *dynjson.Array {
	res := &dynjson.Array{Elements: make([]interface{}, 0, len(arr.Elements))}

	m := &dynjson.Object{}
	for _, v := range arr.Elements {
		switch v.(type) {
		case *dynjson.Object:
			m = mergeObjects(v.(*dynjson.Object), m)
		case *dynjson.Array:
			if valObj, ok := mergeArray(v.(*dynjson.Array)).Elements[0].(*dynjson.Object); ok {
				m = mergeObjects(valObj, m)
			}
		default:
			res.Elements = append(res.Elements, v)
		}
	}

	if len(m.Properties) > 0 {
		res.Elements = append(res.Elements, m)
	}

	return res
}

func mergeObjects(maps ...*dynjson.Object) *dynjson.Object {
	result := &dynjson.Object{
		Properties: []*dynjson.Property{},
	}
	for _, m := range maps {
		for i, v := range m.Properties {

			existsV, exists := result.GetProperty(v.Key)

			switch v.Value.(type) {
			case *dynjson.Array:
				if mergedObj, ok := mergeArray(v.Value.(*dynjson.Array)).Elements[0].(*dynjson.Object); ok {
					result.Properties = append(result.Properties, mergeObjects(v.Value.(*dynjson.Object), mergedObj).Properties...)
				} else {

				}
			case *dynjson.Object:
				if exists {
					result.Properties = append(result.Properties, mergeObjects(v.Value.(*dynjson.Object), existsV.Value.(*dynjson.Object)).Properties...)
				} else {
					result.Properties[i] = v
				}
			default:
				if !exists {
					result.Properties = append(result.Properties, v)
				}
			}
		}
	}
	return result
}
