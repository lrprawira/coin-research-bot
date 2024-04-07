package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
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
	cryptocurrencyList := data.Data.CryptoCurrencyList
	totalFilteredOutCandidates := 0
	coinMarketData := make([]CoinMarketResponseBody, len(cryptocurrencyList))
	coinChartOverviewData := make([]CoinChartOverviewDataPayload, len(cryptocurrencyList))

	var wg sync.WaitGroup

	/* get coin chart */
	if createdAtChart == createdAtMode {
		wg.Add(len(cryptocurrencyList))
		for i, cryptocurrencyData := range cryptocurrencyList {
			go getCoinChartOverviewData(&wg, cryptocurrencyData, coinChartOverviewData, i)
		}
		wg.Wait()
	}

	wg.Add(len(cryptocurrencyList))

	/* get coin market data */
	for i, cryptocurrencyData := range cryptocurrencyList {
		// unixTime, err := strconv.ParseInt(coinChartOverviewData[i].key, 10, 64)
		var createdAt time.Time
		if createdAtChart == createdAtMode {
			createdAt = time.Unix(coinChartOverviewData[i].key, 0)
		} else {
			createdAt, err = time.Parse(time.RFC3339, cryptocurrencyData.DateAdded)
			if err != nil {
				wg.Done()
				continue
			}
		}

		// if createdAt.Before(time.Now().AddDate(-2, 0, 0)) {
		if createdAt.Before(time.Date(2022, 0, 0, 0, 0, 0, 0, time.UTC)) {
			wg.Done()
			totalFilteredOutCandidates++
			continue
		}
		go getCoinMarketData(&wg, cryptocurrencyData, coinMarketData, i, []string{"binance"})
	}

	fmt.Printf("Found %d coins; %d filtered by date\n", len(cryptocurrencyList), totalFilteredOutCandidates)
	fmt.Println("Waiting for every coin to be fetched")
	wg.Wait()

	for i, cryptocurrencyData := range cryptocurrencyList {
		if coinMarketData[i].Data.Id == 0 {
			continue
		}

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
