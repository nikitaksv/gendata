package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var data = []byte(`
{
  "id": 2,
  "first_name": "Tam",
  "last_name": "Le-Good",
  "email": "tlegood1@so-net.ne.jp",
  "gender": "Bigender",
  "ip_address": "2.92.36.184",
  "addresses": [
    {
      "name": "Home",
      "city": "Test",
      "street": "Test st.",
      "house": "1",
      "coordinates": {
        "lon": 77.2123124,
        "lat": 43.2123124
      }
    }
  ],
  "skills": [
    "go",
    "php"
  ]
}
`)

func TestGen_PHP(t *testing.T) {
	tmpl := []byte(`
<?php

namespace common\models;

using yii\base\BaseObject;

/***
 * Class {{ Name }}
 * @package common\models
 */
class {{ Name }} extends BaseObject
{ 
{{ Properties }}
	/**
	 * @var {{ Type.Doc }}
	 */
	public {{ Type }} ${{ Name.CamelCase }};
{{ /Properties }}
}
`)
	out := []byte(`TestGenClass.php

&lt;?php

namespace common\models;

using yii\base\BaseObject;

/***
 * Class TestGenClass
 * @package common\models
 */
class TestGenClass extends BaseObject
{
	/**
	 * @var int
	 */
	public int $id;
	/**
	 * @var string
	 */
	public string $firstName;
	/**
	 * @var string
	 */
	public string $lastName;
	/**
	 * @var string
	 */
	public string $email;
	/**
	 * @var string
	 */
	public string $gender;
	/**
	 * @var string
	 */
	public string $ipAddress;
	/**
	 * @var TestAddressesClass[]
	 */
	public array $testAddressesClass;
	/**
	 * @var string[]
	 */
	public array $skills;
}
TestAddressesClass.php

&lt;?php

namespace common\models;

using yii\base\BaseObject;

/***
 * Class TestAddressesClass
 * @package common\models
 */
class TestAddressesClass extends BaseObject
{
	/**
	 * @var string
	 */
	public string $name;
	/**
	 * @var string
	 */
	public string $city;
	/**
	 * @var string
	 */
	public string $street;
	/**
	 * @var string
	 */
	public string $house;
	/**
	 * @var TestCoordinatesClass
	 */
	public TestCoordinatesClass $testCoordinatesClass;
}
TestCoordinatesClass.php

&lt;?php

namespace common\models;

using yii\base\BaseObject;

/***
 * Class TestCoordinatesClass
 * @package common\models
 */
class TestCoordinatesClass extends BaseObject
{
	/**
	 * @var float64
	 */
	public float64 $lon;
	/**
	 * @var float64
	 */
	public float64 $lat;
}
`)
	srv := NewService(zap.NewNop())
	rsp, err := srv.Gen(context.Background(), &GenRequest{
		Tmpl: tmpl,
		Data: data,
		Config: &Config{
			Lang:            "php",
			DataFormat:      "json",
			RootClassName:   "Gen",
			PrefixClassName: "Test",
			SuffixClassName: "Class",
		},
	})
	assert.NoError(t, err)
	if assert.NotNil(t, rsp) && assert.NotEmpty(t, rsp.RenderedFiles) {
		sb := &strings.Builder{}
		for _, file := range rsp.RenderedFiles {
			bs, err := io.ReadAll(file.Content)
			if assert.NoError(t, err) {
				sb.WriteString(file.FileName)
				sb.WriteString("\n")
				sb.Write(bs)
			}
		}
		assert.Equal(t, string(out), sb.String())
	}
}
func TestGen_Go(t *testing.T) {
	tmpl := []byte(`
package main

{{ SPLIT }}

type {{ Name.PascalCase }} struct {
{{ Properties }}
	{{ Name.PascalCase }} {{ Type.Doc }} ` + "`json:\"{{ Name.CamelCase }}\"`" + `
{{ /Properties }}
}
{{ /SPLIT }}
`)

	out := []byte(`
package main


type Gen struct {
	Id int ` + "`" + `json:"id"` + "`" + `
	FirstName string ` + "`" + `json:"firstName"` + "`" + `
	LastName string ` + "`" + `json:"lastName"` + "`" + `
	Email string ` + "`" + `json:"email"` + "`" + `
	Gender string ` + "`" + `json:"gender"` + "`" + `
	IpAddress string ` + "`" + `json:"ipAddress"` + "`" + `
	Addresses []*Addresses ` + "`" + `json:"addresses"` + "`" + `
	Skills []string ` + "`" + `json:"skills"` + "`" + `
}

type Addresses struct {
	Name string ` + "`" + `json:"name"` + "`" + `
	City string ` + "`" + `json:"city"` + "`" + `
	Street string ` + "`" + `json:"street"` + "`" + `
	House string ` + "`" + `json:"house"` + "`" + `
	Coordinates *Coordinates ` + "`" + `json:"coordinates"` + "`" + `
}

type Coordinates struct {
	Lon float64 ` + "`" + `json:"lon"` + "`" + `
	Lat float64 ` + "`" + `json:"lat"` + "`" + `
}
`)
	srv := NewService(zap.NewNop())
	rsp, err := srv.Gen(context.Background(), &GenRequest{
		Tmpl: tmpl,
		Data: data,
		Config: &Config{
			Lang:          "go",
			DataFormat:    "json",
			RootClassName: "Gen",
		},
	})
	assert.NoError(t, err)
	if assert.NotNil(t, rsp) && assert.NotEmpty(t, rsp.RenderedFiles) && assert.Len(t, rsp.RenderedFiles, 1) {
		assert.Equal(t, "gen.go", rsp.RenderedFiles[0].FileName)
		bs, err := io.ReadAll(rsp.RenderedFiles[0].Content)
		if assert.NoError(t, err) {
			assert.Equal(t, string(out), string(bs))
		}
	}
}

