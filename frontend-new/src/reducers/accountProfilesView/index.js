import { SET_ACTIVE_PROJECT } from 'Reducers/types';
import { SEGMENT_DELETED } from 'Reducers/timelines/types';
import { GROUP_NAME_DOMAINS } from 'Components/GlobalFilter/FilterWrapper/utils';
import { apiStates } from 'Reducers/dashboard/constants';
import { EMPTY_ARRAY } from 'Utils/global';
import MomentTz from 'Components/MomentTz';
import {
  SET_ACCOUNT_PAYLOAD,
  SET_ACCOUNT_SEGMENT_MODAL,
  UPDATE_ACCOUNT_PAYLOAD,
  ENABLE_NEW_SEGMENT_MODE,
  DISABLE_NEW_SEGMENT_MODE,
  SET_FILTERS_DIRTY,
  TOGGLE_ACCOUNTS_TAB,
  ACCOUNTS_INSIGHTS_CONFIG_LOADING,
  ACCOUNTS_INSIGHTS_CONFIG_SUCCESS,
  ACCOUNTS_INSIGHTS_CONFIG_ERROR,
  ACCOUNTS_INSIGHTS_LOADING,
  ACCOUNTS_INSIGHTS_ERROR,
  ACCOUNTS_INSIGHTS_SUCCESS,
  SET_INSIGHTS_DURATION,
  SET_INSIGHTS_COMPARE_SEGMENT,
  EDIT_INSIGHTS_METRIC_LOADING,
  EDIT_INSIGHTS_METRIC_SUCCESS,
  EDIT_INSIGHTS_METRIC_ERROR
} from './types';

export function generateInsightsKey({
  widgetGroupId,
  segmentId,
  dateFrom,
  dateTo
}) {
  return `${widgetGroupId}_${segmentId}_${MomentTz(dateFrom).format(
    'YYYY-MM-DD'
  )}_${MomentTz(dateTo).format('YYYY-MM-DD')}`;
}

export const INITIAL_ACCOUNT_PAYLOAD = {
  source: GROUP_NAME_DOMAINS
};

const initialState = {
  accountPayload: INITIAL_ACCOUNT_PAYLOAD,
  showSegmentModal: false,
  newSegmentMode: false,
  filtersDirty: false,
  activeTab: 'accounts', // accounts | insights,
  insightsConfig: {
    ...apiStates,
    config: EMPTY_ARRAY,
    dateRange: {}
  },
  insights: {},
  insightsCompareConfig: {},
  editInsightsMetric: {
    ...apiStates
  }
};

export default function (state = initialState, action) {
  switch (action.type) {
    case SET_ACCOUNT_PAYLOAD:
      return {
        ...state,
        accountPayload: action.payload
      };
    case UPDATE_ACCOUNT_PAYLOAD:
      return {
        ...state,
        accountPayload: {
          ...state.accountPayload,
          ...action.payload
        }
      };
    case SET_ACCOUNT_SEGMENT_MODAL: {
      return {
        ...state,
        showSegmentModal: action.payload
      };
    }
    case ENABLE_NEW_SEGMENT_MODE: {
      return {
        ...state,
        newSegmentMode: true
      };
    }
    case DISABLE_NEW_SEGMENT_MODE: {
      return {
        ...state,
        newSegmentMode: false
      };
    }
    case SET_ACTIVE_PROJECT: {
      return {
        ...initialState
      };
    }
    case SEGMENT_DELETED: {
      return {
        ...state,
        accountPayload: {
          ...INITIAL_ACCOUNT_PAYLOAD
        }
      };
    }
    case SET_FILTERS_DIRTY: {
      return {
        ...state,
        filtersDirty: action.payload
      };
    }
    case TOGGLE_ACCOUNTS_TAB:
      return {
        ...state,
        activeTab: action.payload
      };
    case ACCOUNTS_INSIGHTS_CONFIG_LOADING: {
      return {
        ...state,
        insightsConfig: {
          ...state.insightsConfig,
          ...apiStates,
          loading: true,
          config: EMPTY_ARRAY
        }
      };
    }
    case ACCOUNTS_INSIGHTS_CONFIG_SUCCESS: {
      return {
        ...state,
        insightsConfig: {
          ...state.insightsConfig,
          ...apiStates,
          completed: true,
          config: action.payload
        }
      };
    }
    case ACCOUNTS_INSIGHTS_CONFIG_ERROR: {
      return {
        ...state,
        insightsConfig: {
          ...state.insightsConfig,
          ...apiStates,
          error: true,
          config: EMPTY_ARRAY
        }
      };
    }
    case ACCOUNTS_INSIGHTS_LOADING: {
      const { widgetGroupId, segmentId, dateFrom, dateTo } = action.payload;
      const key = generateInsightsKey({
        widgetGroupId,
        segmentId,
        dateFrom,
        dateTo
      });
      return {
        ...state,
        insights: {
          ...state.insights,
          [key]: {
            ...apiStates,
            loading: true
          }
        }
      };
    }
    case ACCOUNTS_INSIGHTS_ERROR: {
      const { widgetGroupId, segmentId, dateFrom, dateTo } = action.payload;
      const key = generateInsightsKey({
        widgetGroupId,
        segmentId,
        dateFrom,
        dateTo
      });
      return {
        ...state,
        insights: {
          ...state.insights,
          [key]: {
            ...apiStates,
            error: true
          }
        }
      };
    }
    case ACCOUNTS_INSIGHTS_SUCCESS: {
      const { widgetGroupId, segmentId, dateFrom, dateTo } = action.payload;
      const key = generateInsightsKey({
        widgetGroupId,
        segmentId,
        dateFrom,
        dateTo
      });
      return {
        ...state,
        insights: {
          ...state.insights,
          [key]: {
            ...apiStates,
            completed: true,
            data: action.payload.data
          }
        }
      };
    }
    case SET_INSIGHTS_DURATION: {
      return {
        ...state,
        insightsConfig: {
          ...state.insightsConfig,
          dateRange: {
            [action.payload.segmentId]: action.payload.range
          }
        }
      };
    }
    case SET_INSIGHTS_COMPARE_SEGMENT: {
      return {
        ...state,
        insightsCompareConfig: {
          ...state.insightsCompareConfig,
          [action.payload.segmentId]: action.payload.compareSegmentId
        }
      };
    }
    case EDIT_INSIGHTS_METRIC_LOADING: {
      return {
        ...state,
        editInsightsMetric: {
          ...apiStates,
          loading: true
        }
      };
    }
    case EDIT_INSIGHTS_METRIC_SUCCESS: {
      return {
        ...state,
        editInsightsMetric: {
          ...apiStates,
          completed: true
        }
      };
    }
    case EDIT_INSIGHTS_METRIC_ERROR: {
      return {
        ...state,
        editInsightsMetric: {
          ...apiStates,
          error: true
        }
      };
    }
    default:
      return state;
  }
}
