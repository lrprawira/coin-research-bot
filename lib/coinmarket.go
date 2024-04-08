package lib

import (
	"coin_research_bot/lib/common"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
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

type CoinMarketDataArray []CoinMarketResponseBody

func getCoinMarketEndpoint(cryptoCurrencyData CryptoCurrencyData) string {
	return fmt.Sprintf("%s?slug=%s&start=1&limit=200&category=spot&centerType=all&sort=cmc_rank_advanced&direction=desc&spotUntracked=true", baseCoinMarketEndpoint, cryptoCurrencyData.Slug)
}

func getCoinMarketData(wg *sync.WaitGroup, cryptocurrencyData CryptoCurrencyData, coinMarketDataArray []CoinMarketResponseBody, iter int) {
	defer wg.Done()
	req, err := http.NewRequest("GET", getCoinMarketEndpoint(cryptocurrencyData), nil)
	req.Header = common.CommonHeader
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return
	}

	var coinMarketData CoinMarketResponseBody
	err = json.NewDecoder(res.Body).Decode(&coinMarketData)

	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return
	}

	coinMarketDataArray[iter] = coinMarketData
}

func GetCoinMarketDataArray(cryptoCurrencyList *[]CryptoCurrencyData) CoinMarketDataArray  {
	var wg sync.WaitGroup
	coinMarketDataArray := make(CoinMarketDataArray, len(*cryptoCurrencyList))
	wg.Add(len(*cryptoCurrencyList))
	for i, cryptoCurrencyData := range *cryptoCurrencyList {
		go getCoinMarketData(&wg, cryptoCurrencyData, coinMarketDataArray, i)
	}
	
	wg.Wait()
	return coinMarketDataArray
}

func (coinMarketDataArray CoinMarketDataArray) FilterByExchanges (cryptoCurrencyList *[]CryptoCurrencyData, exchanges []string) []CryptoCurrencyData {
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
