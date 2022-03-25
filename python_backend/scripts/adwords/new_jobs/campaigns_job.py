from .service_fetch_job import ServicesFetch
from .fields_mapping import FieldsMapping

class NewGetCampaignsJob(ServicesFetch):
    
    # Elements in EXTRACT_FIELDS and HEADERS_VMAX are in sync
    # If field is added in any one then add to other too.
    EXTRACT_FIELDS = [
        "campaign.id",
        "campaign.name",
        "campaign.status",
        "campaign.serving_status",
        "campaign.start_date",
        "campaign.end_date",
        "campaign.ad_serving_optimization_status",
        "campaign.advertising_channel_type",
        "campaign.advertising_channel_sub_type",
        "campaign.labels",
        "campaign.experiment_type",
        "campaign.base_campaign",
        "campaign.tracking_url_template",
        "campaign.final_url_suffix",
        "campaign.url_custom_parameters",
        "campaign.selective_optimization.conversion_actions",
    ]

    HEADERS_VMAX = [
        "id",
        "name",
        "status",
        "serving_status",
        "start_date",
        "end_date",
        "ad_serving_optimization_status",
        "advertising_channel_type",
        "advertising_channel_sub_type",
        "labels",
        "campaign_trial_type",
        "base_campaign_id",
        "tracking_url_template",
        "final_url_suffix",
        "url_custom_parameters", 
        "selective_optimization",
    ]

    REPORT = "campaign"

    FIELDS_WITH_STATUS = [
        "status",
    ]

    FIELDS_WITH_RESOURCE_NAME = [
        "base_campaign_id",
    ]

    FIELDS_TO_INTEGER = [
        "id",
        "base_campaign_id",
    ]

    TRANSFORM_FIELDS_V01 = [
        {ServicesFetch.FIELDS: FIELDS_WITH_STATUS, ServicesFetch.OPERATION: FieldsMapping.transform_service_status},
        {ServicesFetch.FIELDS: FIELDS_WITH_RESOURCE_NAME, ServicesFetch.OPERATION: FieldsMapping.transform_resource_name},
        # Keep integer conversion after resource name if there is resource name
        # resouce name -> string id -> integer id
        {ServicesFetch.FIELDS: FIELDS_TO_INTEGER, ServicesFetch.OPERATION: int},
    ]
    TRANSFORM_MAP_V01 = [
        {ServicesFetch.FIELD: "serving_status", ServicesFetch.MAP: FieldsMapping.SERVICE_SERVING_STATUS_MAPPING},
        {ServicesFetch.FIELD: "ad_serving_optimization_status", ServicesFetch.MAP: FieldsMapping.SERVICE_AD_SERVING_OPTIMIZATION_STATUS_MAPPING},
        {ServicesFetch.FIELD: "advertising_channel_type", ServicesFetch.MAP: FieldsMapping.SERVICE_ADVERTISING_CHANNEL_TYPE_MAPPING},
        {ServicesFetch.FIELD: "advertising_channel_sub_type", ServicesFetch.MAP: FieldsMapping.SERVICE_ADVERTISING_CHANNEL_SUB_TYPE_MAPPING},
        {ServicesFetch.FIELD: "campaign_trial_type", ServicesFetch.MAP: FieldsMapping.SERVICE_CAMPAIGN_TRIAL_TYPE_MAPPING},
    ]

    def __init__(self, next_info):
        super().__init__(next_info)