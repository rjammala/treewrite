package main

import (
	"bytes"
	"regexp"
)

// pattern can be used to find occurrences of a list of pattern nodes
// in a list of subject nodes.  It uses a synthesized regular expression
// to match nodes at the top level, and then recursive calls to
// sub-tree patterns to match lower levels.
type pattern struct {
	// List of pattern nodes being matched.
	list []*node

	// Pattern objects per child (nil for tokens)
	childpat []*pattern

	// A regular expression corresponding to list nodes.
	// list[i] will correspond to re sub-match numbered i+1.
	// (regexp sub-matches are numbered starting at 1).
	re *regexp.Regexp

	// Every unique token is represented by a single rune in re.
	runemap map[string]rune
}

// makePattern returns a pattern for a specified pattern tree, If
// fullMatch is true, the subject list must match exactly, otherwise
// pattern list can occur anywhere in subject list.
func makePattern(p *node) *pattern {
	return makeListPattern(p.children, false, make(map[string]rune))
}

func makeListPattern(list []*node, fullMatch bool, runemap map[string]rune) *pattern {
	p := &pattern{
		runemap:  runemap,
		list:     list,
		childpat: make([]*pattern, len(list)),
	}
	var buf bytes.Buffer
	if fullMatch {
		buf.WriteString("^")
	}
	for i, c := range list {
		if c.children != nil {
			// Match any node in subject and use a recursive
			// pattern matcher to match children.
			buf.WriteString("(.)")
			p.childpat[i] = makeListPattern(c.children, true, runemap)
		} else if c.token.ttype == VAR {
			buf.WriteString("(.)") // Match a single item
		} else if c.token.ttype == RVAR {
			buf.WriteString("(.*)") // Match any number of items
		} else {
			// Match specific token text by first mapping the
			// text to a rune and then matching that rune.
			r, ok := p.runemap[c.token.text]
			if !ok {
				// New token; assign it a unique rune.
				// Adding 128 means we never pick a regexp
				// special character.
				r = rune(len(p.runemap) + 128)
				p.runemap[c.token.text] = r
			}

			buf.WriteString("(")
			buf.WriteRune(r)
			buf.WriteString(")")
		}
	}
	if fullMatch {
		buf.WriteString("$")
	}

	p.re = regexp.MustCompile(buf.String())

	return p
}

// match represents the result of a successful pattern match.
// It includes the extent of the matched subject nodes as well as
// variable assignment.
type match struct {
	vars         map[string][]*node
	start, limit int
}

func (p *pattern) match(subject []*node) (match, bool) {
	// Convert subject to a []byte for regexp matching.
	// Also generate an index that maps from []byte index to subject index.
	var byteToSubjectIndex []int
	var buf bytes.Buffer
	for i, c := range subject {
		if c.children != nil {
			buf.WriteString("_")
		} else {
			r, ok := p.runemap[c.token.text]
			if !ok {
				// This token does not occur in pattern,
				// so it should only be matched by wildcards
				// in pattern.
				r = '_'
			}
			buf.WriteRune(r)
		}
		// Update mapping from byte index to slist index.
		for len(byteToSubjectIndex) < buf.Len() {
			byteToSubjectIndex = append(byteToSubjectIndex, i)
		}
	}
	byteToSubjectIndex = append(byteToSubjectIndex, len(subject)) // Sentinel

	sub := buf.Bytes()

outerLoop:
	for _, m := range p.re.FindAllSubmatchIndex(sub, -1) {
		assign := match{
			vars:  make(map[string][]*node),
			start: byteToSubjectIndex[m[0]],
			limit: byteToSubjectIndex[m[1]],
		}

		for i, pnode := range p.list {
			if pnode.children == nil &&
				pnode.token.ttype != VAR &&
				pnode.token.ttype != RVAR {
				// Simple token match
				continue
			}

			// Map sub indices start..limit-1 to subject nodes.
			// p.list[i] corresponds to submatch index i+1,
			// which is stored in m[2*(i+1), 2*(i+1)+1].
			var matched []*node
			var last *node
			start := m[2*(i+1)]
			limit := m[2*(i+1)+1]
			for j := start; j < limit; j++ {
				n := subject[byteToSubjectIndex[j]]
				if n != last {
					matched = append(matched, n)
					last = n
				}
			}

			if pnode.children == nil {
				// Variable assignment
				assign.vars[pnode.token.text] = matched
				continue
			}

			// Match children
			cmatch, ok := p.childpat[i].match(matched[0].children)
			if !ok {
				continue outerLoop
			}
			// Copy variable assignment from recursive match
			for k, v := range cmatch.vars {
				assign.vars[k] = v
			}
		}

		return assign, true
	}
	return match{}, false
}
