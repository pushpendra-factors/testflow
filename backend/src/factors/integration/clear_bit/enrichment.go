package clear_bit

import (
	"factors/util"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/clearbit/clearbit-go/clearbit"
)

const API_CLEARBIT = "clearbit_api"

func ExecuteClearBitEnrich(clearbitKey string, properties *util.PropertiesMap, clientIP string, statusChannel chan int, logCtx *log.Entry) {
	defer close(statusChannel)

	err := EnrichmentUsingclearBit(clearbitKey, properties, clientIP)

	if err != nil {
		logCtx.WithFields(log.Fields{"error": err, "apiKey": clearbitKey}).Warn("clearbit enrichment debug")
		statusChannel <- 0
	}
	statusChannel <- 1
}
func EnrichmentUsingclearBit(clearbitKey string, properties *util.PropertiesMap, clientIP string) error {
	if clientIP == "" {
		return fmt.Errorf("invalid IP, failed adding user properties")
	}

	client := clearbit.NewClient(clearbit.WithAPIKey(clearbitKey), clearbit.WithTimeout(util.TimeoutOneSecond+100*time.Millisecond))
	results, _, err := client.Reveal.Find(clearbit.RevealFindParams{
		IP: clientIP,
	})
	if err != nil {
		return err
	}

	FillEnrichmentPropertiesForClearbit(results, properties)
	(*properties)[util.ENRICHMENT_SOURCE] = API_CLEARBIT

	return nil
}

func FillEnrichmentPropertiesForClearbit(results *clearbit.Reveal, properties *util.PropertiesMap) {

	util.ValidateAndFillEnrichmentPropsForStringValue(results.Domain, util.SIX_SIGNAL_DOMAIN, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Name, util.SIX_SIGNAL_NAME, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Category.Industry, util.SIX_SIGNAL_INDUSTRY, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Geo.City, util.SIX_SIGNAL_CITY, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Geo.PostalCode, util.SIX_SIGNAL_ZIP, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Geo.Country, util.SIX_SIGNAL_COUNTRY, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Metrics.EmployeesRange, util.SIX_SIGNAL_EMPLOYEE_RANGE, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Metrics.EstimatedAnnualRevenue, util.SIX_SIGNAL_REVENUE_RANGE, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Category.NaicsCode, util.SIX_SIGNAL_NAICS, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Category.SicCode, util.SIX_SIGNAL_SIC, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Geo.CountryCode, util.SIX_SIGNAL_COUNTRY_ISO_CODE, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Geo.State, util.SIX_SIGNAL_STATE, properties)
	util.ValidateAndFillEnrichmentPropsForIntegerValue(results.Company.Metrics.AnnualRevenue, util.SIX_SIGNAL_ANNUAL_REVENUE, properties)
	util.ValidateAndFillEnrichmentPropsForIntegerValue(results.Company.Metrics.Employees, util.SIX_SIGNAL_EMPLOYEE_COUNT, properties)

	util.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Type, util.ENRICHED_COMPANY_TYPE, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(results.Company.ID, util.ENRICHED_COMPANY_ID, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(results.Company.LegalName, util.ENRICHED_COMPANY_LEGAL_NAME, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Category.Sector, util.ENRICHED_COMPANY_SECTOR, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Category.IndustryGroup, util.ENRICHED_COMPANY_INDUSTRY_GROUP, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Category.SubIndustry, util.ENRICHED_COMPANY_SUB_INDUSTRY, properties)
	util.ValidateAndFillEnrichmentPropsForIntegerValue(results.Company.Metrics.Raised, util.ENRICHED_COMPANY_FUNDING_RAISED, properties)
	util.ValidateAndFillEnrichmentPropsForIntegerValue(results.Company.Metrics.AlexaUsRank, util.ENRICHED_COMPANY_ALEXA_US_RANK, properties)
	util.ValidateAndFillEnrichmentPropsForIntegerValue(results.Company.Metrics.AlexaGlobalRank, util.ENRICHED_COMPANY_ALEXA_GLOBAL_RANK, properties)
	util.ValidateAndFillEnrichmentPropsForIntegerValue(results.Company.FoundedYear, util.ENRICHED_COMPANY_FOUNDED_YEAR, properties)
	util.ValidateAndFillEnrichmentPropsForIntegerValue(results.Company.Metrics.MarketCap, util.ENRICHED_COMPANY_MARKET_CAP, properties)

	//company object: tech
	if techs := results.Company.Tech; len(techs) > 0 {
		if c, ok := (*properties)[util.ENRICHED_COMPANY_TECH]; !ok || c == "" {
			(*properties)[util.ENRICHED_COMPANY_TECH] = techs
		}
	}

	// company object: tags
	if tags := results.Company.Tags; len(tags) > 0 {
		if c, ok := (*properties)[util.ENRICHED_COMPANY_TAGS]; !ok || c == "" {
			(*properties)[util.ENRICHED_COMPANY_TAGS] = tags
		}
	}
}
