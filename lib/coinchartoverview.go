package lib

import (
	"coin_research_bot/lib/common"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

const baseCoinChartOverviewEndpoint = "https://api.coinmarketcap.com/data-api/v3/cryptocurrency/detail/chart"

type CoinChartOverviewResponseBody struct {
	Data struct {
		Points map[string]interface{}
	} `json:"data"`
	Status common.StatusData `json:"status"`
}

func getCoinChartOverviewEndpoint(cryptoCurrencyData CryptoCurrencyData) string {
	return fmt.Sprintf("%s?id=%d&range=ALL", baseCoinChartOverviewEndpoint, cryptoCurrencyData.Id)
}

type CoinChartOverviewDataPayloadItem struct {
	data CoinChartOverviewResponseBody
	key  int64
}

type CoinChartOverviewDataPayload []CoinChartOverviewDataPayloadItem

func getCoinChartOverviewData(wg *sync.WaitGroup, cryptoCurrencyData CryptoCurrencyData, coinChartOverviewDataArray CoinChartOverviewDataPayload, iter int) {
	defer wg.Done()
	req, err := http.NewRequest("GET", getCoinChartOverviewEndpoint(cryptoCurrencyData), nil)
	req.Header = common.CommonHeader
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
		coinChartOverviewDataArray[iter] = CoinChartOverviewDataPayloadItem{coinChartOverviewData, minKey}
	}
	return
}

func GetCoinChartOverviewDataPayloadArray(cryptoCurrencyList *[]CryptoCurrencyData) CoinChartOverviewDataPayload {
	var wg sync.WaitGroup
	coinChartOverviewData := make(CoinChartOverviewDataPayload, len(*cryptoCurrencyList))
	wg.Add(len(*cryptoCurrencyList))

	for i, cryptocurrencyData := range *cryptoCurrencyList {
		go getCoinChartOverviewData(&wg, cryptocurrencyData, coinChartOverviewData, i)
	}
	wg.Wait()
	return coinChartOverviewData
}

func (coinChartOverviewDataPayloadArray CoinChartOverviewDataPayload) FilterByFirstChartDate(cryptoCurrencyList *[]CryptoCurrencyData, beforeTime time.Time) []CryptoCurrencyData  {
	filtered := make([]CryptoCurrencyData, 0)
	for i, cryptoCurrencyData := range *cryptoCurrencyList {
		startDate := time.Unix(coinChartOverviewDataPayloadArray[i].key, 0)
		if startDate.Before(beforeTime) {
			continue
		}
		filtered = append(filtered, cryptoCurrencyData)
	}
	return filtered
}
