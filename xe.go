package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/gocolly/colly"
)

var version = "0.0.3"

const (
    envFromCcyKey = "XE_CCY_FROM"
    envToCcyKey   = "XE_CCY_TO"
    defaultFromCcy = "RUB"
    defaultToCcy  = "USD"
)

// Valid ISO 4217 currency codes (common ones - expand as needed)
var validCurrencies = map[string]bool{
    "USD": true, "EUR": true, "GBP": true, "JPY": true, "AUD": true,
    "CAD": true, "CHF": true, "CNY": true, "RUB": true, "INR": true,
	"KZT": true, "PLN": true, "RSD": true, "CZK": true,
	// Add more as needed
}

// Options - Command line arguments
type Options struct {
	FromCCY       string  `long:"from" short:"f" description:"Convert FROM, defaults to RUB" default:"RUB"`
	ToCCY         string  `long:"to" short:"t" description:"Convert TO, defaults to USD" default:"USD"`
	RateDate      string  `long:"date" short:"d" description:"Date to get the FX rate for, must be in YYYY-MM-DD format"`
	Strip         bool    `long:"strip" short:"s" description:"Remove all clutter and return only rate"`
	ConvertAmount float64 `long:"amount" short:"a" description:"Amount to covert"`
	Version       bool    `long:"version" short:"v" description:"Show version and quit"`
	Math          string  `long:"math" short:"m" description:"Perform math operations"`
	ShowRate      bool    `long:"show-rate" short:"R" description:"Always show rate only"`
}

// CurrencyPair - structure to hold currency pair info
type CurrencyPair struct {
	CcyFrom  string
	CcyTo    string
	Rate     float64
	RateDate string
}

// Crawl fetches page and parses rate info from it
func Crawl(cf string, ct string, dt string) (CurrencyPair, error) {
	// Validate date format first
	if _, err := time.Parse("2006-01-02", dt); err != nil {
		return CurrencyPair{}, fmt.Errorf("invalid date format: %v", err)
	}

	cf = strings.ToUpper(cf)
	ct = strings.ToUpper(ct)
	xeUrl := "https://www.xe.com/currencytables/?from=" + cf + "&date=" + dt

	c := colly.NewCollector(
		colly.AllowedDomains("www.xe.com"),
	)

	var cp CurrencyPair
	var scrapeErr error

	c.OnHTML("div#table-section > section > div > div > table > tbody > tr", func(e *colly.HTMLElement) {
		if e.ChildText("th > a") == ct {
			cp.CcyFrom = cf
			cp.CcyTo = ct
			cp.RateDate = dt
			rate, err := strconv.ParseFloat(e.ChildText("td:nth-of-type(2)"), 64)
			if err != nil {
				scrapeErr = fmt.Errorf("failed to parse rate: %w", err)
				return
			}
			cp.Rate = rate
		}
	})

	if err := c.Visit(xeUrl); err != nil {
		return cp, fmt.Errorf("failed to fetch %s: %w", xeUrl, err)
	}

	if scrapeErr != nil {
		return cp, scrapeErr
	}

	if cp.CcyFrom == "" {
		return cp, fmt.Errorf("no rate information for date %s", dt)
	}

	return cp, nil
}

// ShowHelp – show the help message and quit
func ShowHelp() {
	fmt.Println("Usage:\n  xe.com [OPTIONS]\n\n" +
		"Application Options:\n" +
		"  -f, --from=       Convert FROM, defaults to RUB (default: RUB)\n" +
		"  -t, --to=         Convert TO, defaults to USD (default: USD)\n" +
		"  -d, --date=       Date to get the rate for, must be in YYYY-MM-DD format\n" +
		"  -s, --strip       Returns only the rate, good for use for shell scripting\n" +
		"  -a, --amount      Optionally, provide amount to convert\n\n" +
		"  -m, --math <expr> Calculate the amount from the expression <expr>, then treat as -a\n\n" +
		"  -R, --show-rate   Always show rate" +
		"Short form of xe.com DATE is also supported")
	os.Exit(0)
}

// ParseDate : While we ask for a YYYY-MM-DD date, we will understand any other sensible date delimiter
// (hyphen, slash, dot). We also verify that date looks valid.
func ParseDate(RateDate string) (string, error) {
	re := regexp.MustCompile(`^(\d{2}|\d{4})[./-]*(\d{2})[./-]*(\d{2})$`)
	// (\d{2} - 2 digits, | - or, \d{4} - 4 digits)
	if re.MatchString(RateDate) {
		match := re.FindStringSubmatch(RateDate)
		if match == nil {
			return RateDate, errors.New("can't find valid date")
		}

		st := match[1]
		if len(st) == 2 { // lazy bitches, dangerous!
			if st > "79" { // 1980...
				st = "19" + st
			} else { // 2000...
				st = "20" + st
			}
		}
		st = st + "-" + match[2] + "-" + match[3]

		_, err := time.Parse("2006-01-02", st)

		if err != nil {
			log.Print(err)
			return "", err
		}

		// set to today's date
		t := time.Now()
		if st <= t.Format("2006-01-02") {
			return st, nil
		} else {
			return "", errors.New("provided date is in future")
		}
	}
	return "", errors.New("wrong date string")
}

