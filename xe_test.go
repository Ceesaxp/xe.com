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

func TestCrawl(t *testing.T) {
	var okPair = CurrencyPair{"RUB", "USD", 12.345, "2003-01-02"}
	var failedPair = CurrencyPair{"RUB", "USD", 0, ""}
	var args = []struct {
		cf   string
		ct   string
		dt   string
		want CurrencyPair
	}{
		{"RUB", "USD", "2003-01-02", okPair},
		{"rub", "usd", "2003-01-02", okPair},
		{"RUB", "USD", "2003-13-02", failedPair},
	}
	for _, arg := range args {
		r, _ := Crawl(arg.cf, arg.ct, arg.dt)
		if r.RateDate != arg.want.RateDate {
			t.Errorf("got %q, wanted %q", r.RateDate, arg.want.RateDate)
		}
	}
}
