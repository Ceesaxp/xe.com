# xe.com

Small utility to fetch xe.com rates for a given currency pair and date. In a strict form expects to be called as: `xe.com --from CCY --to CCY --date YYYY-MM-DD`. But a few short forms are also available. `CCY` is an [ISO-4217](https://en.wikipedia.org/wiki/ISO_4217) alphabetic code.

## Usage
`xe.com [OPTIONS]`

## Application Options

```
  -f, --from-ccy      Convert FROM, defaults to RUB (default: RUB)
  -t, --to-ccy        Convert TO, defaults to USD (default: USD)
  -d, --date          Date to get the rate for, must be in YYYY-MM-DD format
  -s, --strip-extra   Returns only the rate, good for use for shell scripting
  -a, --amount        Optionally, provide amount to convert
  -m, --math <expr>   Calculate the amount from the expression <expr>, then treat as -a
  -R, --show-rate     Always show rate
  -v, --version       Print version information and quit
```

Short form of `xe.com <DATE>` or `xe.com <FROM> <TO> <DATE>` are also supported.

## Environment

You can set variables `XE_CCY_FROM` and `XE_CCY_TO` in order to define your default from/to pair. Add this to
your `.bashrc` to have default pair EUR/USD:

```shell
export XE_CCY_FROM=EUR
export XE_CCY_TO=USD
```

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

## Supported Currencies

The tool supports common ISO 4217 currency codes including:
- USD (US Dollar)
- EUR (Euro) 
- GBP (British Pound)
- JPY (Japanese Yen)
- AUD (Australian Dollar)
- CAD (Canadian Dollar)
- CHF (Swiss Franc)
- CNY (Chinese Yuan)
- RUB (Russian Ruble)
- INR (Indian Rupee)
