package util

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"runtime/debug"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

func CloseReadQuery(rows *sql.Rows, tx *sql.Tx) {
	if rows != nil {
		rows.Close()
	}

	if tx == nil {
		return
	}

	err := tx.Commit()
	if err != nil {
		log.WithError(err).WithField("stack", string(debug.Stack())).Error("Failed to commit on transaction.")
	}
}

// DBReadRows Creates [][]interface{} from sql result rows.
// Ref: https://kylewbanks.com/blog/query-result-to-map-in-golang
func DBReadRows(rows *sql.Rows, tx *sql.Tx, queryID string) ([]string, [][]interface{}, error) {
	defer CloseReadQuery(rows, tx)

	cols, err := rows.Columns()
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	startReadTime := time.Now()
	resultRows := make([][]interface{}, 0, 0)
	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			tx.Rollback()
			return cols, nil, err
		}

		// each row.
		resultRow := make([]interface{}, 0, 0)
		for i := range cols {
			val := columnPointers[i].(*interface{})
			switch (*val).(type) {
			case []byte:
				if b, ok := (*val).([]uint8); ok {
					resultRow = append(resultRow, string(b))
				} else {
					tx.Rollback()
					return cols, nil, errors.New("failed reading row. invalid bytes")
				}
			case int, int32, int64, float32:
				resultRow = append(resultRow, SafeConvertToFloat64(*val))
			default:
				resultRow = append(resultRow, *val)
			}
		}

		resultRows = append(resultRows, resultRow)
	}
	LogReadTimeWithQueryRequestID(startReadTime, queryID, &log.Fields{"function": "DBReadRows"})

	return cols, resultRows, nil
}

func DBDebugPreparedStatement(env, stmnt string, params []interface{}) string {
	if env != "production" {
		return fmt.Sprintf(strings.Replace(stmnt, "?", "'%v'", len(params)), params...)
	}

	// Trimming params and statement for logging.
	limitedParams := TrimQueryParams(env, params)
	stmntWithParams := fmt.Sprintf(strings.Replace(stmnt, "?", "'%v'", len(limitedParams)), limitedParams...)
	return TrimQueryString(env, stmntWithParams)
}

func TrimQueryString(env, stmnt string) string {
	if env != "production" {
		return stmnt
	}

	// Limiting statement length to 500 characters.
	return stmnt[:int(math.Min(float64(len(stmnt)), 500))] + "..."
}

func TrimQueryParams(env string, params []interface{}) []interface{} {
	if env != "production" {
		return params
	}

	// Limiting params to 100.
	return params[:int(math.Min(float64(len(params)), 100))]
}

func IsEmptyPostgresJsonb(jsonb *postgres.Jsonb) bool {
	strJson := string((*jsonb).RawMessage)
	return strJson == "" || strJson == "null"
}

func AreJsonbEqual(jsonb1 postgres.Jsonb, jsonb2 postgres.Jsonb) bool {
	jsonString1, err := json.Marshal(jsonb1)
	if err != nil {
		fmt.Println("Error marshalling jsonb1:", err)
		return false
	}

	jsonString2, err := json.Marshal(jsonb2)
	if err != nil {
		fmt.Println("Error marshalling jsonb2:", err)
		return false
	}

	return string(jsonString1) == string(jsonString2)
}

// AddToJsonb adds key values to the jsonb. To overwrite
// existing keys set overwriteExisting to true.
func AddToPostgresJsonb(sourceJsonb *postgres.Jsonb,
	newKvs map[string]interface{}, overwriteExisting bool) (*postgres.Jsonb, error) {

	var sourceMap map[string]interface{}
	if !IsEmptyPostgresJsonb(sourceJsonb) {
		if err := json.Unmarshal((*sourceJsonb).RawMessage, &sourceMap); err != nil {
			return nil, err
		}
	} else {
		sourceMap = make(map[string]interface{}, 0)
	}

	for k, v := range newKvs {
		_, exists := sourceMap[k]
		if exists && !overwriteExisting {
			continue
		}

		sourceMap[k] = v
	}

	newJsonb, err := json.Marshal(sourceMap)
	if err != nil {
		return nil, err
	}

	return &postgres.Jsonb{RawMessage: newJsonb}, nil
}

// RemoveFromJsonb adds key values to the jsonb.
func RemoveFromJsonb(sourceJsonb *postgres.Jsonb,
	keysToRemove []string) (*postgres.Jsonb, error) {

	var sourceMap map[string]interface{}
	if !IsEmptyPostgresJsonb(sourceJsonb) {
		if err := json.Unmarshal((*sourceJsonb).RawMessage, &sourceMap); err != nil {
			return nil, err
		}
	} else {
		sourceMap = make(map[string]interface{}, 0)
	}

	for _, key := range keysToRemove {
		_, exists := sourceMap[key]
		if exists {
			delete(sourceMap, key)
		}
	}

	newJsonb, err := json.Marshal(sourceMap)
	if err != nil {
		return nil, err
	}

	return &postgres.Jsonb{RawMessage: newJsonb}, nil
}

