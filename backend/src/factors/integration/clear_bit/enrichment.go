package clear_bit

import (
	"encoding/json"
	"factors/config"
	U "factors/util"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

const API_CLEARBIT = "clearbit_api"

type ResultChannel struct {
	ExecuteStatus int
	Domain        string
}

// Introduction of this struct is necessitated by the transition away from the deprecated
// vendor package, which was no longer actively maintained.
type Reveal struct {
	IP      string `json:"ip"`
	Fuzzy   bool   `json:"fuzzy"`
	Domain  string `json:"domain"`
	Company struct {
		ID            string   `json:"id"`
		Name          string   `json:"name"`
		LegalName     string   `json:"legalName"`
		Domain        string   `json:"domain"`
		DomainAliases []string `json:"domainAliases"`
		Site          struct {
			PhoneNumbers   []string `json:"phoneNumbers"`
			EmailAddresses []string `json:"emailAddresses"`
		} `json:"site"`
		Category struct {
			Sector        string `json:"sector"`
			IndustryGroup string `json:"industryGroup"`
			Industry      string `json:"industry"`
			SubIndustry   string `json:"subIndustry"`
			SicCode       string `json:"sicCode"`
			NaicsCode     string `json:"naicsCode"`
		} `json:"category"`
		Tags        []string `json:"tags"`
		Description string   `json:"description"`
		FoundedYear int      `json:"foundedYear"`
		Location    string   `json:"location"`
		TimeZone    string   `json:"timeZone"`
		UtcOffset   int      `json:"utcOffset"`
		Geo         struct {
			StreetNumber string  `json:"streetNumber"`
			StreetName   string  `json:"streetName"`
			SubPremise   string  `json:"subPremise"`
			City         string  `json:"city"`
			PostalCode   string  `json:"postalCode"`
			State        string  `json:"state"`
			StateCode    string  `json:"stateCode"`
			Country      string  `json:"country"`
			CountryCode  string  `json:"countryCode"`
			Lat          float64 `json:"lat"`
			Lng          float64 `json:"lng"`
		} `json:"geo"`
		Logo     string `json:"logo"`
		Facebook struct {
			Handle string `json:"handle"`
			Likes  int    `json:"likes"`
		} `json:"facebook"`
		LinkedIn struct {
			Handle string `json:"handle"`
		} `json:"linkedin"`
		Twitter struct {
			Handle    string `json:"handle"`
			ID        string `json:"id"`
			Bio       string `json:"bio"`
			Followers int    `json:"followers"`
			Following int    `json:"following"`
			Location  string `json:"location"`
			Site      string `json:"site"`
			Avatar    string `json:"avatar"`
		} `json:"twitter"`
		Crunchbase struct {
			Handle string `json:"handle"`
		} `json:"crunchbase"`
		EmailProvider bool   `json:"emailProvider"`
		Type          string `json:"type"`
		Ticker        string `json:"ticker"`
		Identifiers   struct {
			UsEIN string `json:"usEIN"`
		} `json:"identifiers"`
		Phone   string `json:"phone"`
		Metrics struct {
			AlexaUsRank            int    `json:"alexaUsRank"`
			AlexaGlobalRank        int    `json:"alexaGlobalRank"`
			Employees              int    `json:"employees"`
			EmployeesRange         string `json:"employeesRange"`
			MarketCap              int    `json:"marketCap"`
			Raised                 int    `json:"raised"`
			AnnualRevenue          int    `json:"annualRevenue"`
			EstimatedAnnualRevenue string `json:"estimatedAnnualRevenue"`
			FiscalYearEnd          int    `json:"fiscalYearEnd"`
			TrafficRank            string `json:"trafficRank"`
		} `json:"metrics"`
		IndexedAt time.Time `json:"indexedAt"`
		Tech      []string  `json:"tech"`
		Parent    struct {
			Domain string `json:"domain"`
		} `json:"parent"`
	} `json:"company"`
	ConfidenceScore int `json:"confidence_score,omitempty"`
}

func ExecuteClearBitEnrichV1(projectId int64, clearbitKey string, properties *U.PropertiesMap, clientIP string, resultChannel chan ResultChannel, logCtx *log.Entry) {
	defer close(resultChannel)

	domain, err := EnrichmentUsingclearBit(projectId, clearbitKey, properties, clientIP)

	if err != nil {
		logCtx.WithFields(log.Fields{"error": err, "apiKey": clearbitKey}).Warn("clearbit enrichment debug")
		resultChannel <- ResultChannel{ExecuteStatus: 0, Domain: ""}
	}
	resultChannel <- ResultChannel{ExecuteStatus: 1, Domain: domain}
}

func ExecuteClearBitEnrich(projectId int64, clearbitKey string, properties *U.PropertiesMap, clientIP string, statusChannel chan int, logCtx *log.Entry) {
	defer close(statusChannel)

	_, err := EnrichmentUsingclearBit(projectId, clearbitKey, properties, clientIP)

	if err != nil {
		logCtx.WithFields(log.Fields{"error": err, "apiKey": clearbitKey}).Warn("clearbit enrichment debug")
		statusChannel <- 0
	}
	statusChannel <- 1
}

func EnrichmentUsingclearBit(projectId int64, clearbitKey string, properties *U.PropertiesMap, clientIP string) (string, error) {

	if clientIP == "" {
		return "", fmt.Errorf("invalid IP, failed adding user properties")
	}

	res, err := ClearbitHTTPRequest(clientIP, clearbitKey)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var results Reveal
	err = json.NewDecoder(res.Body).Decode(&results)
	if err != nil {
		return "", err
	}
	FillEnrichmentPropertiesForClearbit(results, properties)

	if !config.IsCompanyEnrichmentV1Enabled(projectId) {
		(*properties)[U.ENRICHMENT_SOURCE] = API_CLEARBIT
	}

	return results.Domain, nil
}

func FillEnrichmentPropertiesForClearbit(results Reveal, properties *U.PropertiesMap) {

	U.ValidateAndFillEnrichmentPropsForStringValue(results.Domain, U.SIX_SIGNAL_DOMAIN, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Name, U.SIX_SIGNAL_NAME, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Category.Industry, U.SIX_SIGNAL_INDUSTRY, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Geo.City, U.SIX_SIGNAL_CITY, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Geo.PostalCode, U.SIX_SIGNAL_ZIP, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Geo.Country, U.SIX_SIGNAL_COUNTRY, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Metrics.EmployeesRange, U.SIX_SIGNAL_EMPLOYEE_RANGE, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Metrics.EstimatedAnnualRevenue, U.SIX_SIGNAL_REVENUE_RANGE, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Category.NaicsCode, U.SIX_SIGNAL_NAICS, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Category.SicCode, U.SIX_SIGNAL_SIC, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Geo.CountryCode, U.SIX_SIGNAL_COUNTRY_ISO_CODE, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Geo.State, U.SIX_SIGNAL_STATE, properties)
	U.ValidateAndFillEnrichmentPropsForIntegerValue(results.Company.Metrics.AnnualRevenue, U.SIX_SIGNAL_ANNUAL_REVENUE, properties)
	U.ValidateAndFillEnrichmentPropsForIntegerValue(results.Company.Metrics.Employees, U.SIX_SIGNAL_EMPLOYEE_COUNT, properties)

	U.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Type, U.ENRICHED_COMPANY_TYPE, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(results.Company.ID, U.ENRICHED_COMPANY_ID, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(results.Company.LegalName, U.ENRICHED_COMPANY_LEGAL_NAME, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Category.Sector, U.ENRICHED_COMPANY_SECTOR, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Category.IndustryGroup, U.ENRICHED_COMPANY_INDUSTRY_GROUP, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Category.SubIndustry, U.ENRICHED_COMPANY_SUB_INDUSTRY, properties)
	U.ValidateAndFillEnrichmentPropsForIntegerValue(results.Company.Metrics.Raised, U.ENRICHED_COMPANY_FUNDING_RAISED, properties)
	U.ValidateAndFillEnrichmentPropsForIntegerValue(results.Company.Metrics.AlexaUsRank, U.ENRICHED_COMPANY_ALEXA_US_RANK, properties)
	U.ValidateAndFillEnrichmentPropsForIntegerValue(results.Company.Metrics.AlexaGlobalRank, U.ENRICHED_COMPANY_ALEXA_GLOBAL_RANK, properties)
	U.ValidateAndFillEnrichmentPropsForIntegerValue(results.Company.FoundedYear, U.ENRICHED_COMPANY_FOUNDED_YEAR, properties)
	U.ValidateAndFillEnrichmentPropsForIntegerValue(results.Company.Metrics.MarketCap, U.ENRICHED_COMPANY_MARKET_CAP, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Description, U.ENRICHED_COMPANY_DESCRIPTION, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(results.Company.LinkedIn.Handle, U.ENRICHED_COMPANY_LINKEDIN_URL, properties)
	U.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Metrics.TrafficRank, U.ENRICHED_COMPANY_TRAFFIC_RANK, properties)

	//company object: tech
	if techs := results.Company.Tech; len(techs) > 0 {
		if c, ok := (*properties)[U.ENRICHED_COMPANY_TECH]; !ok || c == "" {
			(*properties)[U.ENRICHED_COMPANY_TECH] = techs
		}
	}

	// company object: tags
	if tags := results.Company.Tags; len(tags) > 0 {
		if c, ok := (*properties)[U.ENRICHED_COMPANY_TAGS]; !ok || c == "" {
			(*properties)[U.ENRICHED_COMPANY_TAGS] = tags
		}
	}
}
