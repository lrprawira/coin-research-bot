package lib

import (
	"bytes"
	"coin_research_bot/lib/common"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"log"
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
	Data CoinChartOverviewResponseBody
	Key  int64
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

	if coinChartOverviewData.Status.ErrorCode != "0" {
		return nil, errors.New(coinChartOverviewData.Status.ErrorMessage)
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

func GetCoinChartOverviewDataPayloadArray(cryptoCurrencyList *CryptoCurrencyList) *CoinChartOverviewDataPayload {
	var wg sync.WaitGroup
	ch := make(chan bool, 16)
	coinChartOverviewDataPayloadArray := make(CoinChartOverviewDataPayload, len(*cryptoCurrencyList))
	cacheKeys := make([]string, len(*cryptoCurrencyList))
	for i, cryptoCurrencyData := range *cryptoCurrencyList {
		cacheKeys[i] = fmt.Sprintf("coinchartoverviewdata:%d", cryptoCurrencyData.Id)
	}
	foundInCache := common.GetCaches(cacheKeys, []string{"key", "value", "timestamp"})
	cacheMap := map[string]*CoinChartOverviewDataPayloadItem{}
	for foundInCache.Next() {
		cacheEntry := common.CacheEntry{}
		if err := foundInCache.Scan(&cacheEntry.Key, &cacheEntry.Value, &cacheEntry.Timestamp); err != nil || !cacheEntry.Timestamp.After(time.Now().Add(time.Duration(-86400)*time.Second))  {
			fmt.Fprintf(os.Stderr, "Cache is expired or broken")
			continue
		}
		cacheMap[cacheEntry.Key] = new(CoinChartOverviewDataPayloadItem)
		bufPtr := bytes.NewBuffer(cacheEntry.Value)
		err := gob.NewDecoder(bufPtr).Decode(cacheMap[cacheEntry.Key])
		if err != nil {
			log.Fatalln(err)
		}
	}
	wg.Add(len(*cryptoCurrencyList) - len(cacheMap))

	for i, cryptoCurrencyData := range *cryptoCurrencyList {
		if cached, ok := cacheMap[fmt.Sprintf("coinchartoverviewdata:%d", cryptoCurrencyData.Id)]; ok {
			coinChartOverviewDataPayloadArray[i] = cached
			continue
		}
		ch <- true
		// Shadow vars to remove warnings of using these inside of the closure
		cryptoCurrencyData := cryptoCurrencyData
		i := i
		go func() {
			defer wg.Done()
			coinChartOverviewData := new(CoinChartOverviewDataPayloadItem)
			coinChartOverviewData, err := common.GetCacheOrRunCallable[CoinChartOverviewDataPayloadItem](coinChartOverviewData, fmt.Sprintf("coinchartoverviewdata:%d", cryptoCurrencyData.Id), 86400, func() CoinChartOverviewDataPayloadItem {
				coinChartOverviewData, err := getCoinChartOverviewData(&cryptoCurrencyData)
				if err != nil {
					log.Fatalln(err)
				}
				return *coinChartOverviewData
			})
			<-ch
			if err != nil {
				log.Fatalln(err)
			}
			coinChartOverviewDataPayloadArray[i] = coinChartOverviewData
		}()
	}
	wg.Wait()
	return &coinChartOverviewDataPayloadArray
}

func (coinChartOverviewDataPayloadArray CoinChartOverviewDataPayload) FilterByFirstChartDate(cryptoCurrencyList *CryptoCurrencyList, beforeTime time.Time) CryptoCurrencyList {
	filtered := make([]CryptoCurrencyData, 0, len(*cryptoCurrencyList))
	for i, cryptoCurrencyData := range *cryptoCurrencyList {
		startDate := time.Unix(coinChartOverviewDataPayloadArray[i].Key, 0)
		if startDate.Before(beforeTime) {
			continue
		}
		filtered = append(filtered, cryptoCurrencyData)
	}
	return filtered
}
