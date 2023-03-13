package delta

import (
	"bufio"
	"bytes"
	"encoding/json"
	"factors/filestore"
	"factors/merge"
	M "factors/model/model"
	"factors/model/store"
	serviceDisk "factors/services/disk"
	U "factors/util"
	"fmt"
	"io/ioutil"
	"strings"

	log "github.com/sirupsen/logrus"
)

var notOperations = []string{M.NotEqualOpStr, M.NotContainsOpStr, M.NotEqualOp}

func createKpiInsights(diskManager *serviceDisk.DiskDriver, archiveCloudManager, tmpCloudManager, sortedCloudManager, cloudManager *filestore.FileManager, periodCodesWithWeekNMinus1 []Period, projectId int64, queryId int64, queryGroup M.KPIQueryGroup,
	topK int, skipWpi, skipWpi2 bool, mailerRun bool, beamConfig *merge.RunBeamConfig, useBucketV2 bool, status map[string]interface{}) error {
	// readEvents := true
	var err error
	var newInsightsList = make([]*WithinPeriodInsightsKpi, 0)
	var oldInsightsList = make([]*WithinPeriodInsightsKpi, 0)

	skipW1 := false
	skipW2 := false
	if skipWpi {
		dateString := U.GetDateOnlyFromTimestampZ(periodCodesWithWeekNMinus1[0].From)
		path, name := (*cloudManager).GetInsightsWpiFilePathAndName(projectId, dateString, queryId, topK, mailerRun)
		if reader, err := (*cloudManager).Get(path, name); err == nil {
			data, err := ioutil.ReadAll(reader)
			if err == nil {
				err := json.Unmarshal(data, &oldInsightsList)
				if err == nil {
					skipW1 = true
				}
			}
		}
	}

	if skipWpi2 {
		dateString := U.GetDateOnlyFromTimestampZ(periodCodesWithWeekNMinus1[1].From)
		path, name := (*cloudManager).GetInsightsWpiFilePathAndName(projectId, dateString, queryId, topK, mailerRun)
		if reader, err := (*cloudManager).Get(path, name); err == nil {
			data, err := ioutil.ReadAll(reader)
			if err == nil {
				err := json.Unmarshal(data, &newInsightsList)
				if err == nil {
					skipW2 = true
				}
			}
		}
	}

	for i, query := range queryGroup.Queries {

		//no within period calculations
		if skipW1 && skipW2 {
			break
		}

		//every query occurs twice so
		if i%2 == 0 {
			continue
		}

		metric := query.Metrics[0]

		statusErrorKey := fmt.Sprintf("error-kpi-%d-%d", queryId, i/2+1)

		//get global + local constraints
		propFilter := append(queryGroup.GlobalFilters, query.Filters...)

		var pageOrChannel string
		var spectrum string
		//get features for insights as a map
		var propsToEval = make([]string, 0)

		{
			//get proper props based on category
			var kpiProperties []map[string]string
			var channelOrEvent string
			kpiProperties, spectrum, channelOrEvent, err = getPropertiesToEvaluateAndInfo(projectId, query.DisplayCategory)
			if err != nil {
				wpi := &WithinPeriodInsightsKpi{Category: spectrum, MetricInfo: &MetricInfo{}, ScaleInfo: &MetricInfo{}}
				newInsightsList = append(newInsightsList, wpi)
				oldInsightsList = append(oldInsightsList, wpi)
				log.WithError(err).Errorf("error getPropertiesToEvaluateAndInfo for metric: %s", metric)
				status[statusErrorKey] = err
				continue
			}
			if spectrum == "custom" {
				kpiProperties, err = getFilteredKpiPropertiesForCustomMetric(kpiProperties, metric, projectId, periodCodesWithWeekNMinus1, archiveCloudManager, tmpCloudManager, sortedCloudManager, diskManager, topK, beamConfig, useBucketV2)
				if err != nil {
					wpi := &WithinPeriodInsightsKpi{Category: spectrum, MetricInfo: &MetricInfo{}, ScaleInfo: &MetricInfo{}}
					newInsightsList = append(newInsightsList, wpi)
					oldInsightsList = append(oldInsightsList, wpi)
					log.WithError(err).Errorf("error getFilteredKpiPropertiesForCustomMetric for metric: %s", metric)
					status[statusErrorKey] = err
					continue
				}
			}

			for _, propMap := range kpiProperties {
				if propMap["data_type"] == U.PropertyTypeCategorical {
					var propType string
					if ent, ok := propMap["entity"]; ok {
						if ent == M.UserEntity {
							propType = "up"
						} else if ent == M.EventEntity {
							propType = "ep"
						} else {
							propType = ent
						}
					}
					propName := strings.Join([]string{propType, propMap["name"]}, "#")
					propsToEval = append(propsToEval, propName)
				}
			}

			pageOrChannel = query.PageUrl
			if channelOrEvent != "" {
				pageOrChannel = channelOrEvent
			}
		}

		//get week 2 metrics by reading file
		if !skipW2 {
			if wpi, err := getMetricEvaluated(spectrum, query.DisplayCategory, metric, pageOrChannel, propFilter, propsToEval, projectId, periodCodesWithWeekNMinus1[1], archiveCloudManager, tmpCloudManager, sortedCloudManager, diskManager, beamConfig, useBucketV2); err != nil {
				wpi = &WithinPeriodInsightsKpi{Category: spectrum, MetricInfo: &MetricInfo{}, ScaleInfo: &MetricInfo{}}
				newInsightsList = append(newInsightsList, wpi)
				log.WithError(err).Errorf("error GetMetricEvaluated for week 2 and metric: %s", metric)
				status[statusErrorKey] = err
				continue
			} else {
				wpi.Category = spectrum
				newInsightsList = append(newInsightsList, wpi)
			}
		}

		//get week 1 metrics by reading file
		if !skipW1 {
			if wpi, err := getMetricEvaluated(spectrum, query.DisplayCategory, metric, pageOrChannel, propFilter, propsToEval, projectId, periodCodesWithWeekNMinus1[0], archiveCloudManager, tmpCloudManager, sortedCloudManager, diskManager, beamConfig, useBucketV2); err != nil {
				wpi = &WithinPeriodInsightsKpi{Category: spectrum, MetricInfo: &MetricInfo{}, ScaleInfo: &MetricInfo{}}
				oldInsightsList = append(oldInsightsList, wpi)
				log.WithError(err).Errorf("error GetMetricEvaluated for week 1 and metric: %s", metric)
				status[statusErrorKey] = err
				continue
			} else {
				wpi.Category = spectrum
				oldInsightsList = append(oldInsightsList, wpi)
			}
		}
	}

	if !skipW2 {
		wpiBytes, err := json.Marshal(newInsightsList)
		if err != nil {
			log.WithError(err).Error("failed to marshal wpi2 Info.")
			return err
		}

		err = writeWpiPath(projectId, periodCodesWithWeekNMinus1[1], queryId, topK, bytes.NewReader(wpiBytes), *cloudManager, mailerRun)
		if err != nil {
			log.WithError(err).Error("write WPI error - ", err)
			return err
		}
	}
	if !skipW1 {
		wpiBytes, err := json.Marshal(oldInsightsList)
		if err != nil {
			log.WithError(err).Error("failed to marshal wpi1 Info.")
			return err
		}
		err = writeWpiPath(projectId, periodCodesWithWeekNMinus1[0], queryId, topK, bytes.NewReader(wpiBytes), *cloudManager, mailerRun)
		if err != nil {
			log.WithError(err).Error("write WPI error - ", err)
			return err
		}
	}

	//get insights between the weeks
	var crossPeriodInsightsList []*CrossPeriodInsightsKpi
	periodPair := PeriodPair{First: periodCodesWithWeekNMinus1[0], Second: periodCodesWithWeekNMinus1[1]}
	if len(newInsightsList) > 0 && len(oldInsightsList) > 0 {
		crossPeriodInsightsList, err = computeCrossPeriodKpiInsights(periodPair, newInsightsList, oldInsightsList)
		if err != nil {
			log.WithError(err).Error("compute cpi for kpi error - ", err)
			return err
		}
	}

	//Create cpi file with insights
	if len(crossPeriodInsightsList) > 0 {
		crossPeriodInsightsBytes, err := json.Marshal(crossPeriodInsightsList)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("failed to unmarshal cpi Info.")
			return err
		}
		err = writeCpiPath(projectId, periodPair.Second, queryId, topK, bytes.NewReader(crossPeriodInsightsBytes), *cloudManager, mailerRun)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("failed to write cpi files to cloud")
			return err
		}
	}
	return nil
}

