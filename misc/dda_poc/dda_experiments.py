"""
# Data-driven attribution (DDA) experiments

## Imports and Dependencies

- `tabulate`
- `tqdm`
- `smart_open[s3]`
- `s3fs`
"""

# pip install tabulate tqdm smart_open[s3] matplotlib s3fs

import os
import pandas as pd
from tqdm.notebook import tqdm
import json
import pickle
from collections import Counter
from tabulate import tabulate
import subprocess
from smart_open import open

"""## Function definitions"""

# Parses each line of the events file.
# line: One line of the events file (one user event string)
# resolve_properties: Whether to append an 'epr'/'upr' to properties or not.
# ignore_properties: Whether to get rid of all properties at all or not.
# keep: Either have this as None, or give a list of all keys that you need to retain.
# keep_events: Either give None, or give a whitelist of events to keep.
#              If there's an event matched to be ignored, the function returns None.
def parse_line(line,
               resolve_properties=False,
               ignore_properties=False,
               keep=None,
               keep_events=None):
    if resolve_properties and ignore_properties:
        raise Exception(
            "Cannot resolve and ignore properties at the same time.")
    line = json.loads(line.strip())
    if keep_events is not None:
        if line['en'] not in keep_events:
            return None
    if ignore_properties:
        del line['epr'], line['upr']
    if resolve_properties:
        line.update({f'epr_{k}': v for k, v in line['epr'].items()})
        line.update({f'upr_{k}': v for k, v in line['upr'].items()})
        del line['epr'], line['upr']
    if keep is not None:
        line = {k: line[k] for k in keep if k in line}
    return line


# Computes division in a more smart fashion.
def smart_divide(a, b):
    if b == 0:
        if a == 0:
            return 0
        return float('Inf')
    return 1.0 * a / b


# Gives conversion summary.
# Gets the summary of the first conversion for every user. Either the user doesn't convert
# at all (i.e., if resolved_ids is None) -- 'not_converted'.
# Or the user converts in Session 0 (a dummy session for cases where there's no session) -- 'converted_wo_session',
# or the user converts for more than one times -- 'converted_in_>=1_session'.
def get_summary(resolved_ids):
    if resolved_ids is None:
        return 'not_converted'
    if resolved_ids == [0]:
        return 'converted_wo_session'
    return 'converted_in_>=1_session'


# Caches and retrieves from S3, the first dataframe "df" (one-touch-per-row)
# that we generate from the raw events.txt file.
# This uses the parse_line function and creates a dataframe out of it.
def get_df(s3_events_file_path, force_fresh=False):
    try:
        if force_fresh:
            raise ImportError
        df = pd.read_csv(s3_events_file_path+'.csv', index_col=0)
    except Exception:
        meta_cols = ['uid', 'en', 'et']
        att_cols = ['epr_$campaign', 'epr_$campaign_id', 'epr_$channel']
        keep_cols = meta_cols + att_cols
        with open(s3_events_file_path, 'r') as f:
            lines = []
            for index, line in tqdm(enumerate(f), total=nlines):
                line = parse_line(line,
                                  resolve_properties=True,
                                  ignore_properties=False,
                                  keep=keep_cols,
                                  keep_events={tgt_event, '$session'})
                if line is None:
                    continue
                lines.append(line)
        df = pd.DataFrame(lines).fillna('$none')
        df['et'] = pd.to_datetime(df['et'], unit='s')
        df.to_csv(s3_events_file_path+'.csv')
    return df

# Caches and retrieves from S3, the grouped-aggregated dataframe "user-df" (one-user-per-row)
# that we generate from df.
def get_user_df(s3_events_file_path, df=None, force_fresh=False):
    user_df_path = s3_events_file_path+'.user'
    try:
        if force_fresh:
            raise ImportError
        user_df = pickle.load(open(user_df_path+'.pkl', 'rb'))
    except Exception:
        if df is None:
            df = get_df(s3_events_file_path)
        user_df = df.fillna('$none').groupby('uid').agg(list)
        pickle.dump(user_df, open(user_df_path+'.pkl', 'wb'))
    return user_df

# Takes in user-df (one-user-per-row) information and returns
# conversion and non-conversion based stats, while retaining or
# losing the sequence information.
def get_final_data(user_df, touch_mode, sequence=True):
    final_df = user_df[user_df['n_uniq_ch'] >= 1].apply(lambda x: ([x[touch_mode][i] for i in x['resolved_ids'] if x[touch_mode][i] != '$none']), axis=1)
    final_df = pd.DataFrame({'seq_str': final_df.apply(' -> '.join), 'seq': final_df.apply(tuple if sequence else frozenset)})
    final_df = final_df[final_df['seq'].apply(lambda x: len(x) > 0)]
    
    final_df0 = user_df[user_df['summary'] == 'not_converted'].apply(lambda x: ([y for y in x[touch_mode] if y != '$none']), axis=1)
    final_df0 = pd.DataFrame({'seq_str': final_df0.apply(' -> '.join), 'seq': final_df0.apply(tuple if sequence else frozenset)})
    final_df0 = final_df0[final_df0['seq'].apply(lambda x: len(x) > 0)]
    # print(tabulate(final_df[['seq_str']], headers='keys', tablefmt='psql'))
    return final_df, final_df0

