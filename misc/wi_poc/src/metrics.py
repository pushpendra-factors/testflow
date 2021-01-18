import numpy as np
from wi_poc.src.utils import smart_divide, frac_change
from wi_poc.src.defaults import DEFAULT_NEGATION_PREDICT_MODE, \
    DEFAULT_DIFFERENCE_METRIC_NAMES

def kl_divergence(p, q):
    return sum(p[i] * np.log2(p[i]/q[i]) for i in range(len(p)))


def js_divergence(p, q):
    p = np.array(p)
    q = np.array(q)
    m = 0.5 * (p + q)
    return 0.5 * kl_divergence(p, m) + 0.5 * kl_divergence(q, m)

def compute_intermediate_metrics(df, negation_predict_mode=DEFAULT_NEGATION_PREDICT_MODE):
    df['smooth_pf1'] = (df['f1'] + 1)/(df['u1'] + 1)
    df['smooth_pf2'] = (df['f2'] + 1)/(df['u2'] + 1)
    df['smooth_pfm1'] = (df['fm1'] + 1)/(df['m1'] + 1)
    df['smooth_pfm2'] = (df['fm2'] + 1)/(df['m2'] + 1)
    df['notf1'] = df['u1'] - df['f1']
    df['notf2'] = df['u2'] - df['f2']
    df['notfm1'] = df['m1'] - df['fm1']
    df['notfm2'] = df['m2'] - df['fm2']
    df['cr_notf1'] = df.apply(lambda x: smart_divide(x['notfm1'], x['notf1']), axis=1)
    df['cr_notf2'] = df.apply(lambda x: smart_divide(x['notfm2'], x['notf2']), axis=1)

    df['m2_pred'] = df['cr1'] * df['u1']
    df['fm2_pred'] = df['crf1'] * df['f2']
    if negation_predict_mode == 'cr-based':
        df['notfm2_pred'] = df['cr_notf1'] * df['notf2']
    elif negation_predict_mode == 'diff-based':
        df['notfm2_pred'] = df['m2_pred'] - df['fm2_pred']


def compute_difference_metrics(df, difference_metric_names=DEFAULT_DIFFERENCE_METRIC_NAMES):
    if 'delta_ratio' in difference_metric_names:
        df['D_fm'] = df['fm2_pred'] - df['fm2']
        df['D_notfm'] = df['notfm2_pred'] - df['notfm2']
        df['delta_ratio'] = df.apply(lambda x: smart_divide(np.abs(x['D_fm']),
                                                            np.abs(x['D_fm'] + x['D_notfm'])),
                                     axis=1)
    if 'kld' in difference_metric_names:
        df['kld'] = df.apply(lambda x: kl_divergence([x['smooth_pf1'], 1-x['smooth_pf1']],
                                                     [x['smooth_pf2'], 1-x['smooth_pf2']]),
                             axis=1)
    if 'jsd_f' in difference_metric_names:
        df['jsd_f'] = df.apply(lambda x: js_divergence([x['smooth_pf1'], 1-x['smooth_pf1']],
                                                     [x['smooth_pf2'], 1-x['smooth_pf2']]),
                             axis=1)
    if 'jsd_fm' in difference_metric_names:
        df['jsd_fm'] = df.apply(lambda x: js_divergence([x['smooth_pfm1'], 1-x['smooth_pfm1']],
                                                     [x['smooth_pfm2'], 1-x['smooth_pfm2']]),
                             axis=1)
    if 'impact' in difference_metric_names:
        df['impact'] = df.apply(lambda x: smart_divide(x['fm2'], x['fm1']), axis=1)
    if 'change_in_scale' in difference_metric_names:
        df['change_in_scale'] = df.apply(lambda x: frac_change(x['f1'], x['f2'], False),
                                         axis=1)
    if 'change_in_cr' in difference_metric_names:
        df['change_in_cr'] = df.apply(lambda x: frac_change(x['crf1'], x['crf2'], False),
                                      axis=1)
