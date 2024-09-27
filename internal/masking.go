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

import (
	"fmt"
	"strings"
)

// Masker masks the byte slice in-place.
type Masker func(b []byte)

// Rule represents a masking rule.
type Rule struct {
	Desc   string
	Masker Masker
	Length int // searching length after key
	Keys   []string
}

var cfg = struct {
	Rules map[string]*Rule
	Trie  *Trie
}{
	Rules: make(map[string]*Rule),
}

// MergeRules merges new rules and reconstructs the trie.
func MergeRules(rules map[string]*Rule) error {

	// check the keys of new rules
	for _, r := range rules {
		for _, key := range r.Keys {
			for j := 0; j < len(key); j++ {
				if key[j] != '*' && charTable[key[j]] == -1 {
					return fmt.Errorf("invalid key '%s'", key)
				}
			}
		}
	}

	ruleNames := OrderedMapKeys(rules)
	for _, name := range ruleNames {
		r := rules[name]
		if t, ok := cfg.Rules[name]; ok { // update existing rules
			if r.Masker != nil {
				t.Masker = r.Masker
			}
			if r.Desc != "" {
				t.Desc = r.Desc
			}
			if r.Length > 0 {
				t.Length = r.Length
			}
			if len(r.Keys) > 0 {
				ks := make(map[string]struct{})
				for _, s := range t.Keys {
					ks[s] = struct{}{}
				}
				for _, s := range r.Keys {
					s = strings.ToLower(s)
					ks[s] = struct{}{}
				}
				t.Keys = OrderedMapKeys(ks)
			}
		} else { // add new rules
			ks := make(map[string]struct{})
			for _, s := range r.Keys {
				s = strings.ToLower(s)
				ks[s] = struct{}{}
			}
			cfg.Rules[name] = &Rule{
				Desc:   r.Desc,
				Masker: r.Masker,
				Length: r.Length,
				Keys:   OrderedMapKeys(ks),
			}
		}
	}

	cfg.Trie = ConstructTrie(cfg.Rules)
	return nil
}

// DumpTrie outputs all keys reverse-parsed from the prefix tree.
// It returns a sorted list of all keys presented in the trie.
func DumpTrie() []string {
	if cfg.Trie == nil {
		return nil
	}
	return cfg.Trie.DumpTrie()
}

// KeyFilter defines a function type that checks whether a matched key is valid.
type KeyFilter func(b []byte, start int, end int, anyStart bool, anyEnd bool) bool

var keyFilter = DefaultKeyFilter

// SetKeyFilter sets the key filter.
func SetKeyFilter(f KeyFilter) {
	keyFilter = f
}

// Mask masks the byte slice in-place. It accepts a maximum tolerable
// time in microseconds, if the operation cost is over the maximum
// tolerable time, then the operation is interrupted and returns true.
func Mask(b []byte, maxTolerable int64) (_ []byte, intercepted bool) {

	defer func() {
		if r := recover(); r != nil {
			intercepted = true
		}
	}()

	arr, intercepted := cfg.Trie.Match(b, keyFilter, maxTolerable)
	if len(arr) == 0 {
		return b, intercepted
	}

	l := len(b)
	for i := len(arr) - 1; i >= 0; i-- {
		p := arr[i]
		maxEnd := p.End + p.Rule.Length + 1
		if maxEnd >= l {
			maxEnd = l
		}
		s := b[p.End+1 : maxEnd]
		p.Rule.Masker(s)
	}

	return b, intercepted
}

func startSplitter(b []byte, start int, anyStart bool) bool {
	if start <= 0 { // no other characters on the left
		return true
	}
	if splitterTable[b[start-1]] {
		return true
	}
	if anyStart { // pattern matching
		return true
	}
	if start >= 3 {
		// %22phone%22
		if b[start-3] == '%' && b[start-2] == '2' && b[start-1] == '2' {
			return true
		}
	}
	return false
}

func endSplitter(t []byte, end int, andEnd bool) bool {
	if splitterTable[t[end+1]] {
		return true
	}
	if andEnd { // pattern matching
		return true
	}
	return false
}

// DefaultKeyFilter is the default key filter.
func DefaultKeyFilter(b []byte, start int, end int, anyStart bool, anyEnd bool) bool {
	return startSplitter(b, start, anyStart) && endSplitter(b, end, anyEnd)
}
