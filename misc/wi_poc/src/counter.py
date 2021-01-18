from wi_poc.src.defaults import NUM_USERS_STR, NUM_EVENTS_STR,\
    DEFAULT_UID_FEAT, DEFAULT_UOI
from wi_poc.src.utils import smart_divide


def compute_match_counts(df, target_df, u, m, feat, v, uoi):
    """
    For base data `df`, and target `target_df`, compute counts
    corresponding to users (or events) matching the feature-value
    combination `feat`=`v`.

    Arguments
    ---------
    df : pd.DataFrame
        Base dataframe.
    target_df : pd.DataFrame
        Target dataframe.
    u : int
        Number of total users.
    m : int
        Number of total leads.
    feat : str
        Feature to be considered.
    v : str
        Value of `feat` to be considered.
    uoi : str
        Unit of interest (one of '#users' or '#events').

    Returns
    -------
    f
        #users (or #events) matching `feat`=`v`
    fm
        #leads matching `feat`=`v`
    pf
        Prevalence of `feat`=`v`
    pfm
        Prevalence of `feat`=`v` within leads
    crf
        Conversion rate of `feat`=`v`
    """
    if uoi == NUM_USERS_STR:
        f = df[[DEFAULT_UID_FEAT]][df[feat] == v][DEFAULT_UID_FEAT].nunique()
        fm = target_df[[DEFAULT_UID_FEAT]][target_df[feat] == v][DEFAULT_UID_FEAT].nunique()
    elif uoi == NUM_EVENTS_STR:
        f = df[df[feat] == v].shape[0]
        fm = target_df[target_df[feat] == v].shape[0]
    pf = smart_divide(f, u)
    pfm = smart_divide(fm, m)
    crf = smart_divide(fm, f)
    return f, fm, pf, pfm, crf


def compute_base_counts(df, target_df, uoi=DEFAULT_UOI):
    """
    For base data `df`, and target `target_df`, compute counts
    of total users (or events).

    Arguments
    ---------
    df : pd.DataFrame
        Base dataframe.
    target_df : pd.DataFrame
        Target dataframe.
    uoi : int
        Unit of interest (one of '#users' or '#events').

    Returns
    -------
    u
        #users (or #events) in `df`
    m
        #leads (i.e., those in `target_df`)
    cr
        Conversion rate
    """
    if uoi == NUM_EVENTS_STR:
        u = df.shape[0]
        m = target_df.shape[0]
    elif uoi == NUM_USERS_STR:
        u = df[DEFAULT_UID_FEAT].nunique()
        m = target_df[DEFAULT_UID_FEAT].nunique()
    cr = smart_divide(m, u)
    return u, m, cr
