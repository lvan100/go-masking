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
	"sort"
	"strings"
)

const SectionCount = 7

type CharState struct {
	Char  uint8     // multi chars may correspond to same trie node
	State *TrieNode // the trie node corresponding to current char
}

type TrieNode struct {
	State       int
	Child       [SectionCount][]CharState
	ChildBitMap int64
	Depth       int
	Rule        *Rule
	End         bool      // whether it's a terminal node
	AnyStart    bool      // whether it's start wildcard
	AnyEnd      bool      // whether it's end wildcard
	LastEnd     *TrieNode // pointers to the last terminal node
}

// Trie represents a trie.
type Trie struct {
	Nodes []*TrieNode // all nodes.
	Trie  *TrieNode   // root node.
}

// setNextNode sets the next TrieNode in the trie for a given char.
func setNextNode(p *TrieNode, c uint8, n *TrieNode) {
	const off = 'a' - 'A'
	section := charTable[c] % SectionCount

	switch {
	case c >= 'a' && c <= 'z':
		p.ChildBitMap |= 1 << charTable[c]
		p.Child[section] = append(p.Child[section], CharState{c, n})
		p.Child[section] = append(p.Child[section], CharState{c - off, n})
	case c >= '0' && c <= '9':
		p.ChildBitMap |= 1 << charTable[c]
		p.Child[section] = append(p.Child[section], CharState{c, n})
	case c == '-' || c == '_' || c == '@':
		p.ChildBitMap |= 1 << charTable[c]
		p.Child[section] = append(p.Child[section], CharState{c, n})
	default:
		return // never reach, filters unsupported ASCII characters.
	}

	sort.Slice(p.Child[section], func(i, j int) bool {
		return p.Child[section][i].Char < p.Child[section][j].Char
	})
}

// getNextNode retrieves the next TrieNode based on the given char.
// It filters out non-ASCII characters and invalid ASCII characters.
func getNextNode(n *TrieNode, c uint8) *TrieNode {
	// filters non-ASCII characters, such as Chinese characters
	if c >= uint8(len(charTable)) {
		return nil
	}
	m := charTable[c]
	if m < 0 { // filters invalid ASCII characters
		return nil
	}
	// filters characters not in the trie
	if n.ChildBitMap&(1<<m) == 0 {
		return nil
	}
	child := n.Child[m%SectionCount]
	if l := len(child); l == 0 {
		return nil
	} else if l == 1 {
		if p := child[0]; p.Char == c {
			return p.State
		}
		return nil
	} else if l == 2 {
		if p := child[0]; p.Char == c {
			return p.State
		}
		if p := child[1]; p.Char == c {
			return p.State
		}
		return nil
	}
	i := binarySearch(child, c)
	if i < 0 {
		return nil
	}
	return child[i].State
}

// binarySearch performs a binary search on an array sorted in ascending order.
// It returns the index of the found element or -1 if the element is not found.
func binarySearch(x []CharState, c uint8) int {
	n := len(x)
	i, j := 0, n
	for i < j {
		h := int(uint(i+j) >> 1)
		t := int(x[h].Char) - int(c)
		if t < 0 {
			i = h + 1
		} else if t > 0 {
			j = h
		} else {
			return h
		}
	}
	return -1
}

type ParsedKey struct {
	key      string // the processed string without wildcards
	anyStart bool   // indicates whether it has a start wildcard
	anyEnd   bool   // indicates whether it has an end wildcard
}

func (t ParsedKey) Key() string {
	s := t.key
	if t.anyStart {
		s = "*" + s
	}
	if t.anyEnd {
		s = s + "*"
	}
	return s
}

// sortRuleKeys sorts keys according to their length and character values.
func sortRuleKeys(size int, iter func() (string, bool)) []ParsedKey {
	p := make([]ParsedKey, 0, size)
	for {
		k, ok := iter()
		if !ok {
			break
		}

		anyEnd := false
		anyStart := false
		{
			if k[0] == '*' {
				anyStart = true
				k = k[1:]
			}
			n := len(k)
			if k[n-1] == '*' {
				anyEnd = true
				k = k[:n-1]
			}
		}

		found := false
		for i, m := range p {
			if m.key != k {
				continue
			}
			found = true
			if !m.anyStart && anyStart {
				p[i].anyStart = anyStart
			}
			if !m.anyEnd && anyEnd {
				p[i].anyEnd = anyEnd
			}
			break
		}
		if found {
			continue
		}

		p = append(p, ParsedKey{
			key:      k,
			anyEnd:   anyEnd,
			anyStart: anyStart,
		})
	}

	return p
}

