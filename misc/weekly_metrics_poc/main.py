import os
import pdb
import json
from collections import defaultdict
import errno    
import os

EVENTS_OCCURRENCE_QUERY_TYPE = 'events_occurrence'
UNIQUE_USERS_QUERY_TYPE = 'unique_users'
DEFAULT_QUERIES_FILE_NAME = "queries.json"
DEFAULT_CLOUD_PATH = "/usr/local/var/factors/cloud_storage/"

def mkdir_p(path):
    try:
        os.makedirs(path)
    except OSError as exc:  # Python ≥ 2.5
        if exc.errno == errno.EEXIST and os.path.isdir(path):
            pass
        else:
            raise

def read_queries_from_json_file(queries_file_name = DEFAULT_QUERIES_FILE_NAME):
    queries = json.load(open(queries_file_name, 'r'))
    return queries

def fetch_queries_from_db(project_id):
    # TODO: Replace this with logic to fetch queries from DB.
    return read_queries_from_json_file()

def fetch_queries(project_id):
    queries = fetch_queries_from_db(project_id)
    return queries

def emulate_weekly_model(project_id, model_id, week_number=1):
    cloud_path = DEFAULT_CLOUD_PATH
    if week_number == 1:
        # First week: expected: eo_tp: 4599, uu_tp: 3539
        week_start_time = 1393632004 # 00:00:01, 1st March, 2014
        week_end_time = 1394216999 # 23:59:59, 7th March, 2014
    elif week_number == 2:
        # Second week: expected: eo_tp: 4664, uu_tp: 3540
        week_start_time = 1394217000 # 00:00:01, 8th March, 2014
        week_end_time = 1394821799 # 23:59:59, 14th March, 2014
    elif week_number == 3:
        # Third week: expected: eo_tp: 4581, uu_tp: 3513
        week_start_time = 1394821800 # 00:00:01, 15th March, 2014
        week_end_time = 1395426599 # 23:59:59, 21st March, 2014
    elif week_number == 4:
        # Fourth week: expected: eo_tp: 4572, uu_tp: 3560
        week_start_time = 1395426600 # 00:00:01, 22nd March, 2014
        week_end_time = 1396031399 # 23:59:59, 28th March, 2014

    events_file_name = "events_{}.txt".format(model_id)
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

def main():
    # Min time: 1393632004
    # Max time: 1396310325
    cloud_path = DEFAULT_CLOUD_PATH
    project_id = "11"
    queries = fetch_queries(project_id)

    model_id = "1604020686857"
    # TODO: Remove this and replace the previous line with weekly model id.
    model_id = emulate_weekly_model(project_id, model_id, 4)

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

if __name__ == "__main__":
    main()
