// Copyright 2024 github.com/lvan100
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

// charTable maps valid ASCII characters to bitmap indexes,
// to quickly determine whether a character is in the trie.
// a~z, A~Z: 0~25
// 0~9: 26~35
// -: 36
// _: 37
// @: 38
var charTable [128]int8

// splitterTable maps valid ASCII characters to whether they
// are splitter characters. All characters are true except:
// a~z, A~Z: false
// 0~9: false
// -: false
// _: false
var splitterTable [256]bool

func init() {
	initCharTable()
	initSplitterTable()
}

// initCharTable inits the character mapping table.
func initCharTable() {
	for i := 0; i < len(charTable); i++ {
		charTable[i] = -1
	}
	index := int8(0)
	for i := 'a'; i <= 'z'; i++ { // a~z, A~Z: 0~25
		charTable[i-'a'+'A'] = index
		charTable[i] = index
		index++
	}
	for i := 0; i <= 9; i++ { // 0~9: 26~35
		charTable[i+'0'] = index
		index++
	}
	charTable['-'] = index
	index++
	charTable['_'] = index
	index++
	charTable['@'] = index
	index++
}

// initSplitterTable inits the splitter table.
func initSplitterTable() {
	for i := 0; i < len(splitterTable); i++ {
		splitterTable[i] = true
	}
	for i := 'a'; i <= 'z'; i++ { // a~z, A~Z: false
		splitterTable[i-'a'+'A'] = false
		splitterTable[i] = false
	}
	for i := 0; i <= 9; i++ { // 0~9: false
		splitterTable[i+'0'] = false
	}
	splitterTable['-'] = false
	splitterTable['_'] = false
}
