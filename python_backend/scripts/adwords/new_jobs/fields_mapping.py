
import re


class FieldsMapping:

    # For Reports
    STATUS_MAPPING = {
        '0' : "unspecified",
        '1' : "unknown",
        '2' : "enabled",
        '3' : "paused",
        '4' : "removed",
    }

    BOOLEAN_MAPPING = {
        "False" : 'false',
        "True" : 'true',
    }

    KEYWORD_MAPPING = {
        '0': "unspecified",
        '1': "unknown",
        '2': "Exact",
        '3': "Phrase",
        '4': "Broad"
    }

    BIDDING_SOURCE_MAPPING = {
        '0': "unspecified",
        '1': "unknown",
        '5': "campaign bidding strategy",
        '6': "ad groups",
        '7': "ad group criteria",
    }

    AD_GROUP_TYPE_MAPPING = {
        '0': "unspecified",
        '1': "Unknown",
        '2': "Standard",
        '3': "Display",
        '4': "Shopping - Product",
        '6': "Hotel",
        '7': "Shopping - Smart",
        '8': "Bumper",
        '9': "In-stream",
        '10': "Video discovery",
        '11': "Standard",
        '12': "Outstream",
        '13': "Search Dynamic Ads",
        '14': "Shopping - Comparison-listing",
        '15': "Hotel Promoted",
        '16': "Video Responsive",
        '17': "Video Efficient Reach",
        '18': "Smart Campaign",
    }

    CAMPAIGN_TRIAL_TYPE_MAPPING = {
        '0' : "unspecified",
        '1' : "unknown",
        '2' : "base campaign",
        '3' : "draft campaign",
        '4' : "trial campaign",
    }

    ADVERTISING_CHANNEL_TYPE_MAPPING = {
        '0' : "unspecified",
        '1' : "unknown",
        '2' : "Search",
        '3' : "Display",
        '4' : "Shopping",
        '5' : "Hotel",
        '6' : "Video",
        '7' : "Multi Channel",
        '8' : "Local",
        '9' : "Smart",
        '10' : "Performance Max",
        '11' : "Local Services",
    }

    ADVERTISING_CHANNEL_SUB_TYPE_MAPPING = {
        '0': "Unspecified",
        '1': "Unknown",
        '2': "Search Mobile App",
        '3': "Display Mobile App",
        '4': "Search Express",
        '5': "Display Express",
        '6': "Shopping Smart Ads",
        '7': "Gmail Ad campaign",
        '8': "Smart display campaign",
        '9': "Video Outstrem",
        '10': "Video Action",
        '11': "Video Non Skippable",
        '12': "App Campaign",
        '13': "App Campaign for Engagement",
        '14': "Local Campaign",
        '15': "Shopping Comparison Listing Ads",
        '16': "Smart Campaign",
        '17': "Video Sequence",
        '18': "App Campaign for Pre Registration",
    }   

    INTERACTION_TYPES_MAPPING = {
        '<InteractionEventType.UNSPECIFIED: 0>' : "unspecified",
        '<InteractionEventType.UNKNOWN: 1>' : "unknown",
        '<InteractionEventType.CLICK: 2>' : "Clicks",
        '<InteractionEventType.ENGAGEMENT: 3>' : "Engagements",
        '<InteractionEventType.VIDEO_VIEW: 4>' : "Video Views",
        '<InteractionEventType.NONE: 5>' : "None",
    }

    AD_NETWORK_TYPE_MAPPING = {
        '0' : "unspecified",
        '1' : "unknown",
        '2' : "Search Network",
        '3' : "Search Partners",
        '4' : "Display Network",
        '5' : "YouTube Search",
        '6' : "YouTube Videos",
        '7' : "Cross-network",
    }

    CLICK_TYPE_MAPPING = {
        '0' : "unspecified",
        '1' : "unknown",
        '2' : "App engagement ad deep link",
        '3' : "Breadcrumbs",
        '4' : "Broadband Plan",
        '5' : "Manually dialed phone calls",
        '6' : "Phone calls",
        '7' : "Click on engagement ad",
        '8' : "Driving direction",
        '9' : "Get location details",
        '10' : "Call",
        '11' : "Directions",
        '12' : "Image(s)",
        '13' : "Go to landing page",
        '14' : "Map",
        '15' : "Go to store info",
        '16' : "Text",
        '17' : "Mobile phone calls",
        '18' : "Print offer",
        '19' : "Other",
        '20' : "Product plusbox offer",
        '21' : "Shopping ad - Standard",
        '22' : "Sitelink",
        '23' : "Show nearby locations",
        '25' : "Headline",
        '26' : "App store",
        '27' : "Call-to-Action overlay",
        '28' : "Video Card Action Headline Clicks",
        '29' : "End cap",
        '30' : "Website",
        '31' : "Visual Sitelinks",
        '32' : "Wireless Plan",
        '33' : "Shopping ad - Local",
        '34' : "Shopping ad - MultiChannel Local",
        '35' : "Shopping ad - MultiChannel Online",
        '36' : "Shopping ad - Coupon",
        '37' : "Shopping ad - Sell on Google",
        '38' : "Shopping ad - App Deeplink",
        '39' : "Shopping - Showcase - Category",
        '40' : "Shopping - Showcase - Local storefront",
        '42' : "Shopping - Showcase - Online product",
        '43' : "Shopping - Showcase - Local product",
        '44' : "Promotion Extension",
        '45' : "Ad Headline",
        '46' : "Swipes",
        '47' : "See More",
        '48' : "Sitelink 1",
        '49' : "Sitelink 2",
        '50' : "Sitelink 3",
        '51' : "Sitelink 4",
        '52' : "Sitelink 5",
        '53' : "Hotel price",
        '54' : "Price Extension",
        '55' : "Hotel Book-on-Google room selectio",
        '56' : "Shopping Comparision Listing",
    }

    DEVICE_MAPPING = {
        '0': "unspecified",
        '1': "Other",
        '2': "Mobile devices with full browsers",
        '3': "Tablets with full browsers",
        '4': "Computers",
        '5': "Other",
        '6': "Devices streaming video content to TV screens",
    }

    SLOT_MAPPING = {
        '0': "unspecified",
        '1': "unknown",
        '2': "Google search: Side",
        '3': "Google search: Top",
        '4': "Google search: Other",
        '5': "Google Display Network",
        '6': "Search partners: Top",
        '7': "Search partners: Other",
        '8': "Cross-network", 
    }

    KEYWORD_MATCH_TYPE_MAPPING = {
        '0': "unspecified",
        '1': "unknown",
        '2': "Exact",
        '3': "Phrase",
        '4': "Broad",
    }

    # For Services
    SERVICE_STATUS_MAPPING = {
        '0' : "UNSPECIFIED",
        '1' : "UNKNOWN",
        '2' : "ENABLED",
        '3' : "PAUSED",
        '4' : "REMOVED",
    }

    SERVICE_AD_GROUP_TYPE_MAPPING = {
        '0': "UNSPECIFIED",
        '1': "UNKNOWN",
        '2': "SEARCH_STANDARD",
        '3': "DISPLAY_STANDARD",
        '4': "SHOPPING_PRODUCT_ADS",
        '6': "HOTEL_ADS",
        '7': "SHOPPING_SMART_ADS",
        '8': "VIDEO_BUMPER",
        '9': "VIDEO_TRUE_VIEW_IN_STREAM",
        '10': "VIDEO_TRUE_VIEW_IN_DISPLAY",
        '11': "VIDEO_NON_SKIPPABLE_IN_STREAM",
        '12': "VIDEO_OUTSTREAM",
        '13': "SEARCH_DYNAMIC_ADS",
        '14': "SHOPPING_COMPARISON_LISTING_ADS",
        '15': "PROMOTED_HOTEL_ADS",
        '16': "VIDEO_RESPONSIVE",
        '17': "VIDEO_EFFICIENT_REACH",
        '18': "SMART_CAMPAIGN_ADS",
    }

    SERVICE_SERVING_STATUS_MAPPING = {
       '0': "UNSPECIFIED",
       '1': "UNKNOWN",
       '2': "SERVING",
       '3': "NONE",
       '4': "ENDED",
       '5': "PENDING",
       '6': "SUSPENDED",
    }

    SERVICE_AD_SERVING_OPTIMIZATION_STATUS_MAPPING = {
        '0': "UNSPECIFIED",
        '1': "UNKNOWN",
        '2': "OPTIMIZE",
        '3': "CONVERSION_OPTIMIZE",
        '4': "ROTATE",
        '5': "ROTATE_INDEFINITELY",
        '6': "UNAVAILABLE",
    }

    SERVICE_ADVERTISING_CHANNEL_TYPE_MAPPING = {
        '0': "UNSPECIFIED",
        '1': "UNKNOWN",
        '2': "SEARCH",
        '3': "DISPLAY",
        '4': "SHOPPING",
        '5': "HOTEL",
        '6': "VIDEO",
        '7': "MULTI_CHANNEL",
        '8': "LOCAL",
        '9': "SMART",
        '10': "PERFORMANCE_MAX ",
        '11': "LOCAL_SERVICES ",
        '12': "DISCOVERY"
    }

    SERVICE_ADVERTISING_CHANNEL_SUB_TYPE_MAPPING = {
        '0': "UNSPECIFIED",
        '1': "UNKNOWN",
        '2': "SEARCH_MOBILE_APP",
        '3': "DISPLAY_MOBILE_APP",
        '4': "SEARCH_EXPRESS",
        '5': "DISPLAY_EXPRESS",
        '6': "SHOPPING_SMART_ADS",
        '7': "DISPLAY_GMAIL_AD",
        '8': "DISPLAY_SMART_CAMPAIGN",
        '9': "VIDEO_OUTSTREAM",
        '10': "VIDEO_ACTION",
        '11': "VIDEO_NON_SKIPPABLE",
        '12': "APP_CAMPAIGN ",
        '13': "APP_CAMPAIGN_FOR_ENGAGEMENT",
        '14': "LOCAL_CAMPAIGN",
        '15': "SHOPPING_COMPARISON_LISTING_ADS",
        '16': "SMART_CAMPAIGN",
        '17': "VIDEO_SEQUENCE",
        '18': "APP_CAMPAIGN_FOR_PRE_REGISTRATION",
    }

    SERVICE_CAMPAIGN_TRIAL_TYPE_MAPPING = {
        '0': "UNSPECIFIED",
        '1': "UNKNOWN",
        '2': "BASE",
        '3': "DRAFT",
        '4': "TRIAL",
    }

    SERVICE_AD_TYPE_MAPPING = {
        '0': "UNSPECIFIED",
        '1': "UNKNOWN",
        '2': "TEXT_AD",
        '3': "EXPANDED_TEXT_AD",
        '7': "EXPANDED_DYNAMIC_SEARCH_AD",
        '8': "HOTEL_AD",
        '9': "SHOPPING_SMART_AD",
        '10': "SHOPPING_PRODUCT_AD",
        '12': "VIDEO_AD",
        '13': "GMAIL_AD",
        '14': "IMAGE_AD",
        '15': "RESPONSIVE_SEARCH_AD",
        '16': "LEGACY_RESPONSIVE_DISPLAY_AD",
        '17': "APP_AD",
        '18': "LEGACY_APP_INSTALL_AD",
        '19': "RESPONSIVE_DISPLAY_AD",
        '20': "LOCAL_AD",
        '21': "HTML5_UPLOAD_AD",
        '22': "DYNAMIC_HTML5_AD",
        '23': "APP_ENGAGEMENT_AD",
        '24': "SHOPPING_COMPARISON_LISTING_AD",
        '25': "VIDEO_BUMPER_AD",
        '26': "VIDEO_NON_SKIPPABLE_IN_STREAM_AD",
        '27': "VIDEO_OUTSTREAM_AD",
        '29': "VIDEO_TRUEVIEW_IN_STREAM_AD",
        '30': "VIDEO_RESPONSIVE_AD",
        '31': "SMART_CAMPAIGN_AD",
        '32': "CALL_AD",
        '33': "APP_PRE_REGISTRATION_AD",
        '34': "IN_FEED_VIDEO_AD",
        '35': "DISCOVERY_MULTI_ASSET_AD",
        '36': "DISCOVERY_CAROUSEL_AD"
    }

    @staticmethod
    def transform_status(field):
        return FieldsMapping.STATUS_MAPPING[field]

    @staticmethod
    def transform_service_status(field):
        return FieldsMapping.SERVICE_STATUS_MAPPING[field]
    
    @staticmethod
    def transform_resource_name(field):
        resource = re.split('/|~', field)
        return resource[len(resource)-1]

    @staticmethod
    def transform_boolean(field):
        return FieldsMapping.BOOLEAN_MAPPING[field]
    
    @staticmethod
    def transform_percentage(field):
        return str(float(field)*100) + '%'

    @staticmethod
    def transform_interaction_types(field):
        old_list = field[1:len(field)-1].split(',')
        new_list = []
        for type in old_list:
            type = type.strip()
            if type != '':
                new_list.append(FieldsMapping.INTERACTION_TYPES_MAPPING[type])
        return new_list