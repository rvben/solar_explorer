package services

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rvben/solar_exporter/models"
)

type OmnikProvider struct {
	pid      string
	base_url string
	site     string
	timeout  int
	db       *models.DataBase
}

func (p *OmnikProvider) Site() string {
	return p.site
}

func (p *OmnikProvider) Timeout() int {
	return p.timeout
}

func (p *OmnikProvider) DB() *models.DataBase {
	return p.db
}

func NewOmnikProvider(site, base_url, pid string, timeout int, db *models.DataBase) *OmnikProvider {
	return &OmnikProvider{site: site, pid: pid, base_url: base_url, timeout: timeout, db: db}
}

func (p *OmnikProvider) GetSolarStatus() (*models.SolarStatus, error) {
	transport := http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{
		Timeout:   60 * time.Second,
		Transport: &transport,
	}

	url := fmt.Sprintf("%s/Terminal/TerminalMain.aspx?pid=%s", p.base_url, p.pid)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request for url [%s]: %s", url, err)
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could succesfully finish request [%s]: %s", url, err)
	}
	defer res.Body.Close()

	url = fmt.Sprintf("%s/AjaxService.ashx?ac=upTerminalMain&psid=%s&random=%f", p.base_url, p.pid, rand.Float32())
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request for url [%s]: %s", url, err)
	}

	for _, cookie := range res.Cookies() {
		if cookie.Name == "ASP.NET_SessionId" {
			req.Header.Set("Cookie", "ASP.NET_SessionId="+cookie.Value)
		}
	}

	res, err = client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could succesfully finish request [%s]: %s", url, err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body from request: %s", err)
	}

	rawStatus := []struct {
		Nowpower     string `json:"nowpower"`
		Daypower     string `json:"daypower"`
		Monthpower   string `json:"monthpower"`
		Yearpower    string `json:"yearpower"`
		Allpower     string `json:"allpower"`
		Lasttime     string `json:"lasttime"`
		Commissioned string `json:"commissioned"`
		Capacity     string `json:"capacity"`
		Installer    string `json:"installer"`
		Peakpower    string `json:"peakpower"`
		Efficiency   string `json:"efficiency"`
		Treesplanted string `json:"treesplanted"`
		Co2          string `json:"co2"`
		Income       string `json:"income"`
	}{}

	jsonErr := json.Unmarshal(bodyBytes, &rawStatus)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to parse body to json: %s", err)
	}

	d := rawStatus[0]
	powerNow, err := convertRawToFloatWatt(d.Nowpower)
	if err != nil {
		return nil, err
	}
	energyToday, err := convertRawToFloatWatt(d.Daypower)
	if err != nil {
		return nil, err
	}
	energyMonth, err := convertRawToFloatWatt(d.Monthpower)
	if err != nil {
		return nil, err
	}
	energyYear, err := convertRawToFloatWatt(d.Yearpower)
	if err != nil {
		energyYear = 0
	}
	energyTotal, err := convertRawToFloatWatt(d.Allpower)
	if err != nil {
		return nil, err
	}

	status := models.SolarStatus{EnergyToday: energyToday, EnergyMonth: energyMonth, EnergyYear: energyYear, EnergyTotal: energyTotal, PowerNow: powerNow}
	return &status, nil
}

func convertRawToFloatWatt(raw string) (float64, error) {
	var multiplier float64
	var valueString string

	if strings.Contains(raw, "kWh") {
		valueString = strings.Replace(raw, " kWh", "", -1)
		multiplier = 1000
	} else if strings.Contains(raw, "MWh") {
		valueString = strings.Replace(raw, " MWh", "", -1)
		multiplier = 1000000
	} else if strings.Contains(raw, "kW") {
		valueString = strings.Replace(raw, " kW", "", -1)
		multiplier = 1000
	}

	value, err := strconv.ParseFloat(valueString, 64)
	if err != nil {
		return -1.0, fmt.Errorf("could not convert [%s] to float: %s", raw, err)
	}
	return value * multiplier, err
}
