package parser

import (
	"encoding/json"
	"github.com/nikitaksv/dynjson"
	"github.com/nikitaksv/gendata/internal/generator/meta"
	"github.com/pkg/errors"
)

type parserJSON struct {
	Options *Options `json:"options"`
}

func NewParserJSON(opts ...Option) (Parser, error) {
	options := &Options{}
	if err := options.apply(opts...); err != nil {
		return nil, err
	}
	return &parserJSON{
		Options: options,
	}, nil
}

func (p *parserJSON) className(name meta.Key) string {
	className := ""
	if p.Options.PrefixClassName != "" {
		className = p.Options.PrefixClassName + name.PascalCase() + p.Options.SuffixClassName
	} else {
		className = name.String() + p.Options.SuffixClassName
	}
	if p.Options.ClassNameFormatter != nil {
		className = p.Options.ClassNameFormatter(meta.Key(className))
	}

	return className
}

func (p *parserJSON) Parse(data []byte) (*meta.Nest, error) {
	j := &dynjson.Json{}
	err := json.Unmarshal(data, j)
	if err != nil {
		return nil, err
	}

	key := meta.Key(p.className(meta.Key(p.Options.RootClassName)))

	// main object
	obj := &meta.Nest{
		Key:        key,
		Type:       meta.TypeOf(key, j.Value, p.Options.TypeFormatters),
		Properties: nil,
	}

	switch vType := j.Value.(type) {
	case *dynjson.Object:
		p.parseMap(obj, vType)
	case *dynjson.Array:
		if valMap, ok := p.mergeArray(vType).Elements[0].(*dynjson.Object); ok {
			p.parseMap(obj, valMap)
		}
	default:
		return nil, errors.New("undefined type json data: " + vType.(string))
	}

	if p.Options.SortProperties {
		obj.Sort()
	}

	return obj, nil
}

func (p *parserJSON) parseMap(obj *meta.Nest, aMap *dynjson.Object) {
	for _, property := range aMap.Properties {
		prop := &meta.Property{
			Key:  meta.Key(property.Key),
			Type: meta.TypeOf(meta.Key(property.Key), property.Value, p.Options.TypeFormatters),
			Nest: nil,
		}
		if prop.Type.IsObject() || prop.Type.Value == meta.TypeArrayObject {
			prop.Key = meta.Key(p.className(prop.Key))
			prop.Type.Key = prop.Key
		}

		switch vType := property.Value.(type) {
		case *dynjson.Object:
			nestedObj := &meta.Nest{
				Key:        prop.Key,
				Type:       meta.TypeOf(prop.Key, property, p.Options.TypeFormatters),
				Properties: nil,
			}
			p.parseMap(nestedObj, vType)
			prop.Nest = nestedObj
		case *dynjson.Array:
			nestedObj := &meta.Nest{
				Key:        prop.Key,
				Type:       meta.TypeOf(prop.Key, property.Value, p.Options.TypeFormatters),
				Properties: nil,
			}
			if valMap, ok := p.mergeArray(vType).Elements[0].(*dynjson.Object); ok {
				p.parseMap(nestedObj, valMap)
				prop.Nest = nestedObj
			}
		}

		obj.Properties = append(obj.Properties, prop)
	}
	if p.Options.SortProperties {
		obj.Sort()
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
			if valMap, ok := p.mergeArray(vType).Elements[0].(*dynjson.Object); ok {
				m = p.mergeMap(valMap, m)
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
				if exists && meta.TypeOf(meta.Key(property.Key), existsProp.Value, p.Options.TypeFormatters).IsObject() {
					dynjsonSetProperty(result, property.Key, p.mergeMap(vType, existsProp.Value.(*dynjson.Object)))
				} else {
					dynjsonSetProperty(result, property.Key, property.Value)
				}
			default:
				if !exists || !meta.TypeOf(meta.Key(property.Key), property, p.Options.TypeFormatters).IsNull() {
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
