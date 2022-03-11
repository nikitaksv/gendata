package parser

import (
	"encoding/json"

	"github.com/nikitaksv/dynjson"
	"github.com/nikitaksv/gendata/internal/generator/meta"
	"github.com/pkg/errors"
)

type parserJSON struct {
	Parser
}

func NewParserJSON() Parser {
	return &parserJSON{}
}

func (p *parserJSON) Parse(data []byte, opts ...Option) (*meta.Nest, error) {
	j := &dynjson.Json{}
	err := json.Unmarshal(data, j)
	if err != nil {
		return nil, err
	}

	options := Options{RootName: "Root"}
	if err := options.apply(opts...); err != nil {
		return nil, err
	}

	// main object
	obj := &meta.Nest{
		Key:        meta.Key(options.RootName),
		Type:       meta.TypeOf(j.Value),
		Properties: nil,
	}

	switch vType := j.Value.(type) {
	case *dynjson.Object:
		parseMap(obj, vType)
	case *dynjson.Array:
		if valMap, ok := mergeArray(vType).Elements[0].(*dynjson.Object); ok {
			parseMap(obj, valMap)
		}
	default:
		return nil, errors.New("undefined type json data: " + vType.(string))
	}

	return obj, nil
}

func parseMap(obj *meta.Nest, aMap *dynjson.Object) {
	for _, property := range aMap.Properties {
		prop := &meta.Property{
			Key:  meta.Key(property.Key),
			Type: meta.TypeOf(property.Value),
			Nest: nil,
		}

		switch vType := property.Value.(type) {
		case *dynjson.Object:
			nestedObj := &meta.Nest{
				Key:        prop.Key,
				Type:       meta.TypeOf(property),
				Properties: nil,
			}
			parseMap(nestedObj, vType)
			prop.Nest = nestedObj
		case *dynjson.Array:
			nestedObj := &meta.Nest{
				Key:        prop.Key,
				Type:       meta.TypeOf(property.Value),
				Properties: nil,
			}
			if valMap, ok := mergeArray(vType).Elements[0].(*dynjson.Object); ok {
				parseMap(nestedObj, valMap)
				prop.Nest = nestedObj
			}
		}

		obj.Properties = append(obj.Properties, prop)
	}
}

func mergeArray(arr *dynjson.Array) *dynjson.Array {
	res := &dynjson.Array{}
	m := &dynjson.Object{}
	for _, v := range arr.Elements {
		switch vType := v.(type) {
		case *dynjson.Object:
			m = mergeMap(vType, m)
		case *dynjson.Array:
			if valMap, ok := mergeArray(vType).Elements[0].(*dynjson.Object); ok {
				m = mergeMap(valMap, m)
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

func mergeMap(maps ...*dynjson.Object) *dynjson.Object {
	result := &dynjson.Object{}
	for _, m := range maps {
		for _, property := range m.Properties {
			existsProp, exists := result.GetProperty(property.Key)

			switch vType := property.Value.(type) {
			case *dynjson.Array:
				dynjsonSetProperty(result, property.Key, mergeArray(vType))
			case *dynjson.Object:
				if exists && meta.TypeOf(existsProp.Value) == meta.TypeObject {
					dynjsonSetProperty(result, property.Key, mergeMap(vType, existsProp.Value.(*dynjson.Object)))
				} else {
					dynjsonSetProperty(result, property.Key, property.Value)
				}
			default:
				if !exists || meta.TypeOf(property) != meta.TypeNull {
					dynjsonSetProperty(result, property.Key, property.Value)
				}
			}
		}
	}
	return result
}

func dynjsonSetProperty(j *dynjson.Object, k string, v interface{}) {
	for _, property := range j.Properties {
		if property.Key == k {
			property.Value = v
		}
	}
}
