package six_signal

import (
	"encoding/json"
	"errors"
	"factors/model/model"
	"factors/util"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type response struct {
	Company struct {
		Zip              string `json:"zip"`
		NaicsDescription string `json:"naics_description"`
		EmployeeCount    string `json:"employee_count"`
		Country          string `json:"country"`
		Address          string `json:"address"`
		City             string `json:"city"`
		EmployeeRange    string `json:"employee_range"`
		Industry         string `json:"industry"`
		Sic              string `json:"sic"`
		RevenueRange     string `json:"revenue_range"`
		CountryIsoCode   string `json:"country_iso_code"`
		Phone            string `json:"phone"`
		Domain           string `json:"domain"`
		Name             string `json:"name"`
		State            string `json:"state"`
		Region           string `json:"region"`
		Naics            string `json:"naics"`
		AnnualRevenue    string `json:"annual_revenue"`
		SicDescription   string `json:"sic_description"`
	} `json:"company"`
}

func ExecuteSixSignalEnrich(projectId int64, sixSignalKey string, properties *util.PropertiesMap, clientIP string, statusChannel chan int) {
	err := enrichUsingSixSignal(projectId, sixSignalKey, properties, clientIP)

	if err != nil {
		statusChannel <- 0
	}
	statusChannel <- 1
}
func enrichUsingSixSignal(projectId int64, sixSignalKey string, properties *util.PropertiesMap, clientIP string) error {
	if clientIP == "" {
		return errors.New("invalid IP, failed adding user properties")
	}

	url := "https://epsilon.6sense.com/v1/company/details"
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}
	sixSignalKey = "Token " + sixSignalKey
	req.Header.Add("Authorization", sixSignalKey)
	req.Header.Add("X-Forwarded-For", clientIP)

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	var result response
	if err := json.Unmarshal(body, &result); err != nil {
		log.WithFields(log.Fields{"Error": err}).Warn("Cannot unmarshal JSON")
		return err
	}

	//log.WithFields(log.Fields{"clientIP": clientIP, "response": result}).Info("Six Signal Data Logs")

	if zip := result.Company.Zip; zip != "" {
		if c, ok := (*properties)[util.SIX_SIGNAL_ZIP]; !ok || c == "" {
			(*properties)[util.SIX_SIGNAL_ZIP] = zip
		}
	}

	if naicsDesc := result.Company.NaicsDescription; naicsDesc != "" {
		if c, ok := (*properties)[util.SIX_SIGNAL_NAICS_DESCRIPTION]; !ok || c == "" {
			(*properties)[util.SIX_SIGNAL_NAICS_DESCRIPTION] = naicsDesc
		}
	}

	if empCount := result.Company.EmployeeCount; empCount != "" {
		if c, ok := (*properties)[util.SIX_SIGNAL_EMPLOYEE_COUNT]; !ok || c == "" {
			(*properties)[util.SIX_SIGNAL_EMPLOYEE_COUNT] = empCount
		}
	}

	if country := result.Company.Country; country != "" {
		if c, ok := (*properties)[util.SIX_SIGNAL_COUNTRY]; !ok || c == "" {
			(*properties)[util.SIX_SIGNAL_COUNTRY] = country
		}
	}

	if address := result.Company.Address; address != "" {
		if c, ok := (*properties)[util.SIX_SIGNAL_ADDRESS]; !ok || c == "" {
			(*properties)[util.SIX_SIGNAL_ADDRESS] = address
		}
	}

	if city := result.Company.City; city != "" {
		if c, ok := (*properties)[util.SIX_SIGNAL_CITY]; !ok || c == "" {
			(*properties)[util.SIX_SIGNAL_CITY] = city
		}
	}

	if empRange := result.Company.EmployeeRange; empRange != "" {
		if c, ok := (*properties)[util.SIX_SIGNAL_EMPLOYEE_RANGE]; !ok || c == "" {
			(*properties)[util.SIX_SIGNAL_EMPLOYEE_RANGE] = empRange
		}
	}

	if industry := result.Company.Industry; industry != "" {
		if c, ok := (*properties)[util.SIX_SIGNAL_INDUSTRY]; !ok || c == "" {
			(*properties)[util.SIX_SIGNAL_INDUSTRY] = industry
		}
	}

	if sic := result.Company.Sic; sic != "" {
		if c, ok := (*properties)[util.SIX_SIGNAL_SIC]; !ok || c == "" {
			(*properties)[util.SIX_SIGNAL_SIC] = sic
		}
	}

	if revenueRange := result.Company.RevenueRange; revenueRange != "" {
		if c, ok := (*properties)[util.SIX_SIGNAL_REVENUE_RANGE]; !ok || c == "" {
			(*properties)[util.SIX_SIGNAL_REVENUE_RANGE] = revenueRange
		}
	}

	if countryIsoCode := result.Company.CountryIsoCode; countryIsoCode != "" {
		if c, ok := (*properties)[util.SIX_SIGNAL_COUNTRY_ISO_CODE]; !ok || c == "" {
			(*properties)[util.SIX_SIGNAL_COUNTRY_ISO_CODE] = countryIsoCode
		}
	}

	if phone := result.Company.Phone; phone != "" {
		if c, ok := (*properties)[util.SIX_SIGNAL_PHONE]; !ok || c == "" {
			(*properties)[util.SIX_SIGNAL_PHONE] = phone
		}
	}

	if domain := result.Company.Domain; domain != "" {
		if c, ok := (*properties)[util.SIX_SIGNAL_DOMAIN]; !ok || c == "" {
			(*properties)[util.SIX_SIGNAL_DOMAIN] = domain
			model.SetSixSignalAPICountCacheResult(projectId, util.TimeZoneStringIST)
		}
	}

	if name := result.Company.Name; name != "" {
		if c, ok := (*properties)[util.SIX_SIGNAL_NAME]; !ok || c == "" {
			(*properties)[util.SIX_SIGNAL_NAME] = name
		}
	}

	if state := result.Company.State; state != "" {
		if c, ok := (*properties)[util.SIX_SIGNAL_STATE]; !ok || c == "" {
			(*properties)[util.SIX_SIGNAL_STATE] = state
		}
	}

	if region := result.Company.Region; region != "" {
		if c, ok := (*properties)[util.SIX_SIGNAL_REGION]; !ok || c == "" {
			(*properties)[util.SIX_SIGNAL_REGION] = region
		}
	}

	if naics := result.Company.Naics; naics != "" {
		if c, ok := (*properties)[util.SIX_SIGNAL_NAICS]; !ok || c == "" {
			(*properties)[util.SIX_SIGNAL_NAICS] = naics
		}
	}

	if annualRevenue := result.Company.AnnualRevenue; annualRevenue != "" {
		if c, ok := (*properties)[util.SIX_SIGNAL_ANNUAL_REVENUE]; !ok || c == "" {
			(*properties)[util.SIX_SIGNAL_ANNUAL_REVENUE] = annualRevenue
		}
	}

	if sicDesc := result.Company.SicDescription; sicDesc != "" {
		if c, ok := (*properties)[util.SIX_SIGNAL_SIC_DESCRIPTION]; !ok || c == "" {
			(*properties)[util.SIX_SIGNAL_SIC_DESCRIPTION] = sicDesc
		}
	}

	return nil
}
