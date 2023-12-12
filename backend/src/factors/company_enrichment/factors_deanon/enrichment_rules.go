package factors_deanon

import (
	"encoding/json"
	"factors/model/model"
	U "factors/util"
	"strings"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

// ApplyFactorsDeanonRules fetches the sixsignal config and checks the country filter on basis of isoCode and pageURL filter.
func ApplyFactorsDeanonRules(factorsDeanonRulesJson *postgres.Jsonb, isoCode, pageURL string) (bool, error) {

	var factorsDeanonRules model.SixSignalConfig
	if factorsDeanonRulesJson != nil {
		err := json.Unmarshal(factorsDeanonRulesJson.RawMessage, &factorsDeanonRules)
		if err != nil {
			log.WithField("six_signal_config", factorsDeanonRulesJson).WithError(err).Error("Failed to decode six signal property")
			return false, err
		}
	}

	//No filter case is true
	if (factorsDeanonRules.CountryInclude == nil || len(factorsDeanonRules.CountryInclude) == 0) &&
		(factorsDeanonRules.CountryExclude == nil || len(factorsDeanonRules.CountryExclude) == 0) &&
		(factorsDeanonRules.PagesInclude == nil || len(factorsDeanonRules.PagesInclude) == 0) &&
		(factorsDeanonRules.PagesExclude == nil || len(factorsDeanonRules.PagesExclude) == 0) {
		return true, nil
	}

	countryFilterPassed := IsCountryRulesPassed(factorsDeanonRules, isoCode)
	if !countryFilterPassed {
		return false, nil
	}

	pageFilterPassed, _ := IsPageUrlRulesPassed(factorsDeanonRules, pageURL)
	if !pageFilterPassed {
		return false, nil
	}

	return true, nil
}

// IsCountryRulesPassed checks whether the country filter is successfully applied or not.
func IsCountryRulesPassed(factorsDeanonRules model.SixSignalConfig, isoCode string) bool {

	isCountryIncluded := (factorsDeanonRules.CountryInclude != nil && len(factorsDeanonRules.CountryInclude) != 0)
	isCountryExcluded := (factorsDeanonRules.CountryExclude != nil && len(factorsDeanonRules.CountryExclude) != 0)

	if !isCountryIncluded && !isCountryExcluded {
		return true
	}

	mapOfCountries := make(map[string]bool)

	if isCountryIncluded {
		for _, filter := range factorsDeanonRules.CountryInclude {
			mapOfCountries[filter.Value] = true
		}

	} else if isCountryExcluded {
		for _, filter := range factorsDeanonRules.CountryExclude {
			mapOfCountries[filter.Value] = true
		}
	}

	contains := mapOfCountries[isoCode]

	if isCountryIncluded {
		return contains
	}
	return !contains
}

// IsPageUrlRulesPassed checks whether the page url filter is successfully applied or not.
func IsPageUrlRulesPassed(factorsDeanonRules model.SixSignalConfig, pageUrl string) (bool, error) {

	pageFilterPassed := true
	isPageUrlIncluded := (factorsDeanonRules.PagesInclude != nil && len(factorsDeanonRules.PagesInclude) != 0)
	if isPageUrlIncluded {
		pageFilterPassed = false
		for _, filter := range factorsDeanonRules.PagesInclude {
			switch filter.Type {
			case model.EqualsOpStr:
				//cleaning incoming page url
				parsedPageUrl, err := U.ParseURLStable(pageUrl)
				if err != nil {
					log.WithField("PageUrl", pageUrl).Error("Error occured while ParseURLStable.")
					continue
				}
				basePageUrl := U.GetURLHostAndPath(parsedPageUrl)

				//cleaning sixsignal config page url
				parsedFilterPageUrl, err := U.ParseURLStable(filter.Value)
				if err != nil {
					log.WithField("PageUrl filter", filter.Value).Error("Error occured while ParseURLStable.")
					continue
				}
				filterPageUrl := U.GetURLHostAndPath(parsedFilterPageUrl)

				if filterPageUrl == basePageUrl {
					pageFilterPassed = true
					break
				}
			case model.ContainsOpStr:
				if strings.Contains(pageUrl, filter.Value) {
					pageFilterPassed = true
					break
				}
			}
		}
		// failed to satisfy page include  filter
		if pageFilterPassed == false {
			return false, nil
		}
	}

	isPageUrlExcluded := (factorsDeanonRules.PagesExclude != nil && len(factorsDeanonRules.PagesExclude) != 0)
	if isPageUrlExcluded {
		for _, filter := range factorsDeanonRules.PagesExclude {
			switch filter.Type {
			case model.EqualsOpStr:
				//cleaning incoming pageUrl
				parsedPageUrl, err := U.ParseURLStable(pageUrl)
				if err != nil {
					log.WithField("PageUrl", pageUrl).Error("Error occured while ParseURLStable.")
					continue
				}
				basePageUrl := U.GetURLHostAndPath(parsedPageUrl)

				//cleaning sixsignal config page url
				parsedFilterPageUrl, err := U.ParseURLStable(filter.Value)
				if err != nil {
					log.WithField("PageUrl filter", filter.Value).Error("Error occured while ParseURLStable.")
					continue
				}
				filterPageUrl := U.GetURLHostAndPath(parsedFilterPageUrl)

				if filterPageUrl == basePageUrl {
					//skip if page name matches
					return false, nil
				}
			case model.ContainsOpStr:
				if strings.Contains(pageUrl, filter.Value) {
					//skip if page contains data matches
					return false, nil
				}
			}
		}
	}

	return pageFilterPassed, nil
}
