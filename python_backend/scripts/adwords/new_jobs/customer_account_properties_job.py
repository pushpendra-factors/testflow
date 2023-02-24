import logging as log

from lib.utils.csv import CsvUtil
from lib.adwords.oauth_service.fetch_service import FetchService
from scripts.adwords.new_jobs.fields_mapping import FieldsMapping

import scripts

from .base_job import BaseJob
from .query import QueryBuilder
from .service_fetch_job import ServicesFetch

class NewGetCustomerAccountPropertiesJob(ServicesFetch):

    EXTRACT_FIELDS = [
        "customer.id",
        "customer.currency_code",
        "customer.time_zone",
        "customer.manager",
        "customer.test_account",
    ]

    HEADERS_VMAX = [
        "customer_id",
        "currency_code",
        "date_timezone",
        "can_manage_clients",
        "test_account"
    ]

    TRANSFORM_MAP_V01 = [
        {ServicesFetch.FIELD: "can_manage_clients", ServicesFetch.MAP: FieldsMapping.BOOLEAN_MAPPING},
        {ServicesFetch.FIELD: "test_account", ServicesFetch.MAP: FieldsMapping.BOOLEAN_MAPPING},
    ]

    REPORT = "customer"

    def __init__(self, next_info):
        super().__init__(next_info)
