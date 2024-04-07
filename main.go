package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const host = "api.coinmarketcap.com"
const origin = "https://coinmarketcap.com"
const referer = "https://coinmarketcap.com/"
const userAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:124.0) Gecko/20100101 Firefox/124.0"

var commonHeader = http.Header{
	"Accept":     {"application/json"},
	"Host":       {host},
	"Origin":     {origin},
	"Referer":    {referer},
	"User-Agent": {userAgent},
}

type StatusData struct {
	Timestamp    string `json:"timestamp"`
	ErrorMessage string `json:"error_message"`
}

const createdAtDateAdded = "dateAdded"
const createdAtChart = "firstTraded"
const createdAtMode = createdAtDateAdded

func main() {
	req, err := http.NewRequest("GET", listingEndpoint, nil)
	req.Header = commonHeader
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	var data ListingResponseBody
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		log.Fatalln(err)
	}
	cryptoCurrencyList := data.Data.CryptoCurrencyList
	totalCryptoCurrency := len(cryptoCurrencyList)
	// coinMarketData := make([]CoinMarketResponseBody, len(cryptoCurrencyList))

	/* get coin chart */
	beforeTime := time.Date(2022, 0, 0, 0, 0, 0, 0, time.UTC)
	if createdAtChart == createdAtMode {
		coinChartOverviewDataPayload := getCoinChartOverviewDataPayloadArray(cryptoCurrencyList)
		cryptoCurrencyList = coinChartOverviewDataPayload.FilterByFirstChartDate(cryptoCurrencyList, beforeTime)
	} else {
		filtered := make([]CryptoCurrencyData, 0)
		for _, cryptoCurrencyData := range cryptoCurrencyList {
			createdAt, err := time.Parse(time.RFC3339, cryptoCurrencyData.DateAdded)
			if err != nil {
				log.Fatalln(err.Error())
			}
			if createdAt.Before(beforeTime) {
				continue
			}
			filtered = append(filtered, cryptoCurrencyData)
		}
		cryptoCurrencyList = filtered
	}
	totalFilteredByDate := totalCryptoCurrency - len(cryptoCurrencyList)
	totalCryptoCurrencyAfterFilteredByDate := len(cryptoCurrencyList)

	coinMarketDataArray := getCoinMarketDataArray(cryptoCurrencyList)
	cryptoCurrencyList = coinMarketDataArray.FilterByExchanges(cryptoCurrencyList, []string{"binance"})

	totalFilteredByExchanges := totalCryptoCurrencyAfterFilteredByDate - len(cryptoCurrencyList)

	fmt.Printf("Found %d coins; %d filtered by date; %d filtered by exchanges. %d coins left\n", totalCryptoCurrency, totalFilteredByDate, totalFilteredByExchanges, len(cryptoCurrencyList))

	for _, cryptocurrencyData := range cryptoCurrencyList {
		createdAt, err := time.Parse(time.RFC3339, cryptocurrencyData.DateAdded)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(cryptocurrencyData.Name, fmt.Sprintf("(%s)", cryptocurrencyData.Symbol), createdAt.Format("January 2006"))
		for j := 0; j < len(cryptocurrencyData.Quotes); j++ {
			quote := cryptocurrencyData.Quotes[j]
			if quote.Name != "USDT" {
				continue
			}
			fmt.Printf("FDMarketCap: %f, 1Y: %.3f %%\n", quote.FullyDilutedMarketCap, quote.PercentChangeOneYear)
		}
	}
}
