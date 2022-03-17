from .service_fetch_job import ServicesFetch
from .fields_mapping import FieldsMapping

class NewGetAdGroupsJob(ServicesFetch):

    # Elements in EXTRACT_FIELDS and HEADERS_VMAX are in sync
    # If field is added in any one then add to other too.
    EXTRACT_FIELDS = [
        "ad_group.id",
        "campaign.id",
        "campaign.name",
        "ad_group.name",
        "ad_group.status",
        "ad_group.labels",
        "campaign.base_campaign",
        "ad_group.base_ad_group",
        "ad_group.type",
    ]

    HEADERS_VMAX = [
        "id",
        "campaign_id",
        "campaign_name",
        "name",
        "status",
        "labels",
        "base_campaign_id",
        "base_ad_group_id",
        "ad_group_type",
    ]

    REPORT = "ad_group"

    FIELDS_WITH_STATUS = [
        "status",
    ]

    FIELDS_WITH_RESOURCE_NAME = [
        "base_campaign_id",
        "base_ad_group_id",
    ]

    FIELDS_TO_INTEGER = [
        "id",
        "campaign_id",
        "base_campaign_id",
        "base_ad_group_id",
    ]

    TRANSFORM_FIELDS_V01 = [
        {ServicesFetch.FIELDS: FIELDS_WITH_STATUS, ServicesFetch.OPERATION: FieldsMapping.transform_service_status},
        {ServicesFetch.FIELDS: FIELDS_WITH_RESOURCE_NAME, ServicesFetch.OPERATION: FieldsMapping.transform_resource_name},
        # Keep integer conversion after resource name if there is resource name
        # resouce name -> string id -> integer id
        {ServicesFetch.FIELDS: FIELDS_TO_INTEGER, ServicesFetch.OPERATION: int},
    ]
    TRANSFORM_MAP_V01 = [
        {ServicesFetch.FIELD: "ad_group_type", ServicesFetch.MAP: FieldsMapping.SERVICE_AD_GROUP_TYPE_MAPPING},
    ]

    def __init__(self, next_info):
        super().__init__(next_info)
        