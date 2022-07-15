package model

import (
	U "factors/util"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

// MergeDataRowsHavingSameKeyKPI merges rows having same key by adding each column value
func MergeDataRowsHavingSameKeyKPI(rows [][]interface{}, keyIndex int, attributionKey string, analyzeType string, conversionFunTypes []string, logCtx log.Entry) [][]interface{} {

	rowKeyMap := make(map[string][]interface{})
	maxRowSize := 0
	for _, row := range rows {
		maxRowSize = U.MaxInt(len(row), maxRowSize)
		if len(row) == 0 || len(row) != maxRowSize {
			continue
		}
		// creating a key for using added keys and index
		key := ""
		for j := 0; j <= keyIndex; j++ {
			val, ok := row[j].(string)
			// Ignore row if key is not proper
			if !ok {
				logCtx.Info("empty key value error. Ignoring row and continuing.")
				continue
			}
			key = key + val
		}
		if _, exists := rowKeyMap[key]; exists {
			rowKeyMap[key] = MergeTwoDataRowsKPI(rowKeyMap[key], row, keyIndex, attributionKey, analyzeType, conversionFunTypes)
		} else {
			rowKeyMap[key] = row
		}
	}
	resultRows := make([][]interface{}, 0)
	for _, mapRow := range rowKeyMap {
		resultRows = append(resultRows, mapRow)
	}
	return resultRows
}

//MergeTwoDataRowsKPI adds values of two data rows for KPI attribution queries
func MergeTwoDataRowsKPI(row1 []interface{}, row2 []interface{}, keyIndex int, attributionKey string, analyzeType string, conversionFunTypes []string) []interface{} {

	if analyzeType != AnalyzeTypeHSDeals && analyzeType != AnalyzeTypeSFOpportunities {
		log.WithFields(log.Fields{"AnalyzeType": analyzeType}).Error("KPI-Attribution invalid method call - analyzeType")
		return row1
	}

	row1[keyIndex+1] = row1[keyIndex+1].(int64) + row2[keyIndex+1].(int64)     // Impressions.
	row1[keyIndex+2] = row1[keyIndex+2].(int64) + row2[keyIndex+2].(int64)     // Clicks.
	row1[keyIndex+3] = row1[keyIndex+3].(float64) + row2[keyIndex+3].(float64) // Spend.

	if float64(row1[keyIndex+8].(int64)+row2[keyIndex+8].(int64)) > 0 {
		row1[keyIndex+10], _ = U.FloatRoundOffWithPrecision((float64(row1[keyIndex+8].(int64))*row1[keyIndex+10].(float64)+float64(row2[keyIndex+8].(int64))*row2[keyIndex+10].(float64))/float64(row1[keyIndex+8].(int64)+row2[keyIndex+8].(int64)), U.DefaultPrecision) //AvgSessionTime.
	} else {
		row1[keyIndex+10] = float64(0)
	}

	row1[keyIndex+8] = row1[keyIndex+8].(int64) + row2[keyIndex+8].(int64)    // Sessions.
	row1[keyIndex+9] = row1[keyIndex+9].(int64) + row2[keyIndex+9].(int64)    // Users.
	row1[keyIndex+11] = row1[keyIndex+11].(int64) + row2[keyIndex+11].(int64) // PageViews.

	for idx, _ := range conversionFunTypes {
		nextConPosition := idx * 6
		row1[keyIndex+12+nextConPosition] = row1[keyIndex+12+nextConPosition].(float64) + row2[keyIndex+12+nextConPosition].(float64) // Conversion.
		row1[keyIndex+15+nextConPosition] = row1[keyIndex+15+nextConPosition].(float64) + row2[keyIndex+15+nextConPosition].(float64) // Compare Conversion.
	}
	impressions := (row1[keyIndex+1]).(int64)
	clicks := (row1[keyIndex+2]).(int64)
	spend := row1[keyIndex+3].(float64)

	if float64(impressions) > 0 {
		row1[keyIndex+4], _ = U.FloatRoundOffWithPrecision(100*float64(clicks)/float64(impressions), U.DefaultPrecision) // CTR.
		row1[keyIndex+6], _ = U.FloatRoundOffWithPrecision(1000*float64(spend)/float64(impressions), U.DefaultPrecision) // CPM.
	} else {
		row1[keyIndex+4] = float64(0) // CTR.
		row1[keyIndex+6] = float64(0) // CPM.
	}
	if float64(clicks) > 0 {
		row1[keyIndex+5], _ = U.FloatRoundOffWithPrecision(float64(spend)/float64(clicks), U.DefaultPrecision)                           // AvgCPC.
		row1[keyIndex+7], _ = U.FloatRoundOffWithPrecision(100*float64(row1[keyIndex+12].(float64))/float64(clicks), U.DefaultPrecision) // ClickConversionRate.
	} else {
		row1[keyIndex+5] = float64(0) // AvgCPC.
		row1[keyIndex+7] = float64(0) // ClickConversionRate.
	}

	for idx, funcType := range conversionFunTypes {
		nextConPosition := idx * 6
		// Normal conversion [12, 13, 14] = [Conversion, CPC, Rate]

		if row1[keyIndex+9].(int64) > 0 {
			row1[keyIndex+14+nextConPosition], _ = U.FloatRoundOffWithPrecision(row1[keyIndex+12+nextConPosition].(float64)/float64(row1[keyIndex+9].(int64))*100, U.DefaultPrecision)
		} else {
			row1[keyIndex+9] = int64(0)
			row1[keyIndex+14+nextConPosition] = float64(0)
		}

		// Compare conversion  = [Conversion, CPC, Rate+nextConPosition]

		if row1[keyIndex+9].(int64) > 0 {
			row1[keyIndex+17+nextConPosition], _ = U.FloatRoundOffWithPrecision(row1[keyIndex+15+nextConPosition].(float64)/float64(row1[keyIndex+9].(int64))*100, U.DefaultPrecision)
		} else {
			row1[keyIndex+9] = int64(0)
			row1[keyIndex+17+nextConPosition] = float64(0)
		}

		if strings.ToLower(funcType) == "sum" {

			if spend > 0 {
				row1[keyIndex+13+nextConPosition], _ = U.FloatRoundOffWithPrecision(row1[keyIndex+12+nextConPosition].(float64)/spend, U.DefaultPrecision) // Conversion - CPC.
			} else {
				row1[keyIndex+13+nextConPosition] = float64(0) // Conversion - CPC.
			}

			if spend > 0 {
				row1[keyIndex+16+nextConPosition], _ = U.FloatRoundOffWithPrecision(row1[keyIndex+15+nextConPosition].(float64)/spend, U.DefaultPrecision) // Compare Conversion - CPC.
			} else {
				row1[keyIndex+16+nextConPosition] = float64(0) // Compare Conversion - CPC.
			}

		} else {

			if row1[keyIndex+12+nextConPosition].(float64) > 0 {
				row1[keyIndex+13+nextConPosition], _ = U.FloatRoundOffWithPrecision(spend/row1[keyIndex+12+nextConPosition].(float64), U.DefaultPrecision) // Conversion - CPC.
			} else {
				row1[keyIndex+12+nextConPosition] = float64(0)
				row1[keyIndex+13+nextConPosition] = float64(0) // Conversion - CPC.
			}

			if row1[keyIndex+15+nextConPosition].(float64) > 0 {
				row1[keyIndex+16+nextConPosition], _ = U.FloatRoundOffWithPrecision(spend/row1[keyIndex+15+nextConPosition].(float64), U.DefaultPrecision) // Compare Conversion - CPC.
			} else {
				row1[keyIndex+15+nextConPosition] = float64(0)
				row1[keyIndex+16+nextConPosition] = float64(0) // Compare Conversion - CPC.
			}
		}
	}
	return row1
}

func AddKPIKeyDataInMap(kpiQueryResult QueryResult, logCtx log.Entry, keyIdx int,
	datetimeIdx int, from int64, to int64, valIdx int, kpiValueHeaders []string,
	kpiAggFunctionType []string, kpiData *map[string]KPIInfo) []string {

	var kpiKeys []string

	for _, row := range kpiQueryResult.Rows {

		logCtx.WithFields(log.Fields{"Row": row}).Info("KPI-Attribution KPI Row")

		var kpiDetail KPIInfo

		// get ID
		kpiID := row[keyIdx].(string)

		// get time
		eventTime, err := time.Parse(time.RFC3339, row[datetimeIdx].(string))
		if err != nil {
			logCtx.WithError(err).WithFields(log.Fields{"timestamp": row[datetimeIdx]}).Error("couldn't parse the timestamp for KPI query, continuing")
			continue
		}
		kpiDetail.Timestamp = eventTime.Unix()

		if kpiDetail.Timestamp > to || kpiDetail.Timestamp < from {
			logCtx.WithFields(log.Fields{"kpi-timestamp": row[datetimeIdx]}).Info("ignoring row as KPI-time not in range, continuing")
			continue
		}

		kpiDetail.TimeString = row[datetimeIdx].(string)

		// add kpi values
		var kpiVals []float64
		for vi := valIdx; vi < len(row); vi++ {
			val := float64(0)
			vInt, okInt := row[vi].(int)
			if !okInt {
				vFloat, okFloat := row[vi].(float64)
				if !okFloat {
					logCtx.WithError(err).WithFields(log.Fields{"value": row[vi]}).Error("couldn't parse the value for KPI query, continuing")
					val = 0.0
				} else {
					val = vFloat
				}
			} else {
				val = float64(vInt)
			}
			kpiVals = append(kpiVals, val)
		}
		kpiDetail.KpiValues = kpiVals

		// add headers
		kpiDetail.KpiHeaderNames = kpiValueHeaders
		// add aggregate function type
		kpiDetail.KpiAggFunctionTypes = kpiAggFunctionType

		if existingDetail, exists := (*kpiData)[kpiID]; exists {
			// for existing kpi detail, add up the values
			for idx, val := range kpiDetail.KpiValues {
				existingDetail.KpiValues[idx] = existingDetail.KpiValues[idx] + val
			}
			(*kpiData)[kpiID] = existingDetail
		} else {
			// map to kpi data to key - final data
			(*kpiData)[kpiID] = kpiDetail
			kpiKeys = append(kpiKeys, kpiID)
		}
	}
	return kpiKeys
}
