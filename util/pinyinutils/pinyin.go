// Copyright 2019 Yunion
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pinyinutils

import (
	"bytes"
	"strings"

	"github.com/mozillazg/go-pinyin"
)

// 获取汉字拼音全拼
func Text2Pinyin(hans string) string {
	args := pinyin.NewArgs()
	out := strings.Builder{}
	for _, runeVal := range hans {
		if runeVal > 0x7f {
			o := pinyin.SinglePinyin(runeVal, args)
			for i := range o {
				out.WriteString(o[i])
			}
		} else {
			out.WriteRune(runeVal)
		}
	}
	return out.String()
}

// 获取汉字拼音首字母
func Text2FirstPinyin(hans string) string {
	a := pinyin.NewArgs()
	rows := pinyin.Pinyin(hans, a)
	strResult := ""
	for i := 0; i < len(rows); i++ {
		if len(rows[i]) != 0 {
			str := rows[i][0]
			pi := str[0:1]
			strResult += string(bytes.ToUpper([]byte(pi)))
		}
	}
	return strResult
}
