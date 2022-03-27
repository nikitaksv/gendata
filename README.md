# GenData

[![Godoc Reference](https://godoc.org/github.com/nikitaksv/gendata?status.svg)](http://godoc.org/github.com/nikitaksv/gendata)
![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/nikitaksv/gendata)
![GitHub Workflow Status](https://img.shields.io/github/workflow/status/nikitaksv/gendata/release)
[![codecov](https://codecov.io/gh/nikitaksv/gendata/branch/main/graph/badge.svg?token=TDDP71X62E)](https://codecov.io/gh/nikitaksv/gendata)
![License](https://img.shields.io/github/license/nikitaksv/gendata)

---

Template Data Generator

## Installation

Follow those steps to install the library:

1. Import the library our code:

```shell
go get github.com/nikitaksv/gendata
```

## Usage

Example PHP Class generator from JSON data:
```go
package main

import (
	"bytes"
	"context"
	"fmt"
	"log"

	gendata "github.com/nikitaksv/gendata/pkg/service"
)

func main() {
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
	data := []byte(`
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

	srv := gendata.NewService(nil)
	rsp, err := srv.Gen(context.Background(), &gendata.GenRequest{
		Tmpl: tmpl,
		Data: data,
		Config: &gendata.Config{
			Lang:            "php",
			DataFormat:      "json",
			RootClassName:   "Gen",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range rsp.RenderedFiles {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(file.Content)
		fmt.Println("Filename: ", file.FileName)
		fmt.Println(buf.String())
	}
}
```



