package delta

import (
	M "factors/model/model"
	"factors/model/store/memsql"
	"fmt"
	"strings"
)

var bingadsRequiredDocumentTypes = []int{1, 2, 3, 4, 5, 6} //Refer M.BingadsDocumentTypeAlias for clarity

// weekly insights calculation info for each bingads metric
var bingadsMetricToCalcInfo = map[string]ChannelMetricCalculationInfo{
	M.Impressions: {
		Props:     []ChannelPropInfo{{Name: M.Impressions}},
		Operation: "sum",
	},
	M.Clicks: {
		Props:     []ChannelPropInfo{{Name: M.Clicks}},
		Operation: "sum",
	},
	"spend": {
		Props:     []ChannelPropInfo{{Name: "spend"}},
		Operation: "sum",
	},
	M.Conversions: {
		Props:     []ChannelPropInfo{{Name: M.Conversions}},
		Operation: "sum",
	},
}

var bingadsConstantInfo = map[string]string{
	memsql.CAFilterCampaign: M.FilterCampaign,
	memsql.CAFilterAdGroup:  M.FilterAdGroup,
	memsql.CAFilterKeyword:  M.FilterKeyword,
}

func getBingadsFilterPropertyReportName(propName string, objectType string) (string, error) {
	propNameTrimmed := strings.TrimPrefix(propName, objectType+"_")

	if _, ok := bingadsConstantInfo[objectType]; !ok {
		return "", fmt.Errorf("unknown object type: %s", objectType)
	}
	objectTypeTmp := bingadsConstantInfo[objectType] + "s"
	if name, ok := M.BingAdsInternalRepresentationToExternalRepresentationForReports[fmt.Sprintf("%s.%s", objectTypeTmp, propNameTrimmed)]; ok {
		return name, nil
	}
	return "", fmt.Errorf("filter property report name not found for %s", propName)
}

func getBingadsPropertyFilterName(prop string) (string, error) {
	propWithType := strings.SplitN(prop, "#", 2)
	objType := propWithType[0]
	name := propWithType[1]

	if _, ok := bingadsConstantInfo[objType]; !ok {
		return prop, fmt.Errorf("unknown object type: %s", objType)
	}

	for k, v := range M.BingAdsInternalRepresentationToExternalRepresentationForReports {
		if v == name {
			tmpProp := strings.SplitN(k, ".", 2)
			if strings.Contains(tmpProp[0], bingadsConstantInfo[objType]) {
				reqName := strings.Join([]string{objType, objType + "_" + tmpProp[1]}, "#")
				return reqName, nil
			}
		}
	}

	return prop, fmt.Errorf("property filter name not found for %s", prop)
}
