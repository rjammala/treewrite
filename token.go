package main

import "fmt"

type tokenType int

const (
	END tokenType = iota
	STRING
	COMMENT
	WORD
	SPACE
	VAR
	RVAR
	OTHER
	OPENER
	CLOSER
)

var ttypeStrings = [...]string{
	"END", "STRING", "COMMENT", "WORD", "SPACE", "VAR", "RVAR", "OTHER", "OPENER", "CLOSER",
}

func (t tokenType) String() string {
	if t >= 0 && int(t) < len(ttypeStrings) {
		return ttypeStrings[int(t)]
	}
	return "?"
}

type token struct {
	ttype  tokenType
	line   int
	column int
	text   string

	// Neighboring comments and spaces attached to this token.
	prefix, suffix []token
}

func (t token) String() string {
	if t.ttype == SPACE {
		// Spaces are quoted for clarity.
		return fmt.Sprintf("(%s %d.%d %#v)", t.ttype, t.line, t.column, t.text)
	} else {
		return fmt.Sprintf("(%s %d.%d %s)", t.ttype, t.line, t.column, t.text)
	}
}

type tokenizer struct {
	input    []byte  // Bytes remaining to be processed.
	buffered []token // Space/comment tokens to attach to next token.
	peek     token   // Next non-space/non-comment token.
	line     int
	column   int
}

func newTokenizer(data []byte) *tokenizer {
	t := &tokenizer{input: data, line: 1}
	for {
		t.peek = t.readRaw()
		if t.peek.ttype != COMMENT && t.peek.ttype != SPACE {
			break
		}
		t.buffered = append(t.buffered, t.peek)
	}
	return t
}

func (t *tokenizer) read() token {
	res := t.peek
	res.prefix = t.buffered
	t.buffered = nil
	onSameLine := true
	for {
		t.peek = t.readRaw()
		if t.peek.ttype != COMMENT && t.peek.ttype != SPACE {
			break
		}
		if onSameLine && t.line == res.line {
			res.suffix = append(res.suffix, t.peek)
		} else {
			onSameLine = false
			t.buffered = append(t.buffered, t.peek)
		}
	}
	if t.peek.ttype == END && len(t.buffered) > 0 {
		// Instead of attaching as prefix to END, attach to res.
		res.suffix = append(res.suffix, t.buffered...)
		t.buffered = nil
	}
	return res
}

// scanner holds a mapping from leading byte of a token to a list of
// potential token types that start with that byte.  Each entry in the
// list holds the suffix that must follow the leading byte and the token
// type.  If an optional function in the entry is non-nil, it is
// invoked and returns the token type and the token length.
type scanner [256][]scanEntry
type scanEntry struct {
	suffix string
	ttype  tokenType
	fn     func([]byte) (tokenType, int)
}

var scandata scanner

func init() {
	// For now just process C/C++ syntax.
	scandata['('] = []scanEntry{{"", OPENER, nil}}
	scandata[')'] = []scanEntry{{"", CLOSER, nil}}
	scandata['['] = []scanEntry{{"", OPENER, nil}}
	scandata[']'] = []scanEntry{{"", CLOSER, nil}}
	scandata['{'] = []scanEntry{{"", OPENER, nil}}
	scandata['}'] = []scanEntry{{"", CLOSER, nil}}

	// Special variable length tokens.
	scandata['$'] = []scanEntry{{fn: readVar}}
	scandata['"'] = []scanEntry{{fn: readDblString}}
	scandata['\''] = []scanEntry{{fn: readSingleString}}
	scandata['/'] = []scanEntry{
		scanEntry{suffix: "/", fn: readLineComment},
		scanEntry{suffix: "*", fn: readMultiLineComment}}

	// Spaces and words.
	for b := range scandata {
		if isSpace(byte(b)) {
			scandata[b] = []scanEntry{{fn: readSpaces}}
		} else if isWordByte(byte(b)) {
			scandata[b] = []scanEntry{{fn: readWord}}
		}
	}

	// Multi-character operators.  If two operators have the same suffix,
	// longer operator must occur first.
	for _, op := range []string{
		"%=", "&=", "*=", "+=", "-=", "<<=", ">>=", "^=", "|=", "/=",
		"&&", "||", "++", "--", "->", "<<", ">>",
		"==", "!=", "<=", ">="} {
		b := int(op[0])
		scandata[b] = append(scandata[b], scanEntry{op[1:], OTHER, nil})
	}
}

