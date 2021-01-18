import numpy as np
from tqdm import tqdm

def simple_impute(x):
    str_nulls = {'', 'NaN', 'null', None}
    int_nulls = {np.nan, float('NaN'), None}
    if isinstance(x, str):
        if x in str_nulls or x is None:
            return None
        else:
            return x
    else:
        if x in int_nulls or x is None or x is np.nan:
            return None
        else:
            return x


def sanitize_data(df, feats_to_sanitize):
    for f in tqdm(feats_to_sanitize, desc='Sanitizing'):
        df[f] = df[f].apply(simple_impute)
    return df


def sanitize_screen_size_params(df,
                                wd_col='$screen_width',
                                ht_col='$screen_height',
                                sz_col='$screen_size'):
    cols = df.columns
    if wd_col in cols and ht_col in cols:
        df[wd_col].fillna(0, inplace=True)
        df[ht_col].fillna(0, inplace=True)
        df[sz_col] = df.apply(lambda x: '{}x{}'.format(
            int(x[wd_col]), int(x[ht_col])), axis=1)
        df.drop([wd_col, ht_col], axis=1, inplace=True)
    return df