import json
import os
import pdb
from collections import defaultdict
from .utils import mkdir_p
from .defaults import DEFAULT_CLOUD_PATH, DEFAULT_MODEL_ID, \
    DEFAULT_PROJECT_ID, DEFAULT_WEEK_NUM, DEFAULT_QUERIES_FILE_NAME, \
    EVENTS_OCCURRENCE_QUERY_TYPE, UNIQUE_USERS_QUERY_TYPE

def read_queries_from_json_file(queries_file_name = DEFAULT_QUERIES_FILE_NAME):
    queries = json.load(open(queries_file_name, 'r'))
    return queries

def fetch_queries_from_db(project_id):
    # TODO: Replace this with logic to fetch queries from DB.
    return read_queries_from_json_file()

def fetch_queries(project_id):
    queries = fetch_queries_from_db(project_id)
    return queries

def get_st_en_times(project_id, week_number):
    week_map = {"11": {1: (1393632004, 1394216999),
                       2: (1394217000, 1394821799),
                       3: (1394821800, 1395426599),
                       4: (1393632004, 1393632004)},
                "399": {1: (1537660800, 1538265599),
                        2: (1538265600, 1538870399)}}
    wk_st_time, wk_end_time = week_map[project_id][week_number]
    return wk_st_time, wk_end_time

def emulate_weekly_model(project_id, model_id, week_number=1):
    cloud_path = DEFAULT_CLOUD_PATH
    week_start_time, week_end_time = get_st_en_times(project_id, week_number)
    events_file_name = "events_{0}.txt".format(model_id)
    events_file_path = os.path.join(cloud_path, "projects", project_id, "models", model_id, events_file_name)
    events_file_handle = open(events_file_path, 'r')
    filtered_line_list = []
    while True:
        line = events_file_handle.readline()
        if not line:
            break
        line = json.loads(line)
        event_time = line['et']
        if event_time >= week_start_time and event_time <= week_end_time:
            filtered_line_list.append(json.dumps(line))
    events_file_handle.close()
    week_id = str(week_start_time)
    model_id = week_id
    filtered_file_directory = os.path.join(cloud_path, "projects", project_id, "models", model_id)
    filtered_file_name = "events_{}.txt".format(model_id)
    filtered_file_path = os.path.join(filtered_file_directory, filtered_file_name)
    print(filtered_file_path)
    mkdir_p(filtered_file_directory)
    f = open(filtered_file_path, 'w')
    f.write('\n'.join([str(x) for x in filtered_line_list]) + '\n')
    f.close()
    return model_id

def perform_count_experiment(week_num=DEFAULT_WEEK_NUM,
                             project_id = DEFAULT_PROJECT_ID,
                             model_id = DEFAULT_MODEL_ID,
                             cloud_path = DEFAULT_CLOUD_PATH):
    # Min time: 1393632004
    # Max time: 1396310325
    queries = fetch_queries(project_id)

    # model_id = DEFAULT_MODEL_ID
    # TODO: Remove this and replace the previous line with weekly model id.
    model_id = emulate_weekly_model(project_id, model_id, week_num)

    events_file_name = "events_{}.txt".format(model_id) # Assuming this is the appropriate week's file.
    events_file_path = os.path.join(cloud_path, "projects", project_id, "models", model_id, events_file_name)
    events_file_handle = open(events_file_path, 'r')

    # pdb.set_trace()
    
    query_metrics_map = defaultdict(int)
    prev_user_id = 0
    while True:
        line = events_file_handle.readline()
        if not line:
            break
        line = json.loads(line)
        line_event_name = line['en']
        for query_name, query in queries.items():
            query_type = query['ty']
            events_in_query = [x['na'] for x in query['ewp']]
            if line_event_name in events_in_query:
                if query_type == EVENTS_OCCURRENCE_QUERY_TYPE:
                    query_metrics_map[query_name] += 1
                elif query_type == UNIQUE_USERS_QUERY_TYPE:
                    user_id = line['uid']
                    if user_id != prev_user_id:
                        query_metrics_map[query_name] += 1
                        prev_user_id = user_id
            # line_matches_query = does_line_match_query(line, query)
    print(query_metrics_map)