// (queryEvent works as channel or page depending on spectrum)
// get wpi for kpi for a week
func getMetricEvaluated(spectrum, category string, metric string, eventOrChannel string, propFilter []M.KPIFilter, propsToEval []string, projectId int64, periodCode Period, archiveCloudManager, tmpCloudManager, sortedCloudManager *filestore.FileManager,
	diskManager *serviceDisk.DiskDriver, beamConfig *merge.RunBeamConfig, useBucketV2 bool) (*WithinPeriodInsightsKpi, error) {

	var insights *WithinPeriodInsightsKpi
	var err error
	var scanner *bufio.Scanner
	if spectrum == "events" {
		if scanner, err = GetEventFileScanner(projectId, periodCode, archiveCloudManager, tmpCloudManager, sortedCloudManager, diskManager, beamConfig, useBucketV2); err != nil {
			log.WithError(err).Error("failed getting event file scanner")
			return nil, err
		}
		insights, err = getEventMetricsInfo(metric, eventOrChannel, scanner, propFilter, propsToEval)
	} else if spectrum == "campaign" {
		if category == M.AllChannelsDisplayCategory {
			insights, err = getAllChannelMetricsInfo(metric, propFilter, propsToEval, projectId, periodCode, archiveCloudManager, tmpCloudManager, sortedCloudManager, diskManager, beamConfig, useBucketV2)
		} else {
			if scanner, err = GetChannelFileScanner(eventOrChannel, projectId, periodCode, archiveCloudManager, tmpCloudManager, sortedCloudManager, diskManager, beamConfig, useBucketV2); err != nil {
				log.WithError(err).Error("failed getting " + eventOrChannel + " file scanner")
				return nil, err
			}
			insights, err = getCampaignMetricsInfo(metric, eventOrChannel, scanner, propFilter, propsToEval)
		}
	} else if spectrum == "custom" {
		insights, err = getCustomMetricsInfo(metric, propFilter, propsToEval, projectId, periodCode, archiveCloudManager, tmpCloudManager, sortedCloudManager, diskManager, beamConfig, useBucketV2)
	} else {
		err = fmt.Errorf("unknown spectrum: %s", spectrum)
	}

	return insights, err
}

