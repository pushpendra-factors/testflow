import openai
import json
import pandas as pd
from pprint import pprint
from functools import reduce
from collections import defaultdict
import bert
import torch
import pdb
import random
import pickle
from datetime import datetime
from data_preparer import get_prepared_data, expand_completion, reduce_completion

EMBEDDING_CACHE_PATH = 'prompt_emb_cache.pkl'

# Set up the ChatGPT API credentials
openai.api_key = json.load(open('key.json', 'r'))['key']

def retrain_fine_tuned_model():
    df = get_prepared_data()
    tr_data = list(df[['prompt', 'completion']].T.to_dict().values())
    for i in range(len(tr_data)):
        tr_data[i]['prompt'] = tr_data[i]['prompt'] + ' ->'
        tr_data[i]['completion'] = ' ' + tr_data[i]['completion'].replace(': ', ':').strip() + '.\n'
    file_name = "training_data.jsonl"

    with open(file_name, "w") as output_file:
        for entry in tr_data:
            json.dump(entry, output_file)
            output_file.write("\n")
    
    upload_response = openai.File.create(
    file=open(file_name, "rb"),
    purpose='fine-tune'
    )
    file_id = upload_response.id
    ft_list = [(x['id'], x['fine_tuned_model'], datetime.fromtimestamp(x['created_at']).strftime('%Y-%m-%d %H:%M:%S')) for x in openai.FineTune.list()['data']]
    model = [x for x in ft_list if x[0] == file_id][0][1]
    return model


def read_one_shot_training_data(data_path="chatgpt_training_data_v1.0.tsv"):
    df = pd.read_csv(data_path, sep='\t')
    df.columns = ['question', 'project_dashboard', 'result', 'query', 'concat']
    return df


def form_examples(df):
    examples = "\n".join(df.apply(lambda x: f"{x['question']}: {x['result']}", axis=1))
    return examples


def form_prelude():
    return '''For query_entity_i for website_session, use the following map:
            'New Users' to 'new_users'
            'Repeat Users' to 'repeat_users'

            For query_breakdown_j for query_filter_i for website_session, use the following map:
            'Source' to '$source'
            'Medium' to '$medium'

            For query_entity_i for form_submission, use the following map:
            'Count' to 'count'
            'Unique users' to 'unique_users'

            For query_breakdown_j for query_filter_i for form_submission, use the following map:
            'Referrer URL' to '$referrer_url'
            'Page URL' to '$page_url'

            For query_entity_i for hubspot_contacts, use the following map:
            'Leads (Interested In Trial)' to 'Leads (Interested In Trial)'
            'Leads (Demo Scheduled)' to 'Leads (Demo Scheduled)'

            For query_breakdown_j for query_filter_i for hubspot_contacts, use the following map:
            'Latest Term' to '$latest_term'
            'Initial Page Domain' to '$initial_page_domain'

            For query_breakdown_j for query_filter_i for hubspot_companies, use the following map:
            'Hubspot Company Number Of Contacts With A Buying Role' to '$hubspot_company_hs_num_contacts_with_buying_roles'
            'Hubspot Company Number Of Blockers' to '$hubspot_company_hs_num_blockers'

            For query_entity_i for hubspot_deals, use the following map:
            'Deals' to 'Deals'
            'Pipeline' to 'Pipeline'

            For query_breakdown_j for query_filter_i for hubspot_deals, use the following map:
            'Hubspot Deal Create Date' to '$hubspot_deal_createdate'
            'Hubspot Deal Last Modified Date' to '$hubspot_deal_hs_lastmodifieddate'

            For query_entity_i for google_ads_metrics, use the following map:
            'Impressions' to 'impressions'
            'Clicks' to 'clicks'

            For query_breakdown_j for query_filter_i for google_ads_metrics, use the following map:
            'Campaign Id' to 'campaign_id'
            'Campaign Name' to 'campaign_name'

            For query_entity_i for google_organic_metrics, use the following map:
            'Click through rate' to 'click_through_rate'
            'Position Avg' to 'position_avg'

            For query_breakdown_j for query_filter_i for google_organic_metrics, use the following map:
            'Organic Property Device' to 'organic_property_device'
            'Organic Property Query' to 'organic_property_query'

            For query_entity_i for linkedin_metrics, use the following map:
            'Impressions' to 'impressions'
            'Clicks' to 'clicks'

            For query_breakdown_j for query_filter_i for linkedin_metrics, use the following map:
            'Campaign Group name' to 'campaign_name'
            'Campaign Group id' to 'campaign_id'

            For query_entity_i for linkedin_company_engagements, use the following map:
            'Impressions' to 'impressions'
            'Clicks' to 'clicks'

            For query_breakdown_j for query_filter_i for linkedin_company_engagements, use the following map:
            'Company Vanity Name' to 'company_vanity_name'
            'Company Preferred Country' to 'company_preferred_country'

            For query_entity_i for all_channels_metrics, use the following map:
            'Impressions' to 'impressions'
            'Clicks' to 'clicks'

            For query_breakdown_j for query_filter_i for all_channels_metrics, use the following map:
            'Channel Name' to 'channel_name'
            'Campaign Name' to 'campaign_name'

            For query_entity_i for bingads_metrics, use the following map:
            'Impressions' to 'impressions'
            'Clicks' to 'clicks'

            For query_breakdown_j for query_filter_i for bingads_metrics, use the following map:
            'Campaign Type' to 'campaign_type'
            'Campaign Name' to 'campaign_name'

            For query_entity_i for page_views, use the following map:
            'Exits' to 'exits'
            'Page Views' to 'page_views'

            For query_breakdown_j for query_filter_i for page_views, use the following map:
            'Referrer URL' to '$referrer_url'
            'Page URL' to '$page_url'

            For query_entity_i for event_based, use the following map:
            'Test Custom Event Order Unit Price' to 'Test Custom Event Order Unit Price'
            'Test Custom Event Order Total Price' to 'Test Custom Event Order Total Price'

            For query_entity_i for others, use the following map:
            'Cost Per User (Paid Search)' to 'Cost Per User (Paid Search)'
            'Pipeline per deal- test' to 'Pipeline per deal- test'
            '''


