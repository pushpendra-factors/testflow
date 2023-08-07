package tests

import (
	"encoding/json"
	C "factors/config"
	"factors/util"
	U "factors/util"
	"fmt"
	"testing"

	"github.com/clearbit/clearbit-go/clearbit"

	"github.com/jinzhu/gorm/dialects/postgres"
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
	to := U.GetBeginningOfDayTimestampIn(U.TimeNowZ().Unix(), U.TimeZoneStringIST) - 1
	from := to - 7*U.SECONDS_IN_A_DAY + 1
	expiry := U.GetDashboardCacheResultExpiryInSeconds(from, to, U.TimeZoneStringIST)
	assert.Equal(t, float64(U.CacheExpiryDashboardMutableDataInSeconds), expiry)

	// Even when from is monthly range, it is mutable since to is yesterday.
	from = to - 30*U.SECONDS_IN_A_DAY + 1
	expiry = U.GetDashboardCacheResultExpiryInSeconds(from, to, U.TimeZoneStringIST)
	assert.Equal(t, float64(U.CacheExpiryDashboardMutableDataInSeconds), expiry)

	// When older than buffer (2) days, expiry should be of weekly range and 0 for monthly range.
	to = to - 2*U.SECONDS_IN_A_DAY
	from = to - 7*U.SECONDS_IN_A_DAY + 1
	expiry = U.GetDashboardCacheResultExpiryInSeconds(from, to, U.TimeZoneStringIST)
	assert.Equal(t, float64(U.CacheExpiryWeeklyRangeInSeconds), expiry)

	from = to - 30*U.SECONDS_IN_A_DAY + 1
	expiry = U.GetDashboardCacheResultExpiryInSeconds(from, to, U.TimeZoneStringIST)
	assert.Equal(t, float64(0), expiry)

	// When to is anytime from today, it should return small Today's expiry.
	from = U.GetBeginningOfDayTimestampIn(U.TimeNowZ().Unix(), U.TimeZoneStringIST)
	to = U.TimeNowIn(U.TimeZoneStringIST).Unix()
	expiry = U.GetDashboardCacheResultExpiryInSeconds(from, to, U.TimeZoneStringIST)
	assert.Equal(t, float64(U.CacheExpiryDashboardTodaysDataInSeconds), expiry)
}

func TestEmailValidations(t *testing.T) {
	assert.Equal(t, true, util.IsEmail("test.o’'brien@test123.com"))
	assert.Equal(t, true, util.IsEmail("test.test1@test123.com-xn"))
	assert.Equal(t, false, util.IsEmail("test.test1@test123.com-xn ping 10.2"))
}

func TestParseProjectIDToStringMapFromConfig(t *testing.T) {
	cMap := C.ParseProjectIDToStringMapFromConfig("1:User Identify1,2:Customer User ID ", "")
	assert.Contains(t, cMap, int64(1))
	assert.Equal(t, "User Identify1", cMap[int64(1)])
	// Test after trimmed space.
	assert.Contains(t, cMap, int64(2))
	assert.Equal(t, "Customer User ID", cMap[int64(2)])
}

