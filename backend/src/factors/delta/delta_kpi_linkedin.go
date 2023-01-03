package delta

import (
	M "factors/model/model"
	"factors/model/store/memsql"
	"fmt"
	"strings"
)

var linkedinRequiredDocumentTypes = []int{1, 2, 3, 4, 5, 6} //Refer memsql.LinkedinDocumentTypeAlias for clarity

var linkedinMetricToCalcInfo = map[string]MetricCalculationInfo{
	M.Impressions: {
		Props:     []ChannelPropInfo{{Name: M.Impressions}},
		Operation: "sum",
	},
	M.Clicks: {
		Props:     []ChannelPropInfo{{Name: M.Clicks}},
		Operation: "sum",
	},
	"spend": {
		Props:     []ChannelPropInfo{{Name: "costInLocalCurrency"}},
		Operation: "sum",
	},
	M.Conversions: {
		Props:     []ChannelPropInfo{{Name: "conversionValueInLocalCurrency"}},
		Operation: "sum",
	},
}

var linkedinConstantInfo = map[string]string{
	memsql.CAFilterCampaign: M.LinkedinCampaignGroup,
	memsql.CAFilterAdGroup:  M.LinkedinCampaign,
	memsql.CAFilterKeyword:  M.LinkedinCreative,
}

func getLinkedinFilterPropertyReportName(propName string, objectType string) (string, error) {
	propNameTrimmed := strings.TrimPrefix(propName, objectType+"_")

	if _, ok := linkedinConstantInfo[objectType]; !ok {
		return "", fmt.Errorf("unknown object type: %s", objectType)
	}
	return fmt.Sprintf("%s_%s", linkedinConstantInfo[objectType], propNameTrimmed), nil
}

func getLinkedinPropertyFilterName(prop string) (string, error) {
	propWithType := strings.SplitN(prop, "#", 2)
	objType := propWithType[0]
	name := propWithType[1]

	nameSplit := strings.Split(name, "_")
	num := len(nameSplit)
	reqName := strings.Join([]string{objType, objType + "_" + nameSplit[num-1]}, "#")
	return reqName, nil
}
