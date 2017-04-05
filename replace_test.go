package main

import "testing"

func TestReplace(t *testing.T) {
	type test struct {
		subject     string
		pattern     string
		replacement string
		output      string
	}
	for _, c := range []test{
		// Simple replacement.
		{"F", "F", "G", "G"},

		// Complicated replacement.
		{"x F(1,2,3) y", "F($a,$b,3)", "G($b,$a)", "x G(2,1) y"},

		// Repeated vars.
		{"f(a,b,c)", "f($x*)", "g($x*,$x*)", "g(a,b,c,a,b,c)"},

		// Multiple non-overlapping matches.
		{"x F F y F", "F", "X Y", "x X Y X Y y X Y"},
		{"x(F)(F)y(F)", "(F)", "(X,Y)", "x(X,Y)(X,Y)y(X,Y)"},

		// Nested match.
		{"(((x+y)))", "$a+$b", "$b+$a", "(((y+x)))"},
		{"(x+y)+1", "$a + 1", "increment($a)", "increment((x+y))"},

		// Should be applied twice bottom-up.
		{"x+(y+z)", "$a+$b", "$b+$a", "(z+y)+x"},
		{"(((x+(y+z))))", "$a+$b", "$b+$a", "((((z+y)+x)))"},

		// Should not replace inside replacement.
		{"x+y", "$a+$b", "$a/$b+0", "x/y+0"},

		// Overlapping match; should only replace once.
		{"x#y#z", "$a#$b", "$b#$a", "y#x#z"},

		// Comment copying.
		{"x/*foo*/+0", "$a+0", "$a", "x/*foo*/"},
		{"x+0/*foo*/", "$a+0", "$a", "x/*foo*/"},
		{"/*foo*/0+x", "0+$a", "$a", "/*foo*/x"},

		// Newlines must be preserved on either side.
		{"\nx", "x", "y", "\ny"},
		{"\nx+0", "$a+0", "$a", "\nx"},
		{"x\n", "x", "y", "y\n"},
		{"\n\n\nx\n\n", "x", "y", "\n\n\ny\n\n"},
		{"\n\n\nx y\n\n", "x y", "z", "\n\n\nz\n\n"},
	} {
		sub := parse([]byte(c.subject))
		pat := parse([]byte(c.pattern))
		rep := parse([]byte(c.replacement))
		replace(sub, pat, rep)
		out := string(sub.serialize())
		//fmt.Println("Result:", out)
		if out != c.output {
			t.Error("\nReplace:\n", c.pattern,
				"\nBy:\n", c.replacement,
				"\nIn:\n", c.subject,
				"\nGot:\n", out,
				"\nExpect:\n", c.output,
				"\n")
		}
	}

}
