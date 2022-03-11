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
	"fmt"
	"log"
	"os"
	"text/template"

	parser2 "github.com/nikitaksv/gendata/internal/generator/parser"
	"github.com/nikitaksv/gendata/internal/lexer"
	"github.com/nikitaksv/gendata/internal/syntax"
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
 * Class {{ Name }}
 * @package common\models
 */
class {{ Name }} extends BaseObject
{ 
{{ Properties }}
	/**
	 * @var {{ Type }}
	 */
	public {{ Type }} ${{ Name.CamelCase }};
{{ /Properties }}
}
`

func main() {
	parser := parser2.NewParserJSON()
	metaObj, err := parser.Parse([]byte(jsonStr))
	if err != nil {
		log.Fatal(err)
	}

	bs, _ := json.Marshal(metaObj)
	fmt.Println(string(bs))

	if err := syntax.Validate([]byte(renderTmpl)); err != nil {
		log.Fatal(err)
	}

	renderTmplDone := []byte(renderTmpl)
	for _, l := range lexer.Lexers {
		renderTmplDone = l.Replace(renderTmplDone)
	}

	tmpl := template.New("page")
	if tmpl, err = tmpl.Parse(string(renderTmplDone)); err != nil {
		log.Fatal(err)
	}
	err = tmpl.Execute(os.Stdout, metaObj)
	if err != nil {
		log.Fatal(err)
	}
}
