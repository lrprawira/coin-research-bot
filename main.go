package main

import (
	"coin_research_bot/filtermodules"
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
	db := common.GetDB()
	defer db.Close()

	common.HandleCacheTableCreation()

	cryptoCurrencyListingResponseBody := lib.GetCoinList()
	cryptoCurrencyList := cryptoCurrencyListingResponseBody.Data.CryptoCurrencyList

	statistics := struct {
		Total               uint
		FilteredByDate      uint
		FilteredByExchanges uint
	}{}

	statistics.Total = uint(len(cryptoCurrencyList))

	/* Filters */
	cryptoCurrencyList = filtermodules.FilterByStartDate(&cryptoCurrencyList, nil, createdAtMode, beforeTime)
	statistics.FilteredByDate = statistics.Total - uint(len(cryptoCurrencyList))

	cryptoCurrencyList = filtermodules.FilterByExchanges(&cryptoCurrencyList, nil, []string{"binance"})
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