# Processes conversion and non-conversion based stats while
# retaining or losing the sequence information.
def get_conversion_info(conv_df, nconv_df, sequence=True):
    col_name = 'seq' if sequence else 'set' 
    if not sequence:
        conv_df[col_name] = conv_df['seq'].apply(frozenset)
        nconv_df[col_name] = nconv_df['seq'].apply(frozenset)

    conv = conv_df.groupby(col_name).agg(len).to_dict()['seq_str']
    nconv = nconv_df.groupby(col_name).agg(len).to_dict()['seq_str']

    conv_dict = {k: (conv[k], nconv.get(k, 0), smart_divide(conv[k], conv[k]+nconv.get(k, 0))) for k in conv}

    return conv_dict

# Computes shapley values from conversion information.
# Given conv_dict, a dictionary where key is a set of touchpoints (called coalition),
# and value is the number of converts exclusively due to that coalition.
def find_shapley_values(conv_dict, tgt_mode):
    synergy_dict = {k: v[0 if tgt_mode is 'c' else -1] for k, v in conv_dict.items()}
    shapley = defaultdict(float)
    for coal, syn in synergy_dict.items():
        size = len(coal)
        for item in coal:
            shapley[item] += syn/size
    scores = shapley
    return scores

def find_shapley_values(conv_dict, tgt_mode):
    synergy_dict = {k: v[0 if tgt_mode is 'c' else -1] for k, v in conv_dict.items()}
    shapley = defaultdict(float)
    for coal, syn in synergy_dict.items():
        total = sum([x[1] for x in coal])
        for item, mult in coal:
            shapley[item] += syn*mult/total
    scores = shapley
    return scores

# Computes linear and weighted attribution from conversion information.
# mode=const: average attribution
# mode=lini: linear-increasing: lower weight to first-touch and higher to last-touch
# mode=lind: linear-decreasing: higher weight to first-touch and lower to last-touch
def find_avg_att(conv_dict, tgt_mode, mode='const'):
    convs = defaultdict(float)
    nconvs = defaultdict(float)
    for k, v in conv_dict.items():
        size = len(k)
        if mode == 'const':
            wts = np.ones((size,))
        elif mode == 'lini':
            wts = np.linspace(0, 1, size+1)[1:]
        elif mode == 'lind':
            wts = np.linspace(1, 0, size+1)[:-1]
        wts = wts / wts.sum()
        for i, key in enumerate(k):
            convs[key] += v[0]/wts[i]
            nconvs[key] += v[1]/wts[i]
    if tgt_mode == 'c':
        return dict(convs)
    if tgt_mode == 'cr':
        scores = {k:smart_divide(convs[k], convs[k]+nconvs[k]) for k in convs}
        return scores

# Computes single-touch, both first-touch and last-touch attributions
def find_single_touch(conv_dict, tgt_mode, mode):
    convs = defaultdict(float)
    nconvs = defaultdict(float)
    for k, v in conv_dict.items():
        convs[k[0 if mode == 'first' else -1]] += v[0]
        nconvs[k[0 if mode == 'first' else -1]] += v[1]
    if tgt_mode == 'c':
        return dict(convs)
    if tgt_mode == 'cr':
        scores = {k:smart_divide(convs[k], convs[k]+nconvs[k]) for k in convs}
        return scores
    
# Find all attribution scores given conversion and non-conversion info
def find_att_scores(conv_df, nconv_df, att_mode, tgt_mode):
    if att_mode.startswith('shap'): # Shapley
        conv_dict = get_conversion_info(conv_df, nconv_df, sequence=False)
        return find_shapley_values(conv_dict, tgt_mode)
    elif att_mode.startswith('ft'): # First touch
        conv_dict = get_conversion_info(conv_df, nconv_df, sequence=True)
        return find_single_touch(conv_dict, tgt_mode, 'first')
    elif att_mode.startswith('lt'): # First touch
        conv_dict = get_conversion_info(conv_df, nconv_df, sequence=True)
        return find_single_touch(conv_dict, tgt_mode, 'last')
    elif att_mode.startswith('avg'):
        conv_dict = get_conversion_info(conv_df, nconv_df, sequence=True)
        return find_avg_att(conv_dict, tgt_mode, 'const')
    elif att_mode.startswith('lini'):
        conv_dict = get_conversion_info(conv_df, nconv_df, sequence=True)
        return find_avg_att(conv_dict, tgt_mode, 'lini')
    elif att_mode.startswith('lind'):
        conv_dict = get_conversion_info(conv_df, nconv_df, sequence=True)
        return find_avg_att(conv_dict, tgt_mode, 'lind')

