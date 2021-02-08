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
	"encoding/json"
	"fmt"
	"testing"
)

func TestMeta_UnmarshalFromJSON(t *testing.T) {
	bs := []byte(`
{
  "a1": "one",
  "a2": -10,
  "a3": "ahead",
  "a4": false,
  "a5": 17,
  "a6": "wet",
  "a7": [
    {
      "b1": "one",
      "c2": "two",
      "c3": false
    },
    {
      "b1": "one",
      "b4": 4,
      "b5": [
        "five"
      ],
      "b6": true
    }
  ],
  "a8": null
}
`)

	m := &Meta{}
	err := json.Unmarshal(bs, &m)
	if err != nil {
		t.Error(err)
	}

	jsonBs, _ := json.MarshalIndent(m, "", "  ")
	fmt.Println(string(jsonBs))
}
