package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/rvben/solar_exporter/models"
)

type GinlongProvider struct {
	site     string
	username string
	password string
	pid      string
	timeout  int
	db       *models.DataBase
}

func (p *GinlongProvider) Site() string {
	return p.site
}

func (p *GinlongProvider) Timeout() int {
	return p.timeout
}

func (p *GinlongProvider) DB() *models.DataBase {
	return p.db
}

func NewGinlongProvider(site, username, password, pid string, timeout int, db *models.DataBase) *GinlongProvider {
	return &GinlongProvider{site: site, username: username, password: password, pid: pid, timeout: timeout, db: db}
}

func (p *GinlongProvider) GetSolarStatus() (*models.SolarStatus, error) {
	params := url.Values{}
	params.Add("userName", p.username)
	params.Add("userNameDisplay", p.username)
	params.Add("password", p.password)
	params.Add("lan", `2`)
	params.Add("userType", `C`)
	body := strings.NewReader(params.Encode())

	req, err := http.NewRequest("POST", "https://m.ginlong.com/cpro/login/validateLogin.json", body)
	if err != nil {
		// handle err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle err
	}
	defer resp.Body.Close()

	jsessionId := ""
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "JSESSIONID" {
			jsessionId = cookie.Value
		}
	}
	if jsessionId == "" {
		return nil, fmt.Errorf("Could not find JSESSIONID in response.")
	}

	params = url.Values{}
	params.Add("plantId", p.pid)
	body = strings.NewReader(params.Encode())

	req, err = http.NewRequest("POST", "https://m.ginlong.com/cpro/epc/plantDetail/showPlantDetailAjax.json", body)
	if err != nil {
		// handle err
	}
	req.Header.Set("Cookie", fmt.Sprintf("JSESSIONID=%s", jsessionId))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		// handle err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body from request: %s", err)
	}

	rawStatus := struct {
		Result struct {
			OwnerUser struct {
				BindRelation int    `json:"bindRelation"`
				CreateTime   int64  `json:"createTime"`
				Email        string `json:"email"`
				IsAccept     int    `json:"isAccept"`
				IsActived    int    `json:"isActived"`
				IsThirdparty int    `json:"isThirdparty"`
				IsVerify     int    `json:"isVerify"`
				LastLogin    int64  `json:"lastLogin"`
				LocaleID     int    `json:"localeId"`
				NickName     string `json:"nickName"`
				Password     string `json:"password"`
				PushSn       string `json:"pushSn"`
				RegDate      int64  `json:"regDate"`
				Terminate    string `json:"terminate"`
				Token        string `json:"token"`
				UID          int    `json:"uid"`
				UpdateTime   int64  `json:"updateTime"`
			} `json:"ownerUser"`
			PlantAllWapper struct {
				Plant struct {
					Address    string `json:"address"`
					Angle      int    `json:"angle"`
					CityID     int    `json:"cityId"`
					CountryID  int    `json:"countryId"`
					CreateDate int64  `json:"createDate"`
					Currency   struct {
						CurrencyCode string `json:"currencyCode"`
						DisplayName  string `json:"displayName"`
						ID           int    `json:"id"`
						NumericCode  string `json:"numericCode"`
						Symbol       string `json:"symbol"`
					} `json:"currency"`
					CurrencyID int     `json:"currencyId"`
					Direction  string  `json:"direction"`
					GridType   int     `json:"gridType"`
					IsDel      int     `json:"isDel"`
					Lat        float64 `json:"lat"`
					Lon        float64 `json:"lon"`
					Minllis    int     `json:"minllis"`
					Name       string  `json:"name"`
					PlantID    int     `json:"plantId"`
					Power      float64 `json:"power"`
					RunDate    int64   `json:"runDate"`
					StateID    int     `json:"stateId"`
					Status     int     `json:"status"`
					TimezoneID int     `json:"timezoneId"`
					Type       int     `json:"type"`
					UpdateDate int64   `json:"updateDate"`
				} `json:"plant"`
				PlantData struct {
					BatterySoc            string  `json:"batterySoc"`
					EnergyMonth           float64 `json:"energyMonth"`
					EnergyToday           float64 `json:"energyToday"`
					EnergyTotal           float64 `json:"energyTotal"`
					EnergyTotalReal       float64 `json:"energyTotalReal"`
					EnergyYear            float64 `json:"energyYear"`
					HoursEnergy           float64 `json:"hoursEnergy"`
					HoursenergyUpdatetime int64   `json:"hoursenergyUpdatetime"`
					IncomeMonth           float64 `json:"incomeMonth"`
					IncomeToday           float64 `json:"incomeToday"`
					IncomeTotal           float64 `json:"incomeTotal"`
					IncomeTotalReal       float64 `json:"incomeTotalReal"`
					IncomeYear            float64 `json:"incomeYear"`
					PlantID               int     `json:"plantId"`
					PlantUpdateTime       int64   `json:"plantUpdateTime"`
					Power                 float64 `json:"power"`
					UpdateTime            int64   `json:"updateTime"`
				} `json:"plantData"`
				PlantDetail struct {
					BenchmarkPrice     float64 `json:"benchmarkPrice"`
					Cost               float64 `json:"cost"`
					EnergyType         int     `json:"energyType"`
					Interest           float64 `json:"interest"`
					Percent            float64 `json:"percent"`
					Pic                string  `json:"pic"`
					PlantID            int     `json:"plantId"`
					Price              float64 `json:"price"`
					PriceNet           float64 `json:"priceNet"`
					Subsidy            float64 `json:"subsidy"`
					SubsidyBuild       float64 `json:"subsidyBuild"`
					SubsidyCity        float64 `json:"subsidyCity"`
					SubsidyCityYears   int     `json:"subsidyCityYears"`
					SubsidyCounty      float64 `json:"subsidyCounty"`
					SubsidyCountyYears int     `json:"subsidyCountyYears"`
					SubsidyLocal       float64 `json:"subsidyLocal"`
					SubsidyLocalYears  int     `json:"subsidyLocalYears"`
					SubsidyYears       int     `json:"subsidyYears"`
					Years              int     `json:"years"`
				} `json:"plantDetail"`
			} `json:"plantAllWapper"`
			Co2            float64 `json:"co2"`
			ParamSelectors struct {
				ParamDaySelectors   string `json:"paramDaySelectors"`
				ParamMonthSelectors string `json:"paramMonthSelectors"`
				ParamYearSelectors  string `json:"paramYearSelectors"`
				ParamAllSelectors   string `json:"paramAllSelectors"`
			} `json:"paramSelectors"`
			Tree float64 `json:"tree"`
			Pic  string  `json:"pic"`
		} `json:"result"`
		State int `json:"state"`
	}{}

	jsonErr := json.Unmarshal(bodyBytes, &rawStatus)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to parse body to json: %s", err)
	}

	d := rawStatus
	powerNow := d.Result.PlantAllWapper.PlantData.Power
	energyToday := d.Result.PlantAllWapper.PlantData.EnergyToday * 1000
	energyMonth := d.Result.PlantAllWapper.PlantData.EnergyMonth * 1000
	energyYear := d.Result.PlantAllWapper.PlantData.EnergyYear * 1000
	energyTotal := d.Result.PlantAllWapper.PlantData.EnergyTotal * 1000

	status := models.SolarStatus{EnergyToday: energyToday, EnergyMonth: energyMonth, EnergyYear: energyYear, EnergyTotal: energyTotal, PowerNow: powerNow}
	return &status, nil
}
