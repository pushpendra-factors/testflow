package integration

import (
	U "factors/util"

	"github.com/jinzhu/gorm/dialects/postgres"
)

// Note:
// Segment(userId) = Factors(customerUserId), Segment(AnonymousId) = Factors(userId).
// Property mappings are defined on corresponding Fill*Properities method.

type SegmentDevice struct {
	ID                string `json:"id"`
	Manufacturer      string `json:"manufacturer"`
	Model             string `json:"model"`
	Type              string `json:"type"`
	Name              string `json:"name"`
	AdvertisingID     string `json:"advertisingId"`
	AdTrackingEnabled bool   `json:"adTrackingEnabled"`
	Token             string `json:"token"`
}

type SegmentPage struct {
	Referrer string `json:"referrer"`
	RawURL   string `json:"url"`
	Title    string `json:"title"`
	// Path     string `json:"path"`
	// Search   string `json:"search"`
}

type SegmentApp struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Build     string `json:"build"`
	Namespace string `json:"namespace"`
}

type SegmentLocation struct {
	City    string  `json:"city"`
	Country string  `json:"country"`
	Region  string  `json:"region"`
	Lat     float64 `json:"latitude"`
	Long    float64 `json:"longitude"`
}

type SegmentOS struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type SegmentScreen struct {
	Width   float64 `json:"width"`
	Height  float64 `json:"height"`
	Density float64 `json:"density"`
}

type SegmentNetwork struct {
	Bluetooth bool   `json:"bluetooth"`
	Carrier   string `json:"carrier"`
	Cellular  bool   `json:"cellular"`
	Wifi      bool   `json:"wifi"`
}

type SegmentCampaign struct {
	Name    string `json:"name"`
	Source  string `json:"source"`
	Medium  string `json:"medium"`
	Term    string `json:"term"`
	Content string `json:"content"`
}

type SegmentContext struct {
	Campaign  SegmentCampaign `json:"campaign"`
	IP        string          `json:"ip"`
	Location  SegmentLocation `json:"location"`
	Page      SegmentPage     `json:"page"`
	UserAgent string          `json:"userAgent"`
	OS        SegmentOS       `json:"os"`
	Screen    SegmentScreen   `json:"screen"`
	Locale    string          `json:"locale"`
	Device    SegmentDevice   `json:"device"`
	Network   SegmentNetwork  `json:"network"`
	App       SegmentApp      `json:"app"`
	Timezone  string          `json:"timezone"`
}

type SegmentEvent struct {
	TrackName   string          `json:"event"`
	ScreenName  string          `json:"name"`
	UserId      string          `json:"userId"`
	AnonymousID string          `json:"anonymousId"`
	Channel     string          `json:"channel"`
	Context     SegmentContext  `json:"context"`
	Timestamp   string          `json:"timestamp"`
	Type        string          `json:"type"`
	Version     float64         `json:"version"`
	Properties  U.PropertiesMap `json:"properties"`
	Traits      postgres.Jsonb  `json:"traits"`
}

func FillSegmentGenericEventProperties(properties *U.PropertiesMap, event *SegmentEvent) {
	if event.Context.Location.Lat != 0 {
		(*properties)[U.EP_LOCATION_LATITUDE] = event.Context.Location.Lat
	}
	if event.Context.Location.Long != 0 {
		(*properties)[U.EP_LOCATION_LONGITUDE] = event.Context.Location.Long
	}
}

func FillSegmentGenericUserProperties(properties *U.PropertiesMap, event *SegmentEvent) {
	(*properties)[U.UP_PLATFORM] = U.PLATFORM_WEB
	if event.Context.UserAgent != "" {
		(*properties)[U.UP_USER_AGENT] = event.Context.UserAgent
	}
	if event.Context.Location.Country != "" {
		(*properties)[U.UP_COUNTRY] = event.Context.Location.Country
	}
	if event.Context.Location.City != "" {
		(*properties)[U.UP_CITY] = event.Context.Location.City
	}
	if event.Context.Location.Region != "" {
		(*properties)[U.UP_REGION] = event.Context.Location.Region
	}

	// Added to generic event though it is mobile specific on segment.
	if event.Context.OS.Name != "" {
		(*properties)[U.UP_OS] = event.Context.OS.Name
	}
	if event.Context.OS.Version != "" {
		(*properties)[U.UP_OS_VERSION] = event.Context.OS.Version
	}
	if event.Context.Screen.Width != 0 {
		(*properties)[U.UP_SCREEN_WIDTH] = event.Context.Screen.Width
	}
	if event.Context.Screen.Height != 0 {
		(*properties)[U.UP_SCREEN_HEIGHT] = event.Context.Screen.Height
	}
}

