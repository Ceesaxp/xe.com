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

var version = "0.0.2"

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

// Crawl fetch page and parse rate info from it
func Crawl(cf string, ct string, dt string) (CurrencyPair, error) {
	cf = strings.ToUpper(cf)
	ct = strings.ToUpper(ct)
	xeUrl := "https://www.xe.com/currencytables/?from=" + cf + "&date=" + dt

	c := colly.NewCollector(
		colly.AllowedDomains("www.xe.com"),
	)

	var cp CurrencyPair

	c.OnHTML("div#table-section > section > div > div > table > tbody > tr", func(e *colly.HTMLElement) {
		if e.ChildText("th > a") == ct {
			cp.CcyFrom = cf
			cp.CcyTo = ct
			cp.RateDate = dt
			rate, err := strconv.ParseFloat(e.ChildText("td:nth-of-type(2)"), 64)
			if err == nil {
				cp.Rate = rate
			} else {
				log.Fatal(err)
			}
		}
	})

	err := c.Visit(xeUrl)
	if err != nil {
		log.Fatal("Unable to fetch ", xeUrl)
	}

	if cp.CcyFrom != "" {
		return cp, nil
	} else {
		return cp, errors.New("no rate information for this date, likely incorrect date: " + dt)
	}
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

// HandleMath - handle math expression
func HandleMath(expr string) float64 {
	return 0.0
}

// main is the entry point of the xe.com currency converter program.
// It parses command line arguments, retrieves currency exchange rates from xe.com,
// and outputs the converted amount or exchange rate.
func main() {
	var opts Options
	var rates CurrencyPair

	envFromCcy := os.Getenv("XE_CCY_FROM")
	if envFromCcy == "" {
		envFromCcy = "RUB"
	}

	envToCcy := os.Getenv("XE_CCY_TO")
	if envToCcy == "" {
		envToCcy = "USD"
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

	if opts.Version {
		fmt.Println("xe.com version", version)
		os.Exit(0)
	}

	if opts.Math != "" {
		opts.ConvertAmount = HandleMath(opts.Math)
	}

	// We expect that when called with option flags we get sensible request, but if a short form is used, we need to guess.
	// So, this here is for catching if options parsing did not yield a sensible result or we have a mixed call,
	// like xe.com -s FROM TO DATE
	if len(opts.RateDate) == 0 { // if date is not specified, we will use today's date
		if len(os.Args) == 2 { // assuming xe.com DATE and default currencies
			opts.RateDate = os.Args[1]
		} else if len(os.Args) == 4 { // xe.com FROM TO DATE
			opts.FromCCY = os.Args[1]
			opts.ToCCY = os.Args[2]
			opts.RateDate = os.Args[3]
		} else if len(os.Args) == 5 && os.Args[1] != "-s" { // xe.com FROM TO DATE AMOUNT
			opts.FromCCY = os.Args[1]
			opts.ToCCY = os.Args[2]
			opts.RateDate = os.Args[3]
			opts.ConvertAmount, _ = strconv.ParseFloat(os.Args[4], 64)
		} else if len(os.Args) == 5 { // xe.com -s FROM TO DATE
			opts.Strip = true
			opts.FromCCY = os.Args[2]
			opts.ToCCY = os.Args[3]
			opts.RateDate = os.Args[4]
		} else if len(os.Args) == 6 { // xe.com -f FROM -t TO DATE
			opts.RateDate = os.Args[5]
		} else { // When all else fails, show help
			ShowHelp()
		}
	}

	rd, err := ParseDate(opts.RateDate)

	if err != nil {
		log.Print(err, "Wrong date format, must be YYYY-MM-DD, got: ", opts.RateDate)
		ShowHelp()
	} else {
		rates, err = Crawl(opts.FromCCY, opts.ToCCY, rd)
		if err != nil {
			log.Print(err)
			ShowHelp()
		}
	}

	// note that we use 8-digit precision here, while we may be getting more than 8 decimals
	if opts.Strip {
		fmt.Printf("%.8f", rates.Rate)
	} else if opts.ConvertAmount > 0 {
		fmt.Printf("%s %s %.2f = %s %.2f", opts.RateDate, opts.FromCCY, opts.ConvertAmount,
			opts.ToCCY, opts.ConvertAmount*rates.Rate)
		if opts.ShowRate {
			fmt.Printf(" (rate: %.8f)\n", rates.Rate)
		}
	} else {
		fmt.Printf("%s rate: %.8f %s per 1 %s\n", rd, rates.Rate, rates.CcyFrom, rates.CcyTo)
	}
}
