package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/gocolly/colly"
	"github.com/jessevdk/go-flags"
)

type Options struct {
	FromCCY  string `long:"from" short:"f" description:"Convert FROM, dafaults to RUB" default:"RUB"`
	ToCCY    string `long:"to" short:"t" description:"Convert TO, dafaults to USD" default:"USD"`
	RateDate string `long:"date" short:"d" description:"Date to get the rate for, must be in YYYY-MM-DD format"`
}

type CurrencyPair struct {
	CcyFrom string
	CcyTo   string
	Rate    float64
}

// The crawler
func crawl(cf string, ct string, dt string) CurrencyPair { //limPages int, limPosts int, db *sql.DB) {
	xeUrl := "https://www.xe.com/currencytables/?from=" + cf + "&date=" + dt

	c := colly.NewCollector(
		colly.AllowedDomains("www.xe.com"),
	)

	var cp CurrencyPair

	c.OnHTML("div#table-section > section > div > div > table > tbody > tr", func(e *colly.HTMLElement) {
		if e.ChildText("th > a") == ct {
			cp.CcyFrom = cf
			cp.CcyTo = ct
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

	return cp
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

func main() {
	var opts Options
	var rates CurrencyPair

	parser := flags.NewParser(&opts, flags.Default)
	_, err := parser.Parse()

	if err != nil {
		log.Fatal(err)
	}

	if len(opts.RateDate) == 0 {
		if len(os.Args) > 1 {
			opts.RateDate = os.Args[1]
		} else {
			showHelp()
		}
	}

	r, _ := regexp.Compile("[0-9]{4}-[0-9]{2}-[0-9]{2}")

	if r.MatchString(opts.RateDate) {
		rates = crawl(opts.FromCCY, opts.ToCCY, opts.RateDate)
	} else {
		showHelp()
		log.Fatal("Wrong date format, must be YYYY-MM-DD, got: ", opts.RateDate)
	}

	fmt.Printf("On %s from: %s to: %s rate: %.10f\n", opts.RateDate, rates.CcyFrom, rates.CcyTo, rates.Rate)
}