func (t *tokenizer) readRaw() token {
	in := t.input
	n := len(in)
	if n < 1 {
		return token{ttype: END, line: t.line, column: t.column + 1}
	}

	end, ttype := 1, OTHER // If no match in scandata, token is next byte
	for _, e := range scandata[in[0]] {
		slen := 1 + len(e.suffix)
		if n >= slen && string(in[1:slen]) == e.suffix {
			if e.fn == nil {
				ttype = e.ttype
				end = slen
			} else {
				ttype, end = e.fn(in)
			}
			break
		}
	}

	tok := token{
		ttype:  ttype,
		line:   t.line,
		column: t.column + 1,
		text:   string(in[:end]),
	}
	t.input = in[end:]

	// Update line and column numbers.
	for _, c := range in[:end] {
		t.column++
		if c == '\t' {
			for t.column%8 != 0 {
				t.column++
			}
		}
		if c == '\n' {
			t.line++
			t.column = 0
		}
	}

	return tok
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r' || b == '\n'
}

func isWordByte(b byte) bool {
	return inRange(b, 'a', 'z') ||
		inRange(b, 'A', 'Z') ||
		inRange(b, '0', '9') ||
		b == '_'
}

func inRange(b, first, last byte) bool {
	return b >= first && b <= last
}

func readDblString(in []byte) (tokenType, int)    { return readString(in, '"') }
func readSingleString(in []byte) (tokenType, int) { return readString(in, '\'') }

func readString(in []byte, delimiter byte) (tokenType, int) {
	// Caller guarantees in starts with string opener
	for i, n := 1, len(in); i < n; i++ {
		b := in[i]
		if b == delimiter {
			return STRING, i + 1
		}
		if b == '\\' {
			i++ // Escape next
		}
	}
	return STRING, len(in)
}

func readLineComment(in []byte) (tokenType, int) {
	for i := range in {
		if in[i] == '\n' {
			return COMMENT, i + 1 // Include \n in comment token
		}
	}
	return COMMENT, len(in)
}

func readMultiLineComment(in []byte) (tokenType, int) {
	// Caller guarantees in starts with "//"
	for i, n := 2, len(in); i+1 < n; i++ {
		if in[i] == '*' && in[i+1] == '/' {
			return COMMENT, i + 2
		}
	}
	return COMMENT, len(in)
}

func readWord(in []byte) (tokenType, int) {
	for i, n := 1, len(in); i < n; i++ {
		if !isWordByte(in[i]) {
			return WORD, i
		}
	}
	return WORD, len(in)
}

func readVar(in []byte) (tokenType, int) {
	ttype := OTHER
	end := 1
	if len(in) > 1 && isWordByte(in[1]) {
		ttype = VAR
		_, end = readWord(in[1:])
		end++
		// Optional trailing '*'
		if end < len(in) && in[end] == '*' {
			ttype = RVAR
			end++
		}
	}
	return ttype, end
}

func readSpaces(in []byte) (tokenType, int) {
	for i, n := 1, len(in); i < n; i++ {
		if !isSpace(in[i]) {
			return SPACE, i
		}
	}
	return SPACE, len(in)
}

func closerFor(opener string) string {
	switch opener {
	case "(":
		return ")"
	case "{":
		return "}"
	case "[":
		return "}"
	default:
		return ""
	}
}
