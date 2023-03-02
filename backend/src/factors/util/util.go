package util

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"factors/filestore"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/mssola/user_agent"
	log "github.com/sirupsen/logrus"
	"github.com/ttacon/libphonenumber"
)

// Int64Tuple To be used at places where an int64 tuple is required to be passed or returned.
type Int64Tuple struct {
	First  int64
	Second int64
}

const (
	DataTypeEvent    = "events"
	DataTypeAdReport = "ad_reports"
	DataTypeUser     = "users"
)

const Per_day_epoch int64 = DayInSecs
const Per_week_epoch int64 = WeekInSecs

// PatternProperties To be used in TakeTopK functions
type PatternProperties interface {
	Get_patternEventNames() []string
	Get_count() uint
	Get_patternType() string
}

const SECONDS_IN_A_DAY int64 = 24 * 60 * 60
const EVENT_USER_CACHE_EXPIRY_SECS = 1728000

// timeoutForClearbitEnrichment
const TimeoutOneSecond = 1 * time.Second
const TimeoutTwoSecond = 2 * time.Second
const TimeoutFiveSecond = 5 * time.Second

const (
	DayInSecs                        = 24 * 60 * 60
	WeekInSecs                       = 7 * DayInSecs
	MonthInSecs                      = 31 * DayInSecs
	Alpha                            = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Numeric                          = "0123456789."
	AllowedDerivedMetricSpecialChars = "*/+-()"
	AllowedDerivedMetricOperator     = "*/+-"
)

type TimeZoneString string

const (
	TimeZoneStringIST TimeZoneString = "Asia/Kolkata"
	TimeZoneStringUTC TimeZoneString = "UTC"

	DateRangePresetLastMonth    = "LAST_MONTH"
	DateRangePresetLastWeek     = "LAST_WEEK"
	DateRangePresetCurrentMonth = "CURRENT_MONTH"
	DateRangePresetCurrentWeek  = "CURRENT_WEEK"
	DateRangePresetYesterday    = "YESTERDAY"
	DateRangePresetToday        = "TODAY"
	DateRangePreset30Minutes    = "30MINS"
	GranularityMonth            = "month"
	GranularityWeek             = "week"
	GranularityDays             = "days"
	GranularityQuarter          = "quarter"
)

// General convention for date Functions - suffix Z if utc based, In if timezone is passed, no suffix if localTime.

// NullcharBytes is null charater bytes, specific for golang json marhsal date
var NullcharBytes = []byte{0x5c, 0x75, 0x30, 0x30, 0x30, 0x30} // \u0000
var spaceBytes = []byte{0x20}

// TODO Handle all errors for location on all functions later.
var QueryDateRangePresets = map[string]func(TimeZoneString) (int64, int64, error){
	DateRangePresetToday:        GetQueryRangePresetTodayIn,
	DateRangePresetYesterday:    GetQueryRangePresetYesterdayIn,
	DateRangePresetCurrentWeek:  GetQueryRangePresetCurrentWeekIn,
	DateRangePresetLastWeek:     GetQueryRangePresetLastWeekIn,
	DateRangePresetCurrentMonth: GetQueryRangePresetCurrentMonthIn,
	DateRangePresetLastMonth:    GetQueryRangePresetLastMonthIn,
}

var PresetLookup = map[string]string{
	"LAST_MONTH":    DateRangePresetLastMonth,
	"LAST_WEEK":     DateRangePresetLastWeek,
	"CURRENT_MONTH": DateRangePresetCurrentMonth,
	"CURRENT_WEEK":  DateRangePresetCurrentWeek,
	"YESTERDAY":     DateRangePresetYesterday,
	"TODAY":         DateRangePresetToday,
}

// WebAnalyticsQueryDateRangePresets Date range presets for website analytics.
var WebAnalyticsQueryDateRangePresets = map[string]func(TimeZoneString) (int64, int64, error){
	DateRangePresetLastMonth:    GetQueryRangePresetLastMonthIn,
	DateRangePresetLastWeek:     GetQueryRangePresetLastWeekIn,
	DateRangePresetCurrentMonth: GetQueryRangePresetCurrentMonthIn,
	DateRangePresetCurrentWeek:  GetQueryRangePresetCurrentWeekIn,
	DateRangePresetYesterday:    GetQueryRangePresetYesterdayIn,
	DateRangePresetToday:        GetQueryRangePresetTodayIn,
	DateRangePreset30Minutes:    GetQueryRangePresetLast30MinutesIn,
}

// Caching related constants for core query and dashboards.
const (
	// ImmutableDataEndDateBufferInSeconds Buffer period after which data is assumed to be immutable. Required to rerun
	// cached queries when add_session / other enrichment jobs were still running or not yet ran when data got cached.
	ImmutableDataEndDateBufferInSeconds      = 2 * SECONDS_IN_A_DAY // 2 Days.
	CacheExpiryDashboardMutableDataInSeconds = 1 * SECONDS_IN_A_DAY // 1 Days.
	CacheExpiryQueryMutableDataInSeconds     = 2 * 60 * 60          // 2 Hours.

	CacheExpiryWeeklyRangeInSeconds = 6 * 7 * SECONDS_IN_A_DAY // 6 Weeks.
	CacheExpiryDefaultInSeconds     = 62 * SECONDS_IN_A_DAY    // 62 Days.

	CacheExpiryQueryMaxInSecondsTwoDays     = 2 * SECONDS_IN_A_DAY
	CacheExpiryQueryMaxInSecondsSevenDays   = 7 * SECONDS_IN_A_DAY
	CacheExpiryQueryTodaysDataInSeconds     = 10 * 60      // 10 minutes.
	CacheExpiryDashboardTodaysDataInSeconds = 12 * 60 * 60 // 12 hours.
	CacheExpiryDashboard30MinutesInSeconds  = 12 * 60 * 60 // 12 hours.
)

// Group Names
var MostRecent string = "Most Recent"
var FrequentlySeen string = "Others"
var SmartEvent string = "Custom Events"

func RandomString(n int) string {
	rand.Seed(time.Now().UnixNano())

	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}
