package gha_tests

// THE TESTS UNDER gha_tests ARE SAMPLE TESTS USED ONLY FOR GITHUB ACTIONS.
// DONOT ADD OTHER TESTS TO THIS PACKAGE.

import (
	U "factors/util"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckUniqueTrans(t *testing.T) {
	array := []string{"str1", "str2", "str3", "str4", "str5"}
	isUnique := U.CheckUniqueTrans(array)
	assert.True(t, isUnique)
}

func TestMakeUniqueTrans(t *testing.T) {
	array := []string{"str1", "str2", "str3", "str4", "str5", "str1"}
	assert.Len(t, array, 6)
	array = U.MakeUniqueTrans(array)
	assert.Len(t, array, 5)
}
