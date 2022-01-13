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

func TestCleanupPostgresJsonStringBytes(t *testing.T) {
	assert.Equal(
		t,
		string(U.CleanupUnsupportedCharOnStringBytes([]byte("🌎💧🍃🌾🏭🔬🚽🚿🇯🇵  Environmental Bioengineering Lab. led by Akihiko Terada and Shohei Riya at Dep. Chem. Eng. in Tokyo Univ. Agri. & Tech., "))),
		"  Environmental Bioengineering Lab. led by Akihiko Terada and Shohei Riya at Dep. Chem. Eng. in Tokyo Univ. Agri. & Tech., ",
	)
}
