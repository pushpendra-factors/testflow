from lib.utils.string import StringUtil
from .multiple_requests_fetch_job import MultipleRequestsFetchJob


class GetAdGroupsJob(MultipleRequestsFetchJob):
    FIELDS = ["Id", "CampaignId", "CampaignName", "Name", "Status", "Settings", "Labels",
              "ContentBidCriterionTypeGroup", "BaseCampaignId", "BaseAdGroupId", "AdGroupType"]

    SERVICE_NAME = "AdGroupService"
    ENTITY_TYPE = "AdGroup"

    def process_entity(self, selector, entity):
        doc = {}
        for field in selector["fields"]:
            field_name = StringUtil.first_letter_to_lower(field)
            if StringUtil.is_valid_value_type(entity[field_name]):
                doc[StringUtil.camel_case_to_snake_case(field_name)] = entity[field_name]

        return doc
