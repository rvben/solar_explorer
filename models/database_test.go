package models

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func prepareDB(t *testing.T) (*DataBase, func()) {
	dbfile, err := ioutil.TempFile("", "testdb.*.db")
	if err != nil {
		t.Fatalf("Error creating temporary database: %v", err)
	}
	dbPath := dbfile.Name()
	dbfile.Close()

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Error creating database: %v", err)
	}

	return db, func() {
		db.DB.Close()
		os.Remove(dbPath)
	}
}

func TestSaveDailyValue(t *testing.T) {
	db, cleanup := prepareDB(t)
	defer cleanup()

	date := time.Now().Format("2006-01-02")
	value := 123.45

	err := db.SaveDailyValue(date, value)
	if err != nil {
		t.Fatalf("Error saving daily value: %v", err)
	}

	retrievedValue, err := db.GetDailyValue(date)
	if err != nil {
		t.Fatalf("Error retrieving daily value: %v", err)
	}
	if retrievedValue != value {
		t.Fatalf("Retrieved value does not match saved value. Expected: %f, Got: %f", value, retrievedValue)
	}
}

func TestGetDayRecord(t *testing.T) {
	db, cleanup := prepareDB(t)
	defer cleanup()

	date := time.Now().Format("2006-01-02")
	value := 123.45

	_, _, err := db.GetDayRecord()
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	if err.Error() != "no records found" {
		t.Fatalf("Expected error message 'no records found', got '%s'", err.Error())
	}

	err = db.SaveDailyValue(date, value)
	if err != nil {
		t.Fatalf("Error saving daily value: %v", err)
	}

	retrievedDate, retrievedValue, err := db.GetDayRecord()
	if err != nil {
		t.Fatalf("Error retrieving daily record: %v", err)
	}
	if retrievedDate != date {
		t.Fatalf("Retrieved date does not match saved date. Expected: %s, Got: %s", date, retrievedDate)
	}
	if retrievedValue != value {
		t.Fatalf("Retrieved value does not match saved value. Expected: %f, Got: %f", value, retrievedValue)
	}
}

func TestGetMonthTotal(t *testing.T) {
	db, cleanup := prepareDB(t)
	defer cleanup()

	month := time.Now().Format("2006-01")
	value1 := 123.45
	value2 := 678.90

	date1 := month + "-01"
	date2 := month + "-02"

	err := db.SaveDailyValue(date1, value1)
	if err != nil {
		t.Fatalf("Error saving daily value: %v", err)
	}
	err = db.SaveDailyValue(date2, value2)
	if err != nil {
		t.Fatalf("Error saving daily value: %v", err)
	}

	retrievedValue, err := db.GetMonthTotal()
	if err != nil {
		t.Fatalf("Error retrieving month total: %v", err)
	}
	expectedValue := value1 + value2
	if retrievedValue != expectedValue {
		t.Fatalf("Retrieved value does not match expected value. Expected: %f, Got: %f", expectedValue, retrievedValue)
	}
}

func TestGetYearTotal(t *testing.T) {
	db, cleanup := prepareDB(t)
	defer cleanup()

	year := time.Now().Format("2006")
	value1 := 123.45
	value2 := 678.90

	date1 := year + "-01-01"
	date2 := year + "-02-01"

	err := db.SaveDailyValue(date1, value1)
	if err != nil {
		t.Fatalf("Error saving daily value: %v", err)
	}
	err = db.SaveDailyValue(date2, value2)
	if err != nil {
		t.Fatalf("Error saving daily value: %v", err)
	}

	retrievedValue, err := db.GetYearTotal()
	if err != nil {
		t.Fatalf("Error retrieving year total: %v", err)
	}
	expectedValue := value1 + value2
	if retrievedValue != expectedValue {
		t.Fatalf("Retrieved value does not match expected value. Expected: %f, Got: %f", expectedValue, retrievedValue)
	}
}

func TestSaveTodayValue(t *testing.T) {
	db, cleanup := prepareDB(t)
	defer cleanup()
	date := time.Now().Format("2006-01-02")
	value1 := 123.45
	value2 := 678.90

	err := db.SaveTodayValue(value1)
	if err != nil {
		t.Fatalf("Error saving today's value: %v", err)
	}
	err = db.SaveTodayValue(value2)
	if err != nil {
		t.Fatalf("Error saving today's value: %v", err)
	}

	retrievedValue, err := db.GetDailyValue(date)
	if err != nil {
		t.Fatalf("Error retrieving daily value: %v", err)
	}
	if retrievedValue != value2 {
		t.Fatalf("Retrieved value does not match expected value. Expected: %f, Got: %f", value2, retrievedValue)
	}
}
