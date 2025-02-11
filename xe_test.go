package main

import (
	"testing"
	"time"
	"strings"
)

func TestParseDate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"empty string", "", "", true},
		{"invalid short", "20", "", true},
		{"basic date", "20200101", "2020-01-01", false},
		{"invalid format", "2020011", "", true},
		{"with hyphens", "2020-01-01", "2020-01-01", false},
		{"with dots", "2020.01.01", "2020-01-01", false},
		{"with slashes", "2020/01/01", "2020-01-01", false},
		{"short 2000s", "200101", "2020-01-01", false},
		{"short 1900s", "880101", "1988-01-01", false},
		{"future date", time.Now().AddDate(1, 0, 0).Format("20060102"), "", true},
		{"far future", "3000-01-01", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateCurrency(t *testing.T) {
	tests := []struct {
		name    string
		ccy     string
		wantErr bool
	}{
		{"valid USD", "USD", false},
		{"valid lowercase", "usd", false},
		{"valid mixed case", "UsD", false},
		{"invalid currency", "XYZ", true},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCurrency(tt.ccy)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCurrency() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr bool
	}{
		{
			name: "valid options",
			opts: Options{
				FromCCY: "USD",
				ToCCY:   "EUR",
				ConvertAmount: 100,
			},
			wantErr: false,
		},
		{
			name: "same currency",
			opts: Options{
				FromCCY: "USD",
				ToCCY:   "USD",
			},
			wantErr: true,
		},
		{
			name: "invalid from currency",
			opts: Options{
				FromCCY: "XYZ",
				ToCCY:   "USD",
			},
			wantErr: true,
		},
		{
			name: "negative amount",
			opts: Options{
				FromCCY: "USD",
				ToCCY:   "EUR",
				ConvertAmount: -100,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOptions(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCrawl(t *testing.T) {
	tests := []struct {
		name    string
		from    string
		to      string
		date    string
		wantErr bool
	}{
		{"valid request", "RUB", "USD", "2023-01-02", false},
		{"case insensitive", "rub", "usd", "2023-01-02", false}, 
		{"invalid date", "RUB", "USD", "2023-13-99", true},
		{"too old date", "RUB", "USD", "2003-01-02", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Crawl(tt.from, tt.to, tt.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("Crawl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if strings.ToUpper(result.CcyFrom) != strings.ToUpper(tt.from) || 
				   strings.ToUpper(result.CcyTo) != strings.ToUpper(tt.to) || 
				   result.RateDate != tt.date {
					t.Errorf("Crawl() = %+v, expected currencies %s/%s and date %s", 
						result, tt.from, tt.to, tt.date)
				}
				if result.Rate == 0 {
					t.Error("Crawl() returned zero rate for valid request")
				}
			}
		})
	}
}
