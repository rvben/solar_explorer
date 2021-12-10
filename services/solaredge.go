package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/rvben/solar_exporter/models"
)

type SolarEdgeProvider struct {
	pid     string
	api_key string
	site    string
	timeout int
}

func (p *SolarEdgeProvider) Site() string {
	return p.site
}

func (p *SolarEdgeProvider) Timeout() int {
	return p.timeout
}

func NewSolarEdgeProvider(site, api_key, pid string, timeout int) *SolarEdgeProvider {
	return &SolarEdgeProvider{site: site, pid: pid, api_key: api_key, timeout: timeout}
}

func (p *SolarEdgeProvider) GetSolarStatus() (*models.SolarStatus, error) {
	url := fmt.Sprintf("https://monitoringapi.solaredge.com/site/%s/overview?api_key=%s", p.pid, p.api_key)
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request for url [%s]: %s", url, err)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could succesfully finish request [%s]: %s", url, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body from request: %s", err)
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}
	// TODO: if status 424 - Too Many Requests
	// bodyStr := string(body)
	// log.Println(bodyStr)

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
		return nil, fmt.Errorf("failed to parse body to json: %s", err)
	}

	d := rawStatus.Overview
	powerNow := d.CurrentPower.Power
	energyToday := d.LastDayData.Energy
	energyMonth := d.LastMonthData.Energy
	energyYear := d.LastYearData.Energy
	energyTotal := d.LifeTimeData.Energy
	status := models.SolarStatus{EnergyToday: energyToday, EnergyMonth: energyMonth, EnergyYear: energyYear, EnergyTotal: energyTotal, PowerNow: powerNow}
	return &status, nil
}
