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

func TestUtilGetURLHostAndPath(t *testing.T) {
	url, err := U.ParseURLStable("https://www.factors.ai/?param=1")
	assert.Nil(t, err)
	p1 := U.GetURLHostAndPath(url)
	assert.Equal(t, "www.factors.ai/", p1)

	// hash should be allowed on path.
	url2, err := U.ParseURLStable("https://app.factors.ai/#/core")
	assert.Nil(t, err)
	p2 := U.GetURLHostAndPath(url2)
	assert.Equal(t, "app.factors.ai/#/core", p2)

	// query params on fragment should not exist.
	url3, err := U.ParseURLStable("https://app.factors.ai/#/core?param=1")
	assert.Nil(t, err)
	p3 := U.GetURLHostAndPath(url3)
	assert.Equal(t, "app.factors.ai/#/core", p3)
}

func TestUtilGetQueryParamsFromURLFragment(t *testing.T) {
	paramsMap := U.GetQueryParamsFromURLFragment("a=10&b=20")
	assert.Len(t, paramsMap, 2)
	assert.NotNil(t, paramsMap["a"])
	assert.NotNil(t, paramsMap["b"])
	assert.Equal(t, "10", paramsMap["a"])
	assert.Equal(t, "20", paramsMap["b"])

	paramsMap = U.GetQueryParamsFromURLFragment("a=10&b=")
	assert.Len(t, paramsMap, 1)
	assert.NotNil(t, paramsMap["a"])
	assert.Nil(t, paramsMap["b"])

	paramsMap = U.GetQueryParamsFromURLFragment("a=&b=20")
	assert.Len(t, paramsMap, 1)
	assert.Nil(t, paramsMap["a"])
	assert.NotNil(t, paramsMap["b"])
}
