from .fields_mapping import FieldsMapping
from .reports_fetch_job import ReportsFetch


# Note: If the number of custom paths exceed 7 in the subClasses. Move it to strategic pattern.
class NewClickPerformanceReportsJob(ReportsFetch):
    
    # Elements in EXTRACT_FIELDS and HEADERS_VMAX are in sync
    # If field is added in any one then add to other too.
    EXTRACT_FIELDS = [
       "ad_group.id",
       "ad_group.name",
       "ad_group.status",
       "segments.ad_network_type",
       "click_view.area_of_interest.most_specific",
       "campaign.id",
       "click_view.campaign_location_target",
       "campaign.name",
       "campaign.status",
       "metrics.clicks",
       "segments.click_type",
       "click_view.ad_group_ad",
       "segments.date",
       "segments.device",
       "customer.id",
       "click_view.gclid",
       "click_view.page_number",
       "segments.slot",
       "click_view.user_list",
       "click_view.keyword",
       "click_view.keyword_info.match_type",
    ]

    HEADERS_V01 =[
        "ad_group_id", 
        "ad_group_name", 
        "ad_group_status", 
        "ad_network_type_1",
        "aoi_most_specific_target_id", 
        "campaign_id", 
        "campaign_location_target_id", 
        "campaign_name",
        "campaign_status", 
        "clicks",
        "click_type", 
        "creative_id", 
        "date", 
        "device",
        "external_customer_id", 
        "gcl_id",
        "page", 
        "slot", 
        "user_list_id",
        "keyword_id",
        "keyword_match_type",
    ]

    HEADERS_V00 = [
        "ad_format", "ad_group_id", "ad_group_name", "ad_group_status", "ad_network_type_1",
        "ad_network_type_2",
        "aoi_most_specific_target_id", "campaign_id", "campaign_location_target_id", "campaign_name",
        "campaign_status", "clicks",
        "click_type", "creative_id", "criteria_id", "criteria_parameters", "date", "device",
        "external_customer_id", "gcl_id",
        "page", "slot", "user_list_id"]
    
    HEADERS_V02 = HEADERS_V01
    HEADERS_VMAX = HEADERS_V01
    
    REPORT = "click_view"

    FIELDS_WITH_STATUS = [
        "ad_group_status",
        "campaign_status",
    ]

    FIELDS_WITH_RESOURCE_NAME = [
        "campaign_location_target_id",
        "creative_id",
        "user_list_id",
        "keyword_id",
    ]

    TRANSFORM_MAP_V01 = [
        {ReportsFetch.FIELD: "ad_network_type_1", ReportsFetch.MAP: FieldsMapping.AD_NETWORK_TYPE_MAPPING},
        {ReportsFetch.FIELD: "click_type", ReportsFetch.MAP: FieldsMapping.CLICK_TYPE_MAPPING},
        {ReportsFetch.FIELD: "device", ReportsFetch.MAP: FieldsMapping.DEVICE_MAPPING},
        {ReportsFetch.FIELD: "slot", ReportsFetch.MAP: FieldsMapping.SLOT_MAPPING},
        {ReportsFetch.FIELD: "keyword_match_type", ReportsFetch.MAP: FieldsMapping.KEYWORD_MATCH_TYPE_MAPPING},
    ] 

    def __init__(self, next_info):
        super().__init__(next_info)   
