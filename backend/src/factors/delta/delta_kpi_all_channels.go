package delta

import (
	"factors/filestore"
	"factors/merge"
	M "factors/model/model"
	serviceDisk "factors/services/disk"
	"strings"

	log "github.com/sirupsen/logrus"
)

var channelValueFilterName = map[string]string{
	M.ADWORDS:        "Google Ads",
	M.BINGADS:        "Bing Ads",
	M.LINKEDIN:       "LinkedIn Ads",
	M.FACEBOOK:       "Facebook Ads",
	M.GOOGLE_ORGANIC: "Google Ads",
}

func getAllChannelMetricsInfo(metric string, propFilter []M.KPIFilter, propsToEval []string, projectId int64, periodCode Period, archiveCloudManager, tmpCloudManager, sortedCloudManager *filestore.FileManager,
	diskManager *serviceDisk.DiskDriver, beamConfig *merge.RunBeamConfig, useBucketV2 bool) (*WithinPeriodInsightsKpi, error) {
	var wpi WithinPeriodInsightsKpi
	wpi.MetricInfo = &MetricInfo{}
	wpi.ScaleInfo = &MetricInfo{}
	wpi.ScaleInfo.Features = make(map[string]map[string]float64)

	for _, channel := range []string{M.ADWORDS, M.BINGADS, M.LINKEDIN, M.FACEBOOK, M.GOOGLE_ORGANIC} {
		passFilter := true
		var newPropFilter []M.KPIFilter
		for _, filter := range propFilter {
			if filter.LogicalOp == "AND" {
				passFilter = false
			}
			if filter.ObjectType == "channel" {
				if ok, err := checkValSatisfiesFilterCondition(filter, channelValueFilterName[channel]); err != nil {
					return &wpi, err
				} else if ok {
					passFilter = true
				}
			} else {
				newPropFilter = append(newPropFilter, filter)
			}
		}
		if !passFilter {
			continue
		}

		scanner, err := GetChannelFileScanner(channel, projectId, periodCode, archiveCloudManager, tmpCloudManager, sortedCloudManager, diskManager, true, beamConfig, useBucketV2)
		if err != nil {
			log.WithError(err).Error("failed getting " + channel + " file scanner for all channel kpi")
			continue
		}

		var newPropsToEval []string
		for _, prop := range propsToEval {
			propWithType := strings.SplitN(prop, "#", 2)
			objType := propWithType[0]
			propName := propWithType[1]
			name, err := getFilterPropertyReportName(channel)(propName, objType)
			if err != nil {
				log.WithError(err).Error("error getting property name for channel " + channel + " for all channel kpi")
				continue
			}
			newName := strings.Join([]string{objType, name}, "#")
			newPropsToEval = append(newPropsToEval, newName)
		}
		wpiTmp, err := getCampaignMetricsInfo(metric, channel, scanner, newPropFilter, newPropsToEval)
		if err != nil {
			log.WithError(err).Error("error GetCampaignMetricInfo for all channel kpi for source " + channel)
			continue
		} else {
			wpi.MetricInfo = addMetricInfoStructForSource(channel, wpi.MetricInfo, wpiTmp.MetricInfo)
			// wpi.ScaleInfo = addMetricInfoStruct(wpi.ScaleInfo, wpiTmp.ScaleInfo)
		}
	}
	return &wpi, nil
}

func addMetricInfoStructForSource(source string, baseInfo *MetricInfo, info2add *MetricInfo) *MetricInfo {
	if info2add == nil {
		return baseInfo
	}
	info := *baseInfo
	info.Global += info2add.Global
	if info.Features == nil {
		info.Features = make(map[string]map[string]float64)
	}
	for key, valMap := range info2add.Features {
		if _, ok := info.Features[key]; !ok {
			info.Features[key] = make(map[string]float64)
		}
		for val, cnt := range valMap {
			info.Features[key][val] += cnt
		}
	}
	if _, ok := info.Features["channel#channel_name"]; !ok {
		info.Features["channel#channel_name"] = make(map[string]float64)
	}
	info.Features["channel#channel_name"][channelValueFilterName[source]] = info2add.Global
	return &info
}

// func getAllChannelPropsToEvalForChannel(channel string, propsToEval []string) ([]string, error) {
// 	newPropsToEval := make([]string, 0)
// 	for _, prop := range propsToEval {
// 		var newName string
// 		propArr := strings.SplitN(prop, "#", 2)
// 		propType, propName := propArr[0], propArr[1]
// 		name := strings.TrimPrefix(propName, propType+"_")
// 		switch channel {
// 		case M.ADWORDS:
// 			newName = M.AdwordsInternalPropertiesToReportsInternal[propType+":"+name]
// 		case M.BINGADS:
// 			var propTypeTmp string
// 			if propType != M.FilterKeyword {
// 				propTypeTmp = propType + "s"
// 			}
// 			newName = M.BingAdsInternalRepresentationToExternalRepresentationForReports[propTypeTmp+"."+name]
// 		case M.FACEBOOK:
// 			var propTypeTmp string
// 			if propType == M.CAFilterAdGroup {
// 				propTypeTmp = "ad_set"
// 			}
// 			newName = M.ObjectToValueInFacebookJobsMapping[propTypeTmp+":"+name]
// 		case M.LINKEDIN:
// 			if propType == M.CAFilterCampaign {
// 				newName = M.LinkedinCampaignGroup + "_" + prop
// 			} else if propType == M.CAFilterAdGroup {
// 				newName = M.LinkedinCampaign + "_" + prop
// 			}
// 		case M.GOOGLE_ORGANIC:
// 			continue
// 		}
// 		newPropsToEval = append(newPropsToEval, propType+"#"+newName)
// 	}
// 	return newPropsToEval, nil
// }
