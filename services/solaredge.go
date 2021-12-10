package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/rvben/solar_exporter/models"
)

type SolarEdgeProvider struct {
	pid     string
	api_key string
	site    string
}

func (p *SolarEdgeProvider) Site() string {
	return p.site
}

func NewSolarEdgeProvider(site, api_key, pid string) *SolarEdgeProvider {
	return &SolarEdgeProvider{site: site, pid: pid, api_key: api_key}
}

func (p *SolarEdgeProvider) GetSolarStatus() (models.SolarStatus, error) {

	url := fmt.Sprintf("https://monitoringapi.solaredge.com/site/%s/overview?api_key=%s", p.pid, p.api_key)

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	res, getErr := client.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	// TODO: if status 424 - Too Many Requests
	bodyStr := string(body)
	log.Println(bodyStr)

	rawStatus := struct {
		Overview struct {
			LastUpdateTime string `json:"lastUpdateTime"`
			LifeTimeData   struct {
				Energy  float64 `json:"energy"`
				Revenue float64 `json:"revenue"`
			} `json:"lifeTimeData"`
			LastYearData struct {
				Energy float64 `json:"energy"`
			} `json:"lastYearData"`
			LastMonthData struct {
				Energy float64 `json:"energy"`
			} `json:"lastMonthData"`
			LastDayData struct {
				Energy float64 `json:"energy"`
			} `json:"lastDayData"`
			CurrentPower struct {
				Power float64 `json:"power"`
			} `json:"currentPower"`
			MeasuredBy string `json:"measuredBy"`
		} `json:"overview"`
	}{}

	jsonErr := json.Unmarshal(body, &rawStatus)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	d := rawStatus.Overview
	powerNow := d.CurrentPower.Power
	energyToday := d.LastDayData.Energy
	energyMonth := d.LastMonthData.Energy
	energyTotal := d.LifeTimeData.Energy
	status := models.SolarStatus{EnergyToday: energyToday, EnergyMonth: energyMonth, EnergyTotal: energyTotal, PowerNow: powerNow}

	return status, nil
}
