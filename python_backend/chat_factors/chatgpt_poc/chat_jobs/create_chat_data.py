import sys
import os

directory_path = 'chat_factors/chatgpt_poc'
cache_data_path = os.path.join(directory_path, "data_cached.csv")
EMBEDDING_CACHE_PATH = 'artifacts/prompt_emb_cache.pkl'
emb_pkl_path = os.path.join(directory_path, EMBEDDING_CACHE_PATH)

from optparse import OptionParser
from data_service import DataService
import logging as log
from google.cloud import storage
import pickle
import io
import os
from tornado.log import logging as log
# sys.path.append('/Users/satyamishra/repos/factors/python_backend/chat_factors/')


from chatgpt_poc.chat import embed_prompts
from chatgpt_poc.data_preparer import prepare_data, remove_data_with_prompt

parser = OptionParser()
parser.add_option('--env', dest='env', default='development')
parser.add_option('--chat_bucket', dest='chat_bucket', default="")
parser.add_option("--developer_token", dest="developer_token", help="", default="")
parser.add_option("--oauth_secret", dest="oauth_secret", help="", default="")
parser.add_option('--data_service_host', dest='data_service_host',
                  help='Data service host', default='http://localhost:8089')
parser.add_option('--scratch', default='False', dest='scratch', help='', type=str)


def uploadFileToGCP(chat_bucket, file_data, destination_blob_name):
    bucket_name = chat_bucket
    print("got bucket")
    storage_client = storage.Client()
    bucket = storage_client.bucket(bucket_name)
    blob = bucket.blob(destination_blob_name)
    generation_match_precondition = 0
    blob.upload_from_file(file_data, content_type='text/csv')

    log.info("file uploaded to GCP")


def uploadStringAsFileInGCP(chat_bucket, file_data, destination_blob_name):
    bucket_name = chat_bucket
    print("got bucket")
    storage_client = storage.Client()
    bucket = storage_client.bucket(bucket_name)
    blob = bucket.blob(destination_blob_name)
    generation_match_precondition = 0
    blob.upload_from_string(file_data, content_type='application/octet-stream')
    log.info("file uploaded to GCP", )


def get_chat_data(exclude_prompts=None):
    if exclude_prompts is None:
        exclude_prompts = []
    dataframe = prepare_data(os.path.join('chat_factors/chatgpt_poc', 'data.json'), abbreviate=True)
    dataframe = remove_data_with_prompt(dataframe, exclude_prompts)
    idx_prompts, idx_prompt_embs, indexed_query = embed_prompts(dataframe)
    return idx_prompts, idx_prompt_embs, indexed_query


if __name__ == '__main__':
    try:
        (options, args) = parser.parse_args()
        data_service = DataService(options)

        if options.scratch == "False":
            db_prompts_response = data_service.get_db_prompts(0)
            db_prompts = db_prompts_response["data"]
            indexed_prompts, indexed_prompt_embs, indexed_query = get_chat_data(db_prompts)
            response = data_service.add_chat_embeddings_new(indexed_prompts, indexed_prompt_embs, indexed_query)

        else:
            log.warning("Successfully created files locally. End of create chat data job.")
            indexed_prompts, indexed_prompt_embs, indexed_query = get_chat_data()
            response = data_service.add_chat_embeddings_scratch(indexed_prompts, indexed_prompt_embs, indexed_query)

        if not response.ok:
            log.error("Failed to add chat embeddings")

    #
    #     if options.env == "development":
    #         # Store CSV content and pickle content locally for development
    #         dataframe.to_csv(cache_data_path)
    #         pickle.dump((indexed_prompts, indexed_prompt_embs), open(emb_pkl_path, 'wb'))
    #         log.warning("Successfully created files locally. End of create chat data job.")
    #
    #     elif options.env == "staging" or options.env == "production":
    #         # Upload files to GCP in staging or production environment
    #         csv_content = io.StringIO()
    #         dataframe.to_csv(csv_content, index=False)
    #         csv_content.seek(0)
    #         uploadFileToGCP(options.chat_bucket, csv_content, "chat/data_cached.csv")
    #         pkl_content = io.BytesIO()
    #         pickle.dump((indexed_prompts, indexed_prompt_embs), pkl_content)
    #         pkl_content.seek(0)
    #         uploadStringAsFileInGCP(options.chat_bucket, pkl_content.getvalue(), "chat/prompt_emb_cache.pkl")
    #         log.warning("Successfully created files on GCP. End of create chat data job.")
    #
    #     else:
    #         log.warning("Files not created. Incorrect environment")
    except Exception as e:
        log.error("Error processing request: %s", str(e))

    sys.exit(0)
