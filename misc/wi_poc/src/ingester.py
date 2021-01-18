from smart_open import smart_open
from tqdm import tqdm
import json
from wi_poc.src.feature_processor import flatten_epr_upr, get_criteria_filter
from wi_poc.src.preprocessor import sanitize_screen_size_params
import pandas as pd
from wi_poc.src.utils import merge_dict
import os
from wi_poc.src.defaults import DEFAULT_DEDUPLICATE_LOGIC, DEFAULT_PROJECT_ID,\
    DEFAULT_MODEL_ID, DEFAULT_BASE
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


def wind_up_user(user_lines, series_list, merge_upr_flag, flatten_properties, deduplicate_logic):
    if merge_upr_flag:
        final_upr = merge_dict([ul['upr'] for ul in user_lines], 'final')
    for ul in user_lines:
        if merge_upr_flag:
            ul['upr'] = dict(final_upr)
        if flatten_properties:
            ul = flatten_epr_upr(ul, deduplicate_logic)
        series_list.append(ul)


def read_and_parse_weekly_data(events_file_path,
                               n_lines=None,
                               flatten_properties=True,
                               deduplicate_logic=DEFAULT_DEDUPLICATE_LOGIC,
                               merge_upr_flag=True):
    events_file_handle = smart_open(events_file_path, 'r')
    series_list = []
    pbar = tqdm(total=n_lines, desc='Read events') if n_lines else tqdm(desc='Read events')
    prev_uid = -1
    curr_uid = -1
    user_lines = []
    while True:
        pbar.update()
        line = events_file_handle.readline()
        if not line:
            break
        try:
            line = json.loads(line)
        except json.decoder.JSONDecodeError:
            print('ERROR PARSING: \n\n', line, '\n\n')
            continue
        curr_uid = line['uid']
        if curr_uid == prev_uid: # The same user is continuing
            user_lines.append(line)
        else: # A new user started
            # Be done with the previous user:
            if prev_uid != -1:
                wind_up_user(user_lines, series_list, 
                             merge_upr_flag, flatten_properties, deduplicate_logic)
            # And start with the new user:
            user_lines = [line]
            prev_uid = curr_uid
    pbar.close()
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
                                    merge_upr_flag)
    df = filter_for_base(df, base)
    if sanitize_screen_size:
        print('Sanitizing screen size...', end='')
        df = sanitize_screen_size_params(df)
    print('Done.')
    return df
