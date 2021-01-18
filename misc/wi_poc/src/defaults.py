# Project IDs of clients.
CHARGEBEE = '399'
HIPPOVIDEO = '427'
LIVSPACE = '386'
IZOOTO = '446'
DESIGNCAFE = '483'

# Week information maps. TODO: To make this more dynamic.
AUG_WKS = ['01-07 Aug', '09-15 Aug', '16-22 Aug', '23-29 Aug', '30 Aug-05 Sep']
NOV_WKS = ['01-07 Nov', '08-14 Nov', '15-21 Nov', '22-28 Nov', '29 Nov-05 Dec']
DEC_WKS = ['29 Nov-05 Dec', '06-12 Dec', '13-19 Dec', '20-26 Dec', '27 Dec-02 Jan']
JAN_WKS = ['27 Dec-02 Jan', '03-09 Jan', '10-16 Jan', '17-23 Jan', '23-29 Jan, 30 Jan-05 Feb']
WKS = {'aug': AUG_WKS, 'nov': NOV_WKS, 'dec': DEC_WKS, 'jan': JAN_WKS}
WKS[8] = WKS['aug']
WKS[11] = WKS['nov']
WKS[12] = WKS['dec']
WKS[1] = WKS['jan']

# Some "base" and "target" event examples.
EVENTS_KEY = 'en'
FORM_SUBMITTED_KEY = '$form_submitted'
SESSION_KEY = '$session'
FORM_SUBMITTED_TARGET = (EVENTS_KEY, FORM_SUBMITTED_KEY, True)
SESSION_BASE = (EVENTS_KEY, SESSION_KEY, True)

# Some units of interest (UOI)
NUM_USERS_STR = '#users'
NUM_EVENTS_STR = '#events'

## SOME DEFAULTS
# Data defaults:
DEFAULT_PROJECT_ID = CHARGEBEE
DEFAULT_MODEL_ID = '1' # A sample test model curated for 5000 rows.
DEFAULT_TARGET = FORM_SUBMITTED_TARGET
DEFAULT_BASE = SESSION_BASE
DEFAULT_WK1_KEY = 'sample_wk1' # A sample test model curated for 5000 rows.
DEFAULT_WK2_KEY = 'sample_wk2' # A sample test model curated for 5000 rows.
DEFAULT_WK_PARAMS = {'project_id': DEFAULT_PROJECT_ID,
                     'model_id': DEFAULT_MODEL_ID}

# Feature-selection defaults:
CUSTOM_BLACKLIST = {'$initial_page_domain',
                             '$session_count',
                             '$initial_page_url',
                             '$page_title',
                             '$browser_version',
                             '$os_version',
                             '$user_agent',
                             '$gclid',
                             '$initial_gclid',
                             '$latest_gclid',
                             '$latest_fbclid',
                             '$fbclid',
                             '$initial_fbclid',
                             '$user_id',
                             '$day_of_first_event',
                             '$day_of_week',
                             '$identifiers',
                             '$initial_referrer',
                             '$initial_referrer_domain',
                             '$latest_page_raw_url',
                             '$initial_page_raw_url',
                             '$latest_referrer',
                             '$latest_referrer_domain',
                             '$name',
                             '$page_raw_url',
                             '$page_domain',
                             '$phone',
                             '$referrer',
                             '$referrer_raw_url',
                             '$referrer_domain',
                             '$salesforce_lead_createddate',
                             '$salesforce_lead_customer_whatsapp_optin__c',
                             '$salesforce_lead_date_when_meeting_is_scheduled__c',
                             '$salesforce_lead_email',
                             '$salesforce_lead_follow_up_date_time__c',
                             '$salesforce_lead_gclid__c',
                             '$salesforce_lead_id',
                             '$salesforce_lead_ipaddress__c',
                             '$salesforce_lead_lastmodifiedbyid',
                             '$salesforce_lead_lastmodifieddate',
                             '$salesforce_lead_lastname',
                             '$salesforce_lead_lead_allocation_time__c',
                             '$salesforce_lead_lead_qualified_date__c',
                             'salesforce_lead_mobile_number_external_field__c',
                             '$salesforce_lead_mobilephone',
                             '$salesforce_lead_mobileym__c',
                             '$salesforce_lead_page__c',
                             '$salesforce_lead_page_url__c',
                             '$salesforce_lead_pre_qualified_date__c',
                             '$session_latest_page_raw_url'}
