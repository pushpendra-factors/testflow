from lib.utils.string import StringUtil
from .multiple_requests_fetch_job import MultipleRequestsFetchJob


class GetAdsJob(MultipleRequestsFetchJob):
    FIELDS = ["AdGroupId", "Status", "BaseCampaignId", "BaseAdGroupId"]

    SERVICE_NAME = 'AdGroupAdService'
    ENTITY_TYPE = 'Ads'

    def __init__(self, next_info):
        super().__init__(next_info)

    def process_entity(self, selector, entity):
        doc = {}
        for field in self.FIELDS:
            field_name = StringUtil.first_letter_to_lower(field)
            if StringUtil.is_valid_value_type(entity[field_name]):
                doc[StringUtil.camel_case_to_snake_case(field_name)] = entity[field_name]

            # Add values form ad object.
            if entity['ad'] is not None:
                for ad_field in entity['ad']:
                    if StringUtil.is_valid_value_type(entity['ad'][ad_field]):
                        doc[StringUtil.camel_case_to_snake_case(ad_field)] = entity['ad'][ad_field]

        return doc