"""## Set up paths from S3
- Change `project_id` to the respective client
- Change `st_date` to the desired period
- Change `mode` to eithe `m` or `w` for monthly or weekly
- Change `tgt_event` to change the target/conversion-event information
"""

s3_bucket = 'data-driven-attribution'
project_id = '399' # ChargeBee
# project_id = '559' # AdPushup

mode = 'm'
st_date = '20210901'

nlines = {'559': {'20210901': 490204},
          '399': {'20210901': 1409686}}[project_id][st_date]

tgt_event = {'559': '$hubspot_contact_created',
             '399': '$hubspot_contact_created'}[project_id]
events_file_name = 'events.txt'
s3_suffix = os.path.join(s3_bucket, 'data', 'cloud_storage', 'projects', project_id, 'events', mode, st_date, events_file_name)
s3_events_file_path = f"s3://{s3_suffix}"

# Checks if a sequence of events ended up in "conversion target",
# and trims it after the *first* target occurrence. If target not
# achieved, returns None.
# This function uses a global variable, tgt_event as the second argument tgt_event.
def get_resolved_event_ids(list_of_events, tgt_event=tgt_event):
    resolved_id_list = []
    tgt_achieved_flag = False
    for i, e in enumerate(list_of_events):
        resolved_id_list.append(i)
        if e == tgt_event:
            tgt_achieved_flag = True
            break
    return resolved_id_list if tgt_achieved_flag else None

"""## Read or process event-level data"""

df = get_df(s3_events_file_path) # Read events file line by line, and retain only relevant events and properties.
print('rows, columns: ', df.shape)
df.head()

# Total conversion ratio.
u = df['uid'].nunique()
c = df[df['en'] == tgt_event]['uid'].nunique()
cr = round(smart_divide(c, u)*100, 4)
print('unique users:', u, 'converts:', c, 'ratio:', cr)

ncamp = df['epr_$campaign'].nunique()
nch = df['epr_$channel'].nunique()
print(ncamp, 'campaigns,', nch, 'channels')

"""## User-level group-aggregated data"""

user_df = get_user_df(s3_events_file_path, df) # Aggregate users on the 'uid' column, and collect the list of events, properties, etc.

print('rows, columns:', user_df.shape)

"""### Conversion stats"""

user_df['resolved_ids'] = user_df['en'].apply(lambda x: get_resolved_event_ids(x)) # Resolve conversion (only consider the first conversion as the convert event)
user_df['summary'] = user_df['resolved_ids'].apply(lambda x: get_summary(x)) # Divide users into 3 classes.
print('***User distribution based on conversion***')
user_df['summary'].value_counts()

"""### Campaign-level stats
This part shows the number of **unique campaign-touchpoints** each **converted user** goes through. What is being shown in the table is the number of unique campaign touch-points (1st column), and number of users with that property (2nd column).
"""

user_df['n_uniq_tp_id'] = user_df.apply(lambda x: None if x['summary'] != 'converted_in_>=1_session' else len(set([i for i in [x['epr_$campaign_id'][j] for j in x['resolved_ids']] if i != '$none'])), axis=1)
user_df['n_uniq_tp'] = user_df.apply(lambda x: None if x['summary'] != 'converted_in_>=1_session' else len(set([i for i in [x['epr_$campaign'][j] for j in x['resolved_ids']] if i != '$none'])), axis=1)
user_df['n_uniq_ch'] = user_df.apply(lambda x: None if x['summary'] != 'converted_in_>=1_session' else len(set([i for i in [x['epr_$channel'][j] for j in x['resolved_ids']] if i != '$none'])), axis=1)
print('***Distribution of number of unique non-$none campaigns***')
user_df[user_df['summary'] == 'converted_in_>=1_session']['n_uniq_tp'].value_counts()

"""### Channel-level stats
This part shows the number of **unique channel-touchpoints** each **converted user** goes through. What is being shown in the table is the number of unique channel touch-points (1st column), and number of users with that property (2nd column)
"""

print('***Distribution of number of unique non-$none channels***')
user_df[user_df['summary'] == 'converted_in_>=1_session']['n_uniq_ch'].value_counts()

"""## Touchpoint conversion stats

Using information from "channel" or "campaign" (a parameter to be set), find conversion stats w.r.t. 'channel' or 'campaign' sequences.
"""

# touch_mode = 'epr_$campaign'
touch_mode = 'epr_$channel'
conv_df, nconv_df = get_final_data(user_df, touch_mode, sequence=True)

