package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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
}

func (p *SemsProvider) Site() string {
	return p.site
}

func NewSemsProvider(site, user, password string) *SemsProvider {
	token, cookie := login(user, password)
	return &SemsProvider{site: site, user: user, password: password, token: token, cookie: cookie}
}

func login(user, password string) (token string, cookie string) {
	log.Printf("Logging in as user [%s]", user)
	endpoint := "https://www.semsportal.com/Home/Login"
	data := url.Values{}
	data.Set("account", user)
	data.Set("pwd", password)

	client := &http.Client{}
	r, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatal(err)
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	res, err := client.Do(r)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
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
		log.Fatalf("Failed to log in as user [%s]", user)
	}
	log.Printf("Succesfully logged in as user [%s]", user)
	return token, cookie
}

func (p *SemsProvider) relogin() {
	token, cookie := login(p.user, p.password)
	p.token = token
	p.cookie = cookie
}

func (p *SemsProvider) GetSolarStatus() (models.SolarStatus, error) {
	p.relogin()
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	endpoint := "https://www.semsportal.com/GopsApi/Post?s=v1/PowerStation/GetMonitorDetailByPowerstationId"
	data := url.Values{}
	data.Set("str", fmt.Sprintf("{\"api\":\"v1/PowerStation/GetMonitorDetailByPowerstationId\",\"param\":{\"powerStationId\":\"%s\"}}", p.token))

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	req.Header.Set("Cookie", "ASP.NET_SessionId="+p.cookie)

	res, err := client.Do(req)
	if err != nil {
		return models.SolarStatus{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	bodyBytes, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	rawStatus := struct {
		Language string   `json:"language"`
		Function []string `json:"function"`
		HasError bool     `json:"hasError"`
		Msg      string   `json:"msg"`
		Code     string   `json:"code"`
		Data     struct {
			Inverter []struct {
				IsStored    bool    `json:"is_stored"`
				Name        string  `json:"name"`
				InPac       float64 `json:"in_pac"`
				OutPac      float64 `json:"out_pac"`
				Eday        float64 `json:"eday"`
				Emonth      float64 `json:"emonth"`
				Etotal      float64 `json:"etotal"`
				Status      int     `json:"status"`
				TurnonTime  string  `json:"turnon_time"`
				ReleationID string  `json:"releation_id"`
				Type        string  `json:"type"`
				Capacity    float64 `json:"capacity"`
			}
		}
	}{}
	json.Unmarshal(bodyBytes, &rawStatus)
	if rawStatus.Code != "0" {
		log.Fatalf("Failed to retrieve status: %s", rawStatus.Msg)
	}
	log.Println("Successfully retrieved status.")

	d := rawStatus.Data
	energyToday := d.Inverter[0].Eday * 1000   // Eday is in kW
	energyMonth := d.Inverter[0].Emonth * 1000 // Emonth is in kW
	energyTotal := d.Inverter[0].Etotal * 1000 // Etotal is in kW
	powerNow := d.Inverter[0].OutPac           // OutPac is in W
	status := models.SolarStatus{EnergyToday: energyToday, EnergyMonth: energyMonth, EnergyTotal: energyTotal, PowerNow: powerNow}
	return status, nil
}
