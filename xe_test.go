package main

import "testing"

func TestParseDate(t *testing.T) {
	arg := ""
	got, err := ParseDate(arg)
	want := ""

	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
	if err == nil {
		t.Errorf("No error on %q", arg)
	}
}