"""### Sequence and multiplicity retained"""

def rep_format(x):
    items = []
    prev_c = x[0]
    count = 1
    for i, c in enumerate(x[1:]):
        if c == prev_c:
            count += 1
        else:
            item = f'{prev_c}({count})' if count > 1 else prev_c
            items.append(item)
            count = 1
        prev_c = c
    item = f'{prev_c}({count})' if count > 1 else prev_c
    items.append(item)
    return tuple(items)

conv_dict = get_conversion_info(conv_df, nconv_df, sequence=True)
conv_dict = {rep_format(k): v for k, v in conv_dict.items()}
conv_dict = {' -> '.join(k): v for k, v in conv_dict.items()}
conv_dict_df = pd.DataFrame(conv_dict).T
conv_dict_df.columns = ['conv', 'non-conv', 'conv-ratio']
conv_dict_df['cr_perc'] = conv_dict_df['conv-ratio'].apply(lambda x: f'{"{:2.2f}".format(round(x*100, 2))}%')
conv_dict_df = conv_dict_df.sort_values('conv-ratio', ascending=False)
conv_dict_df.to_csv(s3_events_file_path+f'.{touch_mode}.conv-seq.csv')
conv_dict_df

print('\n'.join(([str(round(conv_dict[k][-1]*100, 2))+ '%'+ '\t'+ k for k in sorted(conv_dict, key=lambda x: -conv_dict[x][-1])])))

"""### Both sequence and multiplicity ignored (only set retained)"""

conv_dict = get_conversion_info(conv_df, nconv_df, sequence=False)
conv_dict = {', '.join(sorted(k)): v for k, v in conv_dict.items()}
conv_dict_df = pd.DataFrame(conv_dict).T
conv_dict_df.columns = ['conv', 'non-conv', 'conv-ratio']
conv_dict_df['cr_perc'] = conv_dict_df['conv-ratio'].apply(lambda x: f'{"{:2.2f}".format(round(x*100, 2))}%')
conv_dict_df = conv_dict_df.sort_values('conv-ratio', ascending=False)
conv_dict_df.to_csv(s3_events_file_path+f'.{touch_mode}.conv-set.csv')
conv_dict_df

print('\n'.join(([str(round(conv_dict[k][-1]*100, 2))+ '%'+ '\t'+ k for k in sorted(conv_dict, key=lambda x: -conv_dict[x][-1])])))

"""## Computing various attribution scores
- `ft`: First touch
- `lt`: Last touch
- `avg`: Mean (all touch-points equally weighted)
- `lini`: Linear-increasing (touch-points weighted by increasing weights)
- `lind`: Linear-decreasing (touch-points weighted by decreasing weights)
- `shap`: Shapley-value based attribution

Also, two modes of target scoring:
- `c`: Conversions (absolute conversions, e.g., 100, 200, etc.)
- `cr`: Conversion-ratio (conversion ratio: converts/(converts+nonconverts), e.g., 0.1, 0.2, etc.)
"""

from collections import defaultdict
import numpy as np
import pandas as pd
att_modes = ['ft', 'lt', 'avg', 'lini', 'lind', 'shap']
tgt_modes = ['c', 'cr']
att_tgt_map = {}
for att_mode in tqdm(att_modes):
    for tgt_mode in tgt_modes:
        score = find_att_scores(conv_df, nconv_df, att_mode, tgt_mode) # Find the absolute scores.
        att_tgt_map[f'{att_mode}-{tgt_mode}'] = score #

att_df = pd.DataFrame(att_tgt_map)
att_df.to_csv(s3_events_file_path+f'.{touch_mode}.att.csv')
att_df = att_df / att_df.sum() # Normalise the attribution scores.
att_df.to_csv(s3_events_file_path+f'.{touch_mode}.att_norm.csv')
att_df

"""### Plotting all attribution scores against each other.
- On the x-axis, you see channels (or campaigns if you have chosen that `touch_mode`.
- On the y-axis, is the normalized score that each attribution scoring algorithm gives.
"""

from matplotlib import pyplot as plt
fig, axs = plt.subplots(len(tgt_modes), len(att_modes), figsize=(len(att_modes)*5, len(tgt_modes)*5))
for j, a in enumerate(att_modes):
    for i, t in enumerate(tgt_modes):
        ax = axs[i][j]
        att_df[f'{a}-{t}'].plot(kind='bar', ax=ax)
        ax.set_title(f'{a}-{t}')
        ax.set_xticklabels(att_df.index if att_df.shape[0]<25 else ax.get_xticks(), rotation = 45, ha='right')
plt.tight_layout()
plt.show()
print(f'{project_id}.{touch_mode}.att_norm.png')

if __name__ == '__main__':
    # For now, the Python file runs sequentially. Bits could be brought here.
    pass
