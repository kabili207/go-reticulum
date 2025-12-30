package main

import "testing"

func TestSplitCommand(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		want []string
	}{
		{`echo hello`, []string{"echo", "hello"}},
		{`echo "a b"`, []string{"echo", "a b"}},
		{`echo 'a b'`, []string{"echo", "a b"}},
		{`echo a\ b`, []string{"echo", "a b"}},
		{`echo "a\\b"`, []string{"echo", `a\b`}},
		{`echo "a\"b"`, []string{"echo", `a"b`}},
		{`echo 'a\b'`, []string{"echo", `a\b`}}, // backslash is literal inside single quotes
	}

	for _, tc := range cases {
		got, err := splitCommand(tc.in)
		if err != nil {
			t.Fatalf("splitCommand(%q) returned error: %v", tc.in, err)
		}
		if len(got) != len(tc.want) {
			t.Fatalf("splitCommand(%q) = %#v, want %#v", tc.in, got, tc.want)
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Fatalf("splitCommand(%q) = %#v, want %#v", tc.in, got, tc.want)
			}
		}
	}
}

func TestSplitCommandErrors(t *testing.T) {
	t.Parallel()

	for _, in := range []string{`echo "unterminated`, `echo 'unterminated`, `echo a\`} {
		_, err := splitCommand(in)
		if err == nil {
			t.Fatalf("splitCommand(%q) expected error", in)
		}
	}
}