// validateCurrency checks if the provided currency code is valid
func validateCurrency(ccy string) error {
    ccy = strings.ToUpper(ccy)
    if !validCurrencies[ccy] {
        return fmt.Errorf("invalid currency code: %s", ccy)
    }
    return nil
}

// parseCommandLineArgs parses and validates command line arguments
func parseCommandLineArgs() (Options, error) {
    var opts Options

    envFromCcy := os.Getenv(envFromCcyKey)
    if envFromCcy == "" {
        envFromCcy = defaultFromCcy
    }

    envToCcy := os.Getenv(envToCcyKey)
    if envToCcy == "" {
        envToCcy = defaultToCcy
    }

    flag.StringVarP(&opts.ToCCY, "to-ccy", "t", envToCcy, "Convert TO, defaults to USD (short)")
    flag.StringVarP(&opts.FromCCY, "from-ccy", "f", envFromCcy, "Convert FROM, defaults to RUB (short)")
    flag.StringVarP(&opts.RateDate, "date", "d", "", "Date to get the rate for, must be in YYYY-MM-DD format (short)")
    flag.BoolVarP(&opts.Strip, "strip-extra", "s", false, "Returns only the rate, good for use for shell scripting")
    flag.Float64VarP(&opts.ConvertAmount, "amount", "a", 0, "Optionally, provide amount to convert")
    flag.BoolVarP(&opts.Version, "version", "v", false, "Print version information and quit")
    flag.StringVarP(&opts.Math, "math", "m", "", "Calculate the amount from the expression <expr>, then treat as -a")
    flag.BoolVarP(&opts.ShowRate, "show-rate", "R", false, "Always show rate")
    flag.Parse()

    return opts, validateOptions(opts)
}

// validateOptions validates the parsed command line options
func validateOptions(opts Options) error {
    if err := validateCurrency(opts.FromCCY); err != nil {
        return fmt.Errorf("from currency error: %w", err)
    }
    if err := validateCurrency(opts.ToCCY); err != nil {
        return fmt.Errorf("to currency error: %w", err)
    }
    if opts.ConvertAmount < 0 {
        return errors.New("amount cannot be negative")
    }
    if opts.FromCCY == opts.ToCCY {
        return errors.New("from and to currencies must be different")
    }
    return nil
}

// handleShortFormArgs processes arguments when used in short form
func handleShortFormArgs(opts *Options) {
    if len(opts.RateDate) == 0 {
        switch len(os.Args) {
        case 2: // xe.com DATE
            opts.RateDate = os.Args[1]
        case 4: // xe.com FROM TO DATE
            opts.FromCCY = os.Args[1]
            opts.ToCCY = os.Args[2]
            opts.RateDate = os.Args[3]
        case 5:
            if os.Args[1] == "-s" { // xe.com -s FROM TO DATE
                opts.Strip = true
                opts.FromCCY = os.Args[2]
                opts.ToCCY = os.Args[3]
                opts.RateDate = os.Args[4]
            } else { // xe.com FROM TO DATE AMOUNT
                opts.FromCCY = os.Args[1]
                opts.ToCCY = os.Args[2]
                opts.RateDate = os.Args[3]
                opts.ConvertAmount, _ = strconv.ParseFloat(os.Args[4], 64)
            }
        case 6: // xe.com -f FROM -t TO DATE
            opts.RateDate = os.Args[5]
        default:
            ShowHelp()
        }
    }
}

// formatOutput formats and prints the final output based on options
func formatOutput(opts Options, rates CurrencyPair) {
    if opts.Strip {
        fmt.Printf("%.8f", rates.Rate)
    } else if opts.ConvertAmount > 0 {
        fmt.Printf("%s %s %.2f = %s %.2f", opts.RateDate, opts.FromCCY, opts.ConvertAmount,
            opts.ToCCY, opts.ConvertAmount*rates.Rate)
        if opts.ShowRate {
            fmt.Printf(" (rate: %.8f)\n", rates.Rate)
        }
    } else {
        fmt.Printf("%s rate: %.8f %s per 1 %s\n", rates.RateDate, rates.Rate, rates.CcyFrom, rates.CcyTo)
    }
}

// HandleMath - handle math expression
func HandleMath(expr string) float64 {
    return 0.0
}

// main is the entry point of the xe.com currency converter program.
// It parses command line arguments, retrieves currency exchange rates from xe.com,
// and outputs the converted amount or exchange rate.
func main() {
    opts, err := parseCommandLineArgs()
    if err != nil {
        log.Print(err)
        ShowHelp()
    }

    if opts.Version {
        fmt.Println("xe.com version", version)
        os.Exit(0)
    }

    if opts.Math != "" {
        opts.ConvertAmount = HandleMath(opts.Math)
    }

    handleShortFormArgs(&opts)

    rd, err := ParseDate(opts.RateDate)
    if err != nil {
        log.Printf("Wrong date format, must be YYYY-MM-DD, got: %s: %v", opts.RateDate, err)
        ShowHelp()
    }

    rates, err := Crawl(opts.FromCCY, opts.ToCCY, rd)
    if err != nil {
        log.Print(err)
        ShowHelp()
    }

    formatOutput(opts, rates)
}