def form_prelude_old(df):
    keys = df['result'].apply(json.loads).apply(set)
    values = df['result'].apply(json.loads).apply(lambda x: set(x.values()))
    key_values = df['result'].apply(json.loads).to_list()
    all_keys = reduce(lambda x, y: x | y, keys)
    all_values = reduce(lambda x, y: x | y, values)
    all_key_values = defaultdict(set)
    for kvs in key_values:
        for k, v in kvs.items():
            if k.endswith('_1') or k.endswith('_2'):
                k = k[:-1] + 'i'
            all_key_values[k].add(v)
    all_key_values

    json_keys_str = "\n".join([f"K{i+1}. '{k}'" for i, k in enumerate(sorted(all_keys))])
    prelude = f'Allowed JSON keys are K1--K{len(all_keys)}, and range of values for key Ki are Vi.1--Vi.ni, where ni is the number of allowed values of key Ki. If you don\'t find any appropriate key or value, return the whole answer as NA (with reason included):'

    json_kvs_str = '\n'.join([f"K{i+1}. '{k}'\n\t{', '.join(['V' + str(i+1) + '.' + str(j+1) + '. ' + v for j, v in enumerate(sorted(all_key_values[k[:-1] + 'i' if k.endswith('_1') or k.endswith('_2') else k]))])}" for i, k in enumerate(sorted(all_keys))])

    prelude += '\n'+json_kvs_str
    return prelude


def ask_gpt(question=None, prepend_question=False):
    ft_model = json.load(open('ft_model.json', 'r'))['fine_tuned_model']
    if len(question) < 5:
        answer = 'ERROR: Your question should be at least 5 characters long.'
    try:
        answer = ask_fine_tuned_model(prompt=question, fine_tuned_model=ft_model)
        answer = answer['choices'][0]['text'].split('\n')[0].strip(' .')
    except openai.error.AuthenticationError:
        answer = 'ERROR: OpenAI API Key Error. Specify the right one via the key.json file.'
    if prepend_question:
        return f"Q: {question}<br>A: {answer}<br><br>"
    else:
        return answer


def ask_fine_tuned_model(prompt, fine_tuned_model, max_tokens=100, temperature=0, return_prompt=False):
    final_prompt = prompt + ' ->'
    answer = openai.Completion.create(
                model=fine_tuned_model,
                prompt=final_prompt,
                max_tokens=max_tokens,
                temperature=temperature
                )
    returnables = {'answer': answer}
    if return_prompt:
        returnables['prompt'] = final_prompt
    return returnables


def get_ir_params():
    tokenizer = bert.get_tokenizer()
    model = bert.get_model()
    return {'tokenizer': tokenizer,
            'model': model}


def embed_prompts(df, sample_size=None):
    all_prompts = list(df['prompt'])
    if sample_size is None:
        indexed_prompts = all_prompts
    else:
        indexed_prompts = sorted(random.sample(all_prompts, sample_size))
    indexed_prompt_embs = bert.embed_sentences(indexed_prompts, normalise=True)
    return indexed_prompts, indexed_prompt_embs


def get_indexed_prompt_embeddings(df, sample_size=None, cache_path=EMBEDDING_CACHE_PATH, silent=False, force_index=False):    
    try:
        if force_index:
            raise FileNotFoundError
        if not silent:
            print('Attempting to read indexed prompt embeddings...', end='')
        prompts, embs = pickle.load(open(cache_path, 'rb'))
        if not silent:
            print('Done!')
    except FileNotFoundError:
        if not silent:
            print('\nCached embeddings not found or indexing forced. Generating from scratch...', end='')
        prompts, embs = embed_prompts(df, sample_size)
        if not silent:
            print('Done!')
            print('Caching now...', end='')
        pickle.dump((prompts, embs), open(cache_path, 'wb'))
        if not silent:
            print('Done!')
    return prompts, embs
    

