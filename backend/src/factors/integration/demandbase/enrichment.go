package demandbase

import (
	"encoding/json"
	U "factors/util"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

const DEMANDBASE_COMPANY_FIRMOGRAPHICS_URL = "https://api.demandbase.com/api/v3/ip.json"
const EMPLOYEE_RANGE_SMALL = "Small"
const EMPLOYEE_RANGE_MID_MARKET = "Mid-Market"
const EMPLOYEE_RANGE_ENTERPRISE = "Enterprise"
const EMPLOYEE_RANGE_SMALL_NUMERICAL_RANGE = "1-99"
const EMPLOYEE_RANGE_MID_MARKET_NUMERICAL_RANGE = "100-999"
const EMPLOYEE_RANGE_ENTERPRISE_NUMERICAL_RANGE = "1000+"

type ResultChannel struct {
	ExecuteStatus int
	Domain        string
}

type Response struct {
	ID     int    `json:"company_id"`
	Type   string `json:"company_type"`
	Domain string `json:"web_site"`
	Name   string `json:"company_name"`

	Industry    string `json:"industry"`
	SubIndustry string `json:"sub_industry"`

	City    string `json:"city"`
	ZipCode string `json:"zip"`
	Country string `json:"country_name"`
	ISOCode string `json:"country"`
	State   string `json:"state"`
	Address string `json:"street_address"`
	Region  string `json:"region_name"`

	RevenueRange  string `json:"revenue_range"`
	Revenue       int    `json:"annual_sales"`
	EmployeeCount int    `json:"employee_count"`
	EmployeeRange string `json:"employee_range"`

	NAICS    string `json:"primary_naics"`
	SIC      string `json:"primary_sic"`
	Phone    string `json:"phone"`
	Linkedin string `json:"company_linkedin_profile"`
}

func ExecuteDemandbaseEnrich(projectId int64, demandbaseAPIKey string, properties *U.PropertiesMap, clientIP string, resultChannel chan ResultChannel, logCtx *log.Entry) {
	defer close(resultChannel)

	domain, err := enrichUsingDemandbase(demandbaseAPIKey, properties, clientIP, logCtx)
	if err != nil {
		logCtx.WithField("error", err).Warn("Failed to enrich using demandbase.")
		resultChannel <- ResultChannel{ExecuteStatus: 0, Domain: ""}
	}
	resultChannel <- ResultChannel{ExecuteStatus: 1, Domain: domain}
}

func enrichUsingDemandbase(demandbaseAPIKey string, properties *U.PropertiesMap, clientIP string, logCtx *log.Entry) (string, error) {

	res, err := DemandbaseHTTPRequest(DEMANDBASE_COMPANY_FIRMOGRAPHICS_URL, "GET", demandbaseAPIKey, clientIP)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		logCtx.WithField("err", res.StatusCode).Warn("Failed to enrich using demandbase")
		return "", fmt.Errorf("failed to enrich using demandbase")
	}

	var result Response
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	FillEnrichmentPropertiesForDemandbase(result, properties)

	return "", nil
}

func FillEnrichmentPropertiesForDemandbase(result Response, properties *U.PropertiesMap) {

	U.ValidateAndFillEnrichmentPropsForIntegerValue(result.ID, U.ENRICHED_COMPANY_ID, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Type, U.ENRICHED_COMPANY_TYPE, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Domain, U.SIX_SIGNAL_DOMAIN, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Name, U.SIX_SIGNAL_NAME, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Industry, U.SIX_SIGNAL_INDUSTRY, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.SubIndustry, U.ENRICHED_COMPANY_SUB_INDUSTRY, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.City, U.SIX_SIGNAL_CITY, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Country, U.SIX_SIGNAL_COUNTRY, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.ISOCode, U.SIX_SIGNAL_COUNTRY_ISO_CODE, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.State, U.SIX_SIGNAL_STATE, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Region, U.SIX_SIGNAL_REGION, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Address, U.SIX_SIGNAL_ADDRESS, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.ZipCode, U.SIX_SIGNAL_ZIP, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.RevenueRange, U.SIX_SIGNAL_REVENUE_RANGE, properties)
	U.ValidateAndFillEnrichmentPropsForIntegerValue(result.Revenue, U.SIX_SIGNAL_ANNUAL_REVENUE, properties)
	U.ValidateAndFillEnrichmentPropsForIntegerValue(result.EmployeeCount, U.SIX_SIGNAL_EMPLOYEE_COUNT, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.NAICS, U.SIX_SIGNAL_NAICS, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.SIC, U.SIX_SIGNAL_SIC, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Phone, U.SIX_SIGNAL_PHONE, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(result.Linkedin, U.ENRICHED_COMPANY_LINKEDIN_URL, properties)

	empRange := TransformDemandbaseEmployeeRangeIntoNumericalRange(result.EmployeeRange)
	U.ValidateAndFillEnrichmentPropsForStringValue(empRange, U.SIX_SIGNAL_EMPLOYEE_RANGE, properties)

}

func TransformDemandbaseEmployeeRangeIntoNumericalRange(responseValue string) string {

	if responseValue == EMPLOYEE_RANGE_SMALL {
		return EMPLOYEE_RANGE_SMALL_NUMERICAL_RANGE
	} else if responseValue == EMPLOYEE_RANGE_MID_MARKET {
		return EMPLOYEE_RANGE_MID_MARKET_NUMERICAL_RANGE
	} else if responseValue == EMPLOYEE_RANGE_ENTERPRISE {
		return EMPLOYEE_RANGE_ENTERPRISE_NUMERICAL_RANGE
	}

	return ""
}
