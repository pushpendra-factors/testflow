package util

import (
	"database/sql"
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
	return string((*jsonb).RawMessage) == ""
}
