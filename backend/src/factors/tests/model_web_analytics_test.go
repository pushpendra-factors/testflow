package tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	M "factors/model"
	TaskSession "factors/task/session"
	U "factors/util"

	SDK "factors/sdk"
)

func TestExecuteWebAnalyticsQueries(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	timestamp := U.UnixTimeBeforeDuration(time.Minute * 15)

	// Page view event. Should be counted.
	pageURL := "example.com/a"
	trackPayload := SDK.TrackPayload{
		Auto: true,
		Name: pageURL,
		EventProperties: U.PropertiesMap{
			"$page_url":     pageURL,
			"$page_raw_url": pageURL,
			"authorName":    "author1",
		},
		Timestamp: timestamp,
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	assert.NotNil(t, response)
	userId := response.UserId

	// Custom event. Should be counted.
	timestamp = timestamp + 1
	trackPayload = SDK.TrackPayload{
		Name:      U.RandomLowerAphaNumString(5),
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			"$page_url":     pageURL,
			"$page_raw_url": pageURL,
			"authorName":    "author1",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	assert.NotNil(t, response)

	_, err = TaskSession.AddSession([]uint64{project.ID}, timestamp-60, 0, 1)
	assert.Nil(t, err)

	queryResult, errCode := M.ExecuteWebAnalyticsQueries(
		project.ID,
		&M.WebAnalyticsQueries{
			QueryNames: []string{
				M.QueryNameUniqueUsers,
			},
			From: U.UnixTimeBeforeDuration(time.Hour * 1),
			To:   U.TimeNowUnix(),
		},
	)
	assert.NotNil(t, queryResult)
	assert.NotNil(t, queryResult.QueryResult)
	assert.Equal(t, http.StatusOK, errCode)

	unitID := U.GetUUID()
	queryResult, errCode = M.ExecuteWebAnalyticsQueries(
		project.ID,
		&M.WebAnalyticsQueries{
			QueryNames: []string{
				M.QueryNameUniqueUsers,
			},
			CustomGroupQueries: []M.WebAnalyticsCustomGroupQuery{
				M.WebAnalyticsCustomGroupQuery{
					GroupByProperties: []string{
						"authorName",
					},
					Metrics: []string{
						M.WAGroupMetricPageViews,
						M.WAGroupMetricTotalExits,
						M.WAGroupMetricExitPercentage,
					},
					UniqueID: unitID,
				},
			},
			From: U.UnixTimeBeforeDuration(time.Hour * 1),
			To:   U.TimeNowUnix(),
		},
	)
	assert.Equal(t, http.StatusOK, errCode)
	assert.NotNil(t, queryResult)
	assert.True(t, len(queryResult.CustomGroupQueryResult) > 0)

	// Group result should have equal length of
	// headers and individual row.
	for _, groupResult := range queryResult.CustomGroupQueryResult {
		assert.Len(t, groupResult.Headers, 4)
		for _, row := range groupResult.Rows {
			assert.Len(t, row, 4)
		}
	}

	// Group authorName.
	assert.Equal(t, "author1", queryResult.CustomGroupQueryResult[unitID].Rows[0][0])
	// Page views should be 1.
	assert.Equal(t, float64(1), queryResult.CustomGroupQueryResult[unitID].Rows[0][1])
	// Total exits should be 1.
	assert.Equal(t, float64(1), queryResult.CustomGroupQueryResult[unitID].Rows[0][2])
	// Exit percentage should be 100%.
	assert.Equal(t, "100.0%", queryResult.CustomGroupQueryResult[unitID].Rows[0][3])
}

func TestWebAnalyticsCustomGroupSessionBasedMetrics(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	timestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	// Test exit metrics.
	pageURL := "example.com/a"
	trackPayload := SDK.TrackPayload{
		Auto: true,
		Name: pageURL,
		EventProperties: U.PropertiesMap{
			"$page_url":     pageURL,
			"$page_raw_url": pageURL,
			"authorName":    "author1",
		},
		Timestamp: timestamp + 1,
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	assert.NotNil(t, response)

	// Same user and session. Different page and author.
	pageURL1 := "example.com/a/2"
	trackPayload1 := SDK.TrackPayload{
		Auto: true,
		Name: pageURL1,
		EventProperties: U.PropertiesMap{
			"$page_url":     pageURL1,
			"$page_raw_url": pageURL1,
			"authorName":    "author2",
		},
		Timestamp: timestamp + 2,
		UserId:    response.UserId,
	}
	status, response = SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	assert.NotNil(t, response)

	_, err = TaskSession.AddSession([]uint64{project.ID}, timestamp-60, 0, 1)
	assert.Nil(t, err)

	unitID := U.GetUUID()
	queryResult, errCode := M.ExecuteWebAnalyticsQueries(
		project.ID,
		&M.WebAnalyticsQueries{
			QueryNames: []string{
				M.QueryNameUniqueUsers,
			},
			CustomGroupQueries: []M.WebAnalyticsCustomGroupQuery{
				M.WebAnalyticsCustomGroupQuery{
					GroupByProperties: []string{
						"$page_url",
						"authorName",
					},
					Metrics: []string{
						M.WAGroupMetricExitPercentage,
					},
					UniqueID: unitID,
				},
			},
			From: timestamp,
			To:   U.TimeNowUnix(),
		},
	)
	assert.Equal(t, http.StatusOK, errCode)
	assert.NotNil(t, queryResult)
	assert.True(t, len(queryResult.CustomGroupQueryResult) > 0)

	assert.Equal(t, 2, len(queryResult.CustomGroupQueryResult[unitID].Rows))
	// Latest page_url and author of session should only be counted.
	assert.Equal(t, pageURL1, queryResult.CustomGroupQueryResult[unitID].Rows[0][0])  // Group: $page_url.
	assert.Equal(t, "author2", queryResult.CustomGroupQueryResult[unitID].Rows[0][1]) // Group: authorName.
	assert.Equal(t, "100.0%", queryResult.CustomGroupQueryResult[unitID].Rows[0][2])  // Exit percentage.

	assert.Equal(t, pageURL, queryResult.CustomGroupQueryResult[unitID].Rows[1][0])   // Group: $page_url.
	assert.Equal(t, "author1", queryResult.CustomGroupQueryResult[unitID].Rows[1][1]) // Group: authorName.
	assert.Equal(t, "0%", queryResult.CustomGroupQueryResult[unitID].Rows[1][2])      // Exit percentage.

	// Different user and session. Same page and author as track1.
	trackPayload2 := SDK.TrackPayload{
		Auto: true,
		Name: pageURL,
		EventProperties: U.PropertiesMap{
			"$page_url":     pageURL,
			"$page_raw_url": pageURL,
			"authorName":    "author1",
		},
		Timestamp: timestamp + 3,
	}
	status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	assert.NotNil(t, response)

	_, err = TaskSession.AddSession([]uint64{project.ID}, timestamp-60, 0, 1)
	assert.Nil(t, err)

	queryResult, errCode = M.ExecuteWebAnalyticsQueries(
		project.ID,
		&M.WebAnalyticsQueries{
			QueryNames: []string{
				M.QueryNameUniqueUsers,
			},
			CustomGroupQueries: []M.WebAnalyticsCustomGroupQuery{
				M.WebAnalyticsCustomGroupQuery{
					GroupByProperties: []string{
						"$page_url",
						"authorName",
					},
					Metrics: []string{
						M.WAGroupMetricPageViews,
						M.WAGroupMetricExitPercentage,
					},
					UniqueID: unitID,
				},
			},
			From: timestamp,
			To:   U.TimeNowUnix(),
		},
	)
	assert.Equal(t, http.StatusOK, errCode)
	assert.NotNil(t, queryResult)
	assert.True(t, len(queryResult.CustomGroupQueryResult) > 0)

	assert.Equal(t, 2, len(queryResult.CustomGroupQueryResult[unitID].Rows))
	assert.Equal(t, pageURL1, queryResult.CustomGroupQueryResult[unitID].Rows[1][0])  // Group: $page_url.
	assert.Equal(t, "author2", queryResult.CustomGroupQueryResult[unitID].Rows[1][1]) // Group: authorName.
	assert.Equal(t, "50.0%", queryResult.CustomGroupQueryResult[unitID].Rows[1][3])   // Exit percentage.

	assert.Equal(t, pageURL, queryResult.CustomGroupQueryResult[unitID].Rows[0][0])   // Group: $page_url.
	assert.Equal(t, "author1", queryResult.CustomGroupQueryResult[unitID].Rows[0][1]) // Group: authorName.
	assert.Equal(t, "50.0%", queryResult.CustomGroupQueryResult[unitID].Rows[0][3])   // Exit percentage.
}

func TestWebAnalyticsGetFormattedTime(t *testing.T) {
	assert.Equal(t, "1h", M.GetFormattedTime(3600))
	assert.Equal(t, "1m", M.GetFormattedTime(60))
	assert.Equal(t, "1s", M.GetFormattedTime(1))
	assert.Equal(t, "1h 2m", M.GetFormattedTime(3720))
	assert.Equal(t, "1h 2m 1s", M.GetFormattedTime(3721.01))
	// should return only ms, if no seconds available and millseconds available.
	assert.Equal(t, "10ms", M.GetFormattedTime(0.01))
	assert.Equal(t, "950ms", M.GetFormattedTime(0.950))
	// should return 0s if not milliseconds available.
	assert.Equal(t, "0s", M.GetFormattedTime(0.0001))

}
