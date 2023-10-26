package six_signal

import (
	"encoding/json"
	"errors"
	"factors/config"
	"factors/model/model"
	"factors/model/store"
	"factors/util"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

const FACTORS_6SIGNAL = "factors_6sense"
const API_6SIGNAL = "API_6Sense"

type Response struct {
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
	defer close(statusChannel)
	logCtx := log.WithField("project_id", projectId)

	isFactorsAPIKey := false
	if sixSignalKey == config.GetFactorsSixSignalAPIKey() {
		isFactorsAPIKey = true
	}
	err := enrichUsingSixSignal(projectId, sixSignalKey, properties, clientIP, isFactorsAPIKey)

	if err != nil {
		logCtx.WithFields(log.Fields{"error": err, "apiKey": sixSignalKey}).Info("enrich --factors debug")
		statusChannel <- 0
	}
	statusChannel <- 1
}
func enrichUsingSixSignal(projectId int64, sixSignalKey string, properties *util.PropertiesMap, clientIP string, isFactorsAPIKey bool) error {

	logCtx := log.WithField("project_id", projectId)

	if clientIP == "" {
		return errors.New("invalid IP, failed adding user properties")
	}

	url := "https://epsilon.6sense.com/v1/company/details"
	method := "GET"
	client := &http.Client{
		Timeout: util.TimeoutOneSecond + 100*time.Millisecond,
	}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		logCtx.WithFields(log.Fields{"error": err, "apiKey": sixSignalKey}).Info("creating new request --factors debug")
		return err
	}
	sixSignalKey = "Token " + sixSignalKey
	req.Header.Add("Authorization", sixSignalKey)
	req.Header.Add("X-Forwarded-For", clientIP)

	res, err := client.Do(req)
	if err != nil {
		logCtx.WithFields(log.Fields{"error": err, "apiKey": sixSignalKey}).Info("client call --factors debug")
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		log.WithFields(log.Fields{"Error": err}).Warn("Cannot unmarshal JSON")
		return err
	}

	FillEnrichmentPropertiesForSixSignal(result, properties, projectId, isFactorsAPIKey)

	// Adding enrichment source
	if isFactorsAPIKey {
		(*properties)[util.ENRICHMENT_SOURCE] = FACTORS_6SIGNAL
	} else {
		(*properties)[util.ENRICHMENT_SOURCE] = API_6SIGNAL
	}

	return nil
}

func FillEnrichmentPropertiesForSixSignal(result Response, properties *util.PropertiesMap, projectId int64, isFactorsAPIKey bool) {

	util.ValidateAndFillEnrichmentPropsForStringValue(result.Company.Zip, util.SIX_SIGNAL_ZIP, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(result.Company.NaicsDescription, util.SIX_SIGNAL_NAICS_DESCRIPTION, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(result.Company.Country, util.SIX_SIGNAL_COUNTRY, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(result.Company.City, util.SIX_SIGNAL_CITY, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(result.Company.Industry, util.SIX_SIGNAL_INDUSTRY, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(result.Company.Sic, util.SIX_SIGNAL_SIC, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(result.Company.RevenueRange, util.SIX_SIGNAL_REVENUE_RANGE, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(result.Company.CountryIsoCode, util.SIX_SIGNAL_COUNTRY_ISO_CODE, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(result.Company.Phone, util.SIX_SIGNAL_PHONE, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(result.Company.State, util.SIX_SIGNAL_STATE, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(result.Company.Region, util.SIX_SIGNAL_REGION, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(result.Company.Naics, util.SIX_SIGNAL_NAICS, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(result.Company.SicDescription, util.SIX_SIGNAL_SIC_DESCRIPTION, properties)

	empCountInt, err := strconv.Atoi(result.Company.EmployeeCount)
	if err != nil {
		util.ValidateAndFillEnrichmentPropsForStringValue(result.Company.EmployeeCount, util.SIX_SIGNAL_EMPLOYEE_COUNT, properties)
	}
	util.ValidateAndFillEnrichmentPropsForIntegerValue(empCountInt, util.SIX_SIGNAL_EMPLOYEE_COUNT, properties)

	annualRevInt, err := strconv.Atoi(result.Company.AnnualRevenue)
	if err != nil {
		util.ValidateAndFillEnrichmentPropsForStringValue(result.Company.AnnualRevenue, util.SIX_SIGNAL_ANNUAL_REVENUE, properties)
	}
	util.ValidateAndFillEnrichmentPropsForIntegerValue(annualRevInt, util.SIX_SIGNAL_ANNUAL_REVENUE, properties)

	empRange := result.Company.EmployeeRange
	util.ValidateAndFillEnrichmentPropsForStringValue(empRange, util.SIX_SIGNAL_EMPLOYEE_RANGE, properties)

	//Keeping name,address and domain empty, if empRange is equals to "0 - 9"
	if empRange == "0 - 9" {
		(*properties)[util.SIX_SIGNAL_ADDRESS] = ""
		(*properties)[util.SIX_SIGNAL_DOMAIN] = ""
		(*properties)[util.SIX_SIGNAL_NAME] = ""
	} else {

		if address := result.Company.Address; address != "" {
			if c, ok := (*properties)[util.SIX_SIGNAL_ADDRESS]; !ok || c == "" {
				(*properties)[util.SIX_SIGNAL_ADDRESS] = address
			}
		}

		if domain := result.Company.Domain; domain != "" {
			if c, ok := (*properties)[util.SIX_SIGNAL_DOMAIN]; !ok || c == "" {

				(*properties)[util.SIX_SIGNAL_DOMAIN] = domain
				model.SetSixSignalAPICountCacheResult(projectId, util.TimeZoneStringIST)

				if isFactorsAPIKey {
					timeZone, statusCode := store.GetStore().GetTimezoneForProject(projectId)
					if statusCode != http.StatusFound {
						timeZone = util.TimeZoneStringIST
					}
					err := model.SetSixSignalMonthlyUniqueEnrichmentCount(projectId, domain, timeZone)
					if err != nil {
						log.Error("SetSixSignalMonthlyUniqueEnrichmentCount Failed.")
					}
				}

			}
		}

		if name := result.Company.Name; name != "" {
			if c, ok := (*properties)[util.SIX_SIGNAL_NAME]; !ok || c == "" {
				(*properties)[util.SIX_SIGNAL_NAME] = name
			}
		}
	}
}
