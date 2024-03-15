import json
from itertools import product
import os
from tqdm import tqdm
import pandas as pd
DATA_CACHE_FILE = 'data_cached.csv'
directory_path = 'chat_factors/chatgpt_poc'



def get_query_templates():
    query_templates = {
        'uni_metric': {
            'timeless': ['%s'],
            'timeful': [
                'Number of %s in %s',
                'How many %s we had %s',
                'Count of %s in %s'
                # 'How many %s did we have %s',
                # 'How many %s visited our website %s'
                ]
            },
        'bi_metric': {
            'timeless': ['%s and %s'],
            'timeful': ['What\'s the %s and %s in %s']
            },
        'breakdown': {
            'timeless': ['%s by %s'],
            'timeful': ['What\'s the breakdown of %s by %s in %s']
            },
        'funnel': {
            'timeless': ['Conversion rate from %s to %s'],
            'timeful': [
                # '%s that led to %s',
                'What\'s the conversion rate from %s to a %s in %s'
                ]
        }
    }
    return query_templates


def get_json_templates():
    json_templates = {
        'uni_metric': 
            '{"query_type": "kpi", "query_entity_1": "%s", \
              "query_filter_1":"none", "query_breakdown_1": "none", \
              "time_range": "%s", "start_time": "default", "end_time": "default"}',
        'bi_metric':
            '{"query_type": "kpi", "query_entity_1": "%s", \
              "query_filter_1":"none", "query_breakdown_1": "none", \
              "query_entity_2": "%s", \
              "query_filter_2":"none", "query_breakdown_2": "none",\
              "time_range": "%s", "start_time": "default", "end_time": "default"}',
        'breakdown':
            '{"query_type": "kpi", "query_entity_1": "%s", \
              "query_filter_1":"none", "query_breakdown_1": "%s", \
              "time_range": "%s", "start_time": "default", "end_time": "default"}',
        'funnel':
            '{"query_type": "funnel", "query_entity_1": "%s", \
              "query_filter_1":"none", "query_breakdown_1": "none", \
              "query_entity_2": "%s", \
              "query_filter_2":"none", "query_breakdown_2": "none",\
              "time_range": "%s", "start_time": "default", "end_time": "default"}'
        }
    return json_templates


def get_time_specifiers():
    more_times = list(map(lambda x: ' '.join(x), product(['this', 'last'], ['month', 'week', 'year']))) + ['today', 'yesterday']
    less_times = list(map(lambda x: ' '.join(x), product(['this'], ['month']))) + ['today', 'yesterday']
    return less_times


def replace_two_by_one(x, key='qe'):
    k1 = f'{key}1'
    k2 = f'{key}2'
    if k1 in x:
        v1 = x[k1]
        del x[k1]
        if k2 in x:
            v2 = x[k2]
            del x[k2]
            x[key] = f"({v1}, {v2})"
        else:
            x[key] = v1

def replace_one_by_two(x, key='qe'):
    k1 = f'{key}1'
    k2 = f'{key}2'
    if key in x:
        v = x[key]
        del x[key]
        try:
            v1, v2 = v
            x[k1] = v1
            x[k2] = v2
        except ValueError:
            x[k1] = v


def get_reduction_map():
    reduction_map = {'query_type': 'qt',
                 'query_entity_1': 'qe1',
                 'query_filter_1': 'qf1',
                 'query_breakdown_1': 'qb1',
                 'query_entity_2': 'qe2',
                 'query_filter_2': 'qf2',
                 'query_breakdown_2': 'qb2',
                 'time_range': 'time',
                 'start_time': 'st',
                 'end_time': 'et',
                 'default': '-',
                 'none': '-'}
    return reduction_map


def reduce_completion(x):
    reduction_map = get_reduction_map()
    for a, b in reduction_map.items():
        x = x.replace(a, b)
    x = json.loads(x)
    for k in ['qe', 'qf', 'qb']:
        replace_two_by_one(x, key=k)
    del x['st'], x['et']
    x = json.dumps(x)
    x = x.replace('"', '')
    x = x.replace("'", '')
    x = x.replace(' ', '')
    x = x.replace('$', '')
    return x