func DecodePostgresJsonb(sourceJsonb *postgres.Jsonb) (*map[string]interface{}, error) {
	var sourceMap map[string]interface{}
	if !IsEmptyPostgresJsonb(sourceJsonb) {
		if err := json.Unmarshal((*sourceJsonb).RawMessage, &sourceMap); err != nil {
			return nil, err
		}
	} else {
		sourceMap = make(map[string]interface{}, 0)
	}

	return &sourceMap, nil
}

func DecodePostgresJsonbAsPropertiesMap(sourceJsonb *postgres.Jsonb) (*PropertiesMap, error) {
	properties, err := DecodePostgresJsonb(sourceJsonb)
	if err != nil {
		return nil, err
	}

	propertiesMap := PropertiesMap(*properties)
	return &propertiesMap, err
}

func EncodeToPostgresJsonb(sourceMap *map[string]interface{}) (*postgres.Jsonb, error) {
	sourceJsonBytes, err := json.Marshal(sourceMap)
	if err != nil {
		return nil, err
	}

	return &postgres.Jsonb{sourceJsonBytes}, nil
}

func EncodeStructTypeToPostgresJsonb(structType interface{}) (*postgres.Jsonb, error) {
	sourceJsonBytes, err := json.Marshal(structType)
	if err != nil {
		return nil, err
	}

	return &postgres.Jsonb{sourceJsonBytes}, nil
}

// EncodeStructTypeToMap Converts a struct to map[string]interface{}.
// Order of keys remains consistent https://stackoverflow.com/a/18668885/2341189.
func EncodeStructTypeToMap(structType interface{}) (map[string]interface{}, error) {
	encodedMap := make(map[string]interface{})
	jsonValue, err := json.Marshal(structType)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(jsonValue, &encodedMap)
	if err != nil {
		return nil, err
	}
	return encodedMap, nil
}

// DecodePostgresJsonbToStructType Decodes a postgres.Jsonb object to given type.
func DecodePostgresJsonbToStructType(sourceJsonb *postgres.Jsonb, destStruct interface{}) error {
	if IsEmptyPostgresJsonb(sourceJsonb) {
		return fmt.Errorf("Empty jsonb object")
	} else if err := json.Unmarshal((*sourceJsonb).RawMessage, destStruct); err != nil {
		return err
	}
	return nil
}

// DecodeInterfaceMapToStructType Converts a source of type map[string]interface{} to given struct.
func DecodeInterfaceMapToStructType(source map[string]interface{}, destStruct interface{}) error {
	sourceBytes, err := json.Marshal(source)
	if err != nil {
		return err
	}
	err = json.Unmarshal(sourceBytes, destStruct)
	if err != nil {
		return err
	}
	return nil
}

// DecodeJSONStringToStructType Decodes a json string to object of given type.
func DecodeJSONStringToStructType(jsonString string, destStruct interface{}) error {
	if jsonString == "" {
		return fmt.Errorf("Empty json string")
	} else if err := json.Unmarshal([]byte(jsonString), destStruct); err != nil {
		return err
	}
	return nil
}

func IsPostgresIntegrityViolationError(err error) bool {
	// i.e pq: duplicate key value violates unique constraint \"col_unique_idx\"
	return strings.Contains(err.Error(), "violates") && strings.Contains(err.Error(), "constraint")
}

func IsPostgresUniqueIndexViolationError(indexName string, err error) bool {
	if indexName == "" || err == nil {
		return false
	}

	return err.Error() == fmt.Sprintf("pq: duplicate key value violates unique constraint \"%s\"", indexName)
}

func IsPostgresUnsupportedUnicodeError(err error) bool {
	return strings.Contains(err.Error(), "unsupported Unicode escape sequence")
}

// GormCleanupCallback Custom GORM Plugin for cleaning up field values.
func GormCleanupCallback(scope *gorm.Scope) {
	for _, field := range scope.Fields() {
		switch field.Field.Type().String() {
		case "string":
			fieldValue := field.Field.Interface().(string)
			err := field.Set(SanitizeStringValueForUnicode(fieldValue))
			if err != nil {
				log.WithField("fieldValue", fieldValue).WithError(err).Error("Failed to cleanup string field value.")
				return
			}
		case "postgres.Jsonb":
			fieldValue := field.Field.Interface().(postgres.Jsonb)
			fieldValue.RawMessage = CleanupUnsupportedCharOnStringBytes(fieldValue.RawMessage)
			err := field.Set(fieldValue)
			if err != nil {
				log.WithField("fieldValue", fieldValue.RawMessage).WithError(err).Error("Failed to cleanup postgres.Jsonb field value.")
				return
			}

		case "*postgres.Jsonb":
			fieldValue := field.Field.Interface().(*postgres.Jsonb)
			if fieldValue == nil {
				return
			}

			fieldValue.RawMessage = CleanupUnsupportedCharOnStringBytes(fieldValue.RawMessage)
			err := field.Set(fieldValue)
			if err != nil {
				log.WithField("fieldValue", fieldValue.RawMessage).WithError(err).Error("Failed to cleanup *postgres.Jsonb field value.")
				return
			}
		}
	}
}

