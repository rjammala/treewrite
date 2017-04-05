package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestTokenizer(t *testing.T) {
	type test struct {
		input  string
		output string
	}
	for _, c := range []test{
		// Empty input
		{"", ""},

		// Multi character operators:
		{`!=`, `(OTHER 1.1 !=)`},
		{`%=`, `(OTHER 1.1 %=)`},
		{`&&`, `(OTHER 1.1 &&)`},
		{`&=`, `(OTHER 1.1 &=)`},
		{`*=`, `(OTHER 1.1 *=)`},
		{`++`, `(OTHER 1.1 ++)`},
		{`+=`, `(OTHER 1.1 +=)`},
		{`--`, `(OTHER 1.1 --)`},
		{`-=`, `(OTHER 1.1 -=)`},
		{`->`, `(OTHER 1.1 ->)`},
		{`/=`, `(OTHER 1.1 /=)`},
		{`<<`, `(OTHER 1.1 <<)`},
		{`<<=`, `(OTHER 1.1 <<=)`},
		{`<=`, `(OTHER 1.1 <=)`},
		{`==`, `(OTHER 1.1 ==)`},
		{`>=`, `(OTHER 1.1 >=)`},
		{`>>`, `(OTHER 1.1 >>)`},
		{`>>=`, `(OTHER 1.1 >>=)`},
		{`^=`, `(OTHER 1.1 ^=)`},
		{`|=`, `(OTHER 1.1 |=)`},
		{`||`, `(OTHER 1.1 ||)`},

		// Various token types by themselves
		{`"foo"`, `(STRING 1.1 "foo")`},
		{`'foo'`, `(STRING 1.1 'foo')`},
		{"// foo\n", ""},
		{"/* foo\nbar */", ""},
		{"foo", "(WORD 1.1 foo)"},
		{" \t", ""},
		{"$x", "(VAR 1.1 $x)"},
		{"$x*", "(RVAR 1.1 $x*)"},

		// Combination
		{"a1b /*x\ny*/$a$b* 200", "(WORD 1.1 a1b)(VAR 2.4 $a)(RVAR 2.6 $b*)(WORD 2.10 200)"},

		// Early termination
		{`"foo`, `(STRING 1.1 "foo)`},
		{`'foo`, `(STRING 1.1 'foo)`},
		{"/* foo", ""},
	} {
		// Read tokens and produce string.
		tokenizer := newTokenizer([]byte(c.input))
		var buf bytes.Buffer
		for {
			tok := tokenizer.read()
			if tok.ttype == END {
				break
			}
			fmt.Fprint(&buf, tok)
		}
		result := buf.String()
		if result != c.output {
			t.Errorf("Tokenizer(%#v):\nGot:\n%s\nExpect:\n%s\n",
				c.input, result, c.output)
		}
	}
}
