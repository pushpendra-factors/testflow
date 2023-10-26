package clear_bit

import (
	"encoding/json"
	"factors/config"
	"factors/util"
	"fmt"
	"net/http"
	"time"

	"github.com/clearbit/clearbit-go/clearbit"
	log "github.com/sirupsen/logrus"
)

const API_CLEARBIT = "clearbit_api"

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

func ExecuteClearBitEnrich(projectId int64, clearbitKey string, properties *util.PropertiesMap, clientIP string, statusChannel chan int, logCtx *log.Entry) {
	defer close(statusChannel)

	err := EnrichmentUsingclearBit(projectId, clearbitKey, properties, clientIP)

	if err != nil {
		logCtx.WithFields(log.Fields{"error": err, "apiKey": clearbitKey}).Warn("clearbit enrichment debug")
		statusChannel <- 0
	}
	statusChannel <- 1
}
func EnrichmentUsingclearBit(projectId int64, clearbitKey string, properties *util.PropertiesMap, clientIP string) error {
	if clientIP == "" {
		return fmt.Errorf("invalid IP, failed adding user properties")
	}

	if config.IsCompanyPropsV1Enabled(projectId) {
		baseUrl := "https://reveal.clearbit.com/v1/companies/find?ip="
		method := "GET"
		url := baseUrl + clientIP
		client := &http.Client{
			Timeout: util.TimeoutOneSecond + 100*time.Millisecond,
		}
		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			return err
		}
		clearbitKey = "Bearer " + clearbitKey
		req.Header.Add("Authorization", clearbitKey)
		res, err := client.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		var results Reveal
		err = json.NewDecoder(res.Body).Decode(&results)
		if err != nil {
			return err
		}
		FillEnrichmentPropertiesForClearbitV1(results, properties)

	} else {
		client := clearbit.NewClient(clearbit.WithAPIKey(clearbitKey), clearbit.WithTimeout(util.TimeoutOneSecond+100*time.Millisecond))
		results, _, err := client.Reveal.Find(clearbit.RevealFindParams{
			IP: clientIP,
		})
		if err != nil {
			return err
		}
		FillEnrichmentPropertiesForClearbit(results, properties)
	}

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

func FillEnrichmentPropertiesForClearbitV1(results Reveal, properties *util.PropertiesMap) {

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
	util.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Description, util.ENRICHED_COMPANY_DESCRIPTION, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(results.Company.LinkedIn.Handle, util.ENRICHED_COMPANY_LINKEDIN_URL, properties)
	util.ValidateAndFillEnrichmentPropsForStringValue(results.Company.Metrics.TrafficRank, util.ENRICHED_COMPANY_TRAFFIC_RANK, properties)

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
