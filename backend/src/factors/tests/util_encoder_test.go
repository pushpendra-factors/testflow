package tests

import (
	"factors/util"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncoding(t *testing.T) {
	str := "!\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"

	encoded := util.Encode(str, 4)
	afterEncoding := "%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~!\"#$"
	assert.Equal(t, afterEncoding, encoded)

	decoded := util.Decode(encoded, 4)
	assert.Equal(t, str, decoded)
}

func TestEncodingWithRichText(t *testing.T) {
	str := "!\"#$😂%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|😇}~"

	encoded := util.Encode(str, 4)
	afterEncoding := "%&'(😂)*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~!\"😇#$"
	assert.Equal(t, afterEncoding, encoded)

	decoded := util.Decode(encoded, 4)
	assert.Equal(t, str, decoded)
}
