UOI_U = 'users'
UOI_E = 'events'


pid, mid1, mid2, queries = parse_args()

events_data1 = read_events_data(pid, mid1)
events_data2 = read_events_data(pid, mid2)

# Store differently?
delta_data = compute_delta_data(events_data1, events_data2, queries)

delta_insights = compute_delta_insights(delta_data)

store_delta_insights(delta_insights)


def parse_event(e):
    for k, v in e.items():
        if k == 'en':
            en = v
        if k == 'uid':
            uid = v
        if k == 'epr':
            epr = v
            epr_map = {}
            for epr_k, epr_v in epr.items():
                epr_map[prefix_to_str(epr_k, 'epr')] = epr_v
        if k == 'upr':
            upr = v
            upr_map = {}
            for upr_k, upr_v in upr.items():
                upr_map[prefix_to_str(upr_k, 'upr')] = upr_v
    return en, uid, epr_map, upr_map


def wrap_up_user(user_f_map, user_fm_maps, user_f_buffer_map, user_fm_buffer_maps):
    for k in user_f_buffer_map.keys():
        for v in user_f_buffer_map[k].keys():
            update_map(user_f_map, k, v)
    for i in range(len(user_fm_maps)):
        for k in user_fm_buffer_maps[i].keys():
            for v in user_fm_buffer_maps[i][k].keys():
                update_map(user_fm_maps[i], k, v)


def update_map(my_map, k, v):
    if k not in my_map:
        my_map[k] = {v: 0}
    elif v not in my_map[k]:
        my_map[k][v] = 0
    my_map[k][v] += 1


def matches_query(query, k, v):
    match_flag = any([(term.key == k and term.value == v) for term in query])
    return match_flag


def process_properties(pr_map, event_f_map, user_f_buffer_map, queries, event_fm_maps, user_fm_buffer_maps):
    for k, v in pr_map.items():
        update_map(user_f_buffer_map, k, v)
        update_map(event_f_map, k, v)
        for i, q in enumerate(queries):
            if matches_query(q, k, v):
                update_map(user_fm_buffer_maps[i], k, v)
                update_map(event_fm_maps[i], k, v)


def process_events(events_data, queries):
    user_f_map = {}
    user_fm_maps = [{} for _ in queries]
    event_f_map = {}
    event_fm_maps = [{} for _ in queries]
    user_f_buffer_map = {}
    user_fm_buffer_maps = [{} for _ in queries]
    prev_uid = -1
    for e in events_data:
        en, uid, pr_map = parse_event(e)
        if uid != prev_uid:
            wrap_up_user(user_f_map, user_fm_maps, user_f_buffer_map, user_fm_buffer_maps)
            user_f_buffer_map = {}
        process_properties(pr_map, event_f_map, user_f_buffer_map, queries, event_fm_maps, user_fm_buffer_maps)
    return user_f_map, user_fm_maps, event_f_map, event_fm_maps


def compute_delta_data(events_data1, events_data2, queries):
    user_f1_map, user_fm1_maps, event_f1_map, event_fm1_maps = process_events(events_data1, queries)
    user_f2_map, user_fm2_maps, event_f2_map, event_fm2_maps = process_events(events_data2, queries)
    delta_data = {UOI_E: {'f1': event_f1_map,
                          'f2': event_f2_map},
                  UOI_U: {'f1': user_f1_map,
                          'f2': user_f2_map}}
    for i, q in enumerate(queries):
        fm1_key = 'fm1_{}'.format(q.name)
        delta_data[UOI_E][fm1_key] = event_fm1_maps[i]
        delta_data[UOI_U][fm1_key] = user_fm1_maps[i]
        fm2_key = 'fm2_{}'.format(q.name)
        delta_data[UOI_E][fm2_key] = event_fm2_maps[i]
        delta_data[UOI_U][fm2_key] = user_fm2_maps[i]
    return delta_data


def compute_delta_insights(delta_data):
    # compute metrics for all feature-value combinations.
    pass
