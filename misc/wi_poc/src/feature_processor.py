from wi_poc.src.utils import merge_dict, is_iterable
import functools
from tqdm import tqdm
import pandas as pd
from smart_open import smart_open
from wi_poc.src.defaults import DEFAULT_DEDUPLICATE_LOGIC, DEFAULT_BLACKLISTS,\
    DEFAULT_UID_FEAT, DEFAULT_MAX_NULL_PERC, CUSTOM_BLACKLIST

SFL_KEY = 'salesforce_lead'
DATE_KEY = 'date'

def merge_upr(df, logic='final'):
    """
    Merges user-properties of all events matching a particular user based on `logic`.

    Arguments
    ---------
    df : pd.DataFrame
        The data. Should contain a 'uid' and a 'upr' column.
    logic : str
        The merge logic. Only 'final' (i.e., merging all uprs until the final user
        event and copying to all others) supported as of now.

    Returns
    -------
    df
        The upr-merged dataframe with 'upr' replaced with the new column.
    """
    df1 = df[['uid', 'upr']].groupby('uid').agg(lambda x: merge_dict(list(x), logic=logic))
    df = df.merge(df1, how='inner', left_on='uid', right_index=True, suffixes=('_curr', '_merged'))
    df['upr'] = df['upr_merged']
    df.drop(['upr_merged', 'upr_curr'], axis=1, inplace=True)
    return df


def flatten_epr_upr(_dict, deduplicate_logic=DEFAULT_DEDUPLICATE_LOGIC):
    """
    Flattens a parsed JSON (into Python dict) w.r.t. event- and user-properties.

    Arguments
    ---------
    _dict : dict
        The parsed dictionary of one event occurrence.
    deduplicate_logic : str
        The logic to be used to handle a property common between upr and epr.
        Supported values are 'upr' (prioritize upr), 'epr' (prioritize epr), and
        None (don't deduplicate, instead add 'epr_' or 'upr_' prefixes.).
    
    Returns
    -------
    _dict
        The updated flattened and/or deduplicated dictionary.
    """
    epr_dict = dict(_dict['epr'])
    upr_dict = dict(_dict['upr'])
    other_keys = {'uid', 'en'}
    _dict = {k: _dict[k] for k in other_keys}
    if deduplicate_logic == 'upr':
        _dict.update(upr_dict)
        _dict.update({k: epr_dict[k] for k in epr_dict if k not in upr_dict})
    elif deduplicate_logic == 'epr':
        _dict.update(epr_dict)
        _dict.update({k: upr_dict[k] for k in upr_dict if k not in epr_dict})
    elif deduplicate_logic is None:
        _dict.update({'epr_{}'.format(k): v for k, v in epr_dict.items()})
        _dict.update({'upr_{}'.format(k): v for k, v in upr_dict.items()})
    else:
        print('Deduplication logic "{}" not supported.'.format(deduplicate_logic))
    return _dict


def get_criteria_filter(df, criteria=None):
    """
    For a "criteria" (see `criteria_map` in `defaults.py` for an example),
    generate a criteria-filter using the dataframe `df`.

    Arguments
    ---------
    df : pd.DataFrame
        The dataframe from which the criteria-filter has to be generated.
    criteria : tuple
        A triplet tuple used to specify a criteria (see `criteria_map` in
        `defaults.py` for an example).
    
    Returns
    -------
    criteria_filter : pd.Series
        A Pandas boolean Series based on the `criteria`.
    """
    if not criteria:
        return pd.Series([True for _ in range(df.shape[0])])
    criteria_feat, criteria_val, criteria_assert = criteria
    if not is_iterable(criteria_feat):
        criteria_feat = [criteria_feat]
        criteria_val = [criteria_val]
        criteria_assert = [criteria_assert]
        criteria = (criteria_feat, criteria_val, criteria_assert)
    criteria_compliant_users = set()
    first_time_flag = True
    for cf, cv, ca in zip(criteria_feat, criteria_val, criteria_assert):
        this_criteria_filter = (df[cf].isna()) if cv is None else (df[cf]==cv)
        users = set(df[this_criteria_filter]['uid'].unique())
        if first_time_flag:
            first_time_flag = False
            if ca:
                criteria_compliant_users.update(users)
            else:
                criteria_compliant_users.update(set(df[~df['uid'].isin(users)]['uid'].unique()))
        else:
            if ca:
                criteria_compliant_users.intersection_update(users)
            else:
                criteria_compliant_users.difference_update(users)
    criteria_filter = df['uid'].isin(criteria_compliant_users)
    return criteria_filter


