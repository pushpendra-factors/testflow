package tests

import (
	"errors"
	"testing"

	U "factors/util"

	"github.com/stretchr/testify/assert"
)

func TestIsPostgresUniqueIndexViolationError(t *testing.T) {
	assert.True(t, U.IsPostgresUniqueIndexViolationError("column_unique_index",
		errors.New("pq: duplicate key value violates unique constraint \"column_unique_index\"")))
}

func TestIsPostgresIntegrityViolationError(t *testing.T) {
	assert.True(t, U.IsPostgresIntegrityViolationError(
		errors.New("pq: duplicate key value violates unique constraint \"column_unique_index\"")))
}
