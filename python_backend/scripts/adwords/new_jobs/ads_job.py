from tkinter.tix import Tree
from .fields_mapping import FieldsMapping
from .service_fetch_job import ServicesFetch


class NewGetAdsJob(ServicesFetch):

    # Elements in EXTRACT_FIELDS and HEADERS_VMAX are in sync
    # If field is added in any one then add to other too.
    EXTRACT_FIELDS = [
        "ad_group.id",
        "ad_group_ad.status",
        "campaign.base_campaign",
        "ad_group.base_ad_group",
        "ad_group_ad.ad.id",
        "ad_group_ad.ad.type",
    ]

    HEADERS_VMAX = [
        "ad_group_id",
        "status",
        "base_campaign_id",
        "base_ad_group_id",
        "id",
        "type",
    ]

    REPORT = "ad_group_ad"
    PROCESS_JOB = True

    FIELDS_WITH_STATUS = [
        "status",
    ]

    FIELDS_WITH_RESOURCE_NAME = [
        "base_campaign_id",
        "base_ad_group_id",
    ]

    FIELDS_TO_INTEGER = [
        "ad_group_id",
        "base_campaign_id",
        "base_ad_group_id",
        "id",
    ]

    TRANSFORM_FIELDS_V01 = [
        {ServicesFetch.FIELDS: FIELDS_WITH_STATUS, ServicesFetch.OPERATION: FieldsMapping.transform_service_status},
        {ServicesFetch.FIELDS: FIELDS_WITH_RESOURCE_NAME, ServicesFetch.OPERATION: FieldsMapping.transform_resource_name},
        # Keep integer conversion after resource name if there is resource name
        # resouce name -> string id -> integer id
        {ServicesFetch.FIELDS: FIELDS_TO_INTEGER, ServicesFetch.OPERATION: int},
    ]
    TRANSFORM_MAP_V01 = [
        {ServicesFetch.FIELD: "type", ServicesFetch.MAP: FieldsMapping.SERVICE_AD_TYPE_MAPPING},
    ]

    def __init__(self, next_info):
        super().__init__(next_info)