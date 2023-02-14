package models

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

type DataBase struct {
	DB        *sql.DB
	dbPath    string
	stmt      *sql.Stmt
	getStmt   *sql.Stmt
	dayStmt   *sql.Stmt
	monthStmt *sql.Stmt
	yearStmt  *sql.Stmt
}

func NewDB(dbPath string) (*DataBase, error) {
	var err error
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	statement, _ := tx.Prepare("CREATE TABLE IF NOT EXISTS daily (id INTEGER PRIMARY KEY, date TEXT UNIQUE, value REAL);")
	_, err = statement.Exec()
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	log.Printf("Initialized database at [%s]\n", dbPath)
	return &DataBase{DB: db, dbPath: dbPath}, db.Ping()
}

func (d *DataBase) SaveTodayValue(value float64) error {
	date := time.Now().Format("2006-01-02")
	oldValue, err := d.GetDailyValue(date)
	if err != nil {
		return err
	}
	if value != oldValue {
		err = d.SaveDailyValue(date, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DataBase) GetDailyValue(day string) (float64, error) {
	if d.dayStmt == nil {
		var err error
		d.dayStmt, err = d.DB.Prepare("SELECT value FROM daily WHERE date = ?;")
		if err != nil {
			return 0, err
		}
	}

	row := d.dayStmt.QueryRow(day)
	var value float64
	err := row.Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}

	return value, nil
}

func (d *DataBase) SaveDailyValue(day string, value float64) error {
	tx, err := d.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if d.stmt == nil {
		d.stmt, err = d.DB.Prepare("INSERT INTO daily (date, value) VALUES (?,?) ON CONFLICT(date) DO UPDATE SET value=excluded.value;")
		if err != nil {
			return err
		}
	}

	_, err = d.stmt.Exec(day, value)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (d *DataBase) GetDayRecord() (string, float64, error) {
	if d.getStmt == nil {
		var err error
		d.getStmt, err = d.DB.Prepare("SELECT date, value FROM daily WHERE value = (SELECT MAX(value) FROM daily);")
		if err != nil {
			return "", 0, err
		}
	}

	row := d.getStmt.QueryRow()
	var date string
	var value float64
	err := row.Scan(&date, &value)
	if err == sql.ErrNoRows {
		return "", 0, fmt.Errorf("no records found")
	} else if err != nil {
		return "", 0, err
	}

	return date, value, nil
}

func (d *DataBase) GetMonthTotal() (float64, error) {
	month := time.Now().Format("2006-01")
	if d.monthStmt == nil {
		var err error
		d.monthStmt, err = d.DB.Prepare("SELECT COALESCE(SUM(value), 0) FROM daily WHERE date LIKE ?;")
		if err != nil {
			return 0, err
		}
	}

	row := d.monthStmt.QueryRow(fmt.Sprintf("%s%%", month))
	var value float64
	err := row.Scan(&value)
	if err != nil {
		return 0, err
	}

	return value, nil
}

func (d *DataBase) GetYearTotal() (float64, error) {
	year := time.Now().Format("2006")
	if d.yearStmt == nil {
		var err error
		d.yearStmt, err = d.DB.Prepare("SELECT COALESCE(SUM(value), 0) FROM daily WHERE date LIKE ?;")
		if err != nil {
			return 0, err
		}
	}

	row := d.yearStmt.QueryRow(fmt.Sprintf("%s%%", year))
	var value float64
	err := row.Scan(&value)
	if err != nil {
		return 0, err
	}

	return value, nil
}
