package main

import (
	"coin_research_bot/lib/filtermodules"
	"coin_research_bot/lib"
	"coin_research_bot/lib/common"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

/* Config */
const createdAtMode = filtermodules.CreatedAtDateAdded
const beforeTimeString = "2022-01-01T00:00:00Z"

var beforeTime, _ = time.Parse(time.RFC3339, beforeTimeString)

/* End config */

func main() {
	// now := time.Now()
	db := common.GetDB()
	defer db.Close()
	// defer func() {
	// 	log.Println(time.Since(now))
	// }()

	statistics := struct {
		Total               uint
		FilteredByDate      uint
		FilteredByExchanges uint
	}{}

	common.HandleCacheTableCreation()

	cryptoCurrencyListingResponseBody := lib.GetCoinList()
	cryptoCurrencyList := cryptoCurrencyListingResponseBody.Data.CryptoCurrencyList

	statistics.Total = uint(len(cryptoCurrencyList))

	/* query for filters */
	var coinChartOverviewDataPayloadArray *lib.CoinChartOverviewDataPayload = nil
	if filtermodules.CreatedAtChart == createdAtMode {
		coinChartOverviewDataPayloadArray = new(lib.CoinChartOverviewDataPayload)
		var err error
		coinChartOverviewDataPayloadArray, err = common.GetCacheOrRunCallable[lib.CoinChartOverviewDataPayload](coinChartOverviewDataPayloadArray, "coinchartoverviewdatapayloadarray", 3600, func() lib.CoinChartOverviewDataPayload {
			return lib.GetCoinChartOverviewDataPayloadArray(&cryptoCurrencyList)
		})
		if err != nil {
			log.Fatalln(err)
		}
	}
	coinMarketDataArray := new(lib.CoinMarketDataArray)
	coinMarketDataArray, err := common.GetCacheOrRunCallable[lib.CoinMarketDataArray](coinMarketDataArray, "coinmarketdataarray", 86400, func() lib.CoinMarketDataArray {
		return lib.GetCoinMarketDataArray(&cryptoCurrencyList)
	})
	if err != nil {
		log.Fatalln(err)
	}
	/* End query for filters */

	/* Filters */
	cryptoCurrencyList = filtermodules.FilterByStartDate(&cryptoCurrencyList, coinChartOverviewDataPayloadArray, createdAtMode, beforeTime)
	statistics.FilteredByDate = statistics.Total - uint(len(cryptoCurrencyList))

	cryptoCurrencyList = filtermodules.FilterByExchanges(&cryptoCurrencyList, coinMarketDataArray, []string{"binance"})
	statistics.FilteredByExchanges = statistics.Total - statistics.FilteredByDate - uint(len(cryptoCurrencyList))
	/* End Filters */

	fmt.Printf("Found %d coins; %d filtered by date; %d filtered by exchanges. %d coins left.\n\n", statistics.Total, statistics.FilteredByDate, statistics.FilteredByExchanges, len(cryptoCurrencyList))

	/* Print Coin List */
	for _, cryptocurrencyData := range cryptoCurrencyList {
		createdAt, err := time.Parse(time.RFC3339, cryptocurrencyData.DateAdded)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(cryptocurrencyData.Name, fmt.Sprintf("(%s)", cryptocurrencyData.Symbol), createdAt.Format("January 2006"))
		for _, quote := range cryptocurrencyData.Quotes {
			if quote.Name != "USDT" {
				continue
			}
			fmt.Printf("FDMarketCap: %f, 1Y: %.3f %%\n", quote.FullyDilutedMarketCap, quote.PercentChangeOneYear)
		}
	}
	/* End Print Coin List */
}
