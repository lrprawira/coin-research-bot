package main

import (
	"bytes"
	"coin_research_bot/filtermodules"
	"coin_research_bot/lib"
	"database/sql"
	"encoding/gob"
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

type CacheEntry struct {
	Id        uint
	Value     []byte
	Timestamp time.Time
}

func main() {
	db, err := sql.Open("sqlite3", "cache.db")

	if err != nil {
		log.Fatalln(err)
	}

	if _, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS caches (
			id INTEGER NOT NULL PRIMARY KEY,
			key TEXT NOT NULL UNIQUE,
			value BLOB,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		`); err != nil {
		log.Fatalln(err)
	}

	cacheEntry := CacheEntry{}
	cryptoCurrencyListingResponseBody := lib.ListingResponseBody{}
	row := db.QueryRow("SELECT id, value, timestamp FROM caches WHERE key=?;", "listing")
	if err = row.Scan(&cacheEntry.Id, &cacheEntry.Value, &cacheEntry.Timestamp); err == nil && cacheEntry.Timestamp.After(time.Now().Add(time.Duration(-15)*time.Minute)) {
		// Found row
		bufPtr := bytes.NewBuffer(cacheEntry.Value)
		gob.NewDecoder(bufPtr).Decode(&cryptoCurrencyListingResponseBody)
	} else {
		cryptoCurrencyListingResponseBody, err = lib.GetCoinList()
		if err != nil {
			log.Fatalln(err)
		}
		var buf bytes.Buffer
		gob.NewEncoder(&buf).Encode(cryptoCurrencyListingResponseBody)
		if cacheEntry.Id != 0 {
			db.Exec("UPDATE caches SET value = ?, timestamp = ? WHERE id = ?", buf, time.Now(), cacheEntry.Id)
		}
		_, err := db.Exec("INSERT INTO caches (key, value) VALUES (?, ?)", "listing", buf.Bytes())
		if err != nil {
			log.Fatalln(err)
		}
	}

	if err != nil {
		log.Fatalln(err)
	}

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