// compute cross period using within period infos
func computeCrossPeriodKpiInsights(periodPair PeriodPair, newInsightsList, oldInsightsList []*WithinPeriodInsightsKpi) ([]*CrossPeriodInsightsKpi, error) {
	crossPeriodInsightsList := make([]*CrossPeriodInsightsKpi, 0)
	if len(newInsightsList) != len(oldInsightsList) {
		return nil, fmt.Errorf("error computeCrossPeriodKpiInsights: both lists should have same length")
	}
	for i := range newInsightsList {
		var crossPeriodInsights CpiMetricInfo
		newInsights := newInsightsList[i]
		oldInsights := oldInsightsList[i]
		oldInfo := *(*oldInsights).MetricInfo
		newInfo := *(*newInsights).MetricInfo

		if newInfo.Features == nil {
			newInfo.Features = make(map[string]map[string]float64)
		}
		if oldInfo.Features == nil {
			oldInfo.Features = make(map[string]map[string]float64)
		}

		//get union of props
		var allProps = make(map[string]map[string]bool)
		for key, valMap := range oldInfo.Features {
			allProps[key] = make(map[string]bool)
			for val := range valMap {
				allProps[key][val] = true
			}
		}
		for key, valMap := range newInfo.Features {
			if _, ok := allProps[key]; !ok {
				allProps[key] = make(map[string]bool)
			}
			for val := range valMap {
				allProps[key][val] = true
			}
		}

		//global
		first := oldInfo.Global
		second := newInfo.Global
		var percentChange, factor float64
		if first != 0 {
			percentChange = 100 * float64(second-first) / float64(first)
			factor = float64(second) / float64(first)
		} else {
			percentChange = 100
			factor = float64(second)
		}
		crossPeriodInsights.GlobalMetrics = DiffMetric{First: first, Second: second, PercentChange: percentChange, FactorChange: factor}

		//features
		crossPeriodInsights.FeatureMetrics = make(map[string]map[string]DiffMetric)
		for key, valMap := range allProps {
			if _, ok := newInfo.Features[key]; !ok {
				newInfo.Features[key] = make(map[string]float64)
			}
			if _, ok := oldInfo.Features[key]; !ok {
				oldInfo.Features[key] = make(map[string]float64)
			}
			for val := range valMap {
				first := oldInfo.Features[key][val]
				second := newInfo.Features[key][val]
				var percentChange, factor float64
				if first != 0 {
					percentChange = 100 * (second - first) / first
					factor = second / first
				} else {
					percentChange = 100
					factor = second
				}
				if _, ok := crossPeriodInsights.FeatureMetrics[key]; !ok {
					crossPeriodInsights.FeatureMetrics[key] = make(map[string]DiffMetric)
				}
				crossPeriodInsights.FeatureMetrics[key][val] = DiffMetric{First: first, Second: second, PercentChange: percentChange, FactorChange: factor}
			}
		}

		var scaleInfo CpiMetricInfo
		oldScale := *(*oldInsights).ScaleInfo
		newScale := *(*newInsights).ScaleInfo

		if oldScale.Features == nil {
			oldScale.Features = make(map[string]map[string]float64)
		}
		if newScale.Features == nil {
			newScale.Features = make(map[string]map[string]float64)
		}

		//global
		first = oldScale.Global
		second = newScale.Global
		if first != 0 {
			percentChange = 100 * float64(second-first) / float64(first)
			factor = float64(second) / float64(first)
		} else {
			percentChange = 100
			factor = float64(second)
		}
		scaleInfo.GlobalMetrics = DiffMetric{First: first, Second: second, PercentChange: percentChange, FactorChange: factor}

		//features
		scaleInfo.FeatureMetrics = make(map[string]map[string]DiffMetric)
		for key, valMap := range allProps {
			if _, ok := newScale.Features[key]; !ok {
				newScale.Features[key] = make(map[string]float64)
			}
			if _, ok := oldScale.Features[key]; !ok {
				oldScale.Features[key] = make(map[string]float64)
			}
			for val := range valMap {
				first := oldScale.Features[key][val]
				second := newScale.Features[key][val]
				var percentChange, factor float64
				if first != 0 {
					percentChange = 100 * (second - first) / first
					factor = second / first
				} else {
					percentChange = 100
					factor = second
				}
				if _, ok := scaleInfo.FeatureMetrics[key]; !ok {
					scaleInfo.FeatureMetrics[key] = make(map[string]DiffMetric)
				}
				scaleInfo.FeatureMetrics[key][val] = DiffMetric{First: first, Second: second, PercentChange: percentChange, FactorChange: factor}
			}
		}
		var cpiInsightsKpi CrossPeriodInsightsKpi
		cpiInsightsKpi.Category = newInsights.Category
		cpiInsightsKpi.Periods = periodPair
		cpiInsightsKpi.Target = &crossPeriodInsights
		cpiInsightsKpi.BaseAndTarget = &crossPeriodInsights
		cpiInsightsKpi.ScaleInfo = &scaleInfo
		crossPeriodInsightsList = append(crossPeriodInsightsList, &cpiInsightsKpi)
	}
	return crossPeriodInsightsList, nil
}

