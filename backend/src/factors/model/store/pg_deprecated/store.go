package postgres

// Postgres store base struct type.
type Postgres struct{}

// GetStore - Returns an instance of postgres.
func GetStore() *Postgres {
	return &Postgres{}
}