func RandomStringForSharableQuery(n int) string {
	rand.Seed(time.Now().UnixNano())
	timestamp := time.Now().Unix()
	timestampstr := strconv.FormatInt(timestamp, 10)
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	result := string(b)
	length := int32(len(result))
	randIndex := rand.Int31n(length - 1)
	newResult := result[:randIndex] + timestampstr + result[randIndex:]
	return newResult
}

func HashKeyUsingSha256Checksum(data string) string {
	sum := sha256.Sum256([]byte(data))
	encryptData := fmt.Sprintf("%x", sum)
	return encryptData
}

func RandomNumericString(n int) string {
	rand.Seed(time.Now().UnixNano())

	var letter = []rune("0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

func RandomNumericStringNonZeroStart(n int) string {
	rand.Seed(time.Now().UnixNano())

	var letter = []rune("0123456789")
	var letterNonZero = []rune("123456789")

	b := make([]rune, n)
	for i := range b {
		if i == 0 {
			b[i] = letter[rand.Intn(len(letterNonZero))]
		} else {
			b[i] = letter[rand.Intn(len(letter))]
		}
	}
	return string(b)
}

func RandomLowerAphaNumString(n int) string {
	rand.Seed(time.Now().UnixNano())

	var letter = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

// RemoveNullCharacterBytes return bytes by replacing null character bytes by space. Only works for json marshaled data
func RemoveNullCharacterBytes(existingbytes []byte) []byte {
	newBytes := bytes.ReplaceAll(existingbytes, NullcharBytes, spaceBytes)
	return newBytes
}

func RandomUint64() uint64 {
	return rand.Uint64()
}

func RandomInt64() int64 {
	return int64(rand.Uint64())
}

func RandomUint64WithUnixNano() uint64 {
	return uint64(time.Now().UnixNano())
}

// RandomIntInRange Generates a random number in range [min, max).
func RandomIntInRange(min, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(max-min)
}

func UnixTimeBeforeAWeek() int64 {
	return UnixTimeBeforeDuration(168 * time.Hour) // 7 days.
}

func UnixTimeBeforeDuration(duration time.Duration) int64 {
	return time.Now().UTC().Unix() - int64(duration.Seconds())
}

func IsNumber(num string) bool {
	// Use regex.
	_, err := strconv.ParseFloat(num, 64)
	if err != nil {
		return false
	}
	return true
}

func IsEmail(str string) bool {
	regexpEmail := regexp.MustCompile(`^[A-Za-z0-9._%+\-'’^!]+@[A-Za-z0-9.\-]+\.[A-Za-z-]{2,20}$`)
	return regexpEmail.MatchString(str)
}

func IsBetterEmail(str string, str1 string) bool {
	return len(str1) > len(str)
}

func IsValidPhone(str string) bool {
	numbers := regexp.MustCompile("\\d").FindAllString(str, -1)
	if len(numbers) < 5 {
		return false
	}

	return true
}

func IsBetterPhone(str string, str1 string) bool {
	return len(str1) > len(str)
}

func IsPersonalEmail(str string) bool {
	str = strings.ToLower(str)
	personalDomains := []string{
		"gmail.com",
		"yahoo.com",
		"hotmail.com",
		"yahoo.co.in",
		"hey.com",
		"icloud.com",
		"me.com",
		"mac.com",
		"aol.com",
		"abc.com",
		"xyz.com",
		"pqr.com",
		"rediffmail.com",
		"live.com",
		"outlook.com",
		"msn.com",
		"ymail.com",
	}
	for _, domain := range personalDomains {
		if strings.Contains(str, domain) {
			return true
		}
	}
	return false
}

func IsValidUrl(tocheck string) bool {

	r, _ := regexp.Compile("(^(http|https)://)?(www)?.com")
	if strings.HasPrefix(tocheck, "$session") {
		return false
	}
	return r.MatchString(tocheck)
	// return true
}

// GetEmailLowerCase returns email in lower case for consistency
func GetEmailLowerCase(emailI interface{}) string {
	email := GetPropertyValueAsString(emailI)
	if !IsEmail(email) {
		return ""
	}

	return strings.ToLower(email)
}

func GetNumberFromAnyString(str string) float64 {
	strAsBytes := []byte(str)
	re := regexp.MustCompile(`[+-]?([0-9]*[.])?[0-9]+`)
	numStr := string(re.Find(strAsBytes))

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0
	}

	return num
}

func GetSortWeightFromAnyType(value interface{}) float64 {
	if value == nil {
		return 0
	}

	switch valueType := value.(type) {
	case float64:
		return value.(float64)
	case float32:
		return float64(value.(float32))
	case int:
		return float64(value.(int))
	case int32:
		return float64(value.(int32))
	case int64:
		return float64(value.(int64))
	case string:
		return GetNumberFromAnyString(value.(string))
	default:
		log.Info("Unsupported type used on GetSortWeightFromAnyType %+v", valueType)
		return 0
	}
}

// SafeConvertToFloat64 Converts an interface to float64 value.
func SafeConvertToFloat64(value interface{}) float64 {
	return GetSortWeightFromAnyType(value)
}

func TrimQuotes(str string) string {
	return strings.TrimSuffix(strings.TrimPrefix(str, "\""), "\"")
}

func GetValueAsString(value interface{}) (string, error) {
	switch value.(type) {
	case float32, float64:
		return fmt.Sprintf("%0.0f", value), nil
	case int, int32, int64:
		return fmt.Sprintf("%v", value), nil
	case string:
		return value.(string), nil
	case bool:
		return strconv.FormatBool(value.(bool)), nil
	default:
		return "", errors.New("invalid type to convert as string")
	}
}

func GetUUID() string {
	return uuid.New().String()
}

func IsValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

func ConvertIntToUUID(value uint64) (string, error) {
	if value == 0 {
		return "", errors.New("invalid integer for conversion")
	}

	valueAsString := fmt.Sprintf("%v", value)
	return ConvertIntStringToUUID(valueAsString)
}

func ConvertIntStringToUUID(intAsString string) (string, error) {
	valueLen := len(intAsString)
	if valueLen > 12 {
		return "", errors.New("unsupported integer of 12 digit")
	}

	lastOctet := "000000000000"
	lastOctet = lastOctet[:len(lastOctet)-valueLen] + intAsString

	return fmt.Sprintf("00000000-0000-0000-0000-%s", lastOctet), nil
}

// StringSliceDiff Returns sliceA - sliceB set of elements.
func StringSliceDiff(sliceA, sliceB []string) []string {
	if len(sliceA) == 0 || len(sliceB) == 0 {
		return sliceA
	}
	sliceBMap := make(map[string]int)

	var diffSlice []string
	for index, value := range sliceB {
		sliceBMap[value] = index
	}

	for _, value := range sliceA {
		_, found := sliceBMap[value]
		if !found {
			diffSlice = append(diffSlice, value)
		}
	}
	return diffSlice
}

// IsNonEmptyKey returns true if key is anything but not empty ("") or '$none'
func IsNonEmptyKey(key string) bool {

	if key != "" && key != "$none" {
		return true
	}
	return false
}

// IfThenElse Is a hack to get ternary one liners
func IfThenElse(condition bool, a interface{}, b interface{}) interface{} {
	if condition {
		return a
	}
	return b
}

// StringValueIn Returns true if `value` is in `list` else false.
func StringValueIn(value string, list []string) bool {
	for _, val := range list {
		if val == value {
			return true
		}
	}
	return false
}

// Uint64ValueIn Returns true if `value` is in `list` else false.
func Uint64ValueIn(value uint64, list []uint64) bool {
	for _, val := range list {
		if val == value {
			return true
		}
	}
	return false
}

// Uint64ValueIn Returns true if `value` is in `list` else false.
func Int64ValueIn(value int64, list []int64) bool {
	for _, val := range list {
		if val == value {
			return true
		}
	}
	return false
}

// FloatRoundOffWithPrecision Rounds of a float64 value to given precision. Ex: 2.667 with precision 2 -> 2.67.
func FloatRoundOffWithPrecision(value float64, precision int) (float64, error) {
	valueString := fmt.Sprintf("%0.*f", precision, value)
	roundOffValue, err := strconv.ParseFloat(valueString, 64)
	if err != nil {
		log.WithFields(log.Fields{"value": value,
			"precision": precision}).Error("error while rounding off float value")
		return roundOffValue, err
	}
	return roundOffValue, nil
}

// UnixToHumanTimeZ Converts epoch to readable String timestamp.
func UnixToHumanTimeZ(timestamp int64) string {
	return time.Unix(timestamp, 0).UTC().Format(time.RFC3339)
}

// SecondsToHMSString Converts seconds int value to hrs min secs string.
func SecondsToHMSString(totalSeconds int64) string {
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%d hrs %d mins %d secs", hours, minutes, seconds)
}

func Min(a int64, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func Max(a int64, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func MinInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func MaxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func MaxFloat64(a float64, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// returns a string of ('?'), ... values of given batchSize
func GetValuePlaceHolder(batchSize int) string {

	concatenatedIds := ""
	for i := 0; i < batchSize; i++ {
		concatenatedIds = concatenatedIds + " (?),"
	}
	// removing that extra ',' at end
	if len(concatenatedIds) > 0 {
		concatenatedIds = concatenatedIds[0 : len(concatenatedIds)-1]
	} else {
		log.WithFields(log.Fields{}).Info("batchSize is zero in GetValuePlaceHolder")
	}
	return concatenatedIds
}

// returns a interface list from given string list
func GetInterfaceList(list []string) []interface{} {

	listInterface := make([]interface{}, len(list))
	for i, v := range list {
		listInterface[i] = v
	}
	return listInterface
}

// GetStringListAsBatch - Splits string list into multiple lists.
func GetStringListAsBatch(list []string, batchSize int) [][]string {
	batchList := make([][]string, 0, 0)
	listLen := len(list)
	for i := 0; i < listLen; {
		next := i + batchSize
		if next > listLen {
			next = listLen
		}

		batchList = append(batchList, list[i:next])
		i = next
	}

	return batchList
}

// GetUint64ListAsBatch - Returns list of uint64 as batches of uint64 list.
func GetInt64ListAsBatch(list []int64, batchSize int) [][]int64 {
	batchList := make([][]int64, 0, 0)
	listLen := len(list)
	for i := 0; i < listLen; {
		next := i + batchSize
		if next > listLen {
			next = listLen
		}

		batchList = append(batchList, list[i:next])
		i = next
	}

	return batchList
}

func GetInterfaceListAsBatch(list []interface{}, batchSize int) [][]interface{} {
	batchList := make([][]interface{}, 0, 0)
	listLen := len(list)
	for i := 0; i < listLen; {
		next := i + batchSize
		if next > listLen {
			next = listLen
		}

		batchList = append(batchList, list[i:next])
		i = next
	}

	return batchList
}

func GetKeysMapAsArray(keys map[string]bool) []string {
	keysArray := make([]string, 0)
	for key := range keys {
		keysArray = append(keysArray, key)
	}

	return keysArray
}

func GetKeysOfInt64StringMap(m *map[int64]string) []int64 {
	if m == nil {
		return []int64{}
	}

	keys := make([]int64, 0)
	for key := range *m {
		keys = append(keys, key)
	}

	return keys
}

func GetSnakeCaseToTitleString(str string) (title string) {
	if str == "" {
		return
	}

	tokens := strings.Split(str, "_")
	for i, token := range tokens {
		if i == 0 {
			title = strings.Title(token)
			continue
		}

		title = fmt.Sprintf("%s %s", title, strings.Title(token))
	}

	return title
}

func IsContainsAnySubString(src string, sub ...string) bool {
	for _, s := range sub {
		if strings.Contains(src, s) {
			return true
		}
	}

	return false
}

// isPureisPurePhoneNumberNumber checks for pure number string
func isPurePhoneNumber(phoneNo string) bool {
	if _, err := strconv.Atoi(phoneNo); err == nil {
		return true
	}
	return false
}

func getPhoneNoSeparatorIndex(phoneNo *string) []int {
	var separatorsIndex []int
	separators := []string{" ", "-"}
	for _, seperator := range separators {
		if spIndex := strings.Index(*phoneNo, seperator); spIndex != -1 {
			separatorsIndex = append(separatorsIndex, spIndex)
		}
	}
	return separatorsIndex
}

// isPhoneValidCountryCode checks if country code exist in libphonenumber list
func isPhoneValidCountryCode(cCode string) bool {
	if code, err := strconv.Atoi(cCode); err == nil {
		if _, exist := libphonenumber.CountryCodeToRegion[code]; exist {
			return true
		}
	}
	return false
}

// maybeAddInternationalPrefix adds if missing '+' at the beginning for the libphonenumber to work
func maybeAddInternationalPrefix(phoneNo *string) bool {
	// Ex 91 1234567890
	separatorsIndex := getPhoneNoSeparatorIndex(phoneNo)

	for _, indexValue := range separatorsIndex {
		cCode := string((*phoneNo)[:indexValue])
		if isPhoneValidCountryCode(cCode) {
			*phoneNo = libphonenumber.PLUS_CHARS + *phoneNo
			return true
		}
	}
	return false
}

// getInternationalPhoneNoWithoutCountryCode removes country code
func getInternationalPhoneNoWithoutCountryCode(intPhone string) string {
	phoneNo := strings.SplitN(intPhone, " ", 2)
	return strings.Join(phoneNo[1:], "")
}

// GetPossiblePhoneNumber returns all possible phone number for specific phone number pattern
func GetPossiblePhoneNumber(phoneNo string) []string {
	var possiblePhoneNo []string
	var phonePattern string
	possiblePhoneNo = append(possiblePhoneNo, phoneNo)

	// try pure numbers if form submited also had pure numbers
	if isPurePhoneNumber(phoneNo) {
		if !strings.Contains(phoneNo, "+") {
			possiblePhoneNo = append(possiblePhoneNo, "0"+phoneNo)
			possiblePhoneNo = append(possiblePhoneNo, "+"+phoneNo)
		}

		//0123-456-789 or (012)-345-6789
		nationalPhoneNo := strings.TrimPrefix(phoneNo, "+")
		if len(nationalPhoneNo) == 10 {
			phonePattern = fmt.Sprintf("%s-%s-%s", nationalPhoneNo[:3], nationalPhoneNo[3:6], nationalPhoneNo[6:])
			possiblePhoneNo = append(possiblePhoneNo, phonePattern)
			phonePattern = fmt.Sprintf("(%s)-%s-%s", nationalPhoneNo[:3], nationalPhoneNo[3:6], nationalPhoneNo[6:])
			possiblePhoneNo = append(possiblePhoneNo, phonePattern)
			phonePattern = fmt.Sprintf("(%s) %s %s", nationalPhoneNo[:3], nationalPhoneNo[3:6], nationalPhoneNo[6:])
			possiblePhoneNo = append(possiblePhoneNo, phonePattern)
		}
	}

	//phone number having '+' will have country code attached
	num, err := libphonenumber.Parse(phoneNo, "")
	if err != nil {
		// ErrInvalidCountryCode describes missing '+', can be added if phone number in format
		if err == libphonenumber.ErrInvalidCountryCode && maybeAddInternationalPrefix(&phoneNo) {
			num, err = libphonenumber.Parse(phoneNo, "")
		}
	}

	if err == nil {

		//international format +91 1234 567 890
		intFormat := libphonenumber.Format(num, libphonenumber.INTERNATIONAL)
		if intFormat != phoneNo {
			possiblePhoneNo = append(possiblePhoneNo, intFormat)
		}

		//International without country code 1234 567 890
		possiblePhoneNo = append(possiblePhoneNo, getInternationalPhoneNoWithoutCountryCode(intFormat))
		possiblePhoneNo = append(possiblePhoneNo, libphonenumber.Format(num, libphonenumber.NATIONAL))

		//911234567890
		nationalFormat := libphonenumber.Format(num, libphonenumber.E164)[1:]
		possiblePhoneNo = append(possiblePhoneNo, nationalFormat)

		//National format 1234 567 890
		nationalNum := libphonenumber.GetNationalSignificantNumber(num)
		possiblePhoneNo = append(possiblePhoneNo, nationalNum)
		possiblePhoneNo = append(possiblePhoneNo, "0"+nationalNum)
		possiblePhoneNo = append(possiblePhoneNo, "+"+nationalNum)

		standardPhone := libphonenumber.Format(num, libphonenumber.E164)
		if standardPhone != "+91"+nationalNum {
			// standard phone number +911234567890. Assuming input phone number be +91 123 456 7890 or similar
			possiblePhoneNo = append(possiblePhoneNo, standardPhone)
		}

		//+9101234 567 890
		possiblePhoneNo = append(possiblePhoneNo, fmt.Sprintf("+%d%s", num.GetCountryCode(), libphonenumber.Format(num, libphonenumber.NATIONAL)))

	}

	return possiblePhoneNo
}

// BytesToReadableFormat Pretty prints bytes to readable KiB/MiB/GiB.. format.
func BytesToReadableFormat(bytes float64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%.1f B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(bytes)/float64(div), "KMGTPE"[exp])
}

// SanitizePhoneNumber currently removes only leading 0 from phone number. New functionalty will added as when required
func SanitizePhoneNumber(v interface{}) string {
	phoneNo := GetPropertyValueAsString(v)

	if phoneNo == "" || len(phoneNo) < 4 {
		return ""
	}

	if phoneNo[0] == '0' {
		return phoneNo[1:]
	}

	return phoneNo
}

func InsertRepeated(strList []string, index int, value string) []string {
	if len(strList) == index { // nil or empty slice or after last element
		return append(strList, value)
	}
	strList = append(strList[:index+1], strList[index:]...) // index < len(a)
	strList[index] = value
	return strList
}

func RankByWordCount(wordFrequencies map[string]int) PairList {
	pl := make(PairList, len(wordFrequencies))
	i := 0
	for k, v := range wordFrequencies {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

// GenerateHash To generate hash value for given byte array.
func GenerateHash(bytes []byte) string {
	hasher := sha1.New()
	hasher.Write(bytes)
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return sha
}

// GenerateHashStringForStruct Marshals the passed struct and generates a unique hash string.
func GenerateHashStringForStruct(queryPayload interface{}) (string, error) {
	queryCacheBytes, err := json.Marshal(queryPayload)
	if err != nil {
		return "", err
	}
	return GenerateHash(queryCacheBytes), nil
}

// DeepCopy Deep copies a to b using json marshaling.
func DeepCopy(a, b interface{}) {
	byt, _ := json.Marshal(a)
	json.Unmarshal(byt, b)
}

// type Pair struct {
// 	Key   string
// 	Value int
// }

// type PairList []Pair

// func (p PairList) Len() int           { return len(p) }
// func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
// func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// IsStandardEvent check if eventname is standard Event
func IsStandardEvent(eventName string) bool {
	return strings.HasPrefix(eventName, "$")
}

// IsCampaignEvent check if eventname is campaign Event
func IsCampaignEvent(eventName string) bool {
	return strings.HasPrefix(eventName, "$session[campaign")
}

// IsSourceEvent check if eventname is campaign Event
func IsSourceEvent(eventName string) bool {
	return strings.HasPrefix(eventName, "$session[source")
}

// IsMediumEvent check if eventname is campaign Event
func IsMediumEvent(eventName string) bool {
	return strings.HasPrefix(eventName, "$session[medium")
}

// IsReferrerEvent check if eventname is referrer Event
func IsReferrerEvent(eventName string) bool {
	return strings.HasPrefix(eventName, "$session[initial_referrer")
}

func IsAdgroupEvent(eventName string) bool {
	return strings.HasPrefix(eventName, "$session[adgroup")
}

func IsCampaignAnalytics(eventName string) bool {
	if IsCampaignEvent(eventName) == true || IsMediumEvent(eventName) == true || IsSourceEvent(eventName) == true || IsReferrerEvent(eventName) == true || IsAdgroupEvent(eventName) == true {
		return true
	}
	return false
}

func IsItreeCampaignEvent(eventName string) bool {
	return strings.HasPrefix(eventName, "$session[campaign") || strings.HasPrefix(eventName, "$session[source") || strings.HasPrefix(eventName, "$session[medium") || strings.HasPrefix(eventName, "$session[adgroup") || strings.HasPrefix(eventName, "$session[initial_referrer")
}

// GetDashboardCacheResultExpiryInSeconds Returns expiry for cache based on query date range.
func GetDashboardCacheResultExpiryInSeconds(from, to int64, timezoneString TimeZoneString) float64 {
	toStartOfDay := GetBeginningOfDayTimestampIn(to, timezoneString)
	nowStartOfDay := GetBeginningOfDayTimestampIn(TimeNowZ().Unix(), timezoneString)
	if Is30MinutesTimeRange(from, to) {
		return float64(CacheExpiryDashboard30MinutesInSeconds)
	} else if to >= nowStartOfDay {
		// End date is in today's range. Keep small expiry.
		return float64(CacheExpiryDashboardTodaysDataInSeconds)
	} else if nowStartOfDay > toStartOfDay && nowStartOfDay-toStartOfDay > ImmutableDataEndDateBufferInSeconds {
		// Data can be assumed to be immutable here after buffer (2) days.
		if to-from == (7*SECONDS_IN_A_DAY - 1) {
			// Weekly range.
			return float64(CacheExpiryWeeklyRangeInSeconds)
		} else if to-from > (27*SECONDS_IN_A_DAY - 1) {
			// Monthly range. Set no expiry.
			return 0
		} else {
			return float64(CacheExpiryDefaultInSeconds)
		}
	}
	return float64(CacheExpiryDashboardMutableDataInSeconds)
}

// GetQueryCacheResultExpiryInSeconds Returns expiry for cache based on query date range.
func GetQueryCacheResultExpiryInSeconds(from, to int64, timezoneString TimeZoneString) float64 {
	toStartOfDay := GetBeginningOfDayTimestampIn(to, timezoneString)
	nowStartOfDay := GetBeginningOfDayTimestampIn(TimeNowZ().Unix(), timezoneString)
	if to >= nowStartOfDay {
		// End date is in today's range. Keep small expiry.
		return float64(CacheExpiryQueryTodaysDataInSeconds)
	} else if nowStartOfDay > toStartOfDay && nowStartOfDay-toStartOfDay > ImmutableDataEndDateBufferInSeconds {
		// Data can be assumed to be immutable here after buffer (2) days.
		return float64(CacheExpiryQueryMaxInSecondsTwoDays)
	}
	return float64(CacheExpiryQueryMaxInSecondsTwoDays)
}

func GetAggrAsFloat64(aggr interface{}) (float64, error) {
	switch aggr.(type) {
	case int:
		return float64(aggr.(int)), nil
	case int64:
		return float64(aggr.(int64)), nil
	case float32:
		return float64(aggr.(float32)), nil
	case float64:
		return aggr.(float64), nil
	case string:
		aggrInt, err := strconv.ParseInt(aggr.(string), 10, 64)
		return float64(aggrInt), err
	default:
		return float64(0), errors.New("invalid aggregate value type")
	}
}

// CleanSplitByDelimiter Splits a string by delimiter and removes any spaces.
// Ex: "a, b, c" and "a,b,c" will return same ["a", "b", "c"].
func CleanSplitByDelimiter(str string, del string) []string {
	split := strings.Split(str, del)

	cleanSplit := make([]string, 0, 0)
	for _, s := range split {
		cleanSplit = append(cleanSplit, strings.TrimSpace(s))
	}
	return cleanSplit
}

// StringIn Check if string is in given list of strings
func StringIn(strList []string, s string) (bool, string, int) {

	if len(strList) == 0 {
		return false, "", -1
	}

	for idx, v := range strList {

		if strings.Compare(v, s) == 0 {
			return true, s, idx
		}
	}

	return false, "", -1

}

func FloatIn(strList []string, s float64) (bool, float64, int) {

	if len(strList) == 0 {
		return false, 0, -1
	}

	for idx, v := range strList {

		if val, err := strconv.ParseFloat(v, 64); err == nil {
			if val == s {
				return true, s, idx
			}
		}

	}

	return false, 0, -1

}

// Remove  remove string from given index
func Remove(s []string, i int) ([]string, error) {
	if len(s) > 0 {
		s[len(s)-1], s[i] = s[i], s[len(s)-1]
		return s[:len(s)-1], nil
	} else {
		return nil, fmt.Errorf("len of string slice is 0 : %v", s)
	}
}

func TakeTopKUC(allPatterns []PatternProperties, topK int, ptype string) []PatternProperties {

	allPatternsType := make([]PatternProperties, 0)
	for _, pattern := range allPatterns {

		if pattern.Get_patternType() == ptype {
			allPatternsType = append(allPatternsType, pattern)
		}
	}

	if len(allPatternsType) > 0 {
		return TakeTopK(allPatternsType, topK)
	}
	return allPatternsType

}

func TakeTopKpageView(allPatterns []PatternProperties, topK int, ptype1, ptype2 string) []PatternProperties {

	allPatternsType := make([]PatternProperties, 0)
	for _, pattern := range allPatterns {

		if pattern.Get_patternType() == ptype1 || pattern.Get_patternType() == ptype2 {
			allPatternsType = append(allPatternsType, pattern)
		}
	}
	if len(allPatternsType) > 0 {
		return TakeTopK(allPatternsType, topK)
	}
	return allPatternsType

}

func TakeTopKIE(allPatterns []PatternProperties, topK int, ptype string) []PatternProperties {

	allPatternsType := make([]PatternProperties, 0)
	for _, pattern := range allPatterns {

		if pattern.Get_patternType() == ptype {
			allPatternsType = append(allPatternsType, pattern)
		}
	}
	if len(allPatternsType) > 0 {
		return TakeTopK(allPatternsType, topK)
	}
	return allPatternsType

}

func TakeTopKspecialEvents(allPatterns []PatternProperties, topK int) []PatternProperties {

	allPatternsType := make([]PatternProperties, 0)
	for _, pt := range allPatterns {
		ename := pt.Get_patternEventNames()[0]
		if IsStandardEvent(ename) == true && IsCampaignAnalytics(ename) == false {
			allPatternsType = append(allPatternsType, pt)
		}
	}
	if len(allPatternsType) > 0 {
		return TakeTopK(allPatternsType, topK)
	}
	return allPatternsType

}

func TakeTopKAllURL(allPatterns []PatternProperties, topK int) []PatternProperties {

	allPatternsType := make([]PatternProperties, 0)
	for _, pt := range allPatterns {

		if IsValidUrl(pt.Get_patternEventNames()[0]) == true {
			allPatternsType = append(allPatternsType, pt)
		}
	}
	if len(allPatternsType) > 0 {
		return TakeTopK(allPatternsType, topK)
	}
	return allPatternsType

}

func TakeTopK(patterns []PatternProperties, topKPatterns int) []PatternProperties {
	// rewrite with heap. can hog the memory
	if len(patterns) > 0 {
		sort.Slice(patterns, func(i, j int) bool { return patterns[i].Get_count() > patterns[j].Get_count() })
		if len(patterns) > topKPatterns {
			return patterns[0:topKPatterns]
		}
		return patterns

	}
	return patterns
}

func GetFilteredMapBySkipList(sourceMap *map[string]interface{}, propertySkipList []string) *map[string]interface{} {
	if sourceMap == nil {
		return nil
	}

	skipListMap := make(map[string]bool)
	for i := range propertySkipList {
		skipListMap[propertySkipList[i]] = true
	}

	filteredMap := make(map[string]interface{})
	for property := range *sourceMap {
		if skipListMap[property] {
			continue
		}

		filteredMap[property] = (*sourceMap)[property]
	}

	return &filteredMap
}

func FilterOnSupport(itms map[string]int, support float32) map[string]int {
	sumVal := 0
	for _, v := range itms {
		sumVal += v
	}
	filterVal := int(math.Ceil((float64(support) * float64(sumVal) / 100)))

	for k, v := range itms {
		if v < filterVal {
			delete(itms, k)

		}
	}
	return itms
}

func FilterOnFrequency(itms map[string]int, topk int) {
	ll := make([]string, 0)
	for k := range itms {
		ll = append(ll, k)
	}
	sort.Strings(ll)
	pl := SortOnPriority(ll, itms, false)
	// pl := RankByWordCount(itms)
	for idx, v := range pl {
		if idx >= topk {
			delete(itms, v)
		}
	}
}

func FilterOnFrequencySpl(itms map[string]int, topk int) {
	ll := make([]string, 0)
	for k := range itms {
		ll = append(ll, k)
	}
	sort.Strings(ll)
	pl := SortOnPriority(ll, itms, false)
	// pl := RankByWordCount(itms)
	for idx, v := range pl {
		if idx >= topk {
			if _, ok := itms["$others"]; !ok {
				itms["$others"] = 0
			}
			itms["$others"] += itms[v]
			delete(itms, v)
		}
	}
}

func IsDateTime(val string) bool {
	const layout = "2006-01-02T15:04:05-0700"
	_, err := time.Parse(layout, val)
	if err != nil {
		return false
	}
	return true
}
func GetNumberOfDigits(input int64) int {
	if input == 0 {
		return 1
	}
	count := 0
	for input > 0 {
		input = input / 10
		count = count + 1
	}
	return count
}

func CheckAndGetStandardTimestamp(inTimestamp int64) int64 {
	if inTimestamp > int64(10000000000) {
		return int64(inTimestamp / int64(1000))
	}
	return inTimestamp
}

func ConvertPostgresJSONBToMap(sourceJsonb postgres.Jsonb) (map[string]interface{}, error) {
	var targetMap map[string]interface{}
	byteArray, err := json.Marshal(sourceJsonb)
	if err != nil {
		log.WithError(err).Error("Failed to marshal JSONB object to byte array inside ConvertPostgresJSONBToMap.")
		return nil, err
	}

	err = json.Unmarshal(byteArray, &targetMap)
	if err != nil {
		log.WithError(err).Error("Failed to unmarshal user properties inside ConvertPostgresJSONBToMap.")
		return nil, err
	}
	return targetMap, nil
}

func GetStartOfDateEpochInOtherTimezone(value int64, currentTimezone, nextTimezone string) int64 {
	currentLocation, _ := time.LoadLocation(currentTimezone)
	nextLocation, _ := time.LoadLocation(nextTimezone)
	dateTimeInCurrentTimezone := time.Unix(value, 0).In(currentLocation)
	beginningOfDateInNextTimezone := time.Date(dateTimeInCurrentTimezone.Year(), dateTimeInCurrentTimezone.Month(), dateTimeInCurrentTimezone.Day(), 0, 0, 0, 0, nextLocation).Unix()
	return beginningOfDateInNextTimezone
}

func GetEndOfDateEpochInOtherTimezone(value int64, currentTimezone, nextTimezone string) int64 {
	currentLocation, _ := time.LoadLocation(currentTimezone)
	nextLocation, _ := time.LoadLocation(nextTimezone)
	dateTimeInCurrentTimezone := time.Unix(value, 0).In(currentLocation)
	beginningOfDateInNextTimezone := time.Date(dateTimeInCurrentTimezone.Year(), dateTimeInCurrentTimezone.Month(), dateTimeInCurrentTimezone.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), nextLocation).Unix()
	return beginningOfDateInNextTimezone
}

// Most of the rows will have groupBy keys and/or timestamp followed by a single value pattern.
// To get Keys - we have taken (groupBy keys and/or timestamp values) as hashKey and metric value as hashValue against it.
func GetkeyFromRow(row []interface{}) string {
	if len(row) <= 1 {
		return "1"
	}
	var key string
	for _, value := range row[:len(row)-1] {
		if valueTime, ok := (value.(time.Time)); ok {
			valueInUnix := valueTime.Unix()
			key = key + fmt.Sprintf("dat$%v:;", valueInUnix)
		} else {
			key = key + fmt.Sprintf("%v", value) + ":;"
		}
	}

	return key
}

func RemoveElementFromArray(inputArray []string, value string) []string {
	finalArray := make([]string, 0)
	for _, ele := range inputArray {
		if value != ele {
			finalArray = append(finalArray, ele)
		}
	}
	return finalArray
}

func SafeAddition(val1 interface{}, val2 interface{}) (interface{}, error) {
	if reflect.TypeOf(val1) != reflect.TypeOf(val1) {
		return float64(0), errors.New("Wrong types passed.")
	}
	switch val1.(type) {
	case float64:
		return val1.(float64) + val2.(float64), nil
	case float32:
		return val1.(float32) + val2.(float32), nil
	case int:
		return val1.(int) + val2.(int), nil
	case int32:
		return val1.(int32) + val2.(int32), nil
	case int64:
		return val1.(int64) + val2.(int64), nil
	}
	log.WithField("val1", val1).WithField("val2", val2).Error("Default value is taken for val1, val2.")
	return float64(0), errors.New("Default value is taken for val1, val2.")
}

// Sorted the rows on basis of initial values to the final.
// Considering that we have group keys and/or timestamp as initial keys, we might want to get them ordered by groupByKeys first.
func GetSorted2DArrays(rows [][]interface{}) [][]interface{} {
	sort.Slice(rows, func(i, j int) bool {
		for x := range rows[i] {
			if rows[i][x] == rows[j][x] {
				continue
			}
			switch valueType := rows[i][x].(type) {
			case float64:
				return rows[i][x].(float64) < rows[j][x].(float64)
			case float32:
				return rows[i][x].(float32) < rows[j][x].(float32)
			case int:
				return rows[i][x].(int) < rows[j][x].(int)
			case int32:
				return rows[i][x].(int32) < rows[j][x].(int32)
			case int64:
				return rows[i][x].(int64) < rows[j][x].(int64)
			case string:
				return rows[i][x].(string) < rows[j][x].(string)
			case time.Time:
				return rows[i][x].(time.Time).Before(rows[j][x].(time.Time))
			default:
				log.Info("Unsupported type used on sorting %+v", valueType)
				return true
			}
		}
		return false
	})
	return rows
}

func HasPrefixFromList(propKey string, prefixList []string) bool {
	for _, prefix := range prefixList {
		if strings.HasPrefix(propKey, prefix) {
			return true
		}
	}
	return false
}

func HasMaliciousContent(reqPayload string) (bool, error) {
	lcasePayload := strings.ToLower(reqPayload)

	// Add a better SQL matching logic. Below will block all payloads with select, delete etc.,
	hasSQLStatement, _ := regexp.MatchString(
		"(?i)((SELECT|DELETE)\\s+.+\\s+FROM\\s+.+)|(DELETE\\s+FROM\\s+.+)|(UPDATE\\s+.+\\s+SET\\s+.+)|(INSERT\\s+INTO\\s+.+VALUES.+)|((ALTER|DROP)\\s+TABLE\\s+.+)",
		lcasePayload,
	)
	if hasSQLStatement {
		return hasSQLStatement, errors.New("sql on payload")
	}

	hasJSScript := strings.Contains(lcasePayload, "<script") ||
		strings.Contains(lcasePayload, "\\u003cscript") ||
		strings.Contains(lcasePayload, "</script") ||
		strings.Contains(lcasePayload, "\\u003c/script")

	if hasJSScript {
		return hasJSScript, errors.New("jsscript tags on payload")
	}

	return false, nil
}

// The string should contain Alphabets(upper/lower case), number, hyphen, underscore, space
func IsUserOrProjectNameValid(actualString string) bool {
	isValid, _ := regexp.MatchString(
		"^[A-Za-z0-9-_ ]*$",
		actualString,
	)
	return isValid
}

func GetBrowser(p *user_agent.UserAgent) (string, string) {
	if IsPingdomBot(p.UA()) {
		return "PingdomBot", ""
	}

	if IsLighthouse(p.UA()) {
		return "LightHouse", ""
	}

	return p.Browser()
}

// FormatProperty get property split based on demlimiter
func FormatProperty(property string) string {
	// check if first character is digit to support previously built models
	prop := []rune(property)
	if unicode.IsDigit(prop[0]) == true {
		return strings.SplitN(property, ".", 2)[1]
	}
	return property
}

func CreatePropertyNameFromDisplayName(displayName string) string {
	displaySplit := strings.Split(displayName, " ")
	propertyName := strings.ToLower(displaySplit[0])
	for _, splitString := range displaySplit[1:] {
		propertyName = fmt.Sprintf("%v_%v", propertyName, strings.ToLower(splitString))
	}
	return propertyName
}

func CreateVirtualDisplayName(actualName string) string {
	displayName := ""
	if strings.HasPrefix(actualName, "$") {
		actualName := strings.TrimPrefix(actualName, "$")
		actualSplit := strings.Split(actualName, "_")
		for _, splitString := range actualSplit {
			if displayName == "" {
				displayName = fmt.Sprintf("%v", CapitalizeFirstLetter(splitString))
			} else {
				displayName = fmt.Sprintf("%v %v", displayName, CapitalizeFirstLetter(splitString))
			}
		}
		return displayName
	}
	return actualName
}

func CapitalizeFirstLetter(data string) string {
	return strings.Title(strings.ToLower(data))
}
func GetArrayOfTokensFromFormula(formula string) []string {
	arr := make([]string, 0)
	prevSplTokenIndex := -1
	for i, c := range formula {
		ch := string(c)
		if i == 0 {
			if strings.Contains(AllowedDerivedMetricSpecialChars, ch) {
				arr = append(arr, ch)
				prevSplTokenIndex = i
			}
		} else if strings.Contains(AllowedDerivedMetricSpecialChars, ch) {
			token := formula[prevSplTokenIndex+1 : i]
			if len(token) != 0 {
				arr = append(arr, token)
			}
			arr = append(arr, ch)
			prevSplTokenIndex = i
		}
	}
	lastCharOfFormula := string(formula[len(formula)-1])
	if !strings.Contains(AllowedDerivedMetricSpecialChars, lastCharOfFormula) {
		token := formula[prevSplTokenIndex+1:]
		if len(token) != 0 {
			arr = append(arr, token)
		}
	}
	return arr
}

func IsAlphabeticToken(token string) bool {
	for _, c := range token {
		ch := string(c)
		if (ch >= "a" && ch <= "z") || (ch >= "A" && ch <= "Z") {
			continue
		} else {
			return false
		}
	}
	return true
}

func IsNumericToken(token string) bool {
	for _, c := range token {
		ch := string(c)
		if (ch >= "0" && ch <= "9") || (ch == ".") {
			continue
		} else {
			return false
		}
	}
	return true
}

// this checks the validity of arithmetic formula by running the formula in a stack and if we are able to get the result then the formula is correct
func ValidateArithmeticFormula(formula string) bool {
	var valueStack []string
	var operatorStack []string

	formulaArray := GetArrayOfTokensFromFormula(formula)

	for i, currToken := range formulaArray {
		if i == 0 {
			continue
		}
		prevToken := formulaArray[i-1]

		if !(IsAlphabeticToken(currToken) || strings.Contains(AllowedDerivedMetricSpecialChars, strings.ToLower(currToken)) || IsNumericToken(currToken)) {
			return false
		}
		if (IsAlphabeticToken(currToken) || IsNumericToken(currToken)) && !strings.Contains(AllowedDerivedMetricSpecialChars, strings.ToLower(prevToken)) {
			return false
		}
		if strings.Contains(AllowedDerivedMetricOperator, strings.ToLower(currToken)) && strings.Contains(AllowedDerivedMetricOperator, strings.ToLower(prevToken)) {
			return false
		}
	}

	// Explained in the EvaluateKPIExpressionWithBraces func, same logic applied at both the places
	// this run a simulated calculation of the arithmetic formula and let's us know if the formula is correct or not
	for _, token := range formulaArray {

		if token == "(" {
			operatorStack = append(operatorStack, token)
		} else if strings.Contains(Alpha, strings.ToLower(token)) {
			valueStack = append(valueStack, token)

		} else if strings.Contains(Numeric, string(token[0])) || string(token[0]) == "." {
			var err error
			if strings.Contains(token, ".") {
				_, err = strconv.ParseFloat(token, 64)
			} else {
				_, err = strconv.ParseInt(token, 10, 64)
			}
			if err != nil {
				return false
			} else {
				valueStack = append(valueStack, token)
			}

		} else if token == ")" {
			for len(operatorStack) != 0 && operatorStack[len(operatorStack)-1] != "(" {
				if len(valueStack) < 2 {
					return false
				}
				valueStack = valueStack[:len(valueStack)-1]
				operatorStack = operatorStack[:len(operatorStack)-1]
			}
			if len(operatorStack) != 0 {
				operatorStack = operatorStack[:len(operatorStack)-1]
			}
		} else {
			for len(operatorStack) != 0 && Precedence(operatorStack[len(operatorStack)-1]) >= Precedence(token) {
				if len(valueStack) < 2 {
					return false
				}
				valueStack = valueStack[:len(valueStack)-1]
				operatorStack = operatorStack[:len(operatorStack)-1]
			}
			operatorStack = append(operatorStack, token)
		}
	}
	for len(operatorStack) != 0 {
		if len(valueStack) < 2 {
			return false
		}
		valueStack = valueStack[:len(valueStack)-1]
		operatorStack = operatorStack[:len(operatorStack)-1]
	}
	if len(valueStack) == 1 && len(operatorStack) == 0 {
		return true
	} else {
		return false
	}
}

func Precedence(op string) int {
	if op == "+" || op == "-" {
		return 1
	}
	if op == "*" || op == "/" || op == "%" {
		return 2
	}
	return 0
}

func ApplyOp(a, b float64, op string) float64 {
	switch op {
	case "+":
		return a + b
	case "-":
		return a - b
	case "*":
		return a * b
	case "/":
		if b == float64(0) {
			return 0
		}
		return a / b
	default:
		return 0
	}
}

func CheckFileExists(cloudManager *filestore.FileManager, path, name string) (bool, error) {
	if _, err := (*cloudManager).Get(path, name); err != nil {
		log.WithFields(log.Fields{"err": err, "filePath": path,
			"fileName": name}).Info("Failed to fetch from cloud path")
		return false, err
	} else {
		return true, nil
	}
}

func GetModelType(start, end int64) string {
	diff := end - start
	if diff <= Per_day_epoch {
		return "d"
	} else if diff <= Per_week_epoch {
		return "w"
	} else if diff < Per_week_epoch*6 {
		return "m"
	} else {
		return "q"
	}
}

func SetBitAtPosition(input string, replacement string, index int) string {
	return input[:index] + string(replacement) + input[index+1:]

}