// ConstructTrie constructs a trie from a map of rules.
// It ensures that the generated trie is consistent each time.
// When "abc", "*abc", "abc*", and "*abc*" exists, keeps only "*abc*".
func ConstructTrie(rules map[string]*Rule) *Trie {

	krMap := make(map[string]*Rule)
	for _, r := range rules {
		for _, key := range r.Keys {
			krMap[key] = r
		}
	}

	// reorders the rule keys, so the trie is consistent each time.
	p := sortRuleKeys(len(krMap), func() func() (string, bool) {
		keys := OrderedMapKeys(krMap)
		sort.Slice(keys, func(i, j int) bool {
			if len(keys[i]) < len(keys[j]) {
				return true
			}
			if len(keys[i]) > len(keys[j]) {
				return false
			}
			return keys[i] < keys[j]
		})
		i := -1
		return func() (string, bool) { // implements an iterator.
			i++
			if i < len(keys) {
				return keys[i], true
			}
			return "", false
		}
	}())

	trie := &TrieNode{}
	nodes := make([]*TrieNode, 0, 512)
	nodes = append(nodes, trie)

	stateIndex := 1
	for i := 0; i < len(p); i++ {
		var lastEnd *TrieNode
		currState := 0
		j := 0

		// skips the existing prefixes.
		for ; j < len(p[i].key); j++ {
			r := getNextNode(nodes[currState], p[i].key[j])
			if r == nil {
				break
			}
			currState = r.State
			if r.End { // pointers to the last terminal nodes.
				lastEnd = r
			}
		}

		// handles the non-existent parts.
		for ; j < len(p[i].key); j++ {
			n := &TrieNode{
				State:   stateIndex,
				Depth:   j + 1,
				LastEnd: lastEnd, // the nearest terminal node.
			}
			setNextNode(nodes[currState], p[i].key[j], n)
			nodes = append(nodes, n)
			currState = stateIndex
			stateIndex++
		}

		nodes[currState].End = true
		nodes[currState].Rule = krMap[p[i].Key()]
		nodes[currState].AnyStart = p[i].anyStart
		nodes[currState].AnyEnd = p[i].anyEnd
	}

	return &Trie{nodes, trie}
}

// Position represents the start and end positions of a matched rule.
type Position struct {
	Start int
	End   int
	Rule  *Rule
}

// MatchSleep just for test
var MatchSleep func()

// Match performs a match operation on the given byte slice using the trie.
// It returns a list of matched positions and rules, and a boolean indicating
// that whether the operation was intercepted.
func (t *Trie) Match(b []byte, f KeyFilter, maxTime int64) ([]Position, bool) {
	result := make([]Position, 0, 8)
	startTime := MicroNow()
	current := t.Nodes[0]
	tLength := len(b)
	pos := 0
	for {
		count := 0
		for ; pos < tLength; pos++ {
			count++
			if count > 128 {
				break
			}
			n := getNextNode(current, b[pos])
			if n != nil {
				current = n
			} else {
				if current.Depth == 0 {
					continue
				}
				r, ok := testMatch(b, current, pos-1, f)
				if ok {
					result = append(result, r)
				}
				current = t.Nodes[0]
			}
		}
		if pos >= tLength {
			break
		}
		if MatchSleep != nil { // just for test
			MatchSleep()
		}
		// exits if the allowed maximum time is exceeded.
		if MicroNow()-startTime > maxTime {
			return result, true
		}
	}
	return result, false
}

// testMatch checks if the current node matches a rule.
// It returns the matched position and rule if the match is successful.
func testMatch(b []byte, c *TrieNode, pos int, f KeyFilter) (Position, bool) {

	if c.End { // could be an exact match or a prefix match.
		p := Position{pos - c.Depth + 1, pos, c.Rule}
		if f == nil || f(b, p.Start, p.End, c.AnyStart, c.AnyEnd) {
			return p, true
		}
		return Position{}, false
	}

	lastEnd := c.LastEnd // must be a prefix match.
	if lastEnd == nil || !lastEnd.AnyEnd {
		return Position{}, false
	}
	p := Position{pos - c.Depth + 1, pos - (c.Depth - lastEnd.Depth), lastEnd.Rule}
	if f == nil || f(b, p.Start, p.End, lastEnd.AnyStart, lastEnd.AnyEnd) {
		return p, true
	}
	return Position{}, false
}

// DumpTrie outputs all keys reverse-parsed from the prefix tree.
// It returns a sorted list of all keys presented in the trie.
func (t *Trie) DumpTrie() []string {
	m := make(map[string]struct{})
	dumpTrie(t.Trie, "", m)
	return OrderedMapKeys(m)
}

func dumpTrie(n *TrieNode, prefix string, m map[string]struct{}) {
	if n.End {
		s := strings.ToLower(prefix)
		if n.AnyStart {
			s = "*" + s
		}
		if n.AnyEnd {
			s += "*"
		}
		m[s] = struct{}{}
	}
	for _, child := range n.Child {
		for _, c := range child {
			dumpTrie(c.State, prefix+string(c.Char), m)
		}
	}
}
