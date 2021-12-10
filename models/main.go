package models

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB
var dbPath string

func getDB() *sql.DB {
	if db == nil {
		db, _ = sql.Open("sqlite", dbPath)
	}
	return db
}

func InitDB(dataSourceName string) error {
	var err error
	dbPath = dataSourceName

	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	statement, _ := db.Prepare("CREATE TABLE IF NOT EXISTS daily (id INTEGER PRIMARY KEY, date TEXT UNIQUE, value REAL);")
	statement.Exec()

	return db.Ping()
}

func SaveTodayValue(value float64) {
	date := time.Now().Format("2006-01-02")
	SaveDailyValue(date, value)
}

func SaveDailyValue(day string, value float64) {
	DB := getDB()
	statement, _ := DB.Prepare("INSERT INTO daily (date, value) VALUES (?,?) ON CONFLICT(date) DO UPDATE SET value=excluded.value;")
	statement.Exec(day, value)
}

func GetDayRecord() (string, float64) {
	DB := getDB()
	row, err := DB.Query("SELECT date, value FROM daily WHERE value = (SELECT MAX(value) FROM daily);")
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	var date string
	var value float64
	row.Next()
	row.Scan(&date, &value)
	return date, value
}

func GetMonthTotal() float64 {
	DB := getDB()
	month := time.Now().Format("2006-01")
	row, err := DB.Query(fmt.Sprintf("SELECT SUM(value) FROM daily WHERE date LIKE '%s%%';", month))
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()

	var value float64
	row.Next()
	row.Scan(&value)
	return value
}

func GetYearTotal() float64 {
	DB := getDB()
	year := time.Now().Format("2006")
	row, err := DB.Query(fmt.Sprintf("SELECT SUM(value) FROM daily WHERE date LIKE '%s%%';", year))
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()

	var value float64
	row.Next()
	row.Scan(&value)
	return value
}
