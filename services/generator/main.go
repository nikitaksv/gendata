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

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"text/template"
)

var jsonStr = `
{
  "id": 2,
  "first_name": "Tam",
  "last_name": "Le-Good",
  "email": "tlegood1@so-net.ne.jp",
  "gender": "Bigender",
  "ip_address": "2.92.36.184",
  "address" : {
	"city": "",
	"street": "",
	"house": ""
  }
}
`

var renderTmpl = `
<?php

namespace common\models;

using yii\base\BaseObject;

/***
 * Class {{ .Key.PascalCase }}
 * @package common\models
 */
class {{ .Key.PascalCase }} extends BaseObject
{ 
{{ range .Properties }}
	/**
	 * @var {{ .Type }}
	 */
	public ${{ .Key.CamelCase }};
{{end}}
}
`

func main() {
	obj, err := ParseJSON([]byte(jsonStr))
	if err != nil {
		log.Fatal(err)
	}

	bs, _ := json.Marshal(obj)
	fmt.Println(string(bs))

	tmpl := template.New("page")
	if tmpl, err = tmpl.Parse(renderTmpl); err != nil {
		log.Fatal(err)
	}
	err = tmpl.Execute(os.Stdout, obj)
	if err != nil {
		log.Fatal(err)
	}
}

func ParseJSON(data []byte) (*meta.Nest, error) {
	var v interface{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		log.Fatal(err)
	}

	// main object
	obj := &meta.Nest{
		Key:        "rootClass",
		Type:       meta.TypeOf(v),
		Properties: nil,
	}

	switch vType := v.(type) {
	case map[string]interface{}:
		parseMap(obj, v.(map[string]interface{}))
	case []interface{}:
		if valMap, ok := mergeArray(v.([]interface{}))[0].(map[string]interface{}); ok {
			parseMap(obj, valMap)
		}
	default:
		return nil, errors.New("undefined type json data: " + vType.(string))
	}

	obj.Sort()
	return obj, nil
}

func parseMap(obj *meta.Nest, aMap map[string]interface{}) {
	for k, v := range aMap {
		prop := &meta.Property{
			Key:  meta.Key(k),
			Type: meta.TypeOf(v),
			Nest: nil,
		}

		switch v.(type) {
		case map[string]interface{}:
			nestedObj := &meta.Nest{
				Key:        prop.Key,
				Type:       meta.TypeOf(v),
				Properties: nil,
			}
			parseMap(nestedObj, v.(map[string]interface{}))

			prop.Nest = nestedObj
		case []interface{}:
			nestedObj := &meta.Nest{
				Key:        prop.Key,
				Type:       meta.TypeOf(v),
				Properties: nil,
			}
			if valMap, ok := mergeArray(v.([]interface{}))[0].(map[string]interface{}); ok {
				parseMap(nestedObj, valMap)
				prop.Nest = nestedObj
			}
		}

		obj.Properties = append(obj.Properties, prop)
	}
	obj.Sort()
}

func mergeArray(arr []interface{}) []interface{} {
	var res []interface{}
	m := map[string]interface{}{}
	for _, v := range arr {
		switch v.(type) {
		case map[string]interface{}:
			m = mergeMap(v.(map[string]interface{}), m)
		case []interface{}:
			if valMap, ok := mergeArray(v.([]interface{}))[0].(map[string]interface{}); ok {
				m = mergeMap(valMap, m)
			}
		default:
			res = append(res, v)
		}
	}

	if len(m) > 0 {
		res = append(res, m)
	}

	return res
}

func mergeMap(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range maps {
		for k, v := range m {
			existsV, exists := result[k]

			switch v.(type) {
			case []interface{}:
				result[k] = mergeArray(v.([]interface{}))
			case map[string]interface{}:
				if exists && meta.TypeOf(existsV) == meta.TypeObject {
					result[k] = mergeMap(v.(map[string]interface{}), existsV.(map[string]interface{}))
				} else {
					result[k] = v
				}
			default:
				if !exists || meta.TypeOf(v) != meta.TypeNull {
					result[k] = v
				}
			}
		}
	}
	return result
}
