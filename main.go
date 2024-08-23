package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"flag"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rvben/solar_exporter/models"
	"github.com/rvben/solar_exporter/services"
	"gopkg.in/yaml.v2"
)

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

	log.Printf("%s - Start retrieving status from provider %T.\n", Site, p)
	status, err := p.GetSolarStatus()
	if err != nil {
		return err
	}
	log.Printf("%s - Successfully retrieved status from provider %T.\n", Site, p)

	powerNow.WithLabelValues(Site).Set(status.PowerNow)
	energyToday.WithLabelValues(Site).Set(status.EnergyToday)
	energyTotal.WithLabelValues(Site).Set(status.EnergyTotal)

	log.Printf("%s - Synchronizing values with database.\n", Site)
	p.DB().SaveTodayValue(status.EnergyToday)
	monthTotal, err := p.DB().GetMonthTotal()
	if err != nil {
		log.Printf("%s - Error retrieving month total: %s", Site, err)
		return err
	}
	if status.EnergyMonth > monthTotal {
		monthTotal = status.EnergyMonth
	}
	energyMonth.WithLabelValues(Site).Set(monthTotal)

	yearTotal, err := p.DB().GetYearTotal()
	if err != nil {
		log.Printf("%s - Error retrieving year total: %s", Site, err)
		return err
	}
	if status.EnergyYear > yearTotal {
		yearTotal = status.EnergyYear
	}
	energyYear.WithLabelValues(Site).Set(yearTotal)

	record_date, value, err := p.DB().GetDayRecord()
	if err != nil {
		log.Printf("Could not get DayRecord: %s", err)
	}
	dayRecord.DeleteLabelValues(Site)
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
	Sems []struct {
		Site     string `yaml:"site"`
		Account  string `yaml:"account"`
		Password string `yaml:"password"`
		Timeout  int    `yaml:"timeout"`
	} `yaml:"sems"`
	Ginlong []struct {
		Site     string `yaml:"site"`
		Pid      string `yaml:"pid"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Timeout  int    `yaml:"timeout"`
	} `yaml:"ginlong"`
}

func NewConfig(configPath string) (*Config, error) {
	config := &Config{}
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Validate configuration values
	if config.Server.Port == "" {
		return nil, fmt.Errorf("server port is required")
	}
	if config.Server.DbDir == "" {
		return nil, fmt.Errorf("database directory is required")
	}
	// Add more validation as needed

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
	configPath := flag.String("config", "config.yml", "path to config file")
	flag.Parse()

	if err := ValidateConfigPath(*configPath); err != nil {
		log.Fatal(err)
	}

	cfg, err := NewConfig(*configPath)
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
	for _, p := range cfg.Ginlong {
		timeout := p.Timeout
		if timeout == 0 {
			timeout = cfg.Server.DefaultTimeout
		}
		databaseFile := fmt.Sprintf("%s/%s.db", databaseDir, p.Site)
		db, err := models.NewDB(databaseFile)
		if err != nil {
			log.Fatal(err)
		}
		if p.Pid == "" {
			p.Pid = "172533"
		}
		provider := services.NewGinlongProvider(p.Site, p.Username, p.Password, p.Pid, timeout, db)
		providers = append(providers, provider)
	}

	for _, p := range cfg.SolarEdge {
		timeout := p.Timeout
		if timeout == 0 {
			timeout = cfg.Server.DefaultTimeout
		}
		databaseFile := fmt.Sprintf("%s/%s.db", databaseDir, p.Site)
		db, err := models.NewDB(databaseFile)
		if err != nil {
			log.Fatal(err)
		}
		provider := services.NewSolarEdgeProvider(p.Site, p.APIKey, p.Pid, timeout, db)
		providers = append(providers, provider)
	}

	for _, p := range cfg.Sems {
		timeout := p.Timeout
		if timeout == 0 {
			timeout = cfg.Server.DefaultTimeout
		}
		databaseFile := fmt.Sprintf("%s/%s.db", databaseDir, p.Site)
		db, err := models.NewDB(databaseFile)
		if err != nil {
			log.Fatal(err)
		}
		provider := services.NewSemsProvider(p.Site, p.Account, p.Password, timeout, db)
		providers = append(providers, provider)
	}

	// Start Metrics Collection
	for _, p := range providers {
		recordMetrics(p)
	}

	// Start server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: promhttp.Handler(),
	}

	go func() {
		log.Printf("Starting server at :%s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", cfg.Server.Port, err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	if err := server.Shutdown(context.Background()); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exiting")
}