func GormDefault(scope *gorm.Scope) {
	for _, field := range scope.Fields() {
		if field.IsIgnored {
			continue
		}

		switch field.Field.Type().String() {
		case "time.Time":
			fieldValue := field.Field.Interface().(time.Time)
			if !fieldValue.IsZero() {
				continue
			}

			err := field.Set(time.Date(1971, 1, 1, 0, 0, 0, 0, time.UTC))
			if err != nil {
				log.WithError(err).Error("Failed to set default datetime field.")
				return
			}

		}
	}
}

func CleanupUnsupportedCharOnStringBytes(stringBytes []byte) []byte {
	nullRemovedBytes := RemoveNullCharacterBytes(stringBytes)
	return []byte(SanitizeStringValueForUnicode(string(nullRemovedBytes)))
}

// https://stackoverflow.com/a/34863211/2341189
// https://docs.singlestore.com/v7.3/guides/use-memsql/physical-schema-design/using-json/using-json/#unicode-support
// https://go.dev/ref/spec#Conversions_to_and_from_a_string_type
func SanitizeStringValueForUnicode(s string) string {
	runes := make([]rune, 0, 0)
	for _, r := range s {
		if !utf8.ValidRune(r) {
			continue
		} else if r > 65536 {
			// Memsql supports only till 65536. Convert to unicode point text.
			// escaped := []rune(strings.Replace(fmt.Sprintf("%U", r), "U+", "\\u", 1))
			// runes = append(runes, escaped...)
			continue
		} else {
			runes = append(runes, r)
		}
	}
	return string(runes)
}

func SantizePostgresJsonbForUnicode(jsonb *postgres.Jsonb) {
	if jsonb == nil {
		return
	}

	jsonb.RawMessage = json.RawMessage(SanitizeStringValueForUnicode(string(jsonb.RawMessage)))
}

func GetUniqueQueryRequestID() string {
	return RandomLowerAphaNumString(5)
}

func LogReadTimeWithQueryRequestID(startTime time.Time, reqID string, logFields *log.Fields) {
	timeTaken := time.Now().Sub(startTime).Seconds()
	log.WithFields(*logFields).WithField("req_id", reqID).
		WithField("time_in_secs", timeTaken).Info("Query rows read.")
}

func LogExecutionTimeWithQueryRequestID(startTime time.Time, reqID string, logFields *log.Fields) {
	timeTaken := time.Now().Sub(startTime).Seconds()
	log.WithFields(*logFields).WithField("req_id", reqID).
		WithField("time_in_secs", timeTaken).Info("Query executed.")
}

func LogComputeTimeWithQueryRequestID(startTime time.Time, reqID string, logFields *log.Fields) {
	timeTaken := time.Now().Sub(startTime).Seconds()
	log.WithFields(*logFields).WithField("req_id", reqID).
		WithField("time_in_secs", timeTaken).Info("Computations on query results completed.")
}

func isValidType(value interface{}) bool {
	switch value.(type) {
	case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64, float32, float64, string, bool:
		return true
	default:
		return false
	}
}

func DiffPostgresJsonb(projectID int64, oldPostgresJsonb,
	newPostgresJsonb *postgres.Jsonb, caller string) *map[string]interface{} {

	logCtx := log.WithField("project_id", projectID).WithField("caller", caller)

	oldJMap, err1 := DecodePostgresJsonb(oldPostgresJsonb)
	newJMap, err2 := DecodePostgresJsonb(newPostgresJsonb)
	if err1 != nil || err2 != nil {
		logCtx.WithField("error1", err1).
			WithField("error2", err2).
			WithField("old_json", oldPostgresJsonb).
			WithField("new_json", newPostgresJsonb).
			Error("Failed to decode json. Retruning new json as diff.")
		return nil
	}

	diffMap := make(map[string]interface{}, 0)
	for k, v := range *oldJMap {
		// Old keys with new values.
		newV, exists := (*newJMap)[k]
		if exists && isValidType(newV) && isValidType(v) {
			if v != newV {
				diffMap[k] = newV
			}
		}
	}

	// New keys.
	for k, v := range *newJMap {
		if _, exists := (*oldJMap)[k]; !exists && isValidType(v) {
			diffMap[k] = v
		}
	}

	// Whitelisted property of type map.
	vNew, newExists := (*newJMap)[UP_META_OBJECT_IDENTIFIER_KEY]
	vOld, oldExists := (*oldJMap)[UP_META_OBJECT_IDENTIFIER_KEY]
	addProperty := (!oldExists && newExists) || (oldExists && newExists && !reflect.DeepEqual(vNew, vOld))
	if addProperty {
		obj, err := json.Marshal(vNew)
		if err == nil {
			diffMap[UP_META_OBJECT_IDENTIFIER_KEY] = string(obj)
		}
	}

	logCtx.WithField("old_len", len(*oldJMap)).
		WithField("new_len", len(*newJMap)).
		WithField("delta_len", len(diffMap)).
		Info("Diff of properties on overwrite.")

	return &diffMap
}
