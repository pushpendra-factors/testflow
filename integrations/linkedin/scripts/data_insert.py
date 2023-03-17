import logging as log
from constants import *
from data_service import DataService

class DataInsert:
    
    def insert_metadata(options, doc_type, project_id, ad_account, response, timestamp, extraMeta):
        log.warning(INSERTION_LOG.format(doc_type, 'metadata', timestamp))
        for data in response:
            data.update(extraMeta[str(data['id'])])
        add_documents_response = DataService(options).add_all_linkedin_documents(project_id,
                                     ad_account, doc_type, response, timestamp)
        return add_documents_response


    
    def insert_insights(options, doc_type, project_id, ad_account, response, timestamp):
        log.warning(INSERTION_LOG.format(doc_type, 'insights', timestamp))
        if len(response) > 0:
            add_documents_response = DataService(options).add_all_linkedin_documents(project_id,
                                     ad_account, doc_type, response, timestamp)
            if not add_documents_response.ok and add_documents_response.status_code != 409:
                errString = DOC_INSERT_ERROR.format(doc_type,
                                     'insights',add_documents_response.status,
                                         add_documents_response.text,
                                             project_id, ad_account)
                log.error(errString)
                return errString
        log.warning(INSERTION_END_LOG.format(doc_type, 'insights', timestamp))
        return ''