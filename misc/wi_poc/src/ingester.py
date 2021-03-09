from smart_open import smart_open
from tqdm import tqdm
import json
from wi_poc.src.feature_processor import flatten_epr_upr, get_criteria_filter
from wi_poc.src.preprocessor import sanitize_screen_size_params
import pandas as pd
from wi_poc.src.utils import merge_dict
import os
import time
from wi_poc.src.defaults import DEFAULT_DEDUPLICATE_LOGIC, DEFAULT_PROJECT_ID,\
    DEFAULT_MODEL_ID, DEFAULT_BASE, IZOOTO, DEFAULT_MERGE_UPR_LOGIC
from wi_poc.src.config import DEFAULT_CLOUD_PATH

def read_and_parse_weekly_data_without_upr_merge(events_file_path,
                               n_lines=None,
                               flatten_properties=True,
                               deduplicate_logic=DEFAULT_DEDUPLICATE_LOGIC):
    events_file_handle = smart_open(events_file_path, 'r')
    series_list = []
    pbar = tqdm(total=n_lines, desc='Read events') if n_lines else tqdm(desc='Read events')
    while True:
        pbar.update()
        line = events_file_handle.readline()
        if not line:
            break
        line = json.loads(line)
        if flatten_properties:
            line = flatten_epr_upr(line, deduplicate_logic)
        series_list.append(line)
    pbar.close()
    events_file_handle.close()
    df = pd.DataFrame(series_list)
    return df


def is_salesforce_event(event_name):
    return event_name.startswith('$sf_')


def is_hubspot_event(event_name):
    return event_name.startswith('$hubspot_')


def is_crm_event(event_name):
    return is_salesforce_event(event_name) or is_hubspot_event(event_name)


def wind_up_user(user_lines, series_list, merge_upr_flag, flatten_properties, \
        deduplicate_logic, counts, merge_upr_logic=DEFAULT_MERGE_UPR_LOGIC):
    if contains_unsubscribe_event(user_lines):
        counts['unsubscribe_users_removed'] += 1
        return
    if not contains_session_event(user_lines):
        counts['no_session_users_removed'] += 1
        return
    if merge_upr_flag:
        upr_list = [ul['upr'] for ul in user_lines]
        final_merged_upr = merge_dict(upr_list)
        prev_upr = None
        for ul in user_lines:
            if merge_upr_logic == 'cumulative':
                if is_crm_event(ul['en']):
                    ul['upr'] = dict(merge_dict([prev_upr, ul['upr']]))
                prev_upr = dict(ul['upr'])
            elif merge_upr_logic == 'final':
                ul['upr'] = dict(final_merged_upr)
    for ul in user_lines:
        if flatten_properties:
            ul = flatten_epr_upr(ul, deduplicate_logic)
        series_list.append(ul)


def is_izooto(project_id):
    if project_id is None:
        return False
    return project_id == IZOOTO

def contains_event(user_lines, event='$session'):
    return any([l['en'] == event for l in user_lines])

def contains_unsubscribe_event(user_lines):
    unsubscribe_event = 'www.izooto.com/campaign/unsubscribing-from-web-push-notifications'
    return contains_event(user_lines, unsubscribe_event)

def contains_session_event(user_lines):
    session_event = '$session'
    return contains_event(user_lines, session_event)

def aggregate_user_lines(line, user_lines, project_id=None, counts=None):
    # if line['en'] == '$session' and \
    #     is_izooto(project_id) and \
    #         contains_unsubscribe_event(user_lines): # If an unsubscribe event occurs for iZooto
    #     counts['unsubscribe_users_removed'] += 1
    #     user_lines[:] = [] # Makes an in-place change.
    user_lines.append(line)


