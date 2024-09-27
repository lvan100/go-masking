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

package masking_test

import (
	"bytes"
	"github.com/lvan100/go-masking"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

var (
	id15RegExp = regexp.MustCompile("((%22|%3a|%3A)?)(\\d{6})(\\d{5})(\\d{3}(x|X|\\d))")
	id18RegExp = regexp.MustCompile("((%22|%3a|%3A)?)(\\d{6})(\\d{8})(\\d{3}(x|X|\\d))")
)

func IdMaskerByRegexp(t []byte) []byte {
	if id18RegExp.Match(t) {
		return id18RegExp.ReplaceAll(t, []byte("$1$3********$5"))
	}
	if id15RegExp.Match(t) {
		return id15RegExp.ReplaceAll(t, []byte("$1$3*****$5"))
	}
	return t
}

var casesOfID = []struct {
	src  string
	want string
}{
	{"123", "123"},
	{"abcdef123456789", "abcdef123456789"},
	{"1234567t9012345", "1234567t9012345"},
	{"12345678901234a", "12345678901234a"},
	{"12345678901234x", "123456*****234x"},
	{"12345678901234X", "123456*****234X"},
	{"123456789012345", "123456*****2345"},
	{"123456789012345a", "123456*****2345a"},
	{"1234567890123450", "123456*****23450"},
	{"1234567890123450b", "123456*****23450b"},
	{"12345678901234503", "123456*****234503"},
	{"12345678901234567c", "123456*****234567c"},
	{"123456789012345678", "123456********5678"},
	{"12345678901234567x", "123456********567x"},
	{"12345678901234567X", "123456********567X"},
	{"%12345678901234", "%12345678901234"},
	{"%22123456789045678", "%22123456*****5678"},
	{"%22123456789012345678", "%22123456********5678"},
}

func TestSimpleIdMasker(t *testing.T) {
	for _, tt := range casesOfID {
		{
			s := []byte(strings.Clone(tt.src))
			masking.SimpleIdMasker(s)
			if bytes.Compare(s, []byte(tt.want)) != 0 {
				t.Errorf("SimpleIdMasker() = %s, want %s", s, tt.want)
			}
		}
		{
			s := []byte(strings.Clone(tt.src))
			got := IdMaskerByRegexp(s)
			if bytes.Compare(got, []byte(tt.want)) != 0 {
				t.Errorf("IdMaskerByRegexp() = %s, want %s", got, tt.want)
			}
		}
	}
}

func BenchmarkSimpleIdMasker(b *testing.B) {
	for m, tt := range casesOfID {
		src := []byte(tt.src)
		buf := make([]byte, len(src))
		b.Run("simple#"+strconv.Itoa(m), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				copy(buf, src)
				masking.SimpleIdMasker(buf)
			}
		})
	}
	for m, tt := range casesOfID {
		src := []byte(tt.src)
		buf := make([]byte, len(src))
		b.Run("regexp#"+strconv.Itoa(m), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				copy(buf, src)
				IdMaskerByRegexp(buf)
			}
		})
	}
}

var (
	phoneRegExp = regexp.MustCompile("((%22|%3a|%3A)?)((\\+?86)?)(\\d{3})(\\d{4})(\\d{4})")
)

func PhoneMaskerByRegexp(t []byte) []byte {
	if phoneRegExp.Match(t) {
		return phoneRegExp.ReplaceAll(t, []byte("$1$3$5****$7"))
	}
	return t
}

var casesOfPhone = []struct {
	src  string
	want string
}{
	{"123", "123"},
	{"abc12345678", "abc12345678"},
	{"12345f78901", "12345f78901"},
	{"12345678901", "123****8901"},
	{"12345678901a", "123****8901a"},
	{"123456789011", "123****89011"},
	{"+8234567890", "+8234567890"},
	{"+8612345f78901", "+8612345f78901"},
	{"+8612345678901", "+86123****8901"},
	{"+8612345678901a", "+86123****8901a"},
	{"+86123456789011", "+86123****89011"},
	{"%1234567890", "%1234567890"},
	{"%2212345f78901", "%2212345f78901"},
	{"%3A12345678901", "%3A123****8901"},
	{"%2212345678901a", "%22123****8901a"},
}

func TestSimplePhoneMasker(t *testing.T) {
	for _, tt := range casesOfPhone {
		{
			s := []byte(strings.Clone(tt.src))
			masking.SimplePhoneMasker(s)
			if bytes.Compare(s, []byte(tt.want)) != 0 {
				t.Errorf("SimplePhoneMasker() = %s, want %s", s, tt.want)
			}
		}
		{
			got := PhoneMaskerByRegexp([]byte(tt.src))
			if bytes.Compare(got, []byte(tt.want)) != 0 {
				t.Errorf("PhoneMaskerByRegexp() = %s, want %s", got, tt.want)
			}
		}
	}
}

func BenchmarkSimplePhoneMasker(b *testing.B) {
	for m, tt := range casesOfPhone {
		src := []byte(tt.src)
		buf := make([]byte, len(src))
		b.Run("simple#"+strconv.Itoa(m), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				copy(buf, src)
				masking.SimplePhoneMasker(buf)
			}
		})
	}
	for m, tt := range casesOfPhone {
		src := []byte(tt.src)
		buf := make([]byte, len(src))
		b.Run("regexp#"+strconv.Itoa(m), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				copy(buf, src)
				PhoneMaskerByRegexp(buf)
			}
		})
	}
}