def get_matching_examples(query_prompt, df, silent=False, force_index=False):
    # TODO: Optimize matching
    # pdb.set_trace()
    indexed_prompts, indexed_prompt_embs = get_indexed_prompt_embeddings(df, sample_size=100, silent=silent, force_index=force_index)
    prompt_emb = bert.embed_sentence(query_prompt, normalise=True)
    sim_pe_all = torch.mm(prompt_emb, indexed_prompt_embs.transpose(0, 1))
    top_k_i = sim_pe_all.topk(10).indices.reshape(-1).numpy()
    matching_prompts = [indexed_prompts[i] for i in top_k_i]
    matching_df = df[df['prompt'].isin(matching_prompts)]
    matching_examples = "\n".join(matching_df.apply(lambda x: f"{x['prompt']}-> {x['completion']}", axis=1))
    return matching_examples


def ask_ir_based_model(prompt, matching_examples, gpt_params=None, return_prompt=False):
    gpt_params = gpt_params or {'engine': 'text-davinci-003',
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
    if return_prompt:
        returnables['prompt'] = final_prompt
    return returnables


def chat_loop_mode():
    ft_model = json.load(open('ft_model.json', 'r'))['fine_tuned_model']
    continue_flag = True
    while continue_flag:
        print('\nKeep asking questions. Enter `q` or Press Ctrl+C to exit.')
        question = input('Q: ')
        if question == 'q' or question == 'quit':
            break
        if len(question) < 5:
            answer = 'Your question should be at least 5 characters long.'
        else:
            try:
                # answer = ask_gpt(examples=examples, question=question, prelude=prelude, postface=postface, verbosity=1)
                answer = ask_fine_tuned_model(prompt=question, fine_tuned_model=ft_model)
            except openai.error.AuthenticationError:
                openai.api_key = input('API Key Error. Enter correct key: ')
                # answer = ask_gpt(examples=examples, question=question, prelude=prelude, postface=postface, verbosity=1)
                answer = ask_fine_tuned_model(prompt=question, fine_tuned_model=ft_model)
            answer = answer['choices'][0]['text'].split('\n')[0].strip(' .')
        print('A: ', end='')
        try:
            pprint(json.loads(answer))
        except:
            print(answer)


def get_ft_model(retrain=False):
    ft = json.load(open('ft_model.json', 'r'))
    if retrain:
        rt_model = retrain_fine_tuned_model()
        ft['latest'] = rt_model
        ft['historical'].append(rt_model)
        json.dump(open('ft_model.json', 'w'))
    model = ft['latest']
    return model


def chat_once_mode(question, model_type='ft', parser=None, scratch=False, silent=False, return_answer=False, return_prompt=False, reduce=True):
    if len(question) < 5:
        if parser is not None:
            parser.print_help()
        raise Exception('Your question should be at least 5 characters long.')
    returned_prompt = None
    try:
        if model_type == 'ft':
            if not silent:
                print('\nSTEP 1 of 2: Fetching fine-tuned model...', end='')
            ft_model = get_ft_model(scratch)
            if not silent:
                print('Done!\n')
                print('\nSTEP 2 of 2: Seeking answer from the fine-tuned model...', end='')
            ft_response = ask_fine_tuned_model(prompt=question, fine_tuned_model=ft_model, return_prompt=return_prompt)
            answer = ft_response['answer']
            if 'prompt' in ft_response:
                returned_prompt = ft_response['prompt']
            answer = answer['choices'][0]['text'].split('\n')[0].strip(' .')
            if reduce:
                answer = reduce_completion(answer)
            if not silent:
                print('Done!\n')
        elif model_type == 'ir':
            if not silent:
                print('\nSTEP 1 of 3: Getting prepared data (raw text prompt-responses)...', end='')
            df = get_prepared_data(force_prepare=scratch)
            if not silent:
                print('Done!\n')
                print('\nSTEP 2 of 3: Retrieving matches using BERT embeddings...', end='')
            matching_examples = get_matching_examples(question, df, silent, force_index=scratch)
            if not silent:
                print('Done!\n')
                print('\nSTEP 3 of 3: Seeking answer from GPT...', end='')
            ir_response = ask_ir_based_model(question, matching_examples, return_prompt=return_prompt)
            answer = ir_response['answer']
            if 'prompt' in ir_response:
                returned_prompt = ir_response['prompt']
            answer = answer['choices'][0]['text'].split('\n')[0].strip(' .')
            # answer = expand_completion(answer)
            if not silent:
                print('\nDone!\n')
    except openai.error.AuthenticationError:
        parser.print_help()
        raise Exception('OpenAI API Key Error. Specify the right one via the key.json file.')
    if return_answer or return_prompt:
        returnables = {}
        if return_answer:
            returnables['answer'] = answer
        if return_prompt:
            returnables['prompt'] = returned_prompt
        return returnables
    try:
        pprint(json.loads(answer))
    except:
        print(answer)
