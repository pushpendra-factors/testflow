package tests

import (
	"encoding/json"
	U "factors/util"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStringListAsBatch(t *testing.T) {
	batch1 := U.GetStringListAsBatch([]string{"1", "2", "3", "4"}, 2)
	assert.Len(t, batch1, 2)
	assert.Len(t, batch1[0], 2)
	assert.Len(t, batch1[1], 2)

	batch2 := U.GetStringListAsBatch([]string{"1", "2", "3"}, 2)
	assert.Len(t, batch2, 2)
	assert.Len(t, batch2[1], 1)
	assert.Equal(t, "1", batch2[0][0])
	assert.Equal(t, "2", batch2[0][1])
	assert.Equal(t, "3", batch2[1][0])
}

func TestIsStringContainsAny(t *testing.T) {
	contains := U.IsContainsAnySubString("https://facebook.com", "facebook", "twitter")
	assert.True(t, contains)

	contains = U.IsContainsAnySubString("https://twitter.com", "facebook", "twitter")
	assert.True(t, contains)

	contains = U.IsContainsAnySubString("https://tiktok.com", "facebook", "twitter")
	assert.False(t, contains)
}

func TestGenerateHash(t *testing.T) {
	queries := []string{
		`Create a hash for this dummy string`,
		`{"cl": "funnel", "ec": "any_given_event", "fr": 1594492200, "to": 1594578599, "ty": "unique_users", "tz": "Asia/Calcutta", "ewp": [{"na": "$session", "pr": []}, {"na": "www.chargebee.com/schedule-a-demo", "pr": []}], "gbp": [], "gbt": ""}`,
		// Exactly same as above. Only 'to' timestamp is changed.
		`{"cl": "funnel", "ec": "any_given_event", "fr": 1594492200, "to": 1594578598, "ty": "unique_users", "tz": "Asia/Calcutta", "ewp": [{"na": "$session", "pr": []}, {"na": "www.chargebee.com/schedule-a-demo", "pr": []}], "gbp": [], "gbt": ""}`,
		`{"cl": "funnel", "ec": "any_given_event", "fr": 1606588200, "to": 1606927698, "ty": "unique_users", "tz": "Asia/Kolkata", "ewp": [{"na": "www.acme.com/schedule-a-demo", "pr": []}, {"na": "Schedule A Demo Form", "pr": []}, {"na": "Deal Created", "pr": []}], "gbp": [{"en": "user", "pr": "Country", "ena": "$present", "pty": "categorical"}], "gbt": ""}`,
		`{"cl": "funnel", "ec": "any_given_event", "fr": 1603564200, "to": 1604168999, "ty": "events_occurrence", "tz": "", "ewp": [{"na": "connect_button_click", "pr": [{"en": "event", "op": "notContains", "pr": "$page_url", "ty": "categorical", "va": "preprod.yourstory.com", "lop": "AND"}, {"en": "event", "op": "notContains", "pr": "$page_url", "ty": "categorical", "va": "localhost:3005", "lop": "AND"}]}, {"na": "connect_reason_click", "pr": [{"en": "event", "op": "notContains", "pr": "$page_url", "ty": "categorical", "va": "preprod.yourstory.com", "lop": "AND"}, {"en": "event", "op": "notContains", "pr": "$page_url", "ty": "categorical", "va": "localhost:3005", "lop": "AND"}]}, {"na": "get_connected_click", "pr": [{"en": "event", "op": "notContains", "pr": "$page_url", "ty": "categorical", "va": "preprod.yourstory.com", "lop": "AND"}, {"en": "event", "op": "notContains", "pr": "$page_url", "ty": "categorical", "va": "localhost:3005", "lop": "AND"}]}], "gbp": [{"en": "event", "pr": "PageType", "ena": "connect_button_click", "eni": 1, "pty": "categorical"}], "gbt": ""}`,
		`{"cl": "funnel", "ec": "any_given_event", "fr": 1583690263, "to": 1584295063, "ty": "unique_users", "tz": "", "ewp": [{"na": "Customer - Quiz Submitted", "pr": [{"en": "event", "op": "equals", "pr": "question_title", "ty": "categorical", "va": "How can Livspace help with…", "lop": "AND"}, {"en": "event", "op": "equals", "pr": "option_display_value", "ty": "categorical", "va": "Furnishing", "lop": "AND"}, {"en": "event", "op": "equals", "pr": "option_display_value", "ty": "categorical", "va": "Sofas", "lop": "OR"}, {"en": "event", "op": "equals", "pr": "option_display_value", "ty": "categorical", "va": "Beds", "lop": "OR"}, {"en": "event", "op": "equals", "pr": "option_display_value", "ty": "categorical", "va": "Decor items", "lop": "OR"}, {"en": "event", "op": "equals", "pr": "option_display_value", "ty": "categorical", "va": "Seating", "lop": "OR"}]}, {"na": "BC Done Project", "pr": []}], "gbp": [], "gbt": ""}`,
	}

	globalHashValues := make(map[string]bool)
	for i := 0; i < 10; i++ {
		hashValues := make(map[string]bool)
		for index := range queries {
			queryCacheBytes, _ := json.Marshal(queries[index])
			queryHash := U.GenerateHash(queryCacheBytes)
			_, found := hashValues[queryHash]

			// To ensure all hashes are different.
			assert.False(t, found)
			hashValues[queryHash] = true

			if i == 0 {
				globalHashValues[queryHash] = true
			} else {
				// If not first loop, hash values must be available in globalHas.
				// This asserts that values generated are same everytime.
				_, found = globalHashValues[queryHash]
				assert.True(t, found)
			}
		}
	}
}

func TestGetDashboardCacheResultExpiryInSeconds(t *testing.T) {
	// Set 'to' as end of day yesterday, 'from' as 7 days before that. Since 'to' is yesterday, it is mutable range.
	to := U.GetBeginningOfDayTimestampZ(U.TimeNow().Unix(), U.TimeZoneStringIST) - 1
	from := to - 7*U.SECONDS_IN_A_DAY + 1
	expiry := U.GetDashboardCacheResultExpiryInSeconds(from, to)
	assert.Equal(t, float64(U.CacheExpiryDashboardMutableDataInSeconds), expiry)

	// Even when from is monthly range, it is mutable since to is yesterday.
	from = to - 30*U.SECONDS_IN_A_DAY + 1
	expiry = U.GetDashboardCacheResultExpiryInSeconds(from, to)
	assert.Equal(t, float64(U.CacheExpiryDashboardMutableDataInSeconds), expiry)

	// When older than buffer (2) days, expiry should be of weekly range and 0 for monthly range.
	to = to - 2*U.SECONDS_IN_A_DAY
	from = to - 7*U.SECONDS_IN_A_DAY + 1
	expiry = U.GetDashboardCacheResultExpiryInSeconds(from, to)
	assert.Equal(t, float64(U.CacheExpiryWeeklyRangeInSeconds), expiry)

	from = to - 30*U.SECONDS_IN_A_DAY + 1
	expiry = U.GetDashboardCacheResultExpiryInSeconds(from, to)
	assert.Equal(t, float64(0), expiry)

	// When to is anytime from today, it should return small Today's expiry.
	from = U.GetBeginningOfDayTimestampZ(U.TimeNow().Unix(), U.TimeZoneStringIST)
	to = U.TimeNowIn(U.TimeZoneStringIST).Unix()
	expiry = U.GetDashboardCacheResultExpiryInSeconds(from, to)
	assert.Equal(t, float64(U.CacheExpiryTodaysDataInSeconds), expiry)
}
