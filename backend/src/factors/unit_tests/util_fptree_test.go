package unit_tests

import (
	U "factors/util"
	"github.com/stretchr/testify/assert"
	"testing"
)
func TestCheckUniqueTrans(t *testing.T){
	array:= []string{"str1","str2","str3","str4","str5"}
	isUnique:= U.CheckUniqueTrans(array)
	assert.True(t,isUnique)
}

func TestMakeUniqueTrans(t *testing.T){
	array:= []string{"str1","str2","str3","str4","str5","str1"}
	assert.Len(t,array,6)
	array = U.MakeUniqueTrans(array)
	assert.Len(t,array,5)
}