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
parser.add_option('--mode', default='', dest='mode', help='', type=str)
parser.add_option('--project_id', default='0', dest='project_id', help='', type=int)


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


def create_and_add_missing_chat_data(project_id):
    dataframe = prepare_data(os.path.join('chat_factors/chatgpt_poc', 'data.json'), abbreviate=True)

    batch_size = 100

    # Extract prompts from the dataframe
    all_prompts = dataframe['prompt'].tolist()

    for i in range(0, len(all_prompts), batch_size):
        batch_prompts = all_prompts[i:i + batch_size]
        missing_prompts_response = data_service.get_missing_prompts(project_id, batch_prompts)
        missing_prompts = missing_prompts_response["data"]
        if missing_prompts:

            missing_df = dataframe[dataframe['prompt'].isin(missing_prompts)]

            idx_prompts, idx_prompt_embs, indexed_query = embed_prompts(missing_df)

            response = data_service.add_chat_embeddings(project_id, idx_prompts, idx_prompt_embs, indexed_query)

            if not response.ok:
                log.error("Failed to add chat embeddings for batch number %d", i // batch_size + 1)
                return response
            log.info("Successfully added embeddings for batch number %d", i // batch_size + 1)
        else:
            log.info("Embeddings for batch number %d is already present in db", i // batch_size + 1)

    return


def delete_and_add_chat_data(project_id):
    # Delete all existing chat data for the given project ID
    delete_response = data_service.delete_chat_data(project_id)
    if not delete_response:
        log.error("Failed to delete chat data for project ID %s", project_id)
        return delete_response
    log.info("Successfully deleted all chat data for project ID %s", project_id)

    # Prepare the data
    dataframe = prepare_data(os.path.join('chat_factors/chatgpt_poc', 'data.json'), abbreviate=True)

    batch_size = 100

    # Add new chat embeddings in batches
    for i in range(0, len(dataframe), batch_size):
        batch_df = dataframe.iloc[i:i + batch_size]

        # Embed prompts
        idx_prompts, idx_prompt_embs, indexed_query = embed_prompts(batch_df)

        # Add the embeddings
        response = data_service.add_chat_embeddings(project_id, idx_prompts, idx_prompt_embs, indexed_query)
        if not response.ok:
            log.error("Failed to add chat embeddings for batch number %d", i // batch_size + 1)
            return response
        log.info("Successfully added embeddings for batch number %d", i // batch_size + 1)

    return


if __name__ == '__main__':
    try:
        (options, args) = parser.parse_args()
        data_service = DataService(options)

        if options.mode == "scratch":
            delete_and_add_chat_data(options.project_id)

        else:
            create_and_add_missing_chat_data(options.project_id)

    except Exception as e:
        log.error("Error processing request: %s", str(e))

    log.info("Successfully completed chat data job")

    sys.exit(0)
