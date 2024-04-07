package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"sync"
)

const baseCoinChartOverviewEndpoint = "https://api.coinmarketcap.com/data-api/v3/cryptocurrency/detail/chart"

type CoinChartOverviewResponseBody struct {
	Data struct {
		Points map[string] interface {}
	} `json:"data"`
	Status StatusData `json:"status"`
}

func getCoinChartOverviewEndpoint(cryptoCurrencyData CryptoCurrencyData) string {
	return fmt.Sprintf("%s?id=%d&range=ALL", baseCoinChartOverviewEndpoint, cryptoCurrencyData.Id)
}

type CoinChartOverviewDataPayload struct{
	data CoinChartOverviewResponseBody
	key int64
}

func getCoinChartOverviewData(wg *sync.WaitGroup, cryptoCurrencyData CryptoCurrencyData, coinChartOverviewDataArray []CoinChartOverviewDataPayload, iter int) {
	defer wg.Done()
	req, err := http.NewRequest("GET", getCoinChartOverviewEndpoint(cryptoCurrencyData), nil)
	req.Header = commonHeader
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return
	}

	var coinChartOverviewData CoinChartOverviewResponseBody
	err = json.NewDecoder(res.Body).Decode(&coinChartOverviewData)

	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return
	}

	found := false
	minKey := int64(math.MaxInt64)

	for key := range coinChartOverviewData.Data.Points {
		currentKey, err := strconv.ParseInt(key, 10, 64)
		if err != nil {
			continue
		}
		found = true
		if currentKey < minKey {
			minKey = currentKey
		}
	}
	if found {
		coinChartOverviewDataArray[iter] = CoinChartOverviewDataPayload{coinChartOverviewData, minKey}
	}
	return
}
