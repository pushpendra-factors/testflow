import logging as log

import scripts
from lib.adwords.oauth_service.fetch_service import FetchService
from .base_job import BaseJob


# Note: If the number of code paths exceed 7 in the subClasses. Move it to strategic pattern.
class MultipleRequestsFetchJob(BaseJob):
    FIELDS = []
    SERVICE_NAME = ''
    ENTITY_TYPE = ''

    def __init__(self, next_info):
        super().__init__(next_info)

    def process_entity(self, selector, entity):
        """ Override this in the sub classes. """

    def get_entities_of_single_page(self, service, selector):
        rows = []
        page = service.get(selector)

        # Display results.
        if 'entries' in page:
            for entity in page['entries']:
                doc = self.process_entity(selector, entity)
                rows.append(doc)
            log.warning('Entities of %s were fetched at offset: %s', self.ENTITY_TYPE, selector['paging']['startIndex'])
        else:
            log.warning('No more entities of %s were found at offset: %s', self.ENTITY_TYPE, selector['paging']['startIndex'])

        return rows, int(page['totalNumEntries'])

    def get_entities(self, service, selector):
        more_pages = True
        rows = []
        requests = 0
        offset = 0
        while more_pages:
            current_rows, total_no_of_entries = self.get_entities_of_single_page(service, selector)

            offset += self.PAGE_SIZE
            selector['paging']['startIndex'] = str(offset)
            more_pages = (offset < int(total_no_of_entries))
            rows.extend(current_rows)
            requests += 1

        return rows, requests

    def start(self):
        service = FetchService(scripts.adwords.CONFIG.ADWORDS_OAUTH).get_service(self.SERVICE_NAME, self._refresh_token, self._customer_account_id)
        offset = 0
        selector = {
            'fields': self.FIELDS,
            'paging': {
                'startIndex': str(offset),
                'numberResults': str(self.PAGE_SIZE)
            }
        }
        return self.get_entities(service, selector)
