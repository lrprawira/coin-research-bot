package filtermodules

import (
	"coin_research_bot/lib"
	"log"
	"time"
)

type StartDateType uint

const (
	CreatedAtDateAdded StartDateType = iota
	CreatedAtChart
)

func FilterByStartDate(cryptoCurrencyList *lib.CryptoCurrencyList, coinChartOverviewDataPayload *lib.CoinChartOverviewDataPayload, createdAtMode StartDateType, beforeTime time.Time) lib.CryptoCurrencyList {
	if CreatedAtChart == createdAtMode {
		if coinChartOverviewDataPayload == nil {
			coinChartOverviewDataPayload = lib.GetCoinChartOverviewDataPayloadArray(cryptoCurrencyList)
		}
		tmp := coinChartOverviewDataPayload.FilterByFirstChartDate(cryptoCurrencyList, beforeTime)
		cryptoCurrencyList = &tmp
	} else {
		filtered := make(lib.CryptoCurrencyList, 0)
		for _, cryptoCurrencyData := range *cryptoCurrencyList {
			createdAt, err := time.Parse(time.RFC3339, cryptoCurrencyData.DateAdded)
			if err != nil {
				log.Fatalln(err.Error())
			}
			if createdAt.Before(beforeTime) {
				continue
			}
			filtered = append(filtered, cryptoCurrencyData)
		}
		cryptoCurrencyList = &filtered
	}
	return *cryptoCurrencyList
}
