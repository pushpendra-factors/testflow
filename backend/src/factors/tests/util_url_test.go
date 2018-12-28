package tests

import (
	U "factors/util"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUtilURLPopURIBySlash(t *testing.T) {
	afterPopURI, poppedURIPart := U.PopURIBySlash("/u1/u2")
	assert.Equal(t, "/u1", afterPopURI)
	assert.Equal(t, "/u2", poppedURIPart)

	afterPopURI, poppedURIPart = U.PopURIBySlash("/u1")
	assert.Equal(t, "/u1", poppedURIPart)
	assert.Equal(t, "", afterPopURI)

	afterPopURI, poppedURIPart = U.PopURIBySlash("")
	assert.Equal(t, "", poppedURIPart)
	assert.Equal(t, "", afterPopURI)
}

func TestUtilURLParseWithoutProtocol(t *testing.T) {
	p, err := U.ParseURLWithoutProtocol("a.com/u1/u2")
	assert.Nil(t, err)
	assert.Equal(t, p.Path, "/u1/u2") // path
	assert.Equal(t, p.Host, "a.com")  // domain

	// parsing filter_expr uri
	p1, err := U.ParseURLWithoutProtocol("a.com/u1/:v1")
	assert.Nil(t, err)
	assert.Equal(t, p1.Path, "/u1/:v1")

	p2, err := U.ParseURLWithoutProtocol("a.com/u1/:v1/u2")
	assert.Nil(t, err)
	assert.Equal(t, p2.Path, "/u1/:v1/u2")

	// check purpose of triming slash suffix slash after parse.
	p3, err := U.ParseURLWithoutProtocol("a.com/u1/u2/")
	assert.Nil(t, err)
	assert.Equal(t, p3.Path, "/u1/u2/")
	assert.NotEqual(t, p3.Path, "/u1/u2")

	p4, err := U.ParseURLWithoutProtocol("localhost:3030/u1/u2")
	assert.Nil(t, err)
	// For users testing from non-prod env.
	assert.Equal(t, p4.Host, "localhost:3030")
}
