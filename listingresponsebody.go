package main

const listingEndpoint = "https://api.coinmarketcap.com/data-api/v3/cryptocurrency/listing?start=1&limit=1000&sortBy=date_added&sortType=asc&convert=USDT&cryptoType=all&tagType=all&audited=false&aux=ath,atl,high24h,low24h,num_market_pairs,cmc_rank,date_added,max_supply,circulating_supply,total_supply,volume_7d,volume_30d,self_reported_circulating_supply,self_reported_market_cap&category=spot&marketCapRange=100000000~150000000"

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


type ListingResponseBody struct {
	Data struct {
		CryptoCurrencyList []CryptoCurrencyData `json:"cryptoCurrencyList"`
		TotalCount         int                  `json:"total_count"`
	} `json:"data"`
	Status StatusData `json:"status"`
}
