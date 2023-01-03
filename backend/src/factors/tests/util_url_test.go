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

func TestUtilIsBotUserAgent(t *testing.T) {
	assert.True(t, U.IsBotUserAgent("Mozilla/5.0 (Linux; Android 5.0; SM-G920A) AppleWebKit (KHTML, like Gecko) Chrome Mobile Safari (compatible; AdsBot-Google-Mobile; +http://www.google.com/mobile/adsbot.html)"))
	assert.True(t, U.IsBotUserAgent("Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1 (compatible; AdsBot-Google-Mobile; +http://www.google.com/mobile/adsbot.html)"))
	assert.True(t, U.IsBotUserAgent("Googlebot-Image/1.0"))
	assert.True(t, U.IsBotUserAgent("Googlebot-News"))
	assert.True(t, U.IsBotUserAgent("Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"))
	assert.True(t, U.IsBotUserAgent("Mozilla/5.0 (compatible; Bingbot/2.0; +http://www.bing.com/bingbot.htm)"))
	assert.True(t, U.IsBotUserAgent("Mozilla/5.0 (compatible; Yahoo! Slurp; http://help.yahoo.com/help/us/ysearch/slurp)"))
	assert.True(t, U.IsBotUserAgent("DuckDuckBot/1.0; (+http://duckduckgo.com/duckduckbot.html)"))
	assert.True(t, U.IsBotUserAgent("Mozilla/5.0 (compatible; Baiduspider/2.0; +http://www.baidu.com/search/spider.html)"))
	assert.True(t, U.IsBotUserAgent("facebookexternalhit/1.0 (+http://www.facebook.com/externalhit_uatext.php)"))
	assert.True(t, U.IsBotUserAgent("ia_archiver (+http://www.alexa.com/site/help/webmasters; crawler@alexa.com)"))
	assert.True(t, U.IsBotUserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/61.0.3163.100 Chrome/61.0.3163.100 Safari/537.36 PingdomPageSpeed/1.0 (pingbot/2.0; +http://www.pingdom.com/)"))
	assert.True(t, U.IsBotUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) browser/2020.2.1 Chrome/78.0.3904.130 Electron/7.3.2 Safari/537.36 PingdomTMS/2020.2"))
	assert.True(t, U.IsBotUserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/53.0.2785.143 Chrome/53.0.2785.143 Safari/537.36 PingdomPageSpeed/1.0 (pingbot/2.0; +http://www.pingdom.com/)"))
	assert.True(t, U.IsBotUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) WPMRecorder/2020.2.6 Chrome/87.0.4280.141 Electron/11.3.0 Safari/537.36 PingdomTMS/2020.2"))
	assert.True(t, U.IsBotUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) browser/2020.2.2 Chrome/78.0.3904.130 Electron/7.3.2 Safari/537.36 PingdomTMS/2020.2"))
	assert.True(t, U.IsBotUserAgent("Pingdom.com_bot_version_1.4_(http://www.pingdom.com/)"))
	assert.True(t, U.IsBotUserAgent("Mozilla/5.0 (Linux; Android 7.0; Moto G (4)) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4590.2 Mobile Safari/537.36 Chrome-Lighthouse"))
}

func TestIsBotEventByPrefix(t *testing.T) {
	assert.True(t, U.IsBotEventByPrefix("gtm-msr.appspot.com/render2"))
	assert.True(t, U.IsBotEventByPrefix("gtm-msr.appspot.com/xxx"))
	assert.True(t, U.IsBotEventByPrefix("gtm-msr.appspot.com/"))
}

func TestRemoveAllInvalidURLEscapeFromURL(t *testing.T) {
	// 0 invalid escape
	assert.Equal(
		t,
		"http://www.livspace.com/in/magazine/gallery-girls-bedroom-ideas/amp#aoh=16164287572386\u0026csi=0\u0026referrer=https://www.google.com\u0026amp_tf=From",
		U.UnescapeAllInvalidURLEscapeFromURL("http://www.livspace.com/in/magazine/gallery-girls-bedroom-ideas/amp#aoh=16164287572386\u0026csi=0\u0026referrer=https://www.google.com\u0026amp_tf=From"),
	)

	// 1 invalid escape
	assert.Equal(
		t,
		"http://www.livspace.com/in/magazine/gallery-girls-bedroom-ideas/amp#aoh=16164287572386\u0026csi=0\u0026referrer=https://www.google.com\u0026amp_tf=From 1$s",
		U.UnescapeAllInvalidURLEscapeFromURL("http://www.livspace.com/in/magazine/gallery-girls-bedroom-ideas/amp#aoh=16164287572386\u0026csi=0\u0026referrer=https://www.google.com\u0026amp_tf=From %1$s"),
	)

	// 3 invalid escapes.
	assert.Equal(
		t,
		"http://www.livspace.com/in/magazine/gallery-girls-bedroom-ideas/amp#aoh=16164287572386\u0026csi=0\u0026referrer=https://www.google.com\u0026amp_tf=From 1$s 1$s 1$s",
		U.UnescapeAllInvalidURLEscapeFromURL("http://www.livspace.com/in/magazine/gallery-girls-bedroom-ideas/amp#aoh=16164287572386\u0026csi=0\u0026referrer=https://www.google.com\u0026amp_tf=From %1$s %1$s %1$s"),
	)

	// valid escape %3B and invalid escape together.
	assert.Equal(
		t,
		"http://www.livspace.com/in/magazine/gallery-girls-bedroom-ideas/amp#aoh=16164287572386\u0026csi=0\u0026referrer=https://www.google.com\u0026amp_tf=From 1$s %3B",
		U.UnescapeAllInvalidURLEscapeFromURL("http://www.livspace.com/in/magazine/gallery-girls-bedroom-ideas/amp#aoh=16164287572386\u0026csi=0\u0026referrer=https://www.google.com\u0026amp_tf=From %1$s %3B"),
	)
}
