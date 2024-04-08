package common

import "net/http"

const host = "api.coinmarketcap.com"
const origin = "https://coinmarketcap.com"
const referer = "https://coinmarketcap.com/"
const userAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:124.0) Gecko/20100101 Firefox/124.0"

var CommonHeader = http.Header{
	"Accept":     {"application/json"},
	"Host":       {host},
	"Origin":     {origin},
	"Referer":    {referer},
	"User-Agent": {userAgent},
}