def expand_completion(x):
    # TODO: Expand better!
    reduction_map = get_reduction_map()
    expansion_map = {v:k for k, v in reduction_map.items()}
    # x = json.loads(x)
    # for k in ['qe', 'qf', 'qb']:
    #     replace_one_by_two(x, key=k)
    # x = json.dumps(x)
    for a, b in expansion_map.items():
        x = x.replace(a, b)
    return x

def abbreviate_data(df):
    df['completion'] = df['orig_completion'].apply(reduce_completion)


def get_prepared_data(raw_data_path=os.path.join('chat_factors/chatgpt_poc', 'data.json'), abbreviate=True, cache_path=DATA_CACHE_FILE, force_prepare=False):
    try:
        if force_prepare:
            raise FileNotFoundError
        df = pd.read_csv(cache_path)
    except FileNotFoundError:
        df = prepare_data(raw_data_path, abbreviate)
        df.to_csv(cache_path)
    return df

def prepare_data(raw_data_path='data.json', abbreviate=True):
    raw_data = json.load(open(raw_data_path, 'r'))
    qts_map = get_query_templates()
    jts_map = get_json_templates()
    times = get_time_specifiers()
    # TODO: To be implemented for more keys (other than `website_session`) as well.
    metrics_map = raw_data['website_session']['metrics']
    dimensions_map = raw_data['website_session']['dimensions']

    qjs = []
    # UNI-METRIC:
    for k, v in tqdm(metrics_map.items()):
        for qt in qts_map['uni_metric']['timeless']:
            query = qt % k
            _json = jts_map['uni_metric'] % (v, 'default')
            qj = (query, _json)
            qjs.append(qj)
        for qt in qts_map['uni_metric']['timeful']:
            for t in times:
                query = qt % (k, t)
                _json = jts_map['uni_metric'] % (v, t.replace(' ', '_'))
                qj = (query, _json)
                qjs.append(qj)

    # BREAKDOWN:
    for km, vm in tqdm(metrics_map.items()):
        for kd, vd in dimensions_map.items():
            for qt in qts_map['breakdown']['timeless']:
                query = qt % (km, kd)
                _json = jts_map['breakdown'] % (vm, vd, 'default')
                qj = (query, _json)
                qjs.append(qj)
            for qt in qts_map['breakdown']['timeful']:
                for t in times:
                    query = qt % (km, kd, t)
                    _json = jts_map['breakdown'] % (vm, vd, t.replace(' ', '_'))
                    qj = (query, _json)
                    qjs.append(qj)

    # BI-METRIC:
    for k1, v1 in tqdm(metrics_map.items()):
        for k2, v2 in (metrics_map.items()):
            if k1==k2:
                continue
            for qt in qts_map['bi_metric']['timeless']:
                query = qt % (k1, k2)
                _json = jts_map['bi_metric'] % (v1, v2, 'default')
                qj = (query, _json)
                qjs.append(qj)
            for qt in qts_map['bi_metric']['timeful']:
                for t in times:
                    query = qt % (k1, k2, t)
                    _json = jts_map['bi_metric'] % (v1, v2, t.replace(' ', '_'))
                    qj = (query, _json)
                    qjs.append(qj)

    # FUNNEL:
    for k1, v1 in tqdm(metrics_map.items()):
        for k2, v2 in (metrics_map.items()):
            if k1==k2:
                continue
            for qt in qts_map['funnel']['timeless']:
                query = qt % (k1, k2)
                _json = jts_map['funnel'] % (v1, v2, 'default')
                qj = (query, _json)
                qjs.append(qj)
            for qt in qts_map['funnel']['timeful']:
                for t in times:
                    query = qt % (k1, k2, t)
                    _json = jts_map['funnel'] % (v1, v2, t.replace(' ', '_'))
                    qj = (query, _json)
                    qjs.append(qj)
    df = pd.DataFrame(qjs, columns=['prompt', 'orig_completion'])
    if abbreviate:
        abbreviate_data(df)
    return df
