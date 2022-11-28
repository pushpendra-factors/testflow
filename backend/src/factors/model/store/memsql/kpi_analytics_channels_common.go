package memsql

import (
	"errors"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) ExecuteKPIQueryForChannels(projectID int64, reqID string, kpiQuery model.KPIQuery) ([]model.QueryResult, int) {
	logCtx := log.WithField("projectId", projectID).WithField("reqId", reqID)
	channelsV1Query, err := model.TransformKPIQueryToChannelsV1Query(kpiQuery)
	queryResults := make([]model.QueryResult, 0)
	hasAnyGroupByTimestamp := (kpiQuery.GroupByTimestamp != "")
	hasAnyGroupBy := len(kpiQuery.GroupBy) != 0
	var queryResult model.QueryResult
	if err != nil {
		logCtx.Warn(err)
		return queryResults, http.StatusBadRequest
	}
	resultHolder, statusCode := store.ExecuteChannelQueryV1(projectID, &channelsV1Query, reqID)
	queryResult.Headers = model.GetTransformedHeadersForChannels(resultHolder.Headers, hasAnyGroupByTimestamp, hasAnyGroupBy)
	queryResult.Rows = model.TransformDateTypeValueForChannels(resultHolder.Headers, resultHolder.Rows, hasAnyGroupByTimestamp, hasAnyGroupBy, channelsV1Query.Timezone)
	if hasAnyGroupByTimestamp {
		err = sanitizeChannelQueryResult(&queryResult, kpiQuery)
		if err != nil {
			logCtx.Warn(err)
			return queryResults, http.StatusBadRequest
		}
	}
	queryResults = append(queryResults, queryResult)
	return queryResults, statusCode
}

