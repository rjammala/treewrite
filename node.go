package main

import (
	"bytes"
	"fmt"
)

// Node is either a leaf (Token), or a parent with a list of children.
type node struct {
	parent   *node
	depth    int     // Root has depth 0; others have depth == 1 + parent.depth
	children []*node // List of chidlren for non-leaf node.
	token            // Only used if children == nil
}

func (n *node) addChild(c *node) {
	n.children = append(n.children, c)
}

func (n *node) String() string {
	buf := &bytes.Buffer{}

	// Recursive helper function.
	var write func(n *node)
	write = func(n *node) {
		if n.children == nil {
			fmt.Fprint(buf, "[", n.token.text, "]")
			return
		}
		fmt.Fprint(buf, "(")
		sep := ""
		for _, c := range n.children {
			fmt.Fprint(buf, sep)
			write(c)
			sep = " "
		}
		fmt.Fprint(buf, ")")
	}

	write(n)
	return buf.String()
}

func (root *node) serialize() []byte {
	var result []byte
	perNode(root, func(n *node) {
		if n.children != nil {
			return
		}
		for _, t := range n.token.prefix {
			result = append(result, []byte(t.text)...)
		}
		result = append(result, []byte(n.token.text)...)
		for _, t := range n.token.suffix {
			result = append(result, []byte(t.text)...)
		}
	})
	return result
}

func perNode(n *node, fn func(*node)) {
	fn(n)
	for _, c := range n.children {
		perNode(c, fn)
	}
}

// fixFields corrects parent and depth fields in a tree.
func fixFields(n, parent *node, depth int) {
	n.parent = parent
	n.depth = depth
	for _, c := range n.children {
		fixFields(c, n, depth+1)
	}
}
