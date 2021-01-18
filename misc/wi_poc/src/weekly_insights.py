import os
import pandas as pd
from smart_open import smart_open
from tqdm import tqdm
from wi_poc.src.defaults import DEFAULT_WK_PARAMS, MODEL_INFO, \
    DEFAULT_WK1_KEY, DEFAULT_WK2_KEY, DEFAULT_PROJECT_ID, \
    DEFAULT_BASE, DEFAULT_TARGET, DEFAULT_FILTERS_KEYS, \
    DEFAULT_BLACKLISTS, DEFAULT_UNIT_OF_INTEREST, \
    DEFAULT_DIFFERENCE_METRIC_NAMES
from wi_poc.src.config import DEFAULT_FEAT_SCHEMA_FILENAME, DEFAULT_CLOUD_PATH
from wi_poc.src.utils import is_iterable
from wi_poc.src.ingester import get_weekly_data
from wi_poc.src.feature_processor import get_criteria_filter, preselect_features,\
    compute_candidate_values
from wi_poc.src.preprocessor import sanitize_data
from wi_poc.src.counter import compute_base_counts, compute_match_counts
from wi_poc.src.filterer import compute_pass_filters, decide_pass_filters, get_filter_params
from wi_poc.src.metrics import compute_intermediate_metrics, compute_difference_metrics


def prepare_week_pair_params(project_id = None, wk1_key=DEFAULT_WK1_KEY, wk2_key=DEFAULT_WK2_KEY):
    wk1_params = dict(DEFAULT_WK_PARAMS)
    wk1_params['project_id'] = project_id
    wk1_params['model_id'], wk1_params['n_lines'] = MODEL_INFO[wk1_params['project_id']][wk1_key]

    wk2_params = dict(wk1_params)
    wk2_params['model_id'], wk2_params['n_lines'] = MODEL_INFO[wk1_params['project_id']][wk2_key]
    return wk1_params, wk2_params


def populate_filter_reject_stats(u1, u2, m1, m2, f1, f2, fm1, fm2, filters, filters_keys):
    frs = {'u1': u1, 'u2': u2,
            'm1': m1, 'm2': m2,
            'f1': f1, 'f2': f2,
            'fm1': fm1, 'fm2': fm2}
    frs.update({k: filters[k] for k in filters_keys})
    return frs


def format_criteria_name(criteria):
    feat, val, asn = criteria
    if not is_iterable(feat):
        feat = [feat]
        val = [val]
        asn = [asn]
    name_components = ['{}{}{}'.format(f, '=' if a else '!=', v)\
                            for f, v, a in zip(feat, val, asn)]
    name = '&'.join(name_components)
    name = name.replace('/', '_')
    return name
    

def store_insights(df, fr_df, out_folder, project_id, wk1_key, wk2_key, uoi, target):
    target_name = format_criteria_name(target)
    out_file_path = os.path.join(out_folder,
                                 '{}_delta_{}-{}-{}-{}.csv'.\
                                     format(project_id, wk1_key, wk2_key,
                                            uoi[1:], target_name))
    frs_file_path = os.path.join(out_folder,
                                 '{}_delta_{}-{}-{}-{}_filter_reject.csv'.\
                                     format(project_id, wk1_key, wk2_key,
                                            uoi[1:], target_name))
    df.to_csv(smart_open(out_file_path, 'w'))
    fr_df.to_csv(smart_open(frs_file_path, 'w'))


def generate_weekly_insights(project_id=DEFAULT_PROJECT_ID,
                             wk1_key=DEFAULT_WK1_KEY,
                             wk2_key=DEFAULT_WK2_KEY,
                             base=DEFAULT_BASE,
                             target=DEFAULT_TARGET,
                             feat_schema_filename=DEFAULT_FEAT_SCHEMA_FILENAME,
                             filters_keys=DEFAULT_FILTERS_KEYS,
                             filter_params_mode=None,
                             blacklists=DEFAULT_BLACKLISTS,
                             uoi=DEFAULT_UNIT_OF_INTEREST,
                             difference_metric_names=DEFAULT_DIFFERENCE_METRIC_NAMES):
    wk1_params, wk2_params = prepare_week_pair_params(project_id,
        wk1_key, wk2_key)
    print('\nReading week 1 data')
    df1 = get_weekly_data(**wk1_params, base=base)
    print('\nReading week 2 data')
    df2 = get_weekly_data(**wk2_params, base=base)

    feats, target_compliant_fvs, exp_feats = preselect_features(df1, df2, target,
        feat_schema_filename, blacklists)
    df1, df2 = df1[feats], df2[feats]
    df1, df2 = sanitize_data(df1, feats), sanitize_data(df2, feats)

    target_df1 = df1[get_criteria_filter(df1, target)]
    target_df2 = df2[get_criteria_filter(df2, target)]

    u1, m1, cr1 = compute_base_counts(df1, target_df1, uoi)
    u2, m2, cr2 = compute_base_counts(df2, target_df2, uoi)

    filter_params = get_filter_params({'u1':u1, 'u2':u2,
                                       'm1':m1, 'm2':m2,
                                       'cr1': cr1, 'cr2': cr2}, mode='bucketed')
    stat_dict = {}
    fr_stats = {}
    for f in tqdm(exp_feats, 'Final explanations'):
        values = compute_candidate_values(df1, df2, f, target_compliant_fvs)
        for v in values:
            fv_str = '{}={}'.format(f, v)
            f1, fm1, pf1, pfm1, crf1 = compute_match_counts(df1, target_df1, u1, m1, f, v, uoi)
            f2, fm2, pf2, pfm2, crf2 = compute_match_counts(df2, target_df2, u2, m2, f, v, uoi)
            filters = compute_pass_filters(f1, f2, fm1, fm2, filter_params)
            pass_decision = decide_pass_filters(filters, filters_keys)
            if not pass_decision:
                fr_stats[fv_str] = populate_filter_reject_stats(u1, u2, m1, m2,
                                                                f1, f2, fm1, fm2,
                                                                filters, filters_keys)
                continue
            stat_dict[fv_str] = {'u1': u1, 'u2': u2, 'm1': m1, 'm2': m2,
                                'cr1': cr1, 'cr2': cr2, 'f1': f1, 'f2': f2,
                                'pf1': pf1, 'pf2': pf2, 'fm1': fm1, 'fm2': fm2,
                                'crf1': crf1, 'crf2': crf2, 'pfm1': pfm1, 'pfm2': pfm2}
    df = pd.DataFrame(stat_dict).T
    fr_df = pd.DataFrame(fr_stats).T
    if df.shape[0] > 0:
        compute_intermediate_metrics(df)
        compute_difference_metrics(df, difference_metric_names)
    store_insights(df, fr_df, DEFAULT_CLOUD_PATH, project_id,
                   wk1_key, wk2_key, uoi, target)


def main():
    generate_weekly_insights()


if __name__ == '__main__':
    main()
