import { DefaultChartTypes } from '../../utils/constants';
import { EMPTY_ARRAY, EMPTY_OBJECT } from '../../utils/global';
import { PIVOT_SORT_ORDERS } from '../../components/PivotTableControls/pivotTableControls.constants';

export const INITIAL_STATE = {
  loading: false,
  error: false,
  data: null,
  apiCallStatus: { required: true, message: null }
};

export const SET_NAVIGATED_FROM_DASHBOARD = 'SET_NAVIGATED_FROM_DASHBOARD';
export const SET_NAVIGATED_FROM_ANALYSE = 'SET_NAVIGATED_FROM_ANALYSE';
export const SET_COMPARISON_ENABLED = 'SET_COMPARISON_ENABLED';
export const SET_COMPARISON_SUPPORTED = 'SET_COMPARISON_SUPPORTED';
export const COMPARISON_DATA_LOADING = 'COMPARISON_DATA_LOADING';
export const COMPARISON_DATA_FETCHED = 'COMPARISON_DATA_FETCHED';
export const COMPARISON_DATA_ERROR = 'COMPARISON_DATA_ERROR';
export const RESET_COMPARISON_DATA = 'RESET_COMPARISON_DATA';
export const SET_COMPARE_DURATION = 'SET_COMPARE_DURATION';
export const UPDATE_CHART_TYPES = 'UPDATE_CHART_TYPES';
export const SET_SAVED_QUERY_SETTINGS = 'SET_SAVED_QUERY_SETTINGS';
export const UPDATE_PIVOT_CONFIG = 'UPDATE_PIVOT_CONFIG';
export const UPDATE_FUNNEL_TABLE_CONFIG = 'UPDATE_FUNNEL_TABLE_CONFIG';
export const UPDATE_CORE_QUERY_REDUCER = 'UPDATE_CORE_QUERY_REDUCER';

export const DEFAULT_PIVOT_CONFIG = {
  rows: EMPTY_ARRAY,
  cols: EMPTY_ARRAY,
  vals: EMPTY_ARRAY,
  aggregatorName: 'Integer Sum',
  rowOrder: PIVOT_SORT_ORDERS.ASCEND,
  configLoaded: false
};

export const DEFAULT_FUNNEL_TABLE_CONFIG = [
  {
    title: 'Show Count of users',
    enabled: true,
    disabled: true,
    key: 'showCount'
  },
  {
    title: 'Show Conv. from previous step',
    enabled: false,
    disabled: false,
    key: 'showPercentage'
  },
  {
    title: 'Show Duration from previous step',
    enabled: false,
    disabled: false,
    key: 'showDuration'
  }
];

export const DEFAULT_ATTRIBUTION_TABLE_FILTERS = EMPTY_OBJECT;

export const CORE_QUERY_INITIAL_STATE = {
  comparison_data: { ...INITIAL_STATE },
  comparison_supported: false,
  comparison_enabled: false,
  navigatedFromDashboard: false,
  navigatedFromAnalyse: false,
  navigatedFromDashboardExistingReports: false,
  comparison_duration: null,
  chartTypes: DefaultChartTypes,
  savedQuerySettings: EMPTY_OBJECT,
  pivotConfig: DEFAULT_PIVOT_CONFIG,
  funnelTableConfig: DEFAULT_FUNNEL_TABLE_CONFIG,
  funnelConversionDurationNumber: '90',
  funnelConversionDurationUnit: 'D', // D/M/H ---> days/minutes/hours
  attributionTableFilters: DEFAULT_ATTRIBUTION_TABLE_FILTERS
};

export const FILTER_TYPES = {
  CATEGORICAL: 'categorical',
  DATETIME: 'datetime'
};

export const BREAKDOWN_TYPES = {
  DATETIME: 'datetime'
};

export const EVENT_FREQ_OPERATORS = {
  equals: 'equals',
  'lesser than': 'lesserThan',
  'lesser than or equals': 'lesserThanOrEqual',
  'greater than': 'greaterThan',
  'greater than or equals': 'greaterThanOrEqual'
};

export const INITIAL_EVENT_WITH_PROPERTIES_STATE = {
  label: '',
  filters: [],
  group: '',
  isEventPerformed: true,
  frequencyOperator: EVENT_FREQ_OPERATORS['greater than or equals'],
  frequency: 1,
  range: 30
};
