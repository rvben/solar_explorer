package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	neturl "net/url"
	"strconv"
	"strings"
	"time"

	"github.com/rvben/solar_exporter/models"
)

type SemsProvider struct {
	user     string
	password string
	token    string
	cookie   string
	site     string
	timeout  int
	db       *models.DataBase
}

func (p *SemsProvider) Site() string {
	return p.site
}

func (p *SemsProvider) Timeout() int {
	return p.timeout
}

func (p *SemsProvider) DB() *models.DataBase {
	return p.db
}

func NewSemsProvider(site, user, password string, timeout int, db *models.DataBase) *SemsProvider {
	return &SemsProvider{site: site, user: user, password: password, timeout: timeout, db: db}
}

func (p *SemsProvider) login() error {
	var cookie, token string

	log.Printf("Logging in as user [%s]", p.user)
	url := "https://www.semsportal.com/Home/Login"
	data := neturl.Values{}
	data.Set("account", p.user)
	data.Set("pwd", p.password)

	client := &http.Client{}
	r, err := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("could not create request for url [%s]: %s", url, err)
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	res, err := client.Do(r)
	if err != nil {
		return fmt.Errorf("could succesfully finish request [%s]: %s", url, err)
	}
	defer res.Body.Close()
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read body from request: %s", err)
	}
	for _, c := range res.Cookies() {
		if c.Name == "ASP.NET_SessionId" {
			cookie = c.Value
		}
	}
	response := struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Redirect string `json:"redirect"`
		} `json:"data"`
	}{}
	json.Unmarshal(bodyBytes, &response)
	split := strings.Split(response.Data.Redirect, "/")
	token = split[len(split)-1]
	if response.Code != 0 {
		return fmt.Errorf("failed to log in as user [%s]", p.user)
	}
	log.Printf("%s - Succesfully logged in as user [%s]\n", p.site, p.user)
	p.token = token
	p.cookie = cookie
	return nil
}

func (p *SemsProvider) GetSolarStatus() (*models.SolarStatus, error) {
	err := p.login()
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	url := "https://www.semsportal.com/GopsApi/Post?s=v3/PowerStation/GetMonitorDetailByPowerstationId"
	data := neturl.Values{}
	data.Set("str", fmt.Sprintf("{\"api\":\"v3/PowerStation/GetMonitorDetailByPowerstationId\",\"param\":{\"powerStationId\":\"%s\"}}", p.token))

	req, err := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("could not create request for url [%s]: %s", url, err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	req.Header.Set("Cookie", "ASP.NET_SessionId="+p.cookie)

	res, err := client.Do(req)
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
	rawStatus := struct {
		Language string      `json:"language"`
		Function interface{} `json:"function"`
		HasError bool        `json:"hasError"`
		Msg      string      `json:"msg"`
		Code     string      `json:"code"`
		Data     struct {
			Kpi struct {
				MonthGeneration float64 `json:"month_generation"`
				Pac             float64 `json:"pac"`
				Power           float64 `json:"power"`
				TotalPower      float64 `json:"total_power"`
				DayIncome       float64 `json:"day_income"`
				TotalIncome     float64 `json:"total_income"`
				YieldRate       float64 `json:"yield_rate"`
				Currency        string  `json:"currency"`
			} `json:"kpi"`
		}
	}{}
	json.Unmarshal(bodyBytes, &rawStatus)
	if rawStatus.Code != "0" {
		return nil, fmt.Errorf("failed to retrieve status for site [%s]: %s", p.site, rawStatus.Msg)
	}

	d := rawStatus.Data
	energyToday := d.Kpi.Power * 1000           // Eday is in kW
	energyMonth := d.Kpi.MonthGeneration * 1000 // Emonth is in kW
	energyTotal := d.Kpi.TotalPower * 1000      // Etotal is in kW
	powerNow := d.Kpi.Pac                       // Pac is in W
	status := models.SolarStatus{EnergyToday: energyToday, EnergyMonth: energyMonth, EnergyTotal: energyTotal, PowerNow: powerNow}
	return &status, nil
}
