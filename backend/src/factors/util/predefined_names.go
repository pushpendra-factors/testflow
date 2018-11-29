package util

// Predefined Event Properties.
var EP_OCCURRENCE_COUNT string = "$occurrenceCount"

// Default Properties added from javascript sdk.
var DP_REFERRER string = "$referrer"
var DP_BROWSER string = "$browser"
var DP_BROWSER_VERSION string = "$browser_version"
var DP_OS string = "$os"
var DP_OS_VERSION string = "$os_version"
var DP_SCREEN_WIDTH string = "$screen_width"
var DP_SCREEN_HEIGHT string = "$screen_height"

// Default Properties added from backend.
var DP_IP string = "$ip"
var DP_COUNTRY string = "$country"

var DEFAULT_NUMERIC_EVENT_PROPERTIES = [...]string{EP_OCCURRENCE_COUNT}
