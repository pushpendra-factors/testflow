package clear_bit

import (
	"factors/config"
	"factors/util"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/clearbit/clearbit-go/clearbit"
)

const API_CLEARBIT = "clearbit_api"

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

	client := clearbit.NewClient(clearbit.WithAPIKey(clearbitKey), clearbit.WithTimeout(util.TimeoutOneSecond+100*time.Millisecond))
	results, _, err := client.Reveal.Find(clearbit.RevealFindParams{
		IP: clientIP,
	})
	if err != nil {
		return err
	}

	if config.IsCompanyPropsV1Enabled(projectId) {
		FillEnrichmentPropertiesForClearbit(results, properties)
	} else {
		log.Error("Clearbit enrichment via old flow for project id: ", projectId)
		if ip := results.IP; ip != "" {
			if c, ok := (*properties)[util.CLR_IP]; !ok || c == "" {
				(*properties)[util.CLR_IP] = ip
			}
		}

		if domain := results.Domain; domain != "" {
			if c, ok := (*properties)[util.CLR_COMPANY_PARENT_DOMAIN]; !ok || c == "" {
				(*properties)[util.CLR_COMPANY_PARENT_DOMAIN] = domain
			}
		}

		if types := results.Company.Type; types != "" {
			if c, ok := (*properties)[util.CLR_COMPANY_TYPE]; !ok || c == "" {
				(*properties)[util.CLR_COMPANY_TYPE] = types
			}
		}

		if id := results.Company.ID; id != "" {
			if c, ok := (*properties)[util.CLR_COMPANY_ID]; !ok || c == "" {
				(*properties)[util.CLR_COMPANY_ID] = id
			}
		}

		if name := results.Company.Name; name != "" {
			if c, ok := (*properties)[util.CLR_COMPANY_NAME]; !ok || c == "" {
				(*properties)[util.CLR_COMPANY_NAME] = name
			}
		}

		if legalName := results.Company.LegalName; legalName != "" {
			if c, ok := (*properties)[util.CLR_COMPANY_LEGALNAME]; !ok || c == "" {
				(*properties)[util.CLR_COMPANY_LEGALNAME] = legalName
			}
		}

		if sector := results.Company.Category.Sector; sector != "" {
			if c, ok := (*properties)[util.CLR_COMPANY_CATEGORY_SECTOR]; !ok || c == "" {
				(*properties)[util.CLR_COMPANY_CATEGORY_SECTOR] = sector
			}
		}

		if industryGroup := results.Company.Category.IndustryGroup; industryGroup != "" {
			if c, ok := (*properties)[util.CLR_COMPANY_CATEGORY_INDUSTRYGROUP]; !ok || c == "" {
				(*properties)[util.CLR_COMPANY_CATEGORY_INDUSTRYGROUP] = industryGroup
			}
		}

		if industry := results.Company.Category.Industry; industry != "" {
			if c, ok := (*properties)[util.CLR_COMPANY_CATEGORY_INDUSTRY]; !ok || c == "" {
				(*properties)[util.CLR_COMPANY_CATEGORY_INDUSTRY] = industry
			}
		}

		if subIndustry := results.Company.Category.SubIndustry; subIndustry != "" {
			if c, ok := (*properties)[util.CLR_COMPANY_CATEGORY_SUBINDUSTRY]; !ok || c == "" {
				(*properties)[util.CLR_COMPANY_CATEGORY_SUBINDUSTRY] = subIndustry
			}
		}

		if City := results.Company.Geo.City; City != "" {
			if c, ok := (*properties)[util.CLR_COMPANY_GEO_CITY]; !ok || c == "" {
				(*properties)[util.CLR_COMPANY_GEO_CITY] = City
			}
		}
		if postalCode := results.Company.Geo.PostalCode; postalCode != "" {
			if c, ok := (*properties)[util.CLR_COMPANY_GEO_POSTALCODE]; !ok || c == "" {
				(*properties)[util.CLR_COMPANY_GEO_POSTALCODE] = postalCode
			}
		}
		if Country := results.Company.Geo.Country; Country != "" {
			if c, ok := (*properties)[util.CLR_COMPANY_GEO_COUNTRY]; !ok || c == "" {
				(*properties)[util.CLR_COMPANY_GEO_COUNTRY] = Country
			}
		}
		if raised := results.Company.Metrics.Raised; raised > 0 {
			if c, ok := (*properties)[util.CLR_COMPANY_METRICS_RAISED]; !ok || c == "" {
				(*properties)[util.CLR_COMPANY_METRICS_RAISED] = raised
			}
		}
		// company object: metrics.alexaGlobalRank
		if alexaGlobalRank := results.Company.Metrics.AlexaUsRank; alexaGlobalRank > 0 {
			if c, ok := (*properties)[util.CLR_COMPANY_METRICS_ALEXAUSRANK]; !ok || c == "" {
				(*properties)[util.CLR_COMPANY_METRICS_ALEXAUSRANK] = alexaGlobalRank
			}
		}
		// company object: metrics.employeesRange
		if employeesRange := results.Company.Metrics.EmployeesRange; employeesRange != "" {
			if c, ok := (*properties)[util.CLR_COMPANY_METRICS_EMPLOYEESRANGE]; !ok || c == "" {
				(*properties)[util.CLR_COMPANY_METRICS_EMPLOYEESRANGE] = employeesRange
			}
		}
		// company object: metrics.annualRevenue
		if annualRevenue := results.Company.Metrics.AnnualRevenue; annualRevenue > 0 {
			if c, ok := (*properties)[util.CLR_COMPANY_METRICS_ANNUAL_REVENUE]; !ok || c == "" {
				(*properties)[util.CLR_COMPANY_METRICS_ANNUAL_REVENUE] = annualRevenue
			}
		}
		// company object: metrics.estimatedAnnualRevenue
		if estimatedAnnualRevenue := results.Company.Metrics.EstimatedAnnualRevenue; estimatedAnnualRevenue != "" {
			if c, ok := (*properties)[util.CLR_COMPANY_METRICS_ESTIMATED_ANNUAL_REVENUE]; !ok || c == "" {
				(*properties)[util.CLR_COMPANY_METRICS_ESTIMATED_ANNUAL_REVENUE] = estimatedAnnualRevenue
			}
		}
		//company object: tech
		if Tech := results.Company.Tech; len(Tech) > 0 {
			if c, ok := (*properties)[util.CLR_COMPANY_TECH]; !ok || c == "" {
				(*properties)[util.CLR_COMPANY_TECH] = Tech
			}
		}

		// company object: tags
		if tags := results.Company.Tags; len(tags) > 0 {
			if c, ok := (*properties)[util.CLR_COMPANY_TAGS]; !ok || c == "" {
				(*properties)[util.CLR_COMPANY_TAGS] = tags
			}
		}

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
