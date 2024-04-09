package filtermodules

import "coin_research_bot/lib"

func FilterByExchanges(cryptoCurrencyList *[]lib.CryptoCurrencyData, coinMarketDataArray *lib.CoinMarketDataArray, exchanges []string) []lib.CryptoCurrencyData {
	if coinMarketDataArray == nil {
		tmp := lib.GetCoinMarketDataArray(cryptoCurrencyList)
		coinMarketDataArray = &tmp
	}
	return coinMarketDataArray.FilterByExchanges(cryptoCurrencyList, exchanges)
}
