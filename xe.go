package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/gocolly/colly"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Options - Command line arguments
type Options struct {
	FromCCY  string `long:"from" short:"f" description:"Convert FROM, dafaults to RUB" default:"RUB"`
	ToCCY    string `long:"to" short:"t" description:"Convert TO, dafaults to USD" default:"USD"`
	RateDate string `long:"date" short:"d" description:"Date to get the FX rate for, must be in YYYY-MM-DD format"`
}

// CurrencyPair - structure to hold currency pair info
type CurrencyPair struct {
	CcyFrom  string
	CcyTo    string
	Rate     float64
	RateDate string
}

// The crawler -- fetch page and parse rate info from it
func crawl(cf string, ct string, dt string) (CurrencyPair, error) { //limPages int, limPosts int, db *sql.DB) {
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
		log.Fatal("Unable to fetch", xeUrl)
	}

	if cp.CcyFrom != "" {
		return cp, nil
	} else {
		return cp, errors.New("No rate information for this date, likely incorrect date: " + dt)
	}
}

// Show help message and quit
func showHelp() {
	fmt.Println("Usage:\n  xe.com [OPTIONS]\n\n" +
		"Application Options:\n" +
		"  -f, --from= Convert FROM, dafaults to RUB (default: RUB)\n" +
		"  -t, --to=   Convert TO, dafaults to USD (default: USD)\n" +
		"  -d, --date= Date to get the rate for, must be in YYYY-MM-DD format\n\n" +
		"Short form of xe.com DATE is also supported")
	os.Exit(0)
}

// While we ask for a YYYY-MM-DD date, we will understand any other sensible date delimiter (hyphen, slash, dot).
// We also verify that date looks valid.
func parseDate(dt string) (string, error) {
	r, err := regexp.Compile("([0-9]{4})[/.-]*([0-9]{2})[/.-]*([0-9]{2})")
	if err != nil {
		return dt, err
	}

	m := r.FindStringSubmatchIndex(dt)

	if m != nil {
		validDate := dt[m[2]:m[3]] + "-" + dt[m[4]:m[5]] + "-" + dt[m[6]:m[7]]
		_, err = time.Parse("2006-01-02", validDate)

		if err != nil {
			return dt, err
		}

		return validDate, nil

	}

	return dt, errors.New("No valida date provided: " + dt)
}

func main() {
	var opts Options
	var rates CurrencyPair

	flag.StringVar(&opts.ToCCY, "t", "USD", "Convert TO, defaults to USD (short)")
	flag.StringVar(&opts.ToCCY, "to", "USD", "Convert TO, defaults to USD")
	flag.StringVar(&opts.FromCCY, "f", "RUB", "Convert FROM, defaults to RUB (short)")
	flag.StringVar(&opts.FromCCY, "from", "RUB", "Convert FROM, defaults to RUB")
	flag.StringVar(&opts.RateDate, "d", "", "Date to get the rate for, must be in YYYY-MM-DD format (short)")
	flag.StringVar(&opts.RateDate, "date", "", "Date to get the rate for, must be in YYYY-MM-DD format")
	flag.Parse()

	// catching if options parsing did not yield a sensible result
	if len(opts.RateDate) == 0 {
		if len(os.Args) == 2 { // xe.com DATE
			opts.RateDate = os.Args[1]
		} else if len(os.Args) == 4 { // xe.com FROM TO DATE
			opts.FromCCY = os.Args[1]
			opts.ToCCY = os.Args[2]
			opts.RateDate = os.Args[3]
		} else if len(os.Args) == 6 { // xe.com -f FROM -t TO DATE
			opts.RateDate = os.Args[5]
		} else {
			showHelp()
		}
	}

	rd, err := parseDate(opts.RateDate)

	if err != nil {
		showHelp()
		log.Fatal(err, "Wrong date format, must be YYYY-MM-DD, got: ", opts.RateDate)
	} else {
		rates = crawl(opts.FromCCY, opts.ToCCY, rd)
	}

	fmt.Printf("%s rate: %.8f %s per 1 %s\n", rd, rates.Rate, rates.CcyFrom, rates.CcyTo)
}
