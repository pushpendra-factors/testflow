from wi_poc.src.utils import implies, frac_change, INFINITY
import numpy as np
from wi_poc.src.defaults import DEFAULT_FILTER_PARAMS

def find_bucket(bucket_map, query, left_include=True, right_include=True):
    match_flag = False
    for bucket_key, value in bucket_map.items():
        st, en = bucket_key
        match_flag = ((query >= st) if left_include else (query > st)) and \
                     ((query <= en) if right_include else (query < en))
        if match_flag:
            return bucket_key, value
    return None, None

def get_bucketed_filter_params(counts):
    # Buckets define that, in comparison with the base value, what value is negligible?
    # For example, for 101 to 500 leads, less than 4 leads are negligible.

    def bucket_leads(n_leads):
        leads_bucket_map = {(0, 100): 2,
                            (101, 500): 4,
                            (501, 1000): 6,
                            (1001, 5000): 11,
                            (5001, INFINITY): 21}
        key, value = find_bucket(leads_bucket_map, n_leads)
        return value
    def bucket_users(n_users):
        users_bucket_map = {(0, 100): 11,
                            (101, 500): 31,
                            (501, 1000): 51,
                            (1001, 5000): 101,
                            (5001, INFINITY): 201}
        key, value = find_bucket(users_bucket_map, n_users)
        return value
    def bucket_rate(rate):
        rate_bucket_map =  {(0, 0.01): 0.001,
                            (0.01, 0.05): 0.003,
                            (0.05, 0.10): 0.005,
                            (0.10, 0.50): 0.01,
                            (0.50, INFINITY): 0.02}
        key, value = find_bucket(rate_bucket_map, rate, right_include=False)
        return value
    def bucket_fc_users(n1_users, n2_users):
        return 0.10
    def bucket_fc_leads(n1_leads, n2_leads):
        return 0.10
    def bucket_prev():
        return 0.90
    u = min(counts['u1'], counts['u2'])
    m = min(counts['m1'], counts['m2'])
    cr = min(counts['cr1'], counts['cr2'])
    fp = {'min_fm1': bucket_leads(m), 'min_fm2': bucket_leads(m),
          'min_f': bucket_users(u), 'min_crf': bucket_rate(cr),
          'min_fc_f': bucket_fc_users(counts['u1'], counts['u2']),
          'min_fc_fm': bucket_fc_leads(counts['m1'], counts['m2']),
          'min_fm': bucket_leads(m),
          'max_prev': bucket_prev()}
    return fp

def get_filter_params(counts, mode=None):
    if mode is None:
        fp = DEFAULT_FILTER_PARAMS
    elif mode == 'bucketed':
        fp = get_bucketed_filter_params(counts)
    return fp


def compute_pass_filters(f1, f2, fm1, fm2, pf1, pf2, pfm1, pfm2, fp):
    """
    Compute rule-based filters. For any filter, True means accept (pass),
    and False means reject (stop).

    Arguments
    ---------
    f1 : int
        #uoi (matching a feature) for first week
    f2 : int
        #uoi (matching a feature) for second week
    fm1 : int
        #leads (matching a feature) for first week
    fm2 : int
        #leads (matching a feature) for second week
    fp : dict
        Filter parameters dictionary
    
    Returns
    -------
    filters : dict
        Dictionary of all filters.
    """
    filters = {}
    filters['nzc'] = (fm1+fm2) > 0 
    filters['zmc'] = implies(fm1 == 0, fm2 >= fp['min_fm2']) and implies(fm2 == 0, fm1 >= fp['min_fm1'])
    filters['msmcr'] = implies(f1 < fp['min_f'] and f2 < fp['min_f'], (fm1 >= fp['min_fm1'] or fm2 >= fp['min_fm2']))
    filters['mscmcc'] = any([frac_change(f1, f2) >= fp['min_fc_f'],
                                frac_change(fm1, fm2) >= fp['min_fc_fm'],
                                np.abs(fm1 - fm2) >= fp['min_fm'],
                                (fm1-fm2) * (f1-f2) < 0])
    filters['fifth'] = any([frac_change(f1, f2) >= fp['min_fc_f'], 
                            np.abs(fm1 - fm2) >= fp['min_fm']])
    filters['sixth'] = any([fm1 >= fp['min_fm1'],
                            fm2 >= fp['min_fm2']])
    filters['max_prev'] = all([x <= fp['max_prev'] for x in [pf1, pf2, pfm1, pfm2]])
    return filters


def decide_pass_filters(filters, filters_keys):
    """
    Process filters and take a "pass" decision.

    Arguments
    ---------
    filters : dict
        A dictionary of filters.
    filters_keys : set
        A set of filter keys specifying which filters to use.

    Returns
    -------
    decision
        Based on a logic to combine all filters, return a decision.
    """
    decision = all([filters[x] for x in filters_keys])
    return decision