func TestCheckOrConvertStandardTimestamp(t *testing.T) {
	type args struct {
		inTimestamp int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{"Test1", args{int64(1635227139001)}, 1635227139},
		{"Test2", args{int64(1635227139002)}, 1635227139},
		{"Test3", args{int64(1635227139999)}, 1635227139},

		{"Test4", args{int64(1635227001)}, 1635227001},
		{"Test5", args{int64(1635227002)}, 1635227002},
		{"Test6", args{int64(1635227999)}, 1635227999},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := U.CheckAndGetStandardTimestamp(tt.args.inTimestamp); got != tt.want {
				t.Errorf("CheckOrConvertStandardTimestamp() = %v, want %v", got, tt.want)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestConvertPostgresJSONBToMap(t *testing.T) {
	name := U.RandomLowerAphaNumString(7)
	email := getRandomEmail()
	sourceJsonb := postgres.Jsonb{[]byte(fmt.Sprintf(`{"name": "%s", "email": "%s"}`, name, email))}
	targetMap, err := U.ConvertPostgresJSONBToMap(sourceJsonb)
	assert.Nil(t, err)
	assert.Equal(t, targetMap["name"], name)
	assert.Equal(t, targetMap["email"], email)
}

func TestClearbit(t *testing.T) {
	client := clearbit.NewClient(clearbit.WithAPIKey("key"))

	results, resp, err := client.Reveal.Find(clearbit.RevealFindParams{
		IP: "104.193.168.24",
	})

	if err == nil {
		fmt.Println(results, resp)
		fmt.Println(results, resp)
	} else {
		fmt.Println(err)
	}
}

func TestTimeZoneConversion(t *testing.T) {

	deltaIST := U.DateAsFormattedInt(U.TimeNowIn(U.TimeZoneString("Asia/Kolkata")))
	deltaPST := U.DateAsFormattedInt(U.TimeNowIn(U.TimeZoneString("America/Vancouver")))
	deltaUTC := U.DateAsFormattedInt(U.TimeNowIn(U.TimeZoneString("UTC")))
	assert.Equal(t, deltaIST > deltaUTC, true)
	assert.Equal(t, deltaPST < deltaUTC, true)

}

func TestSanitizeWeekStart(t *testing.T) {
	type args struct {
		startTime  int64
		zoneString U.TimeZoneString
	}
	args1 := args{1675199123, U.TimeZoneStringIST}
	want1 := int64(1674930600)
	args2 := args{1675199777, U.TimeZoneStringIST}
	want2 := int64(1674930600)
	args3 := args{1675199898, U.TimeZoneStringIST}
	want3 := int64(1674930600)
	args4 := args{1675199100, U.TimeZoneStringIST}
	want4 := int64(1674930600)

	tests := []struct {
		name string
		args args
		want int64
	}{
		{"T1", args1, want1},
		{"T2", args2, want2},
		{"T3", args3, want3},
		{"T4", args4, want4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := U.SanitizeWeekStart(tt.args.startTime, tt.args.zoneString); got != tt.want {
				t.Errorf("SanitizeWeekStart() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRandomIntInRange(t *testing.T) {
	assert.Less(t, U.RandomIntInRange(1, 3), 3)
	assert.Less(t, U.RandomIntInRange(1, 3), 3)
	assert.Less(t, U.RandomIntInRange(1, 3), 3)
	assert.Less(t, U.RandomIntInRange(1, 3), 3)
	assert.Less(t, U.RandomIntInRange(1, 3), 3)
	assert.Less(t, U.RandomIntInRange(1, 3), 3)
	assert.Less(t, U.RandomIntInRange(1, 3), 3)
	assert.Less(t, U.RandomIntInRange(1, 3), 3)
	assert.Less(t, U.RandomIntInRange(1, 3), 3)
	assert.Less(t, U.RandomIntInRange(1, 3), 3)
	assert.Greater(t, U.RandomIntInRange(1, 3), 0)
	assert.Greater(t, U.RandomIntInRange(1, 3), 0)
	assert.Greater(t, U.RandomIntInRange(1, 3), 0)
	assert.Greater(t, U.RandomIntInRange(1, 3), 0)
	assert.Greater(t, U.RandomIntInRange(1, 3), 0)
	assert.Greater(t, U.RandomIntInRange(1, 3), 0)
	assert.Greater(t, U.RandomIntInRange(1, 3), 0)
	assert.Greater(t, U.RandomIntInRange(1, 3), 0)
	assert.Greater(t, U.RandomIntInRange(1, 3), 0)
	assert.Greater(t, U.RandomIntInRange(1, 3), 0)
}

func TestIsIPV4AddressInCIDRRange(t *testing.T) {
	assert.False(t, U.IsIPV4AddressInCIDRRange("40.94.0.0/16", "40.93.255.255")) // previous ip to cidr range start
	assert.True(t, U.IsIPV4AddressInCIDRRange("40.94.0.0/16", "40.94.0.0"))      // start ip in cidr range
	assert.True(t, U.IsIPV4AddressInCIDRRange("40.94.0.0/16", "40.94.255.255"))  //  end ip in cidr range
	assert.False(t, U.IsIPV4AddressInCIDRRange("40.94.0.0/16", "40.95.0.0"))     //  next ip to cidr range end
}

func TestIsAWeeklyRangeQuery(t *testing.T) {
	type args struct {
		timezoneString U.TimeZoneString
		effectiveFrom  int64
		effectiveTo    int64
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "Test1", args: args{
			timezoneString: U.TimeZoneStringIST,
			effectiveFrom:  1689445800,
			effectiveTo:    1690655399,
		}, want: true},
		{name: "Test2", args: args{
			timezoneString: U.TimeZoneStringIST,
			effectiveFrom:  1690655400,
			effectiveTo:    1691260199,
		}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := U.IsAWeeklyRangeQuery(tt.args.timezoneString, tt.args.effectiveFrom, tt.args.effectiveTo)
			if got != tt.want {
				t.Errorf("IsAWeeklyRangeQuery() got = %v, want %v", got, tt.want)
			}

		})
	}
}
