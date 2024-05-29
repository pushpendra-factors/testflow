import sys
import re
from tornado import gen
#sys.path.append('/Users/satyamishra/repos/factors/python_backend/chat_factors/')
#sys.path.append('/Users/satyamishra/repos/factors/python_backend/')

import os
import json
import re
from .base_handler import BaseHandler
from tornado.log import logging as log
from chatgpt_poc.chat import get_answer_from_ir_model
from chatgpt_poc.chat import get_answer_from_ir_model_local, get_answer_using_ir_model
from chatgpt_poc.bert import embed_sentence
from chat.final_query import get_url_and_query_payload_from_gpt_response, validate_gpt_response, UnexpectedGptResponseError
from chat.kpi import KPIOrPropertyNotFoundError
from google.cloud import storage
from lib.data_services.factors_data_service import FactorsDataService
import io
import pickle
import app

os.environ['TF_CPP_MIN_LOG_LEVEL'] = '1'
os.environ['TOKENIZERS_PARALLELISM'] = 'false'


class ChatHandler(BaseHandler):
    _initialized = False

    @classmethod
    def initialize_variable(cls, value):
        if not cls._initialized:
            storage_client = storage.Client()
            bucket = storage_client.get_bucket(app.CONFIG.CHAT_BUCKET)
            blob = bucket.blob('chat/data_cached.csv')
            prompt_response_data_content = blob.download_as_text()
            cls.prompt_response_data = io.StringIO(prompt_response_data_content)
            # Download binary file (e.g., pickled file) as bytes
            blob_pkl = bucket.blob('chat/prompt_emb_cache.pkl')
            prompt_vector_data_content = blob_pkl.download_as_bytes()
            cls.prompt_vector_data = pickle.load(io.BytesIO(prompt_vector_data_content))
            cls._initialized = True
            log.info("Variable initialized successfully.")
        else:
            log.info("Variable already initialized. Skipping.")

    @gen.coroutine
    def post(cls):
        try:
            prompt = json.loads(cls.request.body)["prompt"]
            pid = json.loads(cls.request.body)["pid"]
            kpi_config = json.loads(cls.request.body)["kpi_config"]
            log.info('prompt: %s', prompt)
            prompt_emb = embed_sentence(prompt, normalise=True)
            matching_embeddings_data =  FactorsDataService.get_matching_chat_embeddings(prompt_emb)
            result = get_answer_using_ir_model(cls, prompt, matching_embeddings_data['data'])
            # if app.CONFIG.ADWORDS_APP.env == "development":
            #     result = get_answer_from_ir_model_local(cls, prompt)
            # elif app.CONFIG.ADWORDS_APP.env == "staging" or app.CONFIG.ADWORDS_APP.env == "production":
            #     ChatHandler.initialize_variable("")
            #     result = get_answer_from_ir_model(prompt, cls.prompt_response_data, cls.prompt_vector_data)

            json_string = result["answer"]
            result_dict = json.loads(json_string)

            validate_gpt_response(result_dict)

            query_payload_and_url = get_url_and_query_payload_from_gpt_response(result_dict, pid, kpi_config)
            log.info("done step 4 \n query_payload_and_url :%s", query_payload_and_url)
            result_json = json.dumps(query_payload_and_url, indent=2)

            cls.write(result_json)

        except KPIOrPropertyNotFoundError as kpnfe:
            # Handle kpi not found error here
            log.error("CustomProcessingError processing request: %s", str(kpnfe))
            cls.set_status(400)  # Bad Request
            cls.write(json.dumps({'error': {'code': 400, 'message': str(kpnfe)}}))
        except Exception as e:
            # Handle other exceptions
            log.error("Error processing request: %s", str(e))
            cls.set_status(500)  # Internal Server Error
            cls.write(json.dumps({'error': {'code': 500, 'message': "Internal Server Error"}}))

    @gen.coroutine
    def options(self):
        self.set_status(200)
        self.finish()