func TestGen_Sort(t *testing.T) {
	tmpl := []byte(`
package main

{{ SPLIT }}

type {{ Name.PascalCase }} struct {
{{ Properties }}
	{{ Name.PascalCase }} {{ Type.Doc }} ` + "`json:\"{{ Name.CamelCase }}\"`" + `
{{ /Properties }}
}
{{ /SPLIT }}
`)

	out := []byte(`
package main


type Gen struct {
	Addresses []*Addresses ` + "`" + `json:"addresses"` + "`" + `
	Email string ` + "`" + `json:"email"` + "`" + `
	FirstName string ` + "`" + `json:"firstName"` + "`" + `
	Gender string ` + "`" + `json:"gender"` + "`" + `
	Id int ` + "`" + `json:"id"` + "`" + `
	IpAddress string ` + "`" + `json:"ipAddress"` + "`" + `
	LastName string ` + "`" + `json:"lastName"` + "`" + `
	Skills []string ` + "`" + `json:"skills"` + "`" + `
}

type Addresses struct {
	City string ` + "`" + `json:"city"` + "`" + `
	Coordinates *Coordinates ` + "`" + `json:"coordinates"` + "`" + `
	House string ` + "`" + `json:"house"` + "`" + `
	Name string ` + "`" + `json:"name"` + "`" + `
	Street string ` + "`" + `json:"street"` + "`" + `
}

type Coordinates struct {
	Lat float64 ` + "`" + `json:"lat"` + "`" + `
	Lon float64 ` + "`" + `json:"lon"` + "`" + `
}
`)
	srv := NewService(zap.NewNop())
	rsp, err := srv.Gen(context.Background(), &GenRequest{
		Tmpl: tmpl,
		Data: data,
		Config: &Config{
			Lang:           "go",
			DataFormat:     "json",
			RootClassName:  "Gen",
			SortProperties: true,
		},
	})
	assert.NoError(t, err)
	if assert.NotNil(t, rsp) && assert.NotEmpty(t, rsp.RenderedFiles) && assert.Len(t, rsp.RenderedFiles, 1) {
		assert.Equal(t, "gen.go", rsp.RenderedFiles[0].FileName)
		bs, err := io.ReadAll(rsp.RenderedFiles[0].Content)
		if assert.NoError(t, err) {
			fmt.Println(string(bs))
			assert.Equal(t, string(out), string(bs))
		}
	}
}

func TestService_GenFile(t *testing.T) {
	srv := NewService(zap.NewNop())
	rsp, err := srv.GenFile(context.Background(), &GenFileRequest{
		TmplFile: "../../testdata/template.txt",
		DataFile: "../../testdata/data.json",
		Config: &Config{
			Lang:          "go",
			DataFormat:    "json",
			RootClassName: "RootClass",
		},
	})
	assert.NoError(t, err)
	if assert.NotEmpty(t, rsp.RenderedFiles) && assert.Len(t, rsp.RenderedFiles, 1) {
		fileBs, err := os.ReadFile("../../testdata/" + rsp.RenderedFiles[0].FileName + ".txt")
		if assert.NoError(t, err) {
			renderBs, err := io.ReadAll(rsp.RenderedFiles[0].Content)
			if assert.NoError(t, err) {
				assert.Equal(t, string(fileBs), string(renderBs))
			}
		}
	}
}

func TestService_PredefinedLangSettings(t *testing.T) {
	expected := []byte(`{"items":[{"code":"go","name":"GoLang","fileExtension":".go","splitObjectByFiles":false,"configMapping":{"typeMapping":{"array":"[]interface{}","arrayBool":"[]bool","arrayFloat":"[]float64","arrayInt":"[]int","arrayObject":"[]*{{ Name.PascalCase }}","arrayString":"[]string","bool":"bool","float":"float64","int":"int","null":"interface{}","object":"*{{ Name.PascalCase}}","string":"string"},"typeDocMapping":{"array":"[]interface{}","arrayBool":"[]bool","arrayFloat":"[]float64","arrayInt":"[]int","arrayObject":"[]*{{ Name.PascalCase }}","arrayString":"[]string","bool":"bool","float":"float64","int":"int","null":"interface{}","object":"*{{ Name.PascalCase}}","string":"string"},"classNameMapping":"{{ Name.PascalCase }}","fileNameMapping":"{{ Name.CamelCase }}"}},{"code":"php","name":"PHP","fileExtension":".php","splitObjectByFiles":true,"configMapping":{"typeMapping":{"array":"array","arrayBool":"array","arrayFloat":"array","arrayInt":"array","arrayObject":"array","arrayString":"array","bool":"bool","float":"float64","int":"int","null":"null","object":"{{ Name.PascalCase}}","string":"string"},"typeDocMapping":{"array":"array","arrayBool":"bool[]","arrayFloat":"float[]","arrayInt":"int[]","arrayObject":"{{ Name.PascalCase }}[]","arrayString":"string[]","bool":"bool","float":"float64","int":"int","null":"null","object":"{{ Name.PascalCase}}","string":"string"},"classNameMapping":"{{ Name.PascalCase }}","fileNameMapping":"{{ Name.PascalCase }}"}}]}`)

	srv := NewService(zap.NewNop())
	rsp, err := srv.PredefinedLangSettings(context.Background(), &PredefinedLangSettingsListRequest{})
	assert.NoError(t, err)
	assert.NotEmpty(t, rsp.Items)

	bs, err := json.Marshal(rsp)
	if assert.NoError(t, err) {
		assert.JSONEq(t, string(expected), string(bs))
	}
}
