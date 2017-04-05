package main

import "testing"

func TestParse(t *testing.T) {
	type test struct {
		input  string
		output string
	}
	for _, c := range []test{
		{"foo", "([foo])"},
		{"/*foo*/", "([])"}, // Empty input requires an END token
		{"a b c", "([a] [b] [c])"},
		{"a+b+c", "(([a] [+] [b]) [+] [c])"},
		{"a+b*c/d", "([a] [+] (([b] [*] [c]) [/] [d]))"},
		{"a*(b+c)", "([a] [*] ([(] ([b] [+] [c]) [)]))"},
		{"a(b)", "([a] [(] [b] [)])"},
		{"a(b,c)", "([a] [(] [b] [,] [c] [)])"},
		{"a(b)(c)", "(([a] [(] [b] [)]) [(] [c] [)])"},
		{"a(b(c))(d)", "(([a] [(] ([b] [(] [c] [)]) [)]) [(] [d] [)])"},
		// Following exhibit right-associativity.
		{"+-a", "([+] ([-] [a]))"},
		{"a=b=c", "([a] [=] ([b] [=] [c]))"},
	} {
		expect := c.output
		root := parse([]byte(c.input))
		out := root.String()
		if out != expect {
			t.Errorf("CParse(%#v):\nGot:\n%s\nExpect:\n%s\n", c.input, out, expect)
		}
		serial := string(root.serialize())
		if serial != c.input {
			t.Errorf("Serialize:\nGot:\n%s\nExpect:\n%s\n", serial, c.input)
		}
	}
}
