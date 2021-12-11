package models

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

type DataBase struct {
	DB     *sql.DB
	dbPath string
}

func NewDB(dbPath string) (*DataBase, error) {
	var err error
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	statement, _ := db.Prepare("CREATE TABLE IF NOT EXISTS daily (id INTEGER PRIMARY KEY, date TEXT UNIQUE, value REAL);")
	statement.Exec()
	// log.Printf("Initialized database at [%s]\n", dbPath)
	return &DataBase{DB: db, dbPath: dbPath}, db.Ping()
}

func (d *DataBase) SaveTodayValue(value float64) {
	// log.Printf("Saving today value [%f] at [%s]\n", value, d.dbPath)
	date := time.Now().Format("2006-01-02")
	d.SaveDailyValue(date, value)
}

func (d *DataBase) SaveDailyValue(day string, value float64) {
	statement, _ := d.DB.Prepare("INSERT INTO daily (date, value) VALUES (?,?) ON CONFLICT(date) DO UPDATE SET value=excluded.value;")
	statement.Exec(day, value)
}

func (d *DataBase) GetDayRecord() (string, float64) {
	row, err := d.DB.Query("SELECT date, value FROM daily WHERE value = (SELECT MAX(value) FROM daily);")
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

func (d *DataBase) GetMonthTotal() float64 {
	month := time.Now().Format("2006-01")
	row, err := d.DB.Query(fmt.Sprintf("SELECT SUM(value) FROM daily WHERE date LIKE '%s%%';", month))
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()

	var value float64
	row.Next()
	row.Scan(&value)
	return value
}

func (d *DataBase) GetYearTotal() float64 {
	year := time.Now().Format("2006")
	row, err := d.DB.Query(fmt.Sprintf("SELECT SUM(value) FROM daily WHERE date LIKE '%s%%';", year))
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()

	var value float64
	row.Next()
	row.Scan(&value)
	return value
}