func FillSegmentMobileEventProperties(properties *U.PropertiesMap, event *SegmentEvent) {
	if event.Context.Device.ID != "" {
		(*properties)[U.EP_DEVICE_ID] = event.Context.Device.ID
	}
	if event.Context.Device.Name != "" {
		(*properties)[U.EP_DEVICE_NAME] = event.Context.Device.Name
	}
	if event.Context.Device.AdvertisingID != "" {
		(*properties)[U.EP_DEVICE_ADVERTISING_ID] = event.Context.Device.AdvertisingID
	}
}

func FillSegmentMobileUserProperties(properties *U.PropertiesMap, event *SegmentEvent) {
	if event.Context.App.Name != "" {
		(*properties)[U.UP_APP_NAME] = event.Context.App.Name
	}
	if event.Context.App.Namespace != "" {
		(*properties)[U.UP_APP_NAMESPACE] = event.Context.App.Namespace
	}
	if event.Context.App.Build != "" {
		(*properties)[U.UP_APP_BUILD] = event.Context.App.Build
	}
	if event.Context.App.Version != "" {
		(*properties)[U.UP_APP_VERSION] = event.Context.App.Version
	}
	if event.Context.Device.Model != "" {
		(*properties)[U.UP_DEVICE_MODEL] = event.Context.Device.Model
	}
	if event.Context.Device.Type != "" {
		(*properties)[U.UP_DEVICE_TYPE] = event.Context.Device.Type
	}
	if event.Context.Device.Manufacturer != "" {
		(*properties)[U.UP_DEVICE_MANUFACTURER] = event.Context.Device.Manufacturer
	}
	if event.Context.Network.Carrier != "" {
		(*properties)[U.UP_NETWORK_CARRIER] = event.Context.Network.Carrier
	}
	if event.Context.Screen.Density != 0 {
		(*properties)[U.UP_SCREEN_DENSITY] = event.Context.Screen.Density
	}
	if event.Context.Timezone != "" {
		(*properties)[U.UP_TIMEZONE] = event.Context.Timezone
	}
	if event.Context.Locale != "" {
		(*properties)[U.UP_LOCALE] = event.Context.Locale
	}

	// Boolean values added without check.
	(*properties)[U.UP_DEVICE_ADTRACKING_ENABLED] = event.Context.Device.AdTrackingEnabled
	(*properties)[U.UP_NETWORK_BLUETOOTH] = event.Context.Network.Bluetooth
	(*properties)[U.UP_NETWORK_CELLULAR] = event.Context.Network.Cellular
	(*properties)[U.UP_NETWORK_WIFI] = event.Context.Network.Wifi
}

func FillSegmentWebEventProperties(properties *U.PropertiesMap, event *SegmentEvent) {
	if event.Context.Page.RawURL != "" {
		(*properties)[U.EP_RAW_URL] = event.Context.Page.RawURL
	}
	if event.Context.Page.Title != "" {
		(*properties)[U.EP_PAGE_TITLE] = event.Context.Page.Title
	}
	if event.Context.Page.Referrer != "" {
		(*properties)[U.EP_REFERRER] = event.Context.Page.Referrer
	}
}

func FillSegmentWebUserProperties(properties *U.PropertiesMap, event *SegmentEvent) {
	if event.Context.Campaign.Name != "" {
		(*properties)[U.UP_CAMPAIGN_NAME] = event.Context.Campaign.Name
	}
	if event.Context.Campaign.Source != "" {
		(*properties)[U.UP_CAMPAIGN_SOURCE] = event.Context.Campaign.Source
	}
	if event.Context.Campaign.Medium != "" {
		(*properties)[U.UP_CAMPAIGN_MEDIUM] = event.Context.Campaign.Medium
	}
	if event.Context.Campaign.Term != "" {
		(*properties)[U.UP_CAMPAIGN_TERM] = event.Context.Campaign.Term
	}
	if event.Context.Campaign.Content != "" {
		(*properties)[U.UP_CAMPAIGN_CONTENT] = event.Context.Campaign.Content
	}
}
