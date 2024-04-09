package common

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

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

func GetCacheOrRunCallable[T any](data *T, cacheKey string, cacheTtl uint, cb func() T) (*T, error) {
	// t0 := time.Now()
	cacheEntry := CacheEntry{}
	db := GetDB()
	// log.Println("Cache", time.Since(t0))
	row := db.QueryRow("SELECT id, value, timestamp FROM caches WHERE key=?;", cacheKey)
	if err := row.Scan(&cacheEntry.Id, &cacheEntry.Value, &cacheEntry.Timestamp); err == nil && cacheEntry.Timestamp.After(time.Now().Add(time.Duration(-cacheTtl)*time.Second)) {
		// Found row
		// log.Println("Cache Hit!", reflect.TypeOf(data))
		bufPtr := bytes.NewBuffer(cacheEntry.Value)
		gob.NewDecoder(bufPtr).Decode(data)
		return data, nil
	}
	tmp := cb() 
	data = &tmp
	var buf bytes.Buffer
	gob.NewEncoder(&buf).Encode(tmp)
	if cacheEntry.Id != 0 {
		// Row already exists
		_, err := db.Exec("UPDATE caches SET value = ?, timestamp = ? WHERE id = ?", buf.Bytes(), time.Now(), cacheEntry.Id)
		if err != nil {
			return data, err
		}
		return data, nil
	}
	// Row has not been created
	_, err := db.Exec("INSERT INTO caches (key, value) VALUES (?, ?)", cacheKey, buf.Bytes())
	if err != nil {
		return data, err
	}
	return data, nil
}
