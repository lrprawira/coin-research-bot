package common

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

func HandleCacheTableCreation()  {
	db := GetDB()
	if _, err := db.Exec(`
		BEGIN;
		CREATE TABLE IF NOT EXISTS caches (
			id INTEGER NOT NULL PRIMARY KEY,
			key TEXT NOT NULL UNIQUE,
			value BLOB,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS timestamp_idx ON caches (timestamp);
		COMMIT;
		`); err != nil {
		log.Fatalln(err)
	}
}

func GetCaches(cacheKeys []string, columns []string) *sql.Rows {
	db := GetDB()
	sqlColumns := strings.Join(columns, ",")
	sqlCacheKeys := make([]interface{}, len(cacheKeys))
	columnParamPlaceholders := make([]string, len(columns))
	for i := range columnParamPlaceholders {
		columnParamPlaceholders[i] = "?"
	}
	for i := range sqlCacheKeys {
		sqlCacheKeys[i] = cacheKeys[i]
	}
	sqlStmt := fmt.Sprintf("SELECT %s FROM caches WHERE key IN (", sqlColumns)
	for i := range cacheKeys {
		sqlStmt += "?"
		if i < len(cacheKeys)-1 {
			sqlStmt += ","
		}
	}
	sqlStmt += ");"
	rows, err := db.Query(sqlStmt, sqlCacheKeys...)
	if err != nil {
		log.Fatalln(err)
	}
	return rows
}

func GetCacheOrRunCallable[T any](data *T, cacheKey string, cacheTtl uint, cb func() T) (*T, error) {
	// t0 := time.Now()
	cacheEntry := CacheEntry{}
	db := GetDB()
	// log.Println("Cache", time.Since(t0))
	row := db.QueryRow("SELECT id, key, value, timestamp FROM caches WHERE key=?;", cacheKey)
	if err := row.Scan(&cacheEntry.Id, &cacheEntry.Key, &cacheEntry.Value, &cacheEntry.Timestamp); err == nil && cacheEntry.Timestamp.After(time.Now().Add(time.Duration(-cacheTtl)*time.Second)) {
		// Found row
		// log.Println("Cache Hit!", reflect.TypeOf(data))
		bufPtr := bytes.NewBuffer(cacheEntry.Value)
		err := gob.NewDecoder(bufPtr).Decode(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cache data malformed!")
			goto cacheDataMalformed
		}
		return data, nil
	}
	cacheDataMalformed:
	tmp := cb()
	data = &tmp
	var buf bytes.Buffer
	gob.NewEncoder(&buf).Encode(tmp)

	// bufptr := bytes.NewBuffer(buf.Bytes())
	// tmp2 := new(T)
	// err2 := gob.NewDecoder(bufptr).Decode(tmp2)
	// fmt.Println("tmp start printing")
	// fmt.Println(tmp2)
	// fmt.Println("tmp done printing")
	// fmt.Println(reflect.TypeOf(tmp2))
	// if err2 != nil {
	// 	log.Panicln("error here", err2)
	// }

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
