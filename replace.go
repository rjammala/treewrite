package main

import "sort"

type replacer struct {
	freq  map[string]int // Frequency of each token text.
	occur map[string][]*node
}

func replace(subject, pattern, replacement *node) {
	// TODO: Verify that pattern does not have same var multiple times.
	r := &replacer{
		freq:  make(map[string]int),
		occur: make(map[string][]*node),
	}

	// Populate frequency table and mapping from potential anchor point
	// to nodes.
	perNode(subject, func(n *node) {
		if n.children == nil {
			r.freq[n.token.text]++
			r.occur[n.token.text] = append(r.occur[n.token.text], n)
		}
	})

	// Pick least frequent anchor point in pattern
	var anchor *node
	anchorCount := 1000000000
	perNode(pattern, func(n *node) {
		if n.children != nil {
			return
		}
		if n.token.ttype == VAR || n.token.ttype == RVAR {
			return
		}
		c := r.freq[n.token.text]
		if c < anchorCount {
			anchor = n
			anchorCount = c
		}
	})
	if anchor == nil {
		// TODO: Either handle patterns with no anchors, or raise error.
		return
	}

	// Order anchor occurences in subject in decreasing depth.
	occ, ok := r.occur[anchor.token.text]
	if !ok {
		return
	}
	sort.Slice(occ, func(i, j int) bool {
		return occ[i].depth > occ[j].depth
	})

	pat := makePattern(pattern)

	var last *node
	for _, sub := range occ {
		// anchor might be deep inside the pattern.  Pop back
		// up until we are at root.
		if sub.depth < anchor.depth {
			continue // Cannot match here
		}
		for i := 0; i < anchor.depth; i++ {
			sub = sub.parent
		}
		// Skip if we are in same list as last iteration.
		if sub == last {
			continue
		}
		last = sub

		start := 0
		for start < len(sub.children) {
			// Look for next match of pat in slist
			src := sub.children
			m, ok := pat.match(src[start:])
			if !ok {
				break
			}

			// Generate replacement nodes.
			result := substitute(replacement, m)
			for _, r := range result {
				r.parent = sub
				r.depth = sub.depth + 1
			}
			result = copyComments(src[start+m.start:start+m.limit], result)

			// Children are ordered as follows:
			//   first start		Not passed to Match
			//   next m.start		Skipped by match
			//   next m.limit - m.start	Replaced
			//   remainder			Kept
			remainder := len(src) - start - m.limit
			dst := make([]*node, 0)
			dst = append(dst, src[:start+m.start]...)
			dst = append(dst, result...)
			dst = append(dst, src[len(src)-remainder:]...)
			sub.children = dst
			fixFields(sub, sub.parent, sub.depth)

			// Continue matching just past replaced nodes.
			start = start + m.start + len(result)
		}
	}
}

func substitute(replacement *node, m match) []*node {
	var res []*node
	if replacement.children != nil {
		for _, c := range replacement.children {
			res = append(res, substitute(c, m)...)
		}
		return res
	}

	tok := replacement.token
	if tok.ttype != VAR && tok.ttype != RVAR {
		return append(res, clone(replacement))
	}

	vals, ok := m.vars[tok.text]
	if !ok {
		// TODO: Check for this and raise error earlier
		panic("variable not found + " + tok.text)
	}
	for _, v := range vals {
		res = append(res, clone(v))
	}
	return res
}

func clone(n *node) *node {
	r := &node{}
	*r = *n
	for i, c := range r.children {
		r.children[i] = clone(c)
	}
	return r
}
