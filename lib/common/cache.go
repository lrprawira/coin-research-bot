package common

import "log"

func HandleCacheTableCreation()  {
	db := GetDB()
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS caches (
			id INTEGER NOT NULL PRIMARY KEY,
			key TEXT NOT NULL UNIQUE,
			value BLOB,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		`); err != nil {
		log.Fatalln(err)
	}
}
