package memsql

import "strings"

// Postgres store base struct type.
type MemSQL struct{}

const (
	// Error 1062: Leaf Error (127.0.0.1:3307): Duplicate entry '5000247-FactorsGoal1' for key 'unique_project_id_name_idx'
	// Error 1062: Leaf Error (127.0.0.1:3307): Duplicate entry '6000762-f6e8c235-7aa0-42fe-b987-137866bcdd8f' for key 'PRIMARY'
	MEMSQL_ERROR_CODE_DUPLICATE_ENTRY = "Error 1062"
)

func IsDuplicateRecordError(err error) bool {
	return strings.HasPrefix(err.Error(), MEMSQL_ERROR_CODE_DUPLICATE_ENTRY)
}

// GetStore - Returns an instance of MemSQL.
func GetStore() *MemSQL {
	return &MemSQL{}
}
