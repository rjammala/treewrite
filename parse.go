package main

// parser implements recursive descent parsing.  We do not really parse
// any particular language, but just look for common expression patterns
// and ensure their structure is reflected in the generated parse tree.
type parser struct {
	tok *tokenizer
}

func parse(input []byte) *node {
	p := &parser{tok: newTokenizer(input)}
	return p.root()
}

func (p *parser) root() *node {
	n := &node{children: make([]*node, 0)}
	p.readExprs(n, "")
	end := p.tok.read()
	if len(end.prefix) > 0 {
		// Need to preserve END token to avoid losing
		// attachments.
		n.addChild(&node{token: end})
	}
	// Strip unnecessary levels.
	for len(n.children) == 1 {
		n = n.children[0]
	}
	// Wrap top-level token in a node.
	if n.children == nil {
		r := &node{}
		r.addChild(n)
		n = r
	}
	fixFields(n, nil, 0)
	return n
}

func (p *parser) readExprs(dst *node, closer string) {
	for p.tok.peek.ttype != END && !p.lookingAt(closer) {
		dst.addChild(p.assign())
	}
}

func (p *parser) assign() *node {
	return p.parseRight(p.oror,
		"=", "+=", "-=", "*=", "/=", "%=", "<<=", ">>=",
		"&=", "^=", "!=")
}

func (p *parser) oror() *node   { return p.parseLeft(p.andand, "||") }
func (p *parser) andand() *node { return p.parseLeft(p.bitor, "&&") }
func (p *parser) bitor() *node  { return p.parseLeft(p.bitxor, "|") }
func (p *parser) bitxor() *node { return p.parseLeft(p.bitand, "^") }
func (p *parser) bitand() *node { return p.parseLeft(p.eq, "&") }
func (p *parser) eq() *node     { return p.parseLeft(p.cmp, "==", "!=") }
func (p *parser) cmp() *node    { return p.parseLeft(p.shift, "<", "<=", ">", ">=") }
func (p *parser) shift() *node  { return p.parseLeft(p.plus, "<<", ">>") }
func (p *parser) plus() *node   { return p.parseLeft(p.mult, "+", "-") }
func (p *parser) mult() *node   { return p.parseLeft(p.unary, "*", "/", "%") }

func (p *parser) unary() *node {
	var n *node
	if p.lookingAt("&", "*", "!", "~", "+", "-", "++", "--") {
		n = &node{}
		n.addChild(&node{token: p.tok.read()})
		n.addChild(p.unary())
	} else {
		n = p.suffix()
	}
	return n
}

func (p *parser) suffix() *node {
	n := p.term()
	for {
		if p.lookingAt("++", "--", ".", "->") {
			parent := &node{}
			parent.addChild(n)
			parent.addChild(&node{token: p.tok.read()})
			parent.addChild(p.term())
			n = parent
		} else if p.lookingAt("(", "[", "{") {
			closer := closerFor(p.tok.peek.text)
			parent := &node{}
			parent.addChild(n)
			parent.addChild(&node{token: p.tok.read()})
			p.readExprs(parent, closer)
			if p.lookingAt(closer) {
				parent.addChild(&node{token: p.tok.read()})
			}
			n = parent
		} else {
			break
		}
	}
	return n
}

func (p *parser) term() *node {
	if p.lookingAt("(", "[", "{") {
		closer := closerFor(p.tok.peek.text)
		n := &node{}
		n.addChild(&node{token: p.tok.read()})
		p.readExprs(n, closer)
		if p.lookingAt(closer) {
			n.addChild(&node{token: p.tok.read()})
		}
		return n
	}
	return &node{token: p.tok.read()}
}

func (p *parser) parseLeft(sub func() *node, tokens ...string) *node {
	left := sub()
	for p.lookingAt(tokens...) {
		n := &node{}
		n.addChild(left)
		n.addChild(&node{token: p.tok.read()})
		n.addChild(sub())
		left = n
	}
	return left
}

func (p *parser) parseRight(sub func() *node, tokens ...string) *node {
	n := sub()
	if p.lookingAt(tokens...) {
		left := n
		n = &node{}
		n.addChild(left)
		n.addChild(&node{token: p.tok.read()})
		n.addChild(p.parseRight(sub, tokens...))
	}
	return n
}

// lookingAt returns true iff next token is in tokens.
func (p *parser) lookingAt(tokens ...string) bool {
	for _, t := range tokens {
		if p.tok.peek.text == t {
			return true
		}
	}
	return false
}
