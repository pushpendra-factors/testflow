package util

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jinzhu/gorm/dialects/postgres"
)

// DBReadRows Creates [][]interface{} from sql result rows.
// Ref: https://kylewbanks.com/blog/query-result-to-map-in-golang
func DBReadRows(rows *sql.Rows) ([]string, [][]interface{}, error) {
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	resultRows := make([][]interface{}, 0, 0)

	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
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
					return cols, nil, errors.New("failed reading row. invalid bytes")
				}
			default:
				resultRow = append(resultRow, *val)
			}
		}

		resultRows = append(resultRows, resultRow)
	}

	return cols, resultRows, nil
}

func DBDebugPreparedStatement(stmnt string, params []interface{}) string {
	return fmt.Sprintf(strings.Replace(stmnt, "?", "'%v'", len(params)), params...)
}

func IsEmptyPostgresJsonb(jsonb *postgres.Jsonb) bool {
	strJson := string((*jsonb).RawMessage)
	return strJson == "" || strJson == "null"
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

	return &postgres.Jsonb{newJsonb}, nil
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
