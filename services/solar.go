package services

import "github.com/rvben/solar_exporter/models"

type SolarStatusProvider interface {
	GetSolarStatus() (*models.SolarStatus, error)
	Site() string
	Timeout() int
	DB() *models.DataBase
}
