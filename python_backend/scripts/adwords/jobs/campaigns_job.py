from lib.utils.string import StringUtil
from .multiple_requests_fetch_job import MultipleRequestsFetchJob


class GetCampaignsJob(MultipleRequestsFetchJob):
    FIELDS = ["Id", "CampaignGroupId", "Name", "Status", "ServingStatus", "StartDate", "EndDate",
              "AdServingOptimizationStatus", "Settings", "AdvertisingChannelType", "AdvertisingChannelSubType",
              "Labels", "CampaignTrialType", "BaseCampaignId", "TrackingUrlTemplate", "FinalUrlSuffix",
              "UrlCustomParameters", "SelectiveOptimization"]

    SERVICE_NAME = "CampaignService"
    ENTITY_TYPE = "Campaign"

    def process_entity(self, selector, entity):
        doc = {}
        for field in selector["fields"]:
            field_name = StringUtil.first_letter_to_lower(field)
            if StringUtil.is_valid_value_type(entity[field_name]):
                doc[StringUtil.camel_case_to_snake_case(field_name)] = entity[field_name]

        return doc
