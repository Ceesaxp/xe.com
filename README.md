# xe.com

Small utility to fetch xe.com rates for a given currency pair and date. In a strict form expects to be called as: `xe.com --from CCY --to CCY --date YYYY-MM-DD`. But a few short forms are also available. `CCY` is an [ISO-4217](https://en.wikipedia.org/wiki/ISO_4217) alphabetic code.

## Usage
`xe.com [OPTIONS]`

## Application Options

```
  -f, --from=   Convert FROM, dafaults to RUB (default: RUB)
  -t, --to=     Convert TO, dafaults to USD (default: USD)
  -d, --date=   Date to get the rate for, must be in YYYY-MM-DD format
  -s, --strip   Strip the trailing zeros from the result
  -a, --amount= Amount to convert, defaults to 1 (default: 1)
```

Short form of `xe.com DATE` or `xe.com FROM TO DATE` are also supported.

## Examples

`xe.com -f USD -t GBP -d 2021-03-17`
> \# will return the XE.COM close rate for USD/GBP pair on 17 Mar 2021

`xe.com usd gbp 2021-03-17`
> \# the same, but shorter

`xe.com -d 2021-03-17`
> \# the same, but for the default pair (RUB/USD)

`xe.com 2021-03-17`
> \# the same in the shortest form possible

`xe.com 21.03.17`
> \# you **can** do this too, but I am not guessing whether this is `YY.MM.DD` or `DD.MM.YY` -- use at your own risk

`xe.com -s FROM TO DATE`

> \# this form allows chaining with the likes of `bc(1)` to make calculations, e.g. `echo "$(~/bin/xe.com -s rub eur 20220510)*1234.56" | bc `

`xe.com rub eur 20220510 1234.56`

> \# will convert from RUB amount of 1234.56 to EUR with a rate as of 10 May 2022