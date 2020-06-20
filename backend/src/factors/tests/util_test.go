package tests

import (
	U "factors/util"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStringListAsBatch(t *testing.T) {
	batch1 := U.GetStringListAsBatch([]string{"1", "2", "3", "4"}, 2)
	assert.Len(t, batch1, 2)
	assert.Len(t, batch1[0], 2)
	assert.Len(t, batch1[1], 2)

	batch2 := U.GetStringListAsBatch([]string{"1", "2", "3"}, 2)
	assert.Len(t, batch2, 2)
	assert.Len(t, batch2[1], 1)
	assert.Equal(t, "1", batch2[0][0])
	assert.Equal(t, "2", batch2[0][1])
	assert.Equal(t, "3", batch2[1][0])
}
