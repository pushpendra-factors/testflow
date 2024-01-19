import sys
sys.path.append('/Users/satyamishra/repos/factors/python_backend/chat_factors/')
import os
import json
from .base_handler import BaseHandler
from tornado.log import logging as log
from chatgpt_poc.chat import get_answer_from_ir_model
from chatgpt_poc.chat import get_answer_from_ir_model_local
from google.cloud import storage
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

    def post(self):
        try:
            prompt = self.get_argument("prompt")
            log.info('prompt: %s', prompt)
            if app.CONFIG.ADWORDS_APP.env == "development":
                result = get_answer_from_ir_model_local(prompt)
            elif app.CONFIG.ADWORDS_APP.env == "staging" or app.CONFIG.ADWORDS_APP.env == "production":
                 ChatHandler.initialize_variable("")
                 result = get_answer_from_ir_model(prompt, self.prompt_response_data, self.prompt_vector_data)

            result_json = json.dumps(result, indent=2)
            log.info('Result_1:', result)
            self.write(result_json)
        except Exception as e:
            log.error("Error processing request: %s", str(e))
            self.set_status(500)  # Internal Server Error
            self.write(json.dumps({'error': {'code': 500, 'message': "Internal Server Error"}}))
