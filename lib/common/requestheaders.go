package common

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
)

const host = "api.coinmarketcap.com"
const origin = "https://coinmarketcap.com"
const referer = "https://coinmarketcap.com/"
const userAgent = "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Mobile Safari/537.3"

var requestId = strings.Replace(uuid.NewString(), "-", "", -1)

var CommonHeader = http.Header{
	"Accept":     {"application/json"},
	"Host":       {host},
	"Origin":     {origin},
	"Referer":    {referer},
	"User-Agent": {userAgent},
	"X-Request-Id": {requestId},
	"Sec-Fetch-Dest": {"empty"},
	"Sec-Fetch-Mode": {"no-cors"},
	"Sec-Fetch-Site": {"same-site"},
	"Connection": {"keep-alive"},
	"TE": {"trailers"},
}
