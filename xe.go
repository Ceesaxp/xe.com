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
	FromCCY  string `long:"from" short:"f" description:"Convert FROM, defaults to RUB" default:"RUB"`
	ToCCY    string `long:"to" short:"t" description:"Convert TO, defaults to USD" default:"USD"`
	RateDate string `long:"date" short:"d" description:"Date to get the FX rate for, must be in YYYY-MM-DD format"`
	Strip    bool   `long:"strip" short:"s" description:"Remove all clutter and return only rate"`
}

// CurrencyPair - structure to hold currency pair info
type CurrencyPair struct {
	CcyFrom  string
	CcyTo    string
	Rate     float64
	RateDate string
}

// Crawl fetch page and parse rate info from it
func Crawl(cf string, ct string, dt string) (CurrencyPair, error) { //limPages int, limPosts int, db *sql.DB) {
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
		log.Fatal("Unable to fetch", xeUrl)
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
		"  -f, --from= Convert FROM, defaults to RUB (default: RUB)\n" +
		"  -t, --to=   Convert TO, defaults to USD (default: USD)\n" +
		"  -d, --date= Date to get the rate for, must be in YYYY-MM-DD format\n" +
		"  -s, --strip Returns only the rate, good for use for shell scripting\n\n" +
		"Short form of xe.com DATE is also supported")
	os.Exit(0)
}

// ParseDate : While we ask for a YYYY-MM-DD date, we will understand any other sensible date delimiter
// (hyphen, slash, dot). We also verify that date looks valid.
func ParseDate(RateDate string) (string, error) {
	re := regexp.MustCompile(`(\d{2,4})[./-]*(\d{2})[./-]*(\d{2})$`)
	if re.MatchString(RateDate) {
		match := re.FindStringSubmatch(RateDate)
		if match == nil {
			return RateDate, errors.New("can't find valid date")
		}

		st := match[1]
		if len(match[1]) == 2 { // lazy bitches, dangerous!
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

		return st, nil
	}
	return "", errors.New("wrong date string")
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
	flag.BoolVar(&opts.Strip, "s", false, "Returns only the rate, good for use for shell scripting")
	flag.Parse()

	// catching if options parsing did not yield a sensible result
	if len(opts.RateDate) == 0 {
		if len(os.Args) == 2 { // xe.com DATE
			opts.RateDate = os.Args[1]
		} else if len(os.Args) == 4 { // xe.com FROM TO DATE
			opts.FromCCY = os.Args[1]
			opts.ToCCY = os.Args[2]
			opts.RateDate = os.Args[3]
		} else if len(os.Args) == 5 { // xe.com -s FROM TO DATE
			opts.Strip = true
			opts.FromCCY = os.Args[2]
			opts.ToCCY = os.Args[3]
			opts.RateDate = os.Args[4]
		} else if len(os.Args) == 6 { // xe.com -f FROM -t TO DATE
			opts.RateDate = os.Args[5]
		} else {
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
	} else {
		fmt.Printf("%s rate: %.8f %s per 1 %s\n", rd, rates.Rate, rates.CcyFrom, rates.CcyTo)
	}
}
