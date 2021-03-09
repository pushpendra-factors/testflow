from datetime import datetime
import numpy as np
INFINITY = INFINITE = np.inf

def implies(a, b):
    return (not a) or b

def frac_change(a1, a2, abs=True):
    num = a2-a1
    fc = smart_divide(np.abs(num) if abs else num, a1)
    if fc is INFINITY:
        fc = 1
    return fc

def is_iterable(x):
    iter_types = [list, dict, set, tuple, frozenset]
    return any([isinstance(x, t) for t in iter_types])

def merge_dict_pair(dict1, dict2):
    keys = set(dict1.keys()).union(set(dict2.keys()))
    merged_dict = {}
    for k in keys:
        try:
            v = dict2[k]
        except KeyError:
            v = dict1[k]
        merged_dict[k] = v
    return merged_dict
        
def merge_dict(seq_of_dict):
    seq_of_dict = [x for x in seq_of_dict if x is not None]
    if len(seq_of_dict) == 0:
        return {}
    if len(seq_of_dict) == 1:
        return seq_of_dict[0]
    merged_dict = dict(seq_of_dict[0])
    for d in seq_of_dict[1:]:
        merged_dict = merge_dict_pair(merged_dict, d)
    return merged_dict

def smart_divide(x, y):
    if x == 0:
        return 0
    if y == 0:
        return INFINITY
    return x * 1.0 / y

def get_sign(x):
    if x > 0:
        return '+'
    if x < 0:
        return '-'
    return ''

def normalize_value(v, base_type=int, base_unit=''):
    if base_unit == '%':
        v *= 100
    v = round(v, 2)
    if base_type == int:
        v = int(v)
    return v
