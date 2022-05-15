package main

import "testing"

func TestParseDate(t *testing.T) {
	args := []struct {
		got  string
		want string
		//err error
	}{
		{"", ""},
		{"20", ""},
		{"20200101", "2020-01-01"},
		{"2020011", ""},
		{"2020-01-01", "2020-01-01"},
		{"2020.01.01", "2020-01-01"},
		{"2020/01/01", "2020-01-01"},
		{"200101", "2020-01-01"},
		{"880101", "1988-01-01"},
		{"790101", "2079-01-01"},
	}

	for _, arg := range args {
		r, _ := ParseDate(arg.got)
		if r != arg.want {
			t.Errorf("got %q, wanted %q", r, arg.want)
		}
	}
}