// get all info regarding displayCategory (properties, spectrum, channel)
func getPropertiesToEvaluateAndInfo(projectID int64, displayCategory string) ([]map[string]string, string, string, error) {
	var kpiProperties []map[string]string
	var spectrum string
	var channelOrEvent string
	if displayCategory == M.WebsiteSessionDisplayCategory {
		kpiProperties = M.KPIPropertiesForWebsiteSessions
		spectrum = "events"
		channelOrEvent = U.EVENT_NAME_SESSION
	} else if displayCategory == M.FormSubmissionsDisplayCategory {
		kpiProperties = M.KPIPropertiesForFormSubmissions
		spectrum = "events"
		channelOrEvent = U.EVENT_NAME_FORM_SUBMITTED
	} else if displayCategory == M.PageViewsDisplayCategory {
		kpiProperties = M.KPIPropertiesForPageViews
		spectrum = "events"
	} else if displayCategory == M.GoogleAdsDisplayCategory || displayCategory == M.AdwordsDisplayCategory {
		for category, propMap := range M.MapOfAdwordsObjectsToPropertiesAndRelated {
			for prop, info := range propMap {
				kpiProperties = append(kpiProperties, map[string]string{
					"name":      M.AdwordsInternalPropertiesToReportsInternal[category+":"+prop],
					"data_type": info.TypeOfProperty,
					"entity":    category,
				})
			}
		}
		spectrum = "campaign"
		channelOrEvent = M.ADWORDS
	} else if displayCategory == M.BingAdsDisplayCategory {
		for category, propMap := range M.MapOfBingAdsObjectsToPropertiesAndRelated {
			category2 := category
			if category != M.FilterKeyword {
				category2 = category + "s"
			}
			for prop, info := range propMap {
				kpiProperties = append(kpiProperties, map[string]string{
					"name":      M.BingAdsInternalRepresentationToExternalRepresentationForReports[category2+"."+prop],
					"data_type": info.TypeOfProperty,
					"entity":    category,
				})
			}
		}
		spectrum = "campaign"
		channelOrEvent = M.BINGADS
	} else if displayCategory == M.FacebookDisplayCategory {
		for category, propMap := range M.MapOfFacebookObjectsToPropertiesAndRelated {
			category2 := category
			if category == M.CAFilterAdGroup {
				category2 = "ad_set"
			}
			for prop, info := range propMap {
				kpiProperties = append(kpiProperties, map[string]string{
					"name":      M.ObjectToValueInFacebookJobsMapping[category2+":"+prop],
					"data_type": info.TypeOfProperty,
					"entity":    category,
				})
			}
		}
		spectrum = "campaign"
		channelOrEvent = M.FACEBOOK
	} else if displayCategory == M.LinkedinDisplayCategory {
		for _, prop := range []string{"id", "name"} {
			kpiProperties = append(kpiProperties, map[string]string{
				"name":      M.LinkedinCampaignGroup + "_" + prop,
				"data_type": U.PropertyTypeCategorical,
				"entity":    M.CAFilterCampaign,
			})
			kpiProperties = append(kpiProperties, map[string]string{
				"name":      M.LinkedinCampaign + "_" + prop,
				"data_type": U.PropertyTypeCategorical,
				"entity":    M.CAFilterAdGroup,
			})
		}
		spectrum = "campaign"
		channelOrEvent = M.LINKEDIN
	} else if displayCategory == M.GoogleOrganicDisplayCategory {
		for categ, propMap := range M.MapOfObjectsToPropertiesAndRelatedGoogleOrganic {
			for prop, info := range propMap {
				kpiProperties = append(kpiProperties, map[string]string{
					"name":      prop,
					"data_type": info.TypeOfProperty,
					"entity":    categ,
				})
			}
		}
		spectrum = "campaign"
		channelOrEvent = M.GOOGLE_ORGANIC
	} else if displayCategory == M.AllChannelsDisplayCategory {
		for _, prop := range []string{"id", "name"} {
			kpiProperties = append(kpiProperties, map[string]string{
				"name":      M.CAFilterCampaign + "_" + prop,
				"data_type": U.PropertyTypeCategorical,
				"entity":    M.CAFilterCampaign,
			})
			kpiProperties = append(kpiProperties, map[string]string{
				"name":      M.CAFilterAdGroup + "_" + prop,
				"data_type": U.PropertyTypeCategorical,
				"entity":    M.CAFilterAdGroup,
			})
		}
		spectrum = "campaign"
		channelOrEvent = "all_ads"
	} else {
		switch displayCategory {
		case M.HubspotDealsDisplayCategory:
			kpiProperties = store.GetStore().GetPropertiesForHubspotDeals(projectID, "")
		case M.HubspotCompaniesDisplayCategory:
			kpiProperties = store.GetStore().GetPropertiesForHubspotCompanies(projectID, "")
		case M.HubspotContactsDisplayCategory:
			kpiProperties = store.GetStore().GetPropertiesForHubspotContacts(projectID, "")
		case M.SalesforceUsersDisplayCategory:
			kpiProperties = store.GetStore().GetPropertiesForSalesforceUsers(projectID, "")
		case M.SalesforceAccountsDisplayCategory:
			kpiProperties = store.GetStore().GetPropertiesForSalesforceAccounts(projectID, "")
		case M.SalesforceOpportunitiesDisplayCategory:
			kpiProperties = store.GetStore().GetPropertiesForSalesforceOpportunities(projectID, "")
		case M.MarketoLeadsDisplayCategory:
			kpiProperties = store.GetStore().GetPropertiesForMarketo(projectID, "")
		case M.LeadSquaredLeadsDisplayCategory:
			kpiProperties = store.GetStore().GetPropertiesForLeadSquared(projectID, "")
		default:
			err := fmt.Errorf("no properties to evaluate for category: %s", displayCategory)
			log.WithError(err).Error("unknown category")
			return nil, "", "", err
		}
		spectrum = "custom"
	}
	return kpiProperties, spectrum, channelOrEvent, nil
}
