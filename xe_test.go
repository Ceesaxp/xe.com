package main

import (
	"math"
	"strings"
	"testing"
)

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
		{"790101", ""}, // Actually, this is 2079-01-01, but we don't support that far in the future
		{"3000-01-01", ""},
	}

	for tc, arg := range args {
		r, _ := ParseDate(arg.got)
		if r != arg.want {
			t.Errorf("[%d]: got %q, wanted %q", tc, r, arg.want)
		}
	}
}

func TestUserAgentIsNotCollyDefault(t *testing.T) {
	t.Setenv("XE_USER_AGENT", "")

	got := userAgent()

	if got == "" {
		t.Fatal("user agent is empty, xe.com's CDN rejects the colly default")
	}
	if strings.Contains(strings.ToLower(got), "colly") {
		t.Errorf("user agent %q looks like the colly default, which xe.com answers with 403", got)
	}
}

func TestUserAgentHonoursEnvironmentOverride(t *testing.T) {
	t.Setenv("XE_USER_AGENT", "my-own-agent/1.0")

	if got := userAgent(); got != "my-own-agent/1.0" {
		t.Errorf("got %q, wanted the value of XE_USER_AGENT", got)
	}
}

// TestCrawl talks to xe.com, so it is skipped under -short.
func TestCrawl(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test that fetches from xe.com")
	}

	var args = []struct {
		name     string
		cf       string
		ct       string
		dt       string
		wantRate float64
		wantErr  bool
	}{
		{"knownPair", "USD", "EUR", "2013-01-02", 0.75779724050467, false},
		{"lowerCaseCurrencies", "usd", "eur", "2013-01-02", 0.75779724050467, false},
		{"impossibleDate", "USD", "EUR", "2013-13-02", 0, true},
		{"unknownTargetCurrency", "USD", "ZZZ", "2013-01-02", 0, true},
		{"currencyXeNoLongerPublishes", "RUB", "USD", "2013-01-02", 0, true},
	}

	for _, arg := range args {
		t.Run(arg.name, func(t *testing.T) {
			got, err := Crawl(arg.cf, arg.ct, arg.dt)

			if arg.wantErr {
				if err == nil {
					t.Fatalf("got rate %v, wanted an error", got.Rate)
				}
				return
			}
			if err != nil {
				t.Fatalf("got error %v, wanted rate %v", err, arg.wantRate)
			}
			if math.Abs(got.Rate-arg.wantRate) > 1e-9 {
				t.Errorf("got rate %v, wanted %v", got.Rate, arg.wantRate)
			}
			if got.RateDate != arg.dt {
				t.Errorf("got date %q, wanted %q", got.RateDate, arg.dt)
			}
		})
	}
}
