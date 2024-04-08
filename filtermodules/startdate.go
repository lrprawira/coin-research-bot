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

func FilterByStartDate(cryptoCurrencyList *[]lib.CryptoCurrencyData, coinChartOverviewDataPayload *lib.CoinChartOverviewDataPayload, createdAtMode StartDateType, beforeTime time.Time) []lib.CryptoCurrencyData {
	if CreatedAtChart == createdAtMode {
		if coinChartOverviewDataPayload == nil {
			tmp := lib.GetCoinChartOverviewDataPayloadArray(cryptoCurrencyList)
			coinChartOverviewDataPayload = &tmp
		}
		tmp := coinChartOverviewDataPayload.FilterByFirstChartDate(cryptoCurrencyList, beforeTime)
		cryptoCurrencyList = &tmp
	} else {
		filtered := make([]lib.CryptoCurrencyData, 0)
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