def process_parsed_line(line, prev_uid, user_lines, series_list, merge_upr_flag,\
                        flatten_properties, deduplicate_logic, project_id, counts):
    start = time.time()
    curr_uid = line['uid']
    if curr_uid == prev_uid: # The same user is continuing
        aggregate_user_lines(line, user_lines, project_id, counts)
    else: # A new user started
        # Be done with the previous user:
        if prev_uid != -1:
            wind_up_user(user_lines, series_list, 
                            merge_upr_flag, flatten_properties, deduplicate_logic, counts)
        # And start with the new user:
        counts['users_encountered'] += 1
        user_lines[:] = [line] # This changes the list in-place!
        prev_uid = curr_uid
    end = (time.time()-start)
    if end > 1:
        print("Time taken: {}".format(end))
    return prev_uid

def read_and_parse_weekly_data(events_file_path,
                               n_lines=None,
                               flatten_properties=True,
                               deduplicate_logic=DEFAULT_DEDUPLICATE_LOGIC,
                               merge_upr_flag=True,
                               project_id = None,
                               base = None):
    events_file_handle = smart_open(events_file_path, 'r')
    series_list = []
    pbar = tqdm(total=n_lines, desc='Read events') if n_lines else tqdm(desc='Read events')
    prev_uid = -1
    counts = {'read_line': 0,
              'parse_success_line': 0,
              'parse_failure_line': 0,
              'unsubscribe_users_removed': 0,
              'users_encountered': 0,
              'no_session_users_removed': 0}
    user_lines = []
    while True:
        pbar.update()
        line = events_file_handle.readline()
        counts['read_line'] += 1
        if not line:
            break
        try:
            line = json.loads(line)
            counts['parse_success_line'] += 1
        except json.decoder.JSONDecodeError:
            counts['parse_failure_line'] += 1
            print('ERROR PARSING: \n\n', line, '\n\n')
            continue
        prev_uid = process_parsed_line(line, prev_uid, user_lines, series_list, \
            merge_upr_flag, flatten_properties, deduplicate_logic, project_id, counts)
    pbar.close()
    print(counts)
    events_file_handle.close()
    df = pd.DataFrame(series_list)
    return df


def filter_for_base(df, base=None):
    if not base:
        return df
    print('Filtering for base: {}'.format(base))
    n_all_users = df['uid'].dropna().nunique()
    n_all_rows = df.shape[0]
    base_filter = get_criteria_filter(df, base)
    base_users = set(df[['uid']][base_filter]
                         ['uid'].dropna().unique())
    n_base_users = len(base_users)
    p_base_users = round(n_base_users * 100.0 / n_all_users, 2)
    print('{} base users ({}% of {}) found.'.format(
        n_base_users, p_base_users, n_all_users))
    df = df[df['uid'].isin(base_users)]
    n_base_rows = df.shape[0]
    p_base_rows = round(n_base_rows * 100.0 / n_all_rows, 2)
    print('{} rows ({}% of {}) of data left.'.format(
        n_base_rows, p_base_rows, n_all_rows))
    return df


def get_weekly_data(cloud_path=DEFAULT_CLOUD_PATH,
                    project_id=DEFAULT_PROJECT_ID,
                    model_id=DEFAULT_MODEL_ID,
                    n_lines=None,
                    base=DEFAULT_BASE,
                    flatten_properties=True,
                    deduplicate_logic=DEFAULT_DEDUPLICATE_LOGIC,
                    sanitize_screen_size=True,
                    merge_upr_flag=True):
    events_file_name = "events_{}.txt".format(model_id)
    events_file_path = os.path.join(cloud_path,
                                    "projects", str(project_id),
                                    "models", str(model_id),
                                    events_file_name)
    print('Reading and parsing events file.')
    df = read_and_parse_weekly_data(events_file_path,
                                    n_lines,
                                    flatten_properties,
                                    deduplicate_logic,
                                    merge_upr_flag,
                                    project_id,
                                    base)
    df = filter_for_base(df, base)
    if sanitize_screen_size:
        print('Sanitizing screen size...', end='')
        df = sanitize_screen_size_params(df)
    print('Done.')
    return df
