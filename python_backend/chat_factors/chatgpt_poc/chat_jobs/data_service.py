import requests
import logging as log
import numpy as np


class DataService:
    data_service_host = ''

    def __init__(self, options):
        self.data_service_host = options.data_service_host

    def add_chat_embeddings_scratch(self, project_id, indexed_prompts, indexed_prompt_embs, indexed_queries):
        uri = '/data_service/chat/job/scratch'

        url = self.data_service_host + uri

        # Convert NumPy arrays to lists
        indexed_prompt_embs = [embedding.tolist() if isinstance(embedding, np.ndarray) else embedding
                               for embedding in indexed_prompt_embs]

        payload = {
            'project_id': project_id,
            'indexed_prompts': indexed_prompts,
            'indexed_prompt_embs': indexed_prompt_embs,
            'indexed_queries': indexed_queries

        }

        response = requests.post(url, json=payload)
        if not response.ok:
            log.error("Failed to add chat embeddings")

        return response

    def add_chat_embeddings(self, project_id, indexed_prompts, indexed_prompt_embs, indexed_queries):
        uri = '/data_service/chat/job'

        url = self.data_service_host + uri

        # Convert NumPy arrays to lists
        indexed_prompt_embs = [embedding.tolist() if isinstance(embedding, np.ndarray) else embedding
                               for embedding in indexed_prompt_embs]

        payload = {
            'indexed_prompts': indexed_prompts,
            'indexed_prompt_embs': indexed_prompt_embs,
            'indexed_queries': indexed_queries,
            'project_id': project_id

        }

        response = requests.post(url, json=payload)
        if not response.ok:
            log.error("Failed to add chat embeddings")

        return response

    def get_db_prompts(self, project_id):

        uri = '/data_service/chat/job'

        url = self.data_service_host + uri

        payload = {
            'project_id': project_id,

        }

        response = requests.get(url, json=payload)
        if not response.ok:
            log.error("Failed to add chat embeddings")

        try:
            prompts = response.json()  # Assuming the response is JSON
            return prompts
        except ValueError:
            log.error("Failed to decode chat embeddings response")
            return None

    def get_missing_prompts(self, project_id, indexed_prompts):
        uri = '/data_service/chat/job/missing'

        url = self.data_service_host + uri

        payload = {
            'indexed_prompts': indexed_prompts,
            'project_id': project_id
        }

        response = requests.get(url, json=payload)
        if not response.ok:
            log.error("Failed to add chat embeddings")

        try:
            prompts = response.json()  # Assuming the response is JSON
            return prompts
        except ValueError:
            log.error("Failed to decode chat embeddings response")
            return None

    def delete_chat_data(self, project_id):
        uri = '/data_service/chat/job'

        url = self.data_service_host + uri

        payload = {
            'project_id': project_id,
        }

        response = requests.delete(url, json=payload)
        if not response.ok:
            log.error("Failed to delete chat data for project ID %s", project_id)
            return None

        try:
            result = response.json()  # Assuming the response is JSON
            return result
        except ValueError:
            log.error("Failed to decode delete chat data response for project ID %s", project_id)
            return None




