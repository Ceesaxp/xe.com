# xe.com

Small utility to fetch xe.com rates

## Usage
`xe.com [OPTIONS]`

## Application Options

```
  -f, --from= Convert FROM, dafaults to RUB (default: RUB)
  -t, --to=   Convert TO, dafaults to USD (default: USD)
  -d, --date= Date to get the rate for, must be in YYYY-MM-DD format
```

Short form of `xe.com DATE` is also supported

## Examples

`xe.com -f USD -t GBP -d 2021-03-17`
> \# will return the XE.COM close rate for USD/GBP pair on 17 Mar 2021

`xe.com -d 2021-03-17`
> \# the same, but for the default pair (RUB/USD)

`xe.com 2021-03-17`
> \# the same in the shortest form possible
