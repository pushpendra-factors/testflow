package unit_tests

import (
	U "factors/util"
	"github.com/stretchr/testify/assert"
	"testing"
)
func TestContainsStringInArray(t *testing.T){
	array:= []string{"str1","str2","str3","str4","str5"}
	result:= U.ContainsStringInArray(array,"str2")
	assert.True(t,result)

	result = U.ContainsStringInArray(array,"str6")
	assert.False(t,result)
}