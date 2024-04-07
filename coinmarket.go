package main

import (
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
	Status StatusData `json:"status"`
}

func getCoinMarketEndpoint(cryptoCurrencyData CryptoCurrencyData) string {
	return fmt.Sprintf("%s?slug=%s&start=1&limit=200&category=spot&centerType=all&sort=cmc_rank_advanced&direction=desc&spotUntracked=true", baseCoinMarketEndpoint, cryptoCurrencyData.Slug)
}

func getCoinMarketData(wg *sync.WaitGroup, cryptocurrencyData CryptoCurrencyData, coinMarketDataArray []CoinMarketResponseBody, iter int, filterExchanges []string) {
	defer wg.Done()
	req, err := http.NewRequest("GET", getCoinMarketEndpoint(cryptocurrencyData), nil)
	req.Header = commonHeader
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

	if len(filterExchanges) > 0 {
		for i := 0; i < len(coinMarketData.Data.MarketPairs); i++ {
			for j := 0; j < len(filterExchanges); j++ {
				if coinMarketData.Data.MarketPairs[i].ExchangeSlug == filterExchanges[j] {
					coinMarketDataArray[iter] = coinMarketData
					return 
				}
			}
		}
		// Return nil if there are filterExchanges
		return
	}

	coinMarketDataArray[iter] = coinMarketData
}