DEFAULT_BLACKLISTS = {'custom', 'sf_lead', 'date'}
DEFAULT_MAX_NULL_PERC = 50
DEFAULT_MAX_UNIQ_PERC = 5
DEFAULT_MIN_UNIQ_COUNT = 2
DEFAULT_FEAT_SEL_PARAMS = {'max_null_perc': DEFAULT_MAX_NULL_PERC,
                           'max_uniq_perc': DEFAULT_MAX_UNIQ_PERC,
                           'min_uniq_count': DEFAULT_MIN_UNIQ_COUNT,
                           'blacklisted_features': DEFAULT_BLACKLISTS}

# Filter defaults:
DEFAULT_FILTERS_KEYS = ['nzc', 'zmc', 'msmcr', 'mscmcc', 'fifth', 'sixth']
DEFAULT_FILTER_PARAMS = {'min_fm1': 10, 'min_fm2': 10,
                         'min_f': 50, 'min_crf': 0.10,
                         'min_fc_f': 0.10, 'min_fc_fm': 0.10, 'min_fm': 10}

# Metric defaults:
DEFAULT_DIFFERENCE_METRIC_NAMES = ['delta_ratio', 'jsd_f', 'jsd_fm',
                                   'impact', 'change_in_scale', 'change_in_cr']
DEFAULT_NEGATION_PREDICT_MODE = 'diff-based'

# Miscelleneous defaults:
DEFAULT_UNIT_OF_INTEREST = NUM_USERS_STR
DEFAULT_UID_FEAT = 'uid'
DEFAULT_DEDUPLICATE_LOGIC = 'upr'
DEFAULT_UOI = NUM_USERS_STR

# Model information maps. TODO: To make this more dynamic.
MODEL_INFO = {
    CHARGEBEE: {
        DEFAULT_WK1_KEY: ('1', 5000),
        DEFAULT_WK2_KEY: ('2', 5000),
        WKS['dec'][1]: ('1608449537302', 229231),
        WKS['dec'][2]: ('1608452239024', 235741)},
    HIPPOVIDEO: {
        WKS['aug'][-3]: ('1598177830048', 72504),
        WKS['aug'][-2]: ('1598801710139', 85222)},
    LIVSPACE: {
        WKS['dec'][0]: ('1608707973413', 1572767),
        WKS['dec'][1]: ('1608727490184', 1527141),
        WKS['dec'][2]: ('1609346149884', 1909362),
        WKS['dec'][3]: ('1609386848912', 1579761),
        WKS['dec'][4]: (None, None)},
    IZOOTO: {
        WKS['dec'][0]: ('1607644474318', 261452),
        WKS['dec'][1]: ('1609783107212', 271565),
        WKS['dec'][2]: ('1609793254072', 243228),
        WKS['dec'][3]: ('1609802262328', 268491),
        WKS['dec'][4]: ('1609811114491', 302561),
        WKS['jan'][1]: ('1610519262142', 358134)},
    DESIGNCAFE: {
        WKS['dec'][0]: ('1607482505880', 98965),
        WKS['dec'][1]: ('1609779343714', 103357),
        WKS['dec'][2]: ('1609780522113', 98972),
        WKS['dec'][3]: ('1609781591132', 70183),
        WKS['dec'][4]: ('1609782315325', 73255),
        WKS['jan'][1]: ('1610439268600', 78418)}}

# Specify criteria to select "base" and "target" users for the insights.
# TODO: Make this more dynamic.
criteria_map = {DESIGNCAFE: {'base': SESSION_BASE,
                             'target': ('en',
                                        '$sf_lead_created',
                                        True)},
                IZOOTO: {'base': (['en', '$page_url'],
                                  ['$session', 'www.izooto.com/campaign/unsubscribing-from-web-push-notifications'],
                                  [True, False]),
                         'target': (['en', '$page_url'],
                                    ['$form_submitted', 'panel.izooto.com/signup'],
                                    [True, True])},
                CHARGEBEE: {'base': SESSION_BASE,
                            'target': FORM_SUBMITTED_TARGET},
                LIVSPACE: {'base': SESSION_BASE,
                           'target': FORM_SUBMITTED_TARGET}}
