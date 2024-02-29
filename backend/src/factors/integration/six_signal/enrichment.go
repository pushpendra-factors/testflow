package six_signal

import (
	"encoding/json"
	"errors"
	U "factors/util"
	"strconv"

	log "github.com/sirupsen/logrus"
)

const FACTORS_6SIGNAL = "factors_6sense"
const API_6SIGNAL = "API_6Sense"

type ResultChannel struct {
	ExecuteStatus int
	Domain        string
}
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

func ExecuteSixSignalEnrichV1(projectId int64, sixSignalAPIKey string, properties *U.PropertiesMap, clientIP string, resultChannel chan ResultChannel) {
	defer close(resultChannel)
	logCtx := log.WithField("project_id", projectId)

	domain, err := enrichUsingSixSignal(projectId, sixSignalAPIKey, properties, clientIP, false)
	if err != nil {
		logCtx.WithFields(log.Fields{"error": err, "apiKey": sixSignalAPIKey}).Info("enrich --factors debug")
		resultChannel <- ResultChannel{ExecuteStatus: 0, Domain: ""}
	}
	resultChannel <- ResultChannel{ExecuteStatus: 1, Domain: domain}
}

func enrichUsingSixSignal(projectId int64, sixSignalAPIKey string, properties *U.PropertiesMap, clientIP string, isFactorsAPIKey bool) (string, error) {

	logCtx := log.WithField("project_id", projectId)

	if clientIP == "" {
		return "", errors.New("invalid IP, failed adding user properties")
	}

	res, err := SixSignalHTTPRequest(clientIP, sixSignalAPIKey)
	if err != nil {
		logCtx.WithFields(log.Fields{"error": err, "apiKey": sixSignalAPIKey}).Info("client call --factors debug")
		return "", err
	}
	defer res.Body.Close()

	var result Response
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	FillEnrichmentPropertiesForSixSignal(result, properties, projectId, isFactorsAPIKey)

	return result.Company.Domain, nil
}

func FillEnrichmentPropertiesForSixSignal(result Response, properties *U.PropertiesMap, projectId int64, isFactorsAPIKey bool) {

	U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.Zip, U.SIX_SIGNAL_ZIP, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.NaicsDescription, U.SIX_SIGNAL_NAICS_DESCRIPTION, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.Country, U.SIX_SIGNAL_COUNTRY, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.City, U.SIX_SIGNAL_CITY, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.Industry, U.SIX_SIGNAL_INDUSTRY, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.Sic, U.SIX_SIGNAL_SIC, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.RevenueRange, U.SIX_SIGNAL_REVENUE_RANGE, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.CountryIsoCode, U.SIX_SIGNAL_COUNTRY_ISO_CODE, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.Phone, U.SIX_SIGNAL_PHONE, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.State, U.SIX_SIGNAL_STATE, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.Region, U.SIX_SIGNAL_REGION, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.Naics, U.SIX_SIGNAL_NAICS, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.SicDescription, U.SIX_SIGNAL_SIC_DESCRIPTION, properties)

	empCountInt, err := strconv.Atoi(result.Company.EmployeeCount)
	if err != nil {
		U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.EmployeeCount, U.SIX_SIGNAL_EMPLOYEE_COUNT, properties)
	}
	U.ValidateAndFillEnrichmentPropsForIntegerValue(empCountInt, U.SIX_SIGNAL_EMPLOYEE_COUNT, properties)

	annualRevInt, err := strconv.Atoi(result.Company.AnnualRevenue)
	if err != nil {
		U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.AnnualRevenue, U.SIX_SIGNAL_ANNUAL_REVENUE, properties)
	}
	U.ValidateAndFillEnrichmentPropsForIntegerValue(annualRevInt, U.SIX_SIGNAL_ANNUAL_REVENUE, properties)

	empRange := result.Company.EmployeeRange
	U.ValidateAndFillEnrichmentPropsForStringValue(empRange, U.SIX_SIGNAL_EMPLOYEE_RANGE, properties)

	//Keeping name,address and domain empty, if empRange is equals to "0 - 9"
	if empRange == "0 - 9" {
		(*properties)[U.SIX_SIGNAL_ADDRESS] = ""
		(*properties)[U.SIX_SIGNAL_DOMAIN] = ""
		(*properties)[U.SIX_SIGNAL_NAME] = ""
	} else {

		if address := result.Company.Address; address != "" {
			if c, ok := (*properties)[U.SIX_SIGNAL_ADDRESS]; !ok || c == "" {
				(*properties)[U.SIX_SIGNAL_ADDRESS] = address
			}
		}

		if domain := result.Company.Domain; domain != "" {
			if c, ok := (*properties)[U.SIX_SIGNAL_DOMAIN]; !ok || c == "" {
				(*properties)[U.SIX_SIGNAL_DOMAIN] = domain
			}
		}

		if name := result.Company.Name; name != "" {
			if c, ok := (*properties)[U.SIX_SIGNAL_NAME]; !ok || c == "" {
				(*properties)[U.SIX_SIGNAL_NAME] = name
			}
		}
	}
}
