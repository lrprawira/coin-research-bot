package lib

import (
	"bytes"
	"coin_research_bot/lib/common"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const baseCoinMarketEndpoint = "https://api.coinmarketcap.com/data-api/v3/cryptocurrency/market-pairs/latest"

type CoinMarketResponseBody struct {
	Data struct {
		Id               uint   `json:"id"`
		Name             string `json:"name"`
		Symbol           string `json:"symbol"`
		MarketPairLength uint   `json:"numMarketPairs"`
		MarketPairs      []struct {
			Rank         uint   `json:"rank"`
			ExchangeId   uint   `json:"exchangeId"`
			ExchangeName string `json:"exchangeName"`
			ExchangeSlug string `json:"exchangeSlug"`
			Category     string `json:"category"`
		} `json:"marketPairs"`
	} `json:"data"`
	Status common.StatusData `json:"status"`
}

type CoinMarketDataArray []*CoinMarketResponseBody

func getCoinMarketEndpoint(cryptoCurrencyData CryptoCurrencyData) string {
	return fmt.Sprintf("%s?slug=%s&start=1&limit=100&category=spot&centerType=all&sort=cmc_rank_advanced&direction=desc&spotUntracked=true", baseCoinMarketEndpoint, cryptoCurrencyData.Slug)
}

func getCoinMarketData(cryptoCurrencyData *CryptoCurrencyData) (*CoinMarketResponseBody, error) {
	req, err := http.NewRequest("GET", getCoinMarketEndpoint(*cryptoCurrencyData), nil)
	req.Header = common.CommonHeader
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	var coinMarketData CoinMarketResponseBody
	err = json.NewDecoder(res.Body).Decode(&coinMarketData)

	if err != nil {
		return nil, err
	}

	if coinMarketData.Status.ErrorCode != "0" {
		return nil, errors.New(coinMarketData.Status.ErrorMessage)
	}

	return &coinMarketData, nil
}

func GetCoinMarketDataArray(cryptoCurrencyList *CryptoCurrencyList) *CoinMarketDataArray {
	var wg sync.WaitGroup
	ch := make(chan bool, 16)
	coinMarketDataArray := make(CoinMarketDataArray, len(*cryptoCurrencyList))
	cacheKeys := make([]string, len(*cryptoCurrencyList))
	for i, cryptoCurrencyData := range *cryptoCurrencyList {
		cacheKeys[i] = fmt.Sprintf("coinmarketdata:%d", cryptoCurrencyData.Id)
	}
	foundInCache := common.GetCaches(cacheKeys, []string{"key", "value", "timestamp"})
	cacheMap := map[string]*CoinMarketResponseBody{}
	for foundInCache.Next() {
		cacheEntry := common.CacheEntry{}
		if err := foundInCache.Scan(&cacheEntry.Key, &cacheEntry.Value, &cacheEntry.Timestamp); err != nil || !cacheEntry.Timestamp.After(time.Now().Add(time.Duration(-86400)*time.Second))  {
			fmt.Fprintf(os.Stderr, "Cache is expired or broken")
			continue
		}
		cacheMap[cacheEntry.Key] = new(CoinMarketResponseBody)
		bufPtr := bytes.NewBuffer(cacheEntry.Value)
		err := gob.NewDecoder(bufPtr).Decode(cacheMap[cacheEntry.Key])
		if err != nil {
			log.Fatalln(err)
		}
	}
	wg.Add(len(*cryptoCurrencyList) - len(cacheMap))

	for i, cryptoCurrencyData := range *cryptoCurrencyList {
		if cached, ok := cacheMap[fmt.Sprintf("coinmarketdata:%d", cryptoCurrencyData.Id)]; ok {
			coinMarketDataArray[i] = cached
			continue
		}
		ch <- true
		// Shadow vars to remove warnings of using these inside of the closure
		cryptoCurrencyData := cryptoCurrencyData
		i := i
		go func() {
			defer wg.Done()
			coinMarketData := new(CoinMarketResponseBody)
			coinMarketData, err := common.GetCacheOrRunCallable[CoinMarketResponseBody](coinMarketData, fmt.Sprintf("coinmarketdata:%d", cryptoCurrencyData.Id), 86400, func() CoinMarketResponseBody {
				coinMarketData, err := getCoinMarketData(&cryptoCurrencyData)
				if err != nil {
					log.Fatalln(err)
				}
				return *coinMarketData
			})
			<-ch
			if err != nil {
				log.Fatalln(err)
			}
			coinMarketDataArray[i] = coinMarketData
		}()
	}

	wg.Wait()
	return &coinMarketDataArray
}

func (coinMarketDataArray CoinMarketDataArray) FilterByExchanges(cryptoCurrencyList *CryptoCurrencyList, exchanges []string) []CryptoCurrencyData {
	filtered := make([]CryptoCurrencyData, 0, len(*cryptoCurrencyList))
	for i, cryptoCurrencyData := range *cryptoCurrencyList {
		for _, marketPair := range coinMarketDataArray[i].Data.MarketPairs {
			for _, exchange := range exchanges {
				if marketPair.ExchangeSlug == exchange {
					filtered = append(filtered, cryptoCurrencyData)
					goto FilterByExchangesContinueCryptoCurrency
				}
			}
		}
	FilterByExchangesContinueCryptoCurrency:
	}
	return filtered
}
