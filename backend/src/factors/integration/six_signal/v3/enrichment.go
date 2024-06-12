package v3

import (
	"encoding/json"
	"factors/integration/six_signal"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
)

const SIX_SIGNAL_IDENTIFICATION_URL_V3 = "https://epsilon.6sense.com/v3/company/details"

type Response struct {
	Company struct {
		CompanyMatch      string       `json:"company_match"`
		AdditionalComment string       `json:"additional_comment"`
		Domain            string       `json:"domain"`
		Name              string       `json:"name"`
		Region            string       `json:"region"`
		Country           string       `json:"country"`
		State             string       `json:"state"`
		City              string       `json:"city"`
		Industry          string       `json:"industry"`
		CountryISOCode    string       `json:"country_iso_code"`
		Address           string       `json:"address"`
		Zip               string       `json:"zip"`
		Phone             string       `json:"phone"`
		EmployeeRange     string       `json:"employee_range"`
		RevenueRange      string       `json:"revenue_range"`
		EmployeeCount     string       `json:"employee_count"`
		AnnualRevenue     string       `json:"annual_revenue"`
		IsBlacklisted     bool         `json:"is_blacklisted"`
		Is6QA             bool         `json:"is_6qa"`
		GeoIPCountry      string       `json:"geoIP_country"`
		GeoIPState        string       `json:"geoIP_state"`
		GeoIPCity         string       `json:"geoIP_city"`
		StateCode         string       `json:"state_code"`
		IndustryV2        []IndustryV2 `json:"industry_v2"`
		SICDescription    string       `json:"sic_description"`
		SIC               string       `json:"sic"`
		NAICS             string       `json:"naics"`
		NAICSDescription  string       `json:"naics_description"`
	} `json:"company"`
}

type IndustryV2 struct {
	Industry    string `json:"industry"`
	Subindustry string `json:"subindustry"`
}

func ExecuteSixSignalEnrichV3(projectId int64, sixSignalAPIKey string, properties *U.PropertiesMap, clientIP string, resultChannel chan six_signal.ResultChannel, logCtx *log.Entry) {
	defer close(resultChannel)
	domain, err := enrichUsingSixSignalV3(projectId, sixSignalAPIKey, properties, clientIP, logCtx)
	if err != nil {
		logCtx.WithField("error", err).Warn("Failed to enrich using sixsignal v3.")
		resultChannel <- six_signal.ResultChannel{ExecuteStatus: 0, Domain: ""}
	}
	resultChannel <- six_signal.ResultChannel{ExecuteStatus: 1, Domain: domain}

}

func enrichUsingSixSignalV3(projectId int64, sixSignalAPIKey string, properties *U.PropertiesMap, clientIP string, logCtx *log.Entry) (string, error) {

	res, err := SixSignalV3HTTPRequest(SIX_SIGNAL_IDENTIFICATION_URL_V3, "GET", sixSignalAPIKey, clientIP)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusBadRequest {
		logCtx.Error("Bad Request from sixsignal v3 API")
		return "", fmt.Errorf("status bad request from sixsignal v3 api")
	} else if res.StatusCode == http.StatusPaymentRequired {
		logCtx.Error(("6Sense Quota Exhausted."))
		return "", fmt.Errorf("6Sense quota exhausted")
	}
	var result Response
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	FillEnrichmentPropertiesForSixSignalV3(result, properties, projectId)

	return result.Company.Domain, nil
}

func FillEnrichmentPropertiesForSixSignalV3(result Response, properties *U.PropertiesMap, projectId int64) {

	U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.Zip, U.SIX_SIGNAL_ZIP, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.NAICSDescription, U.SIX_SIGNAL_NAICS_DESCRIPTION, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.Country, U.SIX_SIGNAL_COUNTRY, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.City, U.SIX_SIGNAL_CITY, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.Industry, U.SIX_SIGNAL_INDUSTRY, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.SIC, U.SIX_SIGNAL_SIC, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.RevenueRange, U.SIX_SIGNAL_REVENUE_RANGE, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.CountryISOCode, U.SIX_SIGNAL_COUNTRY_ISO_CODE, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.Phone, U.SIX_SIGNAL_PHONE, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.State, U.SIX_SIGNAL_STATE, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.Region, U.SIX_SIGNAL_REGION, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.NAICS, U.SIX_SIGNAL_NAICS, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.SICDescription, U.SIX_SIGNAL_SIC_DESCRIPTION, properties)

	empCountInt, err := strconv.Atoi(result.Company.EmployeeCount)
	if err != nil {
		U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.EmployeeCount, U.SIX_SIGNAL_EMPLOYEE_COUNT, properties)
	} else {
		U.ValidateAndFillEnrichmentPropsForIntegerValue(empCountInt, U.SIX_SIGNAL_EMPLOYEE_COUNT, properties)
	}

	annualRevInt, err := strconv.Atoi(result.Company.AnnualRevenue)
	if err != nil {
		U.ValidateAndFillEnrichmentPropsForStringValue(result.Company.AnnualRevenue, U.SIX_SIGNAL_ANNUAL_REVENUE, properties)
	} else {
		U.ValidateAndFillEnrichmentPropsForIntegerValue(annualRevInt, U.SIX_SIGNAL_ANNUAL_REVENUE, properties)
	}

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
