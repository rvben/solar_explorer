package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rvben/solar_exporter/models"
	"github.com/rvben/solar_exporter/services"
)

var ACCOUNT = os.Getenv("SEMS_ACCOUNT")
var PASSWORD = os.Getenv("SEMS_PASSWORD")
var SITE = os.Getenv("SEMS_SITE")

var (
	powerNow = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "solar_power_now",
			Help: "Power Now in W",
		},
		[]string{"site"},
	)
	dayRecord = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "solar_day_record",
			Help: "Day Record in Wh",
		},
		[]string{"site", "date"},
	)
	energyToday = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "solar_energy_today",
			Help: "Today's Energy in Wh",
		},
		[]string{"site"},
	)
	energyMonth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "solar_energy_month",
			Help: "Monthly Energy in Wh",
		},
		[]string{"site"},
	)
	energyYear = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "solar_energy_year",
			Help: "Yearly Energy in Wh",
		},
		[]string{"site"},
	)
	energyTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "solar_energy_total",
			Help: "Total Energy in Wh",
		},
		[]string{"site"},
	)
)

func retrieveMetrics(p services.SolarStatusProvider) (err error) {
	Site := p.Site()
	status, err := p.GetSolarStatus()
	if err != nil {
		return
	}
	powerNow.WithLabelValues(Site).Set(status.PowerNow)
	energyToday.WithLabelValues(Site).Set(status.EnergyToday)
	energyTotal.WithLabelValues(Site).Set(status.EnergyTotal)

	models.SaveTodayValue(status.EnergyToday)
	monthTotal := models.GetMonthTotal()
	energyMonth.WithLabelValues(Site).Set(monthTotal)
	yearTotal := models.GetYearTotal()
	energyYear.WithLabelValues(Site).Set(yearTotal)

	record_date, value := models.GetDayRecord()
	dayRecord.Reset()
	dayRecord.WithLabelValues(Site, record_date).Set(value)
	return
}

func recordMetrics(provider services.SolarStatusProvider) {
	go func() {
		for {
			err := retrieveMetrics(provider)
			if err != nil {
				log.Println("Could not retrieve metrics.")
			}
			time.Sleep(15 * time.Second)
		}
	}()
}

func main() {
	prometheus.MustRegister(powerNow)
	prometheus.MustRegister(energyToday)
	prometheus.MustRegister(energyMonth)
	prometheus.MustRegister(energyYear)
	prometheus.MustRegister(energyTotal)
	prometheus.MustRegister(dayRecord)

	databaseDir := "/tmp/data"
	databaseFile := fmt.Sprintf("%s/%s.db", databaseDir, SITE)

	_, err := os.Stat(databaseDir)
	if os.IsNotExist(err) {
		log.Fatalf("Folder [%s] does not exist.\n", databaseDir)
	}
	err = models.InitDB(databaseFile)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/metrics", promhttp.Handler())

	PORT, exists := os.LookupEnv("SOLAR_EXP_PORT")
	if !exists {
		PORT = "2121"
	}
	provider := services.NewSemsProvider(SITE, ACCOUNT, PASSWORD)
	recordMetrics(provider)

	log.Println(fmt.Sprintf("Starting server at :%s", PORT))
	http.ListenAndServe(fmt.Sprintf(":%s", PORT), nil)
}
