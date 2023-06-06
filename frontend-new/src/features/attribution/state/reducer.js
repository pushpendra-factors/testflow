import unset from 'lodash/unset';
import { DefaultDateRangeFormat } from 'Views/CoreQuery/utils';
import { getRearrangedData } from 'Reducers/dashboard/utils';
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
  ATTRIBUTION_DASHBOARD_UNITS_UPDATED,
  SET_ATTR_DATE_RANGE,
  SET_ATTR_QUERIES,
  INITIALIZE_ATTRIBUTION_STATE,
  ATTRIBUTION_DASHBOARD_LOADING,
  ATTRIBUTION_DASHBOARD_LOADED,
  ATTRIBUTION_DASHBOARD_FAILED,
  ATTRIBUTION_QUERIES_LOADED,
  ATTRIBUTION_QUERIES_LOADING,
  ATTRIBUTION_QUERIES_FAILED,
  ATTRIBUTION_WIDGET_DELETED,
  ATTRIBUTION_QUERY_DELETED,
  ATTRIBUTION_QUERY_CREATED,
  ATTRIBUTION_QUERY_UPDATED,
  ATTRIBUTION_DASHBOARD_UNIT_ADDED
} from './action.constants';
import { SET_ACTIVE_PROJECT } from 'Reducers/types';

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
  },
  attributionQueries: {
    loading: false,
    error: false,
    data: []
  },
  dashboardLoading: false,
  dashboardLoadFailed: false,
  dashboard: null
};

export default function (state = defaultState, action) {
  switch (action.type) {
    case ATTRIBUTION_DASHBOARD_UNITS_LOADING:
      return {
        ...state,
        attributionDashboardUnits: {
          ...state.attributionDashboardUnits,
          loading: true
        }
      };

    case ATTRIBUTION_DASHBOARD_UNITS_LOADED:
      return {
        ...state,
        attributionDashboardUnits: {
          ...state.attributionDashboardUnits,
          loading: false,
          data: getRearrangedData(action.payload, state.dashboard)
        }
      };

    case ATTRIBUTION_DASHBOARD_UNITS_FAILED:
      return {
        ...state,
        attributionDashboardUnits: {
          ...state.attributionDashboardUnits,
          loading: false,
          error: true
        }
      };

    case ATTRIBUTION_DASHBOARD_UNITS_UPDATED: {
      return {
        ...state,
        attributionDashboardUnits: {
          ...state.attributionDashboardUnits,
          data: [...action.payload]
        },
        dashboard: {
          ...state.dashboard,
          units_position: action.units_position
        }
      };
    }

    case ATTRIBUTION_DASHBOARD_UNIT_ADDED: {
      return {
        ...state,
        attributionDashboardUnits: {
          ...state.attributionDashboardUnits,
          data: [...state.attributionDashboardUnits.data, action.payload]
        }
      };
    }

    case ATTRIBUTION_WIDGET_DELETED: {
      const updatedUnitsPosition = { ...state.dashboard.units_position };
      unset(updatedUnitsPosition, `position.${action.payload}`);
      unset(updatedUnitsPosition, `size.${action.payload}`);
      return {
        ...state,
        attributionDashboardUnits: {
          ...state.attributionDashboardUnits,
          data: state.attributionDashboardUnits.data.filter(
            (elem) => elem.id !== action.payload
          )
        },
        dashboard: {
          ...state.dashboard,
          units_position: updatedUnitsPosition
        }
      };
    }

    case ATTRIBUTION_DASHBOARD_LOADING:
      return {
        ...state,
        dashboardLoading: true
      };

    case ATTRIBUTION_DASHBOARD_LOADED:
      return {
        ...state,
        dashboardLoading: false,
        dashboard: action.payload
      };
    case ATTRIBUTION_DASHBOARD_FAILED:
      return {
        ...state,
        dashboardLoading: false,
        dashboardLoadFailed: true
      };

    case ATTRIBUTION_QUERIES_LOADING:
      return {
        ...state,
        attributionQueries: {
          ...state.attributionQueries,
          loading: true
        }
      };
    case ATTRIBUTION_QUERIES_FAILED:
      return {
        ...state,
        attributionQueries: {
          ...state.attributionQueries,
          loading: false,
          error: true
        }
      };
    case ATTRIBUTION_QUERIES_LOADED:
      return {
        ...state,
        attributionQueries: {
          ...state.attributionQueries,
          loading: false,
          data: action.payload
        }
      };

    case ATTRIBUTION_QUERY_CREATED:
      return {
        ...state,
        attributionQueries: {
          ...state.attributionQueries,
          data: [action.payload, ...state.attributionQueries.data]
        }
      };

    case ATTRIBUTION_QUERY_UPDATED: {
      const queries = state.attributionQueries.data;
      const updatedQueryIndex = queries.findIndex(
        (d) => d.id === action.queryId
      );
      if (updatedQueryIndex > -1) {
        return {
          ...state,
          attributionQueries: {
            ...state.attributionQueries,
            data: [
              ...queries.slice(0, updatedQueryIndex),
              { ...queries[updatedQueryIndex], ...action.payload },
              ...queries.slice(updatedQueryIndex + 1)
            ]
          }
        };
      }
      return state;
    }

    case ATTRIBUTION_QUERY_DELETED: {
      const queries = state.attributionQueries.data;
      let queryIndex = queries.findIndex((d) => d.id === action.payload);
      return {
        ...state,
        attributionQueries: {
          ...state.attributionQueries,
          data:
            queryIndex > -1
              ? [
                  ...queries.slice(0, queryIndex),
                  ...queries.slice(queryIndex + 1)
                ]
              : queries
        }
      };
    }

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
    case SET_ACTIVE_PROJECT: {
      return {
        ...defaultState
      };
    }
    default: {
      return { ...state };
    }
  }
}
