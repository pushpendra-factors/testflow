package tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gin-gonic/gin"

	H "factors/handler"
	M "factors/model"
	U "factors/util"
)

func TestExecuteWebAnalyticsQueries(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	w := ServePostRequestWithHeaders(r, uri,
		[]byte(`{"event_name": "example.com/a", "event_properties": {"$page_url": "example.com/a", "$page_raw_url": "example.com./a", "authorName": "xxx"}, "user_properties": {"$os": "Mac OSX"}}`),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	resultByQueryName, _, errCode := M.ExecuteWebAnalyticsQueries(
		project.ID,
		&M.WebAnalyticsQueries{
			QueryNames: []string{
				M.QueryNameUniqueUsers,
			},
			From: U.UnixTimeBeforeDuration(time.Hour * 1),
			To:   U.TimeNowUnix(),
		},
	)
	assert.NotNil(t, resultByQueryName)
	assert.Equal(t, http.StatusOK, errCode)

	resultByQueryName, customGroupResult, errCode := M.ExecuteWebAnalyticsQueries(
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
						M.WACustomMetricPageViews,
					},
				},
			},
			From: U.UnixTimeBeforeDuration(time.Hour * 1),
			To:   U.TimeNowUnix(),
		},
	)
	assert.NotNil(t, resultByQueryName)
	assert.Equal(t, http.StatusOK, errCode)
	assert.True(t, len(customGroupResult) > 0)

	// Group result should have equal length of
	// headers and individual row.
	for _, groupResult := range customGroupResult {
		assert.Len(t, groupResult.Headers, 2)
		for _, row := range groupResult.Rows {
			assert.Len(t, row, 2)
		}
	}
}
