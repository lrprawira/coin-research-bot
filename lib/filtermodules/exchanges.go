package filtermodules

import "coin_research_bot/lib"

func FilterByExchanges(cryptoCurrencyList *lib.CryptoCurrencyList, coinMarketDataArray *lib.CoinMarketDataArray, exchanges []string) lib.CryptoCurrencyList {
	if coinMarketDataArray == nil {
		coinMarketDataArray = lib.GetCoinMarketDataArray(cryptoCurrencyList)
	}
	return coinMarketDataArray.FilterByExchanges(cryptoCurrencyList, exchanges)
}
