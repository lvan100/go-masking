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

package masking

var idMaskerTable = [256]uint8{
	'0': 1,
	'1': 1,
	'2': 1,
	'3': 1,
	'4': 1,
	'5': 1,
	'6': 1,
	'7': 1,
	'8': 1,
	'9': 1,
	'x': 2,
	'X': 2,
	'%': 3,
}

// SimpleIdMasker a simple ID card number masking function.
func SimpleIdMasker(t []byte) {
	n := len(t)
	for i := 0; i < n; i++ {
		// 0. 最短必须是15位
		if n-i < 15 {
			return // 长度不够
		}

		// 1. 第1位必须是数字
		c := idMaskerTable[t[i]]
		switch c {
		case 0, 2: // 其他字符
			continue
		case 3: // %22、%3A
			i += 2
			continue
		}

		l := i + 14
		i++

		// 2. 接下来13位都是数字，第2~14位
		flag := true
		for ; i < l; i++ {
			if idMaskerTable[t[i]] != 1 {
				flag = false
				break
			}
		}
		if !flag {
			continue
		}

		// 3. 第15位是数字或者X
		switch idMaskerTable[t[i]] {
		case 2: // x 或者 X
			copy(t[i-8:], "*****")
			return // 15位，最后是X
		case 1: // 数字，继续向后找
			break
		default:
			continue
		}

		i++
		if i < n { // 16位
			if idMaskerTable[t[i]] != 1 {
				copy(t[i-9:], "*****")
				return // 15位，最后是数字
			}
		} else {
			copy(t[i-9:], "*****")
			return // 15位，最后是数字
		}

		i++
		if i < n { // 17位
			if idMaskerTable[t[i]] != 1 {
				copy(t[i-10:], "*****")
				return // 15位，最后是数字
			}
		} else {
			copy(t[i-10:], "*****")
			return // 15位，最后是数字
		}

		i++
		if i < n { // 18位
			switch idMaskerTable[t[i]] {
			case 1, 2:
				copy(t[i-11:], "********")
				return // 18位，最后是数字或X
			}
		}

		copy(t[i-11:], "*****")
		return // 15位，最后是数字
	}
}

var phoneMaskerTable = [256]uint8{
	'0': 1,
	'1': 1,
	'2': 1,
	'3': 1,
	'4': 1,
	'5': 1,
	'6': 1,
	'7': 1,
	'8': 1,
	'9': 1,
	'+': 2,
	'%': 3,
}

// SimplePhoneMasker a simple phone number masking function.
func SimplePhoneMasker(t []byte) {
	n := len(t)
	for i := 0; i < n; i++ {
		if n-i < 11 {
			return // 长度不够
		}

		c := phoneMaskerTable[t[i]]
		switch c {
		case 0: // 其他字符
			continue
		case 2: // +86
			i += 2
			continue
		case 3: // %22、%3A
			i += 2
			continue
		}

		l := i + 10
		i++

		// 2. 接下来10位都是数字
		flag := true
		for ; i < l; i++ {
			if phoneMaskerTable[t[i]] != 1 {
				flag = false
				break
			}
		}
		if !flag {
			continue
		}

		copy(t[i-7:], "****")
		return // 找到疑似手机号
	}
}
