package tests

import (
	C "factors/config"
	"factors/model/model"
	SDK "factors/sdk"

	U "factors/util"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCacheResultByUserAgent(t *testing.T) {

	C.GetConfig().DeviceServiceURL = "http://0.0.0.0:3000/device_service"

	userProperties := make(U.PropertiesMap, 0)
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64; XBOX_ONE_ED) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.79 Safari/537.36 Edge/14.14393"

	// response shall be nil as device info is not set in cache for given user agent
	resp, errCode, err := model.GetCacheResultByUserAgent(userAgent)
	if resp == nil {
		assert.NotEqual(t, errCode, http.StatusFound)
		assert.Nil(t, resp)
	}
	// set device info from device service
	SDK.FillDeviceInfoFromDeviceService(&userProperties, userAgent)
	assert.NotNil(t, userProperties[U.UP_USER_AGENT])
	assert.NotNil(t, userProperties[U.UP_DEVICE_BRAND])
	assert.NotNil(t, userProperties[U.UP_DEVICE_MODEL])
	assert.NotNil(t, userProperties[U.UP_DEVICE_TYPE])
	assert.NotEqual(t, "Bot", userProperties[U.UP_BROWSER])

	// Cached response retrieved
	resp, errCode, err = model.GetCacheResultByUserAgent(userAgent)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, errCode, http.StatusFound)
	assert.Equal(t, resp.IsBot, false)

	assert.NotNil(t, resp.ClientInfo)
	assert.Equal(t, resp.ClientInfo.Name, "Microsoft Edge")
	assert.Equal(t, resp.ClientInfo.Family, "Internet Explorer")
	assert.Equal(t, resp.ClientInfo.Version, "14.14393")

	assert.NotNil(t, resp.OsInfo)
	assert.Equal(t, resp.OsInfo.Name, "Windows")
	assert.Equal(t, resp.OsInfo.Version, "10")

	assert.Equal(t, resp.DeviceType, "console")
	assert.Equal(t, resp.DeviceBrand, "Microsoft")
	assert.Equal(t, resp.DeviceModel, "Xbox One S")

	userAgent2 := "Mozilla/5.0 (compatible; Yahoo! Slurp; http://help.yahoo.com/help/us/ysearch/slurp)"
	userProperties2 := make(U.PropertiesMap, 0)
	resp, errCode, err = model.GetCacheResultByUserAgent(userAgent2)

	if resp == nil {
		assert.NotEqual(t, errCode, http.StatusFound)
		assert.Nil(t, resp)
	}

	SDK.FillDeviceInfoFromDeviceService(&userProperties2, userAgent2)
	assert.NotNil(t, userProperties2[U.UP_USER_AGENT])
	assert.Equal(t, userAgent2, userProperties2[U.UP_USER_AGENT])
	assert.NotNil(t, userProperties2[U.UP_BROWSER])
	assert.Equal(t, "Bot", userProperties2[U.UP_BROWSER])

	resp, errCode, err = model.GetCacheResultByUserAgent(userAgent2)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, errCode, http.StatusFound)
	assert.Equal(t, resp.IsBot, true)

}

func TestSetCacheResultByUserAgent(t *testing.T) {

	userAgent := "Mozilla/5.0 (iPhone; CPU iPhone OS 12_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148"

	deviceInfo := &model.DeviceInfo{
		IsBot: false,
		ClientInfo: model.ClientInfo{
			Type:          "browser",
			Name:          "Mobile Safari",
			ShortName:     "MF",
			Version:       "",
			Engine:        "WebKit",
			EngineVersion: "605.1.15",
			Family:        "Safari",
		},
		OsInfo: model.OsInfo{
			Name:      "iOS",
			ShortName: "IOS",
			Version:   "12.2",
			Platform:  "",
			Family:    "iOS",
		},
		DeviceType:  "smartphone",
		DeviceBrand: "Apple",
		DeviceModel: "iPhone",
	}

	// set cache for user agent
	model.SetCacheResultByUserAgent(userAgent, deviceInfo)

	// Cached response retrieved
	resp, errCode, err := model.GetCacheResultByUserAgent(userAgent)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, errCode, http.StatusFound)
	assert.Equal(t, resp.IsBot, deviceInfo.IsBot)

	assert.NotNil(t, resp.ClientInfo)
	assert.Equal(t, resp.ClientInfo.Name, deviceInfo.ClientInfo.Name)
	assert.Equal(t, resp.ClientInfo.Version, deviceInfo.ClientInfo.Version)
	assert.Equal(t, resp.ClientInfo.Engine, deviceInfo.ClientInfo.Engine)
	assert.Equal(t, resp.ClientInfo.EngineVersion, deviceInfo.ClientInfo.EngineVersion)

	assert.NotNil(t, resp.OsInfo)
	assert.Equal(t, resp.OsInfo.Name, deviceInfo.OsInfo.Name)
	assert.Equal(t, resp.OsInfo.Family, deviceInfo.OsInfo.Family)
	assert.Equal(t, resp.OsInfo.Version, deviceInfo.OsInfo.Version)

	assert.Equal(t, resp.DeviceType, deviceInfo.DeviceType)
	assert.Equal(t, resp.DeviceModel, deviceInfo.DeviceModel)
	assert.Equal(t, resp.DeviceType, deviceInfo.DeviceType)

}
