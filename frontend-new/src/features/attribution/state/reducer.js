import { DefaultDateRangeFormat } from 'Views/CoreQuery/utils';
import {
  SET_EVENT_GOAL,
  SET_TOUCHPOINTS,
  INITIALIZE_TOUCHPOINT_DIMENSIONS,
  INITIALIZE_CONTENT_GROUPS,
  SET_TOUCHPOINT_FILTERS,
  SET_ATTR_QUERY_TYPE,
  SET_ATTRIBUTION_MODEL,
  SET_ATTRIBUTION_WINDOW,
  SET_ATTR_LINK_EVENTS,
  SET_TACTIC_OFFER_TYPE,
  ATTRIBUTION_DASHBOARD_UNITS_FAILED,
  ATTRIBUTION_DASHBOARD_UNITS_LOADED,
  ATTRIBUTION_DASHBOARD_UNITS_LOADING,
  SET_ATTR_DATE_RANGE,
  SET_ATTR_QUERIES,
  INITIALIZE_ATTRIBUTION_STATE
} from './action.constants';

const defaultState = {
  eventGoal: { filters: [] },
  touchpoint: '',
  attr_dimensions: [],
  content_groups: [],
  touchpoint_filters: [],
  attr_query_type: 'EngagementBased',
  models: [],
  window: null,
  linkedEvents: [],
  tacticOfferType: 'TacticOffer',
  attr_dateRange: {
    ...DefaultDateRangeFormat,
    dateStr: ''
  },
  attrQueries: [],
  attributionDashboardUnits: {
    loading: false,
    error: false,
    data: []
  }
};

export default function (state = defaultState, action) {
  switch (action.type) {
    case ATTRIBUTION_DASHBOARD_UNITS_LOADING:
      return {
        ...state,
        attributionDashboardUnits: {
          ...defaultState.attributionDashboardUnits,
          loading: true
        }
      };

    case ATTRIBUTION_DASHBOARD_UNITS_LOADED:
      return {
        ...state,
        attributionDashboardUnits: {
          ...defaultState.attributionDashboardUnits,
          data: action.payload
        }
      };

    case ATTRIBUTION_DASHBOARD_UNITS_FAILED:
      return {
        ...state,
        attributionDashboardUnits: {
          ...defaultState.attributionDashboardUnits,
          error: true
        }
      };

    case SET_EVENT_GOAL: {
      return {
        ...state,
        eventGoal: action.payload
      };
    }
    case SET_TOUCHPOINTS: {
      return {
        ...state,
        touchpoint: action.payload
      };
    }
    case INITIALIZE_TOUCHPOINT_DIMENSIONS: {
      return {
        ...state,
        attr_dimensions: action.payload
      };
    }
    case INITIALIZE_CONTENT_GROUPS: {
      return {
        ...state,
        content_groups: action.payload
      };
    }
    case SET_TOUCHPOINT_FILTERS: {
      return {
        ...state,
        touchpoint_filters: action.payload
      };
    }
    case SET_ATTR_QUERY_TYPE: {
      return {
        ...state,
        attr_query_type: action.payload
      };
    }
    case SET_ATTRIBUTION_MODEL: {
      if (action.payload.length > 1) {
        return {
          ...state,
          models: action.payload,
          linkedEvents: [] // clear linked events if comparison model is added
        };
      }
      return {
        ...state,
        models: action.payload
      };
    }
    case SET_ATTRIBUTION_WINDOW: {
      return {
        ...state,
        window: action.payload
      };
    }
    case SET_ATTR_LINK_EVENTS: {
      return {
        ...state,
        linkedEvents: action.payload
      };
    }
    case SET_TACTIC_OFFER_TYPE: {
      return {
        ...state,
        tacticOfferType: action.payload
      };
    }
    case SET_ATTR_DATE_RANGE: {
      return {
        ...state,
        attr_dateRange: action.payload
      };
    }
    case SET_ATTR_QUERIES: {
      return {
        ...state,
        attrQueries: action.payload
      };
    }

    case INITIALIZE_ATTRIBUTION_STATE: {
      return {
        ...state,
        ...action.payload
      };
    }
    default: {
      return { ...state };
    }
  }
}
