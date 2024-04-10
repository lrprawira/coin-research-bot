package lib

import (
	"coin_research_bot/lib/common"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const listingLimit = 10000
var marketCapRange = struct {min uint; max uint}{100_000_000, 200_000_000}
var listingEndpoint = fmt.Sprintf("https://api.coinmarketcap.com/data-api/v3/cryptocurrency/listing?start=1&limit=%d&sortBy=date_added&sortType=asc&convert=USDT&cryptoType=all&tagType=all&audited=false&aux=ath,atl,high24h,low24h,num_market_pairs,cmc_rank,date_added,max_supply,circulating_supply,total_supply,volume_7d,volume_30d,self_reported_circulating_supply,self_reported_market_cap&category=spot&marketCapRange=%d~%d", listingLimit, marketCapRange.min, marketCapRange.max)

type CryptoCurrencyData struct {
	Id        uint    `json:"id"`
	Name      string  `json:"name"`
	Symbol    string  `json:"symbol"`
	Slug      string  `json:"slug"`
	Ath       float64 `json:"ath"`
	Atl       float64 `json:"atl"`
	DateAdded string  `json:"dateAdded"`
	Quotes    []struct {
		Name                    string  `json:"name"`
		MarketCap               float64 `json:"marketCap"`
		FullyDilutedMarketCap   float64 `json:"fullyDilluttedMarketCap"`
		PercentChangeOneYear    float64 `json:"percentChange1y"`
		PercentChangeYearToDate float64 `json:"ytdPriceChangePercentage"`
		PercentChange90D        float64 `json:"percentChange90d"`
		PercentChange60D        float64 `json:"percentChange60d"`
		PercentChange30D        float64 `json:"percentChange30d"`
		PercentChange7D         float64 `json:"percentChange7d"`
		PercentChange24H        float64 `json:"percentChange24h"`
		PercentChange1H         float64 `json:"percentChange1h"`
	}
}

type CryptoCurrencyList []CryptoCurrencyData

type ListingResponseBody struct {
	Data struct {
		CryptoCurrencyList CryptoCurrencyList `json:"cryptoCurrencyList"`
		TotalCount         int                `json:"total_count"`
	} `json:"data"`
	Status common.StatusData `json:"status"`
}

func getCoinList() (ListingResponseBody, error) {

	req, err := http.NewRequest("GET", listingEndpoint, nil)
	req.Header = common.CommonHeader
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return ListingResponseBody{}, err
	}
	var data ListingResponseBody
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return ListingResponseBody{}, err
	}
	return data, nil
}

func GetCoinList() *ListingResponseBody {
	cryptoCurrencyListingResponseBody := new(ListingResponseBody)
	cryptoCurrencyListingResponseBody, err := common.GetCacheOrRunCallable[ListingResponseBody](cryptoCurrencyListingResponseBody, "listing", 86400, func() ListingResponseBody {
		res, err := getCoinList()
		if err != nil {
			log.Fatalln(err)
		}
		return res
	})
	if err != nil {
		log.Fatalln(err)
	}
	return cryptoCurrencyListingResponseBody
}

func (cryptoCurrencyList *CryptoCurrencyList) PrintCoins() {
	for _, coin := range *cryptoCurrencyList {
		fmt.Print(coin.Symbol, " ")

	}
	fmt.Println()
}
