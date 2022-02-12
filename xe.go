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
)

type Options struct {
	FromCCY  string `long:"from" short:"f" description:"Convert FROM, dafaults to RUB" default:"RUB"`
	ToCCY    string `long:"to" short:"t" description:"Convert TO, dafaults to USD" default:"USD"`
	RateDate string `long:"date" short:"d" description:"Date to get the rate for, must be in YYYY-MM-DD format"`
}

type CurrencyPair struct {
	CcyFrom  string
	CcyTo    string
	Rate     float64
	RateDate string
}

// The crawler
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

func showHelp() {
	fmt.Println("Usage:\n  xe.com [OPTIONS]\n\n" +
		"Application Options:\n" +
		"  -f, --from= Convert FROM, dafaults to RUB (default: RUB)\n" +
		"  -t, --to=   Convert TO, dafaults to USD (default: USD)\n" +
		"  -d, --date= Date to get the rate for, must be in YYYY-MM-DD format\n\n" +
		"Short form of xe.com DATE is also supported")
	os.Exit(0)
}

func CheckValidDate(RateDate string) (string, error) {
	// first replace 'wrong' delimiters, e.g. slashes or dots, with a hyphen
	re := regexp.MustCompile(`[/.]`)
	st := re.ReplaceAllString(RateDate, "-")

	re = regexp.MustCompile(`(\d{2,4})-\d{2}-\d{2}$`)
	if re.MatchString(st) {
		match := re.FindStringSubmatch(st)
		if len(match[1]) == 2 { // lazy bitches, dangerous!
			if st > "79" { // 1980...
				st = "19" + st
			} else { // 2000...
				st = "20" + st
			}
		}
		return st, nil
	} else {
		return "", errors.New("Cannot parse date as string, check format: " + RateDate)
	}
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

	RateDate, err := CheckValidDate(opts.RateDate)
	if err != nil {
		log.Fatal(err)
	}

	rates, err = crawl(strings.ToUpper(opts.FromCCY), strings.ToUpper(opts.ToCCY), RateDate)
	if err != nil {
		log.Print(err)
		showHelp()
		return
	} else {
		fmt.Printf("%s to %s on %s at: %.10f\n", rates.CcyFrom, rates.CcyTo, rates.RateDate, rates.Rate)
	}
}
