package service

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestGen(t *testing.T) {
	data := []byte(`
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
	"house": "",
	"coordinates": {
		"lon": 77.2123124,
		"lat": 43.2123124
	}
  },
  "skills": ["go","php"]
}
`)
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
	srv := NewService(nil)
	rsp, err := srv.Gen(context.Background(), &GenRequest{
		Tmpl: tmpl,
		Data: data,
		Config: &Config{
			Lang:            "go",
			DataFormat:      "json",
			RootClassName:   "Gen",
			PrefixClassName: "Test",
			SuffixClassName: "Class",
			SortProperties:  true,
		},
	})
	assert.NoError(t, err)
	if assert.NotNil(t, rsp) && assert.NotEmpty(t, rsp.RenderedFiles) {
		for _, file := range rsp.RenderedFiles {
			bs, err := io.ReadAll(file.Content)
			if assert.NoError(t, err) {
				fmt.Println("File name: " + file.FileName)
				fmt.Println(string(bs))
			}
		}
	}
}
