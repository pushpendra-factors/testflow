package clear_bit

import (
	"factors/util"
	"fmt"
	"time"

	"github.com/clearbit/clearbit-go/clearbit"
)

func ExecuteClearBitEnrich(clearbitKey string, properties *util.PropertiesMap, clientIP string, statusChannel chan int) {
	defer close(statusChannel)
	err := EnrichmentUsingclearBit(clearbitKey, properties, clientIP)

	if err != nil {
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

	return nil
}
