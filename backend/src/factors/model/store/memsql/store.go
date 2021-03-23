package memsql

// Postgres store base struct type.
type MemSQL struct{}

// GetStore - Returns an instance of MemSQL.
func GetStore() *MemSQL {
	return &MemSQL{}
}