def get_value_stats(df, feat_type_map=None, target=None):
    """
    Find statistics for each feature in the data. To be used for pre-pruning features.

    Arguments
    ---------
    df : pd.DataFrame
        The data.
    feat_type_map : dict
        A dictionary of feature types (read out from the feat-type-map file).
    target : tuple
        A triplet tuple used to specify a target criteria (see `criteria_map` in
        `defaults.py` for an example).
    
    Returns
    -------
    vsdf
        Value statistics dataframe
    target_compliant_fvs
        Those feature-value combinations that appear in target events.
    """
    nvalues = {}
    feat_type_map = feat_type_map or {}
    target_uids = df[get_criteria_filter(df, target)]['uid'].unique()
    n_uniq_uids = df['uid'].nunique()
    target_compliant_fvs = {}
    for c in tqdm(df.columns, desc='Getting value-stats'):
        if len(df[c].shape) == 2:
            df_c = df[c].iloc[:, 0]
        else:
            df_c = df[c]
        p_nonnull = round(df_c.dropna().shape[0] / df.shape[0], 2)
        n_uniq = df_c.dropna().nunique()
        uniq_values_target = set(
            df[[c]][df['uid'].isin(target_uids)][c].dropna().unique())
        n_uniq_target = len(uniq_values_target)
        if n_uniq_target > 0:
            target_compliant_fvs[c] = uniq_values_target
        p_uniq = round(df_c.nunique()/df.shape[0], 2)
        p_nonnull_uniq = round(df_c.nunique()/df_c.dropna().shape[0], 2) if df_c.dropna().shape[0] > 0 \
            else (float('Inf') if df_c.nunique() > 0 else 0)
        p_uniq_wrt_uid = round(df_c.nunique()/n_uniq_uids, 2)
        _type = feat_type_map.get(c, None)
        nvalues[c] = {'p_nonnull': p_nonnull,
                      'n_uniq': n_uniq,
                      'p_uniq': p_uniq,
                      'p_nonnull_uniq': p_nonnull_uniq,
                      'p_uniq_wrt_uid': p_uniq_wrt_uid,
                      'type': _type,
                      'n_uniq_target': n_uniq_target}
    vsdf = pd.DataFrame(nvalues).T
    return vsdf, target_compliant_fvs


def decide_on_feat(row):
    """
    Decide whether to accept or reject feature.

    Arguments
    ---------
    row : pd.Series
        A feature with its value-stats.
    
    Returns
    -------
    decision
        Whether or not to keep the feature.
    """
    decision = 'accept'
    if row['type'] == 'float' or row['type'] == 'int':
        if row['n_uniq'] > 10:
            decision = 'reject'
        else:
            decision = 'accept'
    elif row['type'] == 'str':
        if row['p_uniq_wrt_uid'] > 0.8:
            decision = 'reject'
        else:
            decision = 'accept'
    return decision


def get_blacklisted_feats(features, blacklists=DEFAULT_BLACKLISTS):
    bl_feats = set()
    if 'custom' in blacklists:
        bl_feats.update(CUSTOM_BLACKLIST)
    if 'sf_lead' in blacklists:
        sf_lead_feats = {f for f in features if SFL_KEY in f}
        bl_feats.update(sf_lead_feats)
    if 'date' in blacklists:
        date_feats = {f for f in features if DATE_KEY in f}
        bl_feats.update(date_feats)
    return bl_feats


def preselect_features(df1, df2, target,
                       feat_schema_filename,
                       blacklists=DEFAULT_BLACKLISTS):
    """
    Before processing data of both weeks, select some features.

    Arguments
    ---------
    df1 : pd.DataFrame
        First week's data.
    df2 : pd.DataFrame
        Second week's data.
    target : tuple
        A triplet tuple used to specify a target criteria (see `criteria_map` in
        `defaults.py` for an example).
    feat_schema_filename: str
        CSV file specifying feature schemas (if available).
    blacklists: set
        A set of blacklist keys specifying what blacklists to use.
    
    Returns
    -------
    feats
        The shortlist of features.
    target_compliant_fvs
        Those feature-value combinations that appear at least once in target events.
    exp_feats
        Candidate explanation features.
    """
    feat_df = pd.read_csv(smart_open(feat_schema_filename, 'r'))
    feat_type_map = dict(zip(feat_df['feature'], feat_df['type']))

    vsdf1, target_compliant_fvs1 = get_value_stats(df1, feat_type_map, target)
    vsdf2, target_compliant_fvs2 = get_value_stats(df2, feat_type_map, target)

    vsdf1['decision'] = vsdf1.apply(decide_on_feat, axis=1)
    vsdf2['decision'] = vsdf2.apply(decide_on_feat, axis=1)
    def union_merge_maps(x, y): return {f: x.get(f, set()).union(
        y.get(f, set())) for f in set(x).union(set(y))}
    target_compliant_fvs = union_merge_maps(
        target_compliant_fvs1, target_compliant_fvs2)
    feats = set.intersection(set(vsdf1[vsdf1['decision'] == 'accept'].index),
                             set(vsdf2[vsdf2['decision'] == 'accept'].index))
    blacklisted_feats = get_blacklisted_feats(feats, blacklists)
    feats = feats.difference(blacklisted_feats)
    target_feat, target_val, _ = target
    feats.update({DEFAULT_UID_FEAT})
    feats.update(set(target_feat) if is_iterable(target_feat) else {target_feat})

    if is_iterable(target_feat):
        exp_feats = list(sorted(feats.difference(set(target_feat)).difference({DEFAULT_UID_FEAT})))
    else:
        exp_feats = list(sorted(feats.difference({target_feat, DEFAULT_UID_FEAT})))
    return feats, target_compliant_fvs, exp_feats


def compute_candidate_values(df1, df2, f, target_compliant_fvs):
    """
    For a given feature `f`, compute values on which to iterate upon.

    Arguments
    ---------
    df1 : pd.DataFrame
        First week's data
    df2 : pd.DataFrame
        Second week's data
    f : str
        Feature in consideration
    target_compliant_fvs : dict
        Those feature-value combinations that appear at least once in target events.
    
    Returns
    -------
    values
        Candidate values for feature `f`.
    """
    values1 = set(df1[[f]][~(df1[f].isna())][f].unique())
    values2 = set(df2[[f]][~(df2[f].isna())][f].unique())
    values = values1.union(values2)
    values = values.intersection(target_compliant_fvs.get(f, set()))
    return values
