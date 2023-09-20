package parser

import (
	"encoding/json"

	"github.com/nikitaksv/dynjson"
	"github.com/nikitaksv/gendata/pkg/meta"
	"github.com/pkg/errors"
)

type parserJSON struct{}

func NewParserJSON() (Parser, error) {
	return &parserJSON{}, nil
}

func (p *parserJSON) Parse(data []byte, _ ...Option) (*meta.Meta, error) {
	j := &dynjson.Json{}
	err := json.Unmarshal(data, j)
	if err != nil {
		return nil, err
	}

	key := meta.Key("")

	// main object
	obj := &meta.Meta{
		Key:        key,
		Type:       meta.TypeOf(key, j.Value),
		Properties: nil,
	}

	switch vType := j.Value.(type) {
	case *dynjson.Object:
		p.parseMap(obj, vType)
	case *dynjson.Array:
		mergedArr := p.mergeArray(vType)
		if len(mergedArr.Elements) > 0 {
			if valMap, ok := mergedArr.Elements[0].(*dynjson.Object); ok {
				p.parseMap(obj, valMap)
			}
		}
	default:
		return nil, errors.New("undefined type json data: " + vType.(string))
	}

	return obj, nil
}

func (p *parserJSON) parseMap(obj *meta.Meta, aMap *dynjson.Object) {
	for _, property := range aMap.Properties {
		prop := &meta.Property{
			Key:  meta.Key(property.Key),
			Type: meta.TypeOf(meta.Key(property.Key), property.Value),
			Nest: nil,
		}
		if prop.Type.IsObject() || prop.Type.Value == meta.TypeArrayObject {
			prop.Type.Key = prop.Key
		}

		switch vType := property.Value.(type) {
		case *dynjson.Object:
			nestedObj := &meta.Meta{
				Key:        prop.Key,
				Type:       meta.TypeOf(prop.Key, property),
				Properties: nil,
			}
			p.parseMap(nestedObj, vType)
			prop.Nest = nestedObj
		case *dynjson.Array:
			nestedObj := &meta.Meta{
				Key:        prop.Key,
				Type:       meta.TypeOf(prop.Key, property.Value),
				Properties: nil,
			}
			mergedArr := p.mergeArray(vType)
			if len(mergedArr.Elements) > 0 {
				if valMap, ok := mergedArr.Elements[0].(*dynjson.Object); ok {
					p.parseMap(nestedObj, valMap)
					prop.Nest = nestedObj
				}
			}
		}

		obj.Properties = append(obj.Properties, prop)
	}
}

func (p *parserJSON) mergeArray(arr *dynjson.Array) *dynjson.Array {
	res := &dynjson.Array{}
	m := &dynjson.Object{}
	for _, v := range arr.Elements {
		switch vType := v.(type) {
		case *dynjson.Object:
			m = p.mergeMap(vType, m)
		case *dynjson.Array:
			mergedArr := p.mergeArray(vType)
			if len(mergedArr.Elements) > 0 {
				if valMap, ok := mergedArr.Elements[0].(*dynjson.Object); ok {
					m = p.mergeMap(valMap, m)
				}
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

func (p *parserJSON) mergeMap(maps ...*dynjson.Object) *dynjson.Object {
	result := &dynjson.Object{}
	for _, m := range maps {
		for _, property := range m.Properties {
			existsProp, exists := result.GetProperty(property.Key)

			switch vType := property.Value.(type) {
			case *dynjson.Array:
				dynjsonSetProperty(result, property.Key, p.mergeArray(vType))
			case *dynjson.Object:
				if exists && meta.TypeOf(meta.Key(property.Key), existsProp.Value).IsObject() {
					dynjsonSetProperty(result, property.Key, p.mergeMap(vType, existsProp.Value.(*dynjson.Object)))
				} else {
					dynjsonSetProperty(result, property.Key, property.Value)
				}
			default:
				if !exists || !meta.TypeOf(meta.Key(property.Key), property).IsNull() {
					dynjsonSetProperty(result, property.Key, property.Value)
				}
			}
		}
	}
	return result
}

func dynjsonSetProperty(j *dynjson.Object, k string, v interface{}) {
	_, ok := j.GetProperty(k)
	if len(j.Properties) == 0 || !ok {
		j.Properties = append(j.Properties, &dynjson.Property{Key: k, Value: v})
	} else {
		for _, property := range j.Properties {
			if property.Key == k {
				property.Value = v
			}
		}
	}
}
