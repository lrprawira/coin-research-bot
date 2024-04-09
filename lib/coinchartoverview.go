package lib

import (
	"coin_research_bot/lib/common"
	"encoding/json"
	"errors"
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

type CoinChartOverviewDataPayload []*CoinChartOverviewDataPayloadItem

func getCoinChartOverviewData(cryptoCurrencyData *CryptoCurrencyData) (*CoinChartOverviewDataPayloadItem, error) {
	req, err := http.NewRequest("GET", getCoinChartOverviewEndpoint(*cryptoCurrencyData), nil)
	req.Header = common.CommonHeader
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	var coinChartOverviewData CoinChartOverviewResponseBody
	err = json.NewDecoder(res.Body).Decode(&coinChartOverviewData)

	if err != nil {
		return nil, err
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
		return &CoinChartOverviewDataPayloadItem{coinChartOverviewData, minKey}, nil
	}
	return nil, errors.New("Coin chart not found")
}

func GetCoinChartOverviewDataPayloadArray(cryptoCurrencyList *[]CryptoCurrencyData) CoinChartOverviewDataPayload {
	var wg sync.WaitGroup
	ch := make(chan bool, 8)
	coinChartOverviewDataPayloadArray := make(CoinChartOverviewDataPayload, len(*cryptoCurrencyList))
	wg.Add(len(*cryptoCurrencyList))

	for i, cryptoCurrencyData := range *cryptoCurrencyList {
		ch <- true
		// Shadow vars to remove warnings of using these inside of the closure
		cryptoCurrencyData := cryptoCurrencyData
		i := i
		go func() {
			defer wg.Done()
			coinChartOverviewData, err := getCoinChartOverviewData(&cryptoCurrencyData)
			if err != nil {
				fmt.Fprintf(os.Stderr, err.Error())
				<-ch
			}
			coinChartOverviewDataPayloadArray[i] = coinChartOverviewData
			<-ch
		}()
	}
	wg.Wait()
	return coinChartOverviewDataPayloadArray
}

func (coinChartOverviewDataPayloadArray CoinChartOverviewDataPayload) FilterByFirstChartDate(cryptoCurrencyList *[]CryptoCurrencyData, beforeTime time.Time) []CryptoCurrencyData {
	filtered := make([]CryptoCurrencyData, 0, len(*cryptoCurrencyList))
	for i, cryptoCurrencyData := range *cryptoCurrencyList {
		startDate := time.Unix(coinChartOverviewDataPayloadArray[i].key, 0)
		if startDate.Before(beforeTime) {
			continue
		}
		filtered = append(filtered, cryptoCurrencyData)
	}
	return filtered
}
