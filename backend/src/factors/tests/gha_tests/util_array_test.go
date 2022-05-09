package gha_tests

// THE TESTS UNDER gha_tests ARE SAMPLE TESTS USED ONLY FOR GITHUB ACTIONS.
// DONOT ADD OTHER TESTS TO THIS PACKAGE.

import (
	U "factors/util"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsStringInArray(t *testing.T) {
	array := []string{"str1", "str2", "str3", "str4", "str5"}
	result := U.ContainsStringInArray(array, "str2")
	assert.True(t, result)

	result = U.ContainsStringInArray(array, "str6")
	assert.False(t, result)
}