//NOTE: The following methods can be merged with methods from event_analytics later on
func sanitizeChannelQueryResult(result *model.QueryResult, query model.KPIQuery) error {
	logFields := log.Fields{
		"query": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	aggrIndex, timeIndex, err := GetTimstampAndAggregateIndexOnQueryResult(result.Headers)
	if err != nil {
		return err
	}
	if len(query.GroupBy) == 0 {
		err = addMissingTimestampsOnChannelResultWithoutGroupByProps(result, &query, aggrIndex, timeIndex, true)
	} else {
		err = addMissingTimestampsOnChannelResultWithGroupByProps(result, &query, aggrIndex, timeIndex, true)
	}

	if err != nil {
		return err
	}

	sortResultRowsByTimestamp(result.Rows, timeIndex)

	return nil
}

func addMissingTimestampsOnChannelResultWithoutGroupByProps(result *model.QueryResult,
	query *model.KPIQuery, aggrIndex int, timestampIndex int, isTimezoneEnabled bool) error {
	logFields := log.Fields{
		"query":               query,
		"aggr_index":          aggrIndex,
		"timestamp_index":     timestampIndex,
		"is_timezone_enabled": isTimezoneEnabled,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	rowsByTimestamp := make(map[string][]interface{}, 0)
	for _, row := range result.Rows {
		sTime := fmt.Sprintf("%v", row[timestampIndex])
		ts, tErr := time.Parse("2006-01-02T15:04:05-07:00", sTime)
		if tErr != nil {
			return tErr
		}
		rowsByTimestamp[U.GetTimestampAsStrWithTimezone(ts, query.Timezone)] = row
	}

	timestamps, offsets := getAllTimestampsAndOffsetBetweenByType(query.From, query.To,
		query.GroupByTimestamp, query.Timezone)

	filledResult := make([][]interface{}, 0, 0)
	// range over timestamps between given from and to.
	// uses timestamp string for comparison.
	for index, ts := range timestamps {

		if row, exists := rowsByTimestamp[U.GetTimestampAsStrWithTimezoneGivenOffset(ts, offsets[index])]; exists {
			// overrides timestamp with user timezone as sql results doesn't
			// return timezone used to query.
			row[timestampIndex] = ts
			filledResult = append(filledResult, row)
		} else {
			newRow := make([]interface{}, 3, 3)
			newRow[timestampIndex] = ts
			newRow[aggrIndex] = 0
			filledResult = append(filledResult, newRow)
		}
	}

	result.Rows = filledResult
	return nil
}

func addMissingTimestampsOnChannelResultWithGroupByProps(result *model.QueryResult,
	query *model.KPIQuery, aggrIndex int, timestampIndex int, isTimezoneEnabled bool) error {
	logFields := log.Fields{
		"query":               query,
		"aggr_index":          aggrIndex,
		"timestamp_index":     timestampIndex,
		"is_timezone_enabled": isTimezoneEnabled,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	gkStart, gkEnd, err := getChannelGroupKeyIndexesForSlicing(result.Headers)
	if err != nil {
		return err
	}

	filledResult := make([][]interface{}, 0, 0)

	rowsByGroupAndTimestamp := make(map[string]bool, 0)
	for _, row := range result.Rows {
		encCols := make([]interface{}, 0, 0)
		encCols = append(encCols, row[gkStart:gkEnd]...)

		sTime := fmt.Sprintf("%v", row[timestampIndex])
		ts, tErr := time.Parse("2006-01-02T15:04:05-07:00", sTime)
		if tErr != nil {
			return tErr
		}
		timestampWithTimezone := U.GetTimestampAsStrWithTimezone(ts, query.Timezone)
		// encoded key with group values and timestamp from db row.
		encCols = append(encCols, timestampWithTimezone)
		encKey := getEncodedKeyForCols(encCols)
		rowsByGroupAndTimestamp[encKey] = true

		// overrides timestamp with user timezone as sql results doesn't
		// return timezone used to query.
		row[timestampIndex] = U.GetTimeFromTimestampStr(timestampWithTimezone)
		filledResult = append(filledResult, row)
	}

	timestamps, offsets := getAllTimestampsAndOffsetBetweenByType(query.From, query.To,
		query.GroupByTimestamp, query.Timezone)

	for _, row := range result.Rows {
		for index, ts := range timestamps {
			encCols := make([]interface{}, 0, 0)
			encCols = append(encCols, row[gkStart:gkEnd]...)
			// encoded key with generated timestamp.
			encCols = append(encCols, U.GetTimestampAsStrWithTimezoneGivenOffset(ts, offsets[index]))
			encKey := getEncodedKeyForCols(encCols)

			_, exists := rowsByGroupAndTimestamp[encKey]
			if !exists {
				// create new row with group values and missing date
				// for those group combination and aggr 0.
				rowLen := len(result.Headers)
				newRow := make([]interface{}, rowLen, rowLen)
				groupValues := row[gkStart:gkEnd]

				for i := 0; i < rowLen; {
					if i == gkStart {
						for _, gv := range groupValues {
							newRow[i] = gv
							i++
						}
					}

					if i == aggrIndex {
						newRow[i] = 0
						i++
					}

					if i == timestampIndex {
						newRow[i] = ts
						i++
					}
				}
				rowsByGroupAndTimestamp[encKey] = true
				filledResult = append(filledResult, newRow)
			}
		}
	}

	result.Rows = filledResult
	return nil
}

func getChannelGroupKeyIndexesForSlicing(cols []string) (int, int, error) {
	logFields := log.Fields{
		"cols": cols,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	start := -1
	end := -1

	index := 0
	for _, col := range cols {
		if strings.HasPrefix(col, "campaign_") || strings.HasPrefix(col, "ad_group_") || strings.HasPrefix(col, "keyword_") || col == "datetime" {
			if start == -1 {
				start = index
			} else {
				end = index
			}
		}
		index++
	}

	// single element.
	if start > -1 && end == -1 {
		end = start
	}

	if start == -1 {
		return start, end, errors.New("no group keys found")
	}

	// end index + 1 reads till end index on slice.
	end = end + 1

	return start, end, nil
}
