package main

import "testing"

func TestHumanBytes(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "0.0B"},
		{512, "512.0B"},
		{1024, "1.0KB"},
		{1536, "1.5KB"},
		{1048576, "1.0MB"},
		{1073741824, "1.0GB"},
	}
	for _, c := range cases {
		if got := humanBytes(c.in); got != c.want {
			t.Errorf("humanBytes(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestIsRedirect(t *testing.T) {
	for _, code := range []int{301, 302, 303, 307, 308} {
		if !isRedirect(code) {
			t.Errorf("isRedirect(%d) = false, want true", code)
		}
	}
	for _, code := range []int{200, 204, 400, 404, 500, 0} {
		if isRedirect(code) {
			t.Errorf("isRedirect(%d) = true, want false", code)
		}
	}
}
