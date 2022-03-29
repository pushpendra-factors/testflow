import { ATTRIBUTION_METRICS, DefaultChartTypes } from '../../utils/constants';
import { EMPTY_ARRAY, EMPTY_OBJECT } from '../../utils/global';
import { PIVOT_SORT_ORDERS } from '../../components/PivotTableControls/pivotTableControls.constants';

export const INITIAL_STATE = {
  loading: false,
  error: false,
  data: null,
  apiCallStatus: { required: true, message: null },
};

export const SET_ATTRIBUTION_METRICS = 'SET_ATTRIBUTION_METRICS';
export const SET_NAVIGATED_FROM_DASHBOARD = 'SET_NAVIGATED_FROM_DASHBOARD';
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

export const DEFAULT_PIVOT_CONFIG = {
  rows: EMPTY_ARRAY,
  cols: EMPTY_ARRAY,
  vals: EMPTY_ARRAY,
  aggregatorName: 'Sum',
  rowOrder: PIVOT_SORT_ORDERS.ASCEND,
  configLoaded: false,
};

export const CORE_QUERY_INITIAL_STATE = {
  comparison_data: { ...INITIAL_STATE },
  comparison_supported: false,
  comparison_enabled: false,
  navigatedFromDashboard: false,
  comparison_duration: null,
  attributionMetrics: [...ATTRIBUTION_METRICS],
  chartTypes: DefaultChartTypes,
  savedQuerySettings: EMPTY_OBJECT,
  pivotConfig: DEFAULT_PIVOT_CONFIG,
};

export const FILTER_TYPES = {
  CATEGORICAL: 'categorical',
  DATETIME: 'datetime',
};
