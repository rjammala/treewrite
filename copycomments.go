package main

// position identifies a particular token by combining line and column info,
type position struct{ line, column int }

func pos(t token) position { return position{t.line, t.column} }

// boundary contains info to be added as either prefix or suffix.
type boundary struct {
	tokens []token
}

// copier holds state needed to copy comments without duplication.
type copier struct {
	seen           map[position]bool // Tokens already in destination.
	prefix, suffix boundary          // Info to be added on either side
}

// copyComments copies all comments in src to dst that are not already
// present in dst.  It also adds leading and trailing whitespace if necessary.
func copyComments(src []*node, dst []*node) []*node {
	// Could enhance this by attaching comments found on a token T in src
	// to the first untouched occurrence of T in dst.
	c := &copier{seen: make(map[position]bool)}

	// Find all comments that have already been copied, perhaps
	// because a variable assignment copied some portion of src.
	for _, child := range dst {
		perNode(child, func(n *node) {
			for _, t := range n.token.prefix {
				c.seen[pos(t)] = true
			}
			for _, t := range n.token.suffix {
				c.seen[pos(t)] = true
			}
		})
	}

	// Now walk through src, copying all uncopied comments.
	for i, n := range src {
		c.copyNodeComments(n, i == 0, i == len(src)-1)
	}

	if len(c.prefix.tokens)+len(c.suffix.tokens) == 0 {
		// Nothing to copy.
		return dst
	}

	if len(dst) == 0 {
		// Add a dummy node to which we can attach comments.
		dst = []*node{&node{token: token{ttype: OTHER}}}
	}

	// Copy prefix and suffix to first and last dst nodes respectively.
	first, last := dst[0], dst[len(dst)-1]
	first.token.prefix = append(c.prefix.tokens, first.token.prefix...)
	last.token.suffix = append(last.token.suffix, c.suffix.tokens...)
	return dst
}

func (c *copier) copyNodeComments(n *node, leftSide, rightSide bool) {
	num := len(n.children)
	c.copyCommentTokens(n.token.prefix, leftSide, false)
	for i, child := range n.children {
		c.copyNodeComments(child, leftSide && (i == 0), rightSide && (i == num-1))
	}
	c.copyCommentTokens(n.token.suffix, false, rightSide)
}

func (c *copier) copyCommentTokens(list []token, leftSide, rightSide bool) {
	for _, t := range list {
		if c.seen[pos(t)] {
			continue
		}
		if t.ttype == SPACE && !leftSide && !rightSide {
			// Skip adding spaces found in middle of src
			continue
		}

		// Attach to appropriate boundary (prefix or suffix)
		b := &c.suffix
		if leftSide {
			b = &c.prefix
		}
		b.tokens = append(b.tokens, t)
		c.seen[pos(t)] = true
	}
}
