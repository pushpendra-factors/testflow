package tests

import (
	"bufio"
	D "factors/delta"
	M "factors/model/model"
	"factors/model/store/memsql"
	U "factors/util"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	log "github.com/sirupsen/logrus"
)

var filePath = "./data/delta_test.txt"

func TestGetCampaignMetricSimple(t *testing.T) {
	// var wpi *D.WithinPeriodInsightsKpi
	metric := "testMetric"

	propFilter := []M.KPIFilter{{
		ObjectType:       "campaign",
		PropertyName:     "prop1",
		PropertyDataType: U.PropertyTypeCategorical,
		Condition:        M.EqualsOpStr,
		Value:            "val11",
		LogicalOp:        M.LOGICAL_OP_AND,
	}}
	propsToEvalFiltered := []string{"campaign#prop1", "campaign#prop2", "campaign#prop3", "campaign#prop4", "ad_group#prop4", "ad_group#prop5", "ad_group#prop6", "ad#prop6"}
	queryLevel := 3
	infoMap := map[string]string{
		memsql.CAFilterCampaign: "x",
		memsql.CAFilterAdGroup:  "y",
		memsql.CAFilterKeyword:  "z",
		memsql.CAFilterAd:       "a",
	}
	docTypeAlias := map[string]int{
		"x":                    1,
		"y":                    2,
		"a":                    3,
		"x_performance_report": 4,
		"y_performance_report": 5,
		"a_insights":           6,
		"account":              7,
	}
	requiredDocTypes := []int{1, 2, 3, 4, 5, 6}

	metricCalcInfo1 := D.ChannelMetricCalculationInfo{
		Props:     []D.ChannelPropInfo{{Name: "numProp1", DependentKey: "prop1"}},
		Operation: "sum",
		Constants: map[string]float64{"quotient": 10000},
	}

	metricCalcInfo2 := D.ChannelMetricCalculationInfo{
		Props:     []D.ChannelPropInfo{{Name: "numProp2", DependentKey: "prop3", DependentValue: 31, DependentOperation: "="}},
		Operation: "sum",
		Constants: map[string]float64{},
	}

	metricCalcInfo3 := D.ChannelMetricCalculationInfo{
		Props:     []D.ChannelPropInfo{{Name: "numProp1", DependentKey: "prop6"}, {Name: "numProp2", ReplaceValue: map[float64]float64{0: 1000}}},
		Operation: "quotient",
		Constants: map[string]float64{"product": 100},
	}

	metricCalcInfo4 := D.ChannelMetricCalculationInfo{
		Props:     []D.ChannelPropInfo{{Name: "numProp2", DependentKey: "numProp3", DependentValue: 1, DependentOperation: "!="}, {Name: "numProp1", DependentKey: "prop5", ReplaceValue: map[float64]float64{0: 1000}}},
		Operation: "quotient",
		Constants: map[string]float64{"product": 100},
	}

	simpleMetrics := []D.ChannelMetricCalculationInfo{metricCalcInfo1, metricCalcInfo2}
	globalExpected := []float64{0.015, 1}

	for i, metricCalcInfo := range simpleMetrics {
		file, err := os.Open(filePath)
		if err != nil {
			log.Fatal(err)
		}
		scanner := bufio.NewScanner(file)
		if info, _, err := D.GetCampaignMetricSimple(scanner, propFilter, propsToEvalFiltered, queryLevel, metricCalcInfo, docTypeAlias, requiredDocTypes, infoMap, 20220400, 20220800); err != nil {
			log.WithError(err).Error("error GetCampaignMetric for kpi " + metric)
		} else {
			assert.Equal(t, globalExpected[i], info.Global)
			t.Logf("info: %v, err: %v", info, err)
		}
		file.Close()
	}

	complexMetrics := []D.ChannelMetricCalculationInfo{metricCalcInfo3, metricCalcInfo4}
	globalExpected = []float64{5000, 2.5}

	for i, metricCalcInfo := range complexMetrics {
		file, err := os.Open(filePath)
		if err != nil {
			log.Fatal(err)
		}
		scanner := bufio.NewScanner(file)
		if info, _, err := D.GetCampaignMetricComplex(scanner, propFilter, propsToEvalFiltered, queryLevel, metricCalcInfo, docTypeAlias, requiredDocTypes, infoMap, 20220400, 20220800); err != nil {
			log.WithError(err).Error("error GetCampaignMetric for kpi " + metric)
		} else {
			assert.Equal(t, globalExpected[i], info.Global)
			t.Logf("info: %v", info)
		}
		file.Close()
	}

}
