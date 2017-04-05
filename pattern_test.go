package main

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

func TestMatch(t *testing.T) {
	type test struct {
		subject string
		pattern string
		start   int
		limit   int
		assign  []string
	}

	// Helper to create a test case.
	tcase := func(sub, pattern string, start, limit int, assign ...string) test {
		return test{sub, pattern, start, limit, assign}
	}

	for _, c := range []test{
		// Token match
		tcase("x", "x", 0, 1),
		tcase("x y z", "x y z", 0, 3),
		tcase("x y z", "x y", 0, 2),
		tcase("x y z", "y", 1, 2),

		// Non-repeating variable match
		tcase("x", "$a", 0, 1, "$a => x"),

		// Match variable with context
		tcase("x y z", "x $a z", 0, 3, "$a => y "),

		// Match variable list
		tcase("x y", "x $a*", 0, 2, "$a* => y"),
		tcase("x y z", "x $a*", 0, 3, "$a* => y z"),
		tcase("x", "x $a*", 0, 1, "$a* => "),

		// Match multiple variables and lists
		tcase("x 1 y 2 3 z 4 5", "x $a y $b* z $c*", 0, 8,
			"$a => 1 ", "$b* => 2 3 ", "$c* => 4 5"),

		// Subtree failure
		tcase("x(y,z)", "x($a,w)", -1, -1),

		// Subtree match
		tcase("x(y,z)", "x($a,z)", 0, 6, "$a => y"),

		// Deep subtree
		tcase("x(y(z))", "x($a)", 0, 4, "$a => y(z)"),
	} {
		//fmt.Fprintln(os.Stderr, "X", c.subject, c.pattern)
		expect := strings.Join(c.assign, "\n")

		s := parse([]byte(c.subject))
		if len(s.children) == 0 {
			s = &node{children: []*node{s}}
		}
		p := parse([]byte(c.pattern))
		if len(p.children) == 0 {
			p = &node{children: []*node{p}}
		}
		pat := makePattern(p)

		assign, ok := pat.match(s.children)
		if !ok {
			if c.start >= 0 {
				t.Errorf("Match(%s, %s):\nActual: no match\nExpect: %s\n",
					c.subject, c.pattern, expect)
			}
			continue
		}
		if assign.start != c.start || assign.limit != c.limit {
			t.Errorf("Match(%s, %s):\nExtent: %d..%d\nExpect: %d..%d\n",
				c.subject, c.pattern, assign.start, assign.limit, c.start, c.limit)
			continue
		}

		// Convert assignment into a sorted list of elements of
		// the form "$var => ...assignment...".
		vars := make([]string, 0)
		for k, v := range assign.vars {
			m := &node{children: v}
			str := fmt.Sprintf("%s => %s", k, m.serialize())
			vars = append(vars, str)
		}
		sort.Strings(vars)
		got := strings.Join(vars, "\n")

		if got != expect {
			t.Errorf("Match(%s, %s):\nActual:\n%s\nExpect:\n%s\n",
				c.subject, c.pattern, got, expect)
		}
	}
}
