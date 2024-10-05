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
	"errors"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lvan100/go-masking"
	"github.com/lvan100/go-masking/internal"
)

var ruleMerged atomic.Bool

func testMergeRules(t testing.TB) {
	if ruleMerged.Load() {
		return
	}
	ruleMerged.Store(true)

	{
		keys := masking.DumpTrie()
		if len(keys) > 0 {
			t.Fatalf("got %v, expect empty", keys)
		}
	}

	{
		gotErr := masking.MergeRules(map[string]*masking.Rule{
			"phone": {
				Keys: []string{"cell "},
			},
		})
		expectErr := errors.New("invalid key 'cell '")
		if gotErr == nil {
			t.Fatalf("expect error %v, got nil", expectErr)
		}
		if gotErr.Error() != expectErr.Error() {
			t.Fatalf("got error %v, expect %v", gotErr, expectErr)
		}
	}

	{
		err := masking.MergeRules(map[string]*masking.Rule{
			"phone": {
				Desc: "手机号",
				Keys: []string{
					"phone", "phone1", "mobile", "telephone",
					"p_prefix_", "p_prefix_*",
					"_suffix_p", "*_suffix_p",
					"_content_", "*_content_*",
				},
				Length: 30,
				Masker: masking.SimplePhoneMasker,
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		gotKeys := masking.DumpTrie()
		expectKeys := []string{
			"*_content_*", "*_suffix_p", "mobile",
			"p_prefix_*", "phone", "phone1", "telephone",
		}
		if slices.Compare(gotKeys, expectKeys) != 0 {
			t.Fatalf("got %v, expect %v", gotKeys, expectKeys)
		}
	}

	{
		err := masking.MergeRules(map[string]*masking.Rule{
			"phone": {
				Desc: "手机号",
				Keys: []string{
					"cell", "driver_phone", "spec-cell",
					"p_prefix_other_*",
				},
				Length: 30,
				Masker: masking.SimplePhoneMasker,
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		gotKeys := masking.DumpTrie()
		expectKeys := []string{
			"*_content_*", "*_suffix_p", "cell", "driver_phone", "mobile",
			"p_prefix_*", "p_prefix_other_*", "phone", "phone1",
			"spec-cell", "telephone",
		}
		if slices.Compare(gotKeys, expectKeys) != 0 {
			t.Fatalf("got %v, expect %v", gotKeys, expectKeys)
		}
	}
}

var casesOfMask = []struct {
	src  string
	want string
}{
	{
		src:  "cell",
		want: "cell",
	},
	{
		src:  "cell123:12345678900",
		want: "cell123:12345678900",
	},
	{
		src:  "123cell:12345678900",
		want: "123cell:12345678900",
	},
	{
		src:  "cell:12345678900",
		want: "cell:123****8900",
	},
	{
		src:  "{CELL:12345678900}",
		want: "{CELL:123****8900}",
	},
	{
		src:  "{KELL:12345678900}",
		want: "{KELL:12345678900}",
	},
	{
		src:  "123_suffix_p:12345678900",
		want: "123_suffix_p:123****8900",
	},
	{
		src:  "%22cell%22:12345678900",
		want: "%22cell%22:123****8900",
	},
	{
		src:  "p_prefix_123:12345678900",
		want: "p_prefix_123:123****8900",
	},
	{
		src:  "123_content_123:12345678900",
		want: "123_content_123:123****8900",
	},
	{
		src:  "p_prefix_oth:12345678900",
		want: "p_prefix_oth:123****8900",
	},
	{
		src:  "123p_prefix_oth:12345678900",
		want: "123p_prefix_oth:12345678900",
	},
	{
		src:  "cell:12345678900后面还有中文",
		want: "cell:123****8900后面还有中文",
	},
	{
		src:  "PASSENGER_PHONE:1234567890",
		want: "PASSENGER_PHONE:1234567890",
	},
}

func TestMask(t *testing.T) {
	testMergeRules(t)

	{
		masking.SetKeyFilter(func(b []byte, start int, end int, anyStart bool, anyEnd bool) bool {
			panic(nil)
		})
		_, intercepted := masking.Mask([]byte("cell:123"), 2000)
		if !intercepted {
			t.Fatalf("expect intercepted, got not")
		}
	}

	{
		internal.MatchSleep = func() {
			time.Sleep(10 * time.Millisecond)
		}
		s := make([]byte, 256)
		for i := 0; i < len(s); i++ {
			s[i] = byte(int('0') + i%10)
		}
		_, intercepted := masking.Mask(s, 2000)
		if !intercepted {
			t.Fatalf("expect intercepted, got not")
		}
		internal.MatchSleep = nil
	}

	masking.SetKeyFilter(internal.DefaultKeyFilter)
	for _, tt := range casesOfMask {
		s := []byte(strings.Clone(tt.src))
		masking.Mask(s, 2000)
		if bytes.Compare(s, []byte(tt.want)) != 0 {
			t.Errorf("Mask() = %s, want %s", s, tt.want)
		}
	}
}

func BenchmarkMask(b *testing.B) {
	testMergeRules(b)

	fileNames := []string{
		"50K.txt", "100K.txt", "150K.txt", "200K.txt", "300K.txt",
	}

	type FileData struct {
		name string
		data []byte
	}

	var (
		srcFiles    []FileData
		maskedFiles []FileData
	)

	for _, name := range fileNames {
		data, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			panic(err)
		}
		srcFiles = append(srcFiles, FileData{name, data})
		data, err = os.ReadFile(filepath.Join("testdata", "Masked_"+name))
		if err != nil {
			panic(err)
		}
		maskedFiles = append(maskedFiles, FileData{name, data})
	}
	for i, file := range srcFiles {
		maskedData, _ := masking.Mask(file.data, math.MaxInt)
		if bytes.Compare(maskedData, maskedFiles[i].data) != 0 {
			b.Errorf("%d Mask() = %s, want %s", i, maskedData, maskedFiles[i].data)
		}
	}

	for _, file := range srcFiles {
		b.Run(file.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				src := bytes.Clone(file.data)
				masking.Mask(src, math.MaxInt)
			}
		})
	}
}
