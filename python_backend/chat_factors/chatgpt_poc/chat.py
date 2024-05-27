import openai
import json
import pandas as pd
import random
import pickle
import os
import numpy as np
from tornado.log import logging as log
from .bert import embed_sentence
from .bert import embed_sentences
from .chat_jobs.data_service import DataService
# from lib.data_services.factors_data_service import FactorsDataService
import app

directory_path = 'chat_factors/chatgpt_poc'
cache_data_path = os.path.join(directory_path, "data_cached.csv")
EMBEDDING_CACHE_PATH = os.path.join(directory_path, 'artifacts/prompt_emb_cache.pkl')

key_file_path = os.path.join('chat_factors/chatgpt_poc', 'key.json')

# Load the API key from the key.json file
openai.api_key = json.load(open(key_file_path, 'r'))['key']


def ask_ir_based_model(prompt, matching_examples, gpt_params=None):
    gpt_params = gpt_params or {'engine': 'gpt-3.5-turbo-instruct',
                                'max_tokens': 100,
                                'temperature': 0}
    final_prompt = f'{matching_examples}\n{prompt}->'
    answer = openai.Completion.create(
        engine=gpt_params['engine'],
        prompt=final_prompt,
        max_tokens=gpt_params['max_tokens'],
        n=1,
        stop=None,
        temperature=gpt_params['temperature'])
    returnables = {'answer': answer}
    return returnables


def get_answer_from_ir_model(question, prompt_response_data, prompt_vector_data):
    try:
        log.info('running get_answer_from_ir_model')
        if len(question) < 5:
            raise Exception('Your question should be at least 5 characters long.')
        prompt_response_data.seek(0)
        df = pd.read_csv(prompt_response_data)

        log.info('\n Getting matches using BERT embeddings...')
        matching_examples = get_matching_examples_from_file(question, df, prompt_vector_data)
        log.info("done step 1 \n matching_examples :\n %s", matching_examples)
        log.info('\nSeeking answer from GPT..')
        ir_response = ask_ir_based_model(question, matching_examples)
        answer = ir_response['answer']
        answer = answer['choices'][0]['text'].split('\n')[0].strip(' .')

    except openai.error.AuthenticationError:
        raise Exception('OpenAI API Key Error. Specify the right one via the key.json file.')
    except Exception as e:
        # Handle other exceptions
        log.error("Error processing request in : get_answer_from_ir_model")

    returnables = {'answer': answer}
    log.info("done step 2  \n response from gpt :%s", answer)
    return returnables


def get_answer_from_ir_model_local(question):
    try:
        log.info('running get_answer_from_ir_model_local')
        if len(question) < 5:
            raise Exception('Your question should be at least 5 characters long.')
        df = pd.read_csv(cache_data_path)
        indexed_prompts, indexed_prompt_embs = pickle.load(open(EMBEDDING_CACHE_PATH, 'rb'))
        prompt_emb = embed_sentence(question, normalise=True)

        sim_pe_all = np.matmul(prompt_emb, indexed_prompt_embs.transpose())
        top_k_i = np.argsort(sim_pe_all, axis=1)[:, -10:].reshape(-1)

        matching_prompts = [indexed_prompts[i] for i in top_k_i]
        matching_df = df[df['prompt'].isin(matching_prompts)]
        matching_examples = "\n".join(matching_df.apply(lambda x: f"{x['prompt']}-> {x['completion']}", axis=1))
        log.info("done step 1 \n matching_examples :\n%s", matching_examples)
        log.info('\nSeeking answer from GPT..')
        ir_response = ask_ir_based_model(question, matching_examples)
        answer = ir_response['answer']
        answer = answer['choices'][0]['text'].split('\n')[0].strip(' .')

    except openai.error.AuthenticationError:
        raise Exception('OpenAI API Key Error. Specify the right one via the key.json file.')
    except Exception as e:
        # Handle other exceptions
        log.error("Error processing request in get_answer_from_ir_model_local with error: %s", str(e))

    returnables = {'answer': answer}
    log.info("done step 2 \n response from gpt :%s", answer)
    return returnables


def get_answer_using_ir_model(self, question, embeddings):
    try:
        log.info('running get_answer_from_ir_model_local')

        # Check if the question is at least 5 characters long
        if len(question) < 5:
            raise Exception('Your question should be at least 5 characters long.')

        indexed_prompts = embeddings['indexed_prompts']
        indexed_queries = embeddings['indexed_queries']

        # Format matching examples directly from the provided embeddings
        matching_examples = "\n".join(
            [f"{prompt}-> {query}" for prompt, query in zip(indexed_prompts, indexed_queries)])

        log.info("done step 1 \n matching_examples :\n%s", matching_examples)
        log.info('\nSeeking answer from GPT..')

        # Query the IR-based model with the question and matching examples
        ir_response = ask_ir_based_model(question, matching_examples)
        answer = ir_response['answer']
        answer = answer['choices'][0]['text'].split('\n')[0].strip(' .')

    except openai.error.AuthenticationError:
        raise Exception('OpenAI API Key Error. Specify the right one via the key.json file.')
    except Exception as e:
        # Handle other exceptions
        log.error("Error processing request in get_answer_from_ir_model_local with error: %s", str(e))
        answer = None

    # Prepare the response
    returnables = {'answer': answer}
    log.info("done step 2 \n response from gpt :%s", answer)
    return returnables


def embed_prompts(df, sample_size=None):
    if df.empty:
        return [], [], []
    all_prompts = list(df['prompt'])
    if sample_size is None:
        indexed_prompts = all_prompts
    else:
        indexed_prompts = sorted(random.sample(all_prompts, sample_size))
    indexed_query = df[df['prompt'].isin(indexed_prompts)]['completion'].tolist()
    indexed_prompt_embs = embed_sentences(indexed_prompts, normalise=True)
    return indexed_prompts, indexed_prompt_embs, indexed_query


def get_matching_examples_from_file(query_prompt, df, prompt_vector_data):
    log.info('Attempting to read indexed prompt embeddings...')
    indexed_prompts, indexed_prompt_embs = prompt_vector_data
    log.info('Done getting indexed prompt embeddings!')
    prompt_emb = embed_sentence(query_prompt, normalise=True)

    sim_pe_all = np.matmul(prompt_emb, indexed_prompt_embs.transpose())
    top_k_i = np.argsort(sim_pe_all, axis=1)[:, -10:].reshape(-1)

    matching_prompts = [indexed_prompts[i] for i in top_k_i]
    matching_df = df[df['prompt'].isin(matching_prompts)]
    matching_examples = "\n".join(matching_df.apply(lambda x: f"{x['prompt']}-> {x['completion']}", axis=1))
    return matching_examples
