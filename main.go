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
	"gopkg.in/yaml.v2"
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

func retrieveMetrics(p services.SolarStatusProvider) error {
	Site := p.Site()

	log.Printf("%s - Start retrieving status.\n", Site)
	status, err := p.GetSolarStatus()
	dayRecord.Reset()
	if err != nil {
		return err
	}
	log.Printf("%s - Successfully retrieved status.\n", Site)

	powerNow.WithLabelValues(Site).Set(status.PowerNow)
	energyToday.WithLabelValues(Site).Set(status.EnergyToday)
	energyTotal.WithLabelValues(Site).Set(status.EnergyTotal)

	log.Printf("%s - Synchronizing values with database.\n", Site)
	models.SaveTodayValue(status.EnergyToday)
	monthTotal := models.GetMonthTotal()
	if status.EnergyMonth > monthTotal {
		monthTotal = status.EnergyMonth
	}
	energyMonth.WithLabelValues(Site).Set(monthTotal)
	yearTotal := models.GetYearTotal()
	if status.EnergyMonth > yearTotal {
		yearTotal = status.EnergyYear
	}
	energyYear.WithLabelValues(Site).Set(yearTotal)
	record_date, value := models.GetDayRecord()
	dayRecord.WithLabelValues(Site, record_date).Set(value)
	log.Printf("%s - Synchronized with database.\n", Site)
	return nil
}

func recordMetrics(p services.SolarStatusProvider) {
	go func() {
		for {
			err := retrieveMetrics(p)
			if err != nil {
				log.Printf("%s - Could not retrieve metrics: %s", p.Site(), err)
			}
			time.Sleep(time.Second * time.Duration(p.Timeout()))
		}
	}()
}

type Config struct {
	Server struct {
		Port           string `yaml:"port"`
		DbDir          string `yaml:"db_dir"`
		DefaultTimeout int    `yaml:"default_timeout"`
	} `yaml:"server"`
	SolarEdge []struct {
		Site    string `yaml:"site"`
		APIKey  string `yaml:"api_key"`
		Pid     string `yaml:"pid"`
		Timeout int    `yaml:"timeout"`
	} `yaml:"solaredge"`
	Omnik []struct {
		Site    string `yaml:"site"`
		Pid     string `yaml:"pid"`
		BaseURL string `yaml:"base_url"`
		Timeout int    `yaml:"timeout"`
	} `yaml:"omnik"`
	Sems []struct {
		Site     string `yaml:"site"`
		Account  string `yaml:"account"`
		Password string `yaml:"password"`
		Timeout  int    `yaml:"timeout"`
	} `yaml:"sems"`
}

func NewConfig(configPath string) (*Config, error) {
	config := &Config{}
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

func ValidateConfigPath(path string) error {
	s, err := os.Stat(path)
	if os.IsNotExist(err) {
		log.Fatalf("File [%s] does not exist.\n", path)
	}
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file", path)
	}
	return nil
}

func main() {
	cfg, err := NewConfig("config.yml")
	if err != nil {
		log.Fatal(err)
	}

	prometheus.MustRegister(powerNow)
	prometheus.MustRegister(energyToday)
	prometheus.MustRegister(energyMonth)
	prometheus.MustRegister(energyYear)
	prometheus.MustRegister(energyTotal)
	prometheus.MustRegister(dayRecord)

	databaseDir := cfg.Server.DbDir

	_, err = os.Stat(databaseDir)
	if os.IsNotExist(err) {
		log.Fatalf("Folder [%s] does not exist.\n", databaseDir)
	}

	var providers []services.SolarStatusProvider

	// Load all providers
	for _, p := range cfg.Omnik {
		timeout := p.Timeout
		if timeout == 0 {
			timeout = cfg.Server.DefaultTimeout
		}
		provider := services.NewOmnikProvider(p.Site, p.BaseURL, p.Pid, timeout)
		providers = append(providers, provider)
	}

	for _, p := range cfg.SolarEdge {
		timeout := p.Timeout
		if timeout == 0 {
			timeout = cfg.Server.DefaultTimeout
		}
		provider := services.NewSolarEdgeProvider(p.Site, p.APIKey, p.Pid, timeout)
		providers = append(providers, provider)
	}

	for _, p := range cfg.Sems {
		timeout := p.Timeout
		if timeout == 0 {
			timeout = cfg.Server.DefaultTimeout
		}
		provider := services.NewSemsProvider(p.Site, p.Account, p.Password, timeout)
		providers = append(providers, provider)
	}

	for _, p := range providers {
		databaseFile := fmt.Sprintf("%s/%s.db", databaseDir, p.Site())
		err = models.InitDB(databaseFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Start Metrics Collection
	for _, p := range providers {
		recordMetrics(p)
	}

	// Start server
	http.Handle("/metrics", promhttp.Handler())
	log.Println(fmt.Sprintf("Starting server at :%s", cfg.Server.Port))
	http.ListenAndServe(fmt.Sprintf(":%s", cfg.Server.Port), nil)
}
