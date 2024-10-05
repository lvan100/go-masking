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

import (
	"github.com/lvan100/go-masking/internal"
)

// Rule represents a masking rule.
type Rule = internal.Rule

// Masker masks the byte slice in-place.
type Masker = internal.Masker

// MergeRules merges new rules and reconstructs the trie.
func MergeRules(rules map[string]*Rule) error {
	return internal.MergeRules(rules)
}

// Mask masks the byte slice in-place. It accepts a maximum tolerable
// time in microseconds, if the operation cost is over the maximum
// tolerable time, then the operation is interrupted and returns true.
func Mask(t []byte, maxTolerable int64) (_ []byte, intercepted bool) {
	return internal.Mask(t, maxTolerable)
}

// KeyFilter defines a function type that checks whether a matched key is valid.
type KeyFilter = internal.KeyFilter

// SetKeyFilter sets the key filter.
func SetKeyFilter(f KeyFilter) {
	internal.SetKeyFilter(f)
}

// DumpTrie outputs all keys reverse-parsed from the prefix tree.
// It returns a sorted list of all keys presented in the trie.
func DumpTrie() []string {
	return internal.DumpTrie()
}
