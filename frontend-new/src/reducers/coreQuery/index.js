/* eslint-disable */

import {
  FETCH_EVENTS,
  FETCH_EVENT_PROPERTIES,
  FETCH_USER_PROPERTIES,
  SET_GROUPBY,
  DEL_GROUPBY,
  INITIALIZE_GROUPBY,
  DEL_GROUPBY_EVENT,
  SET_TOUCHPOINTS,
  SET_TOUCHPOINT_FILTERS,
  SET_ATTR_QUERY_TYPE,
  SET_TACTIC_OFFER_TYPE,
  SET_ATTR_DATE_RANGE,
  SET_ATTRIBUTION_MODEL,
  SET_ATTRIBUTION_WINDOW,
  SET_ATTR_LINK_EVENTS,
  SET_EVENT_GOAL,
  SET_CAMP_CHANNEL,
  SET_CAMP_MEASURES,
  SET_CAMP_FILTERS,
  FETCH_CAMP_CONFIG,
  SET_CAMP_GROUBY,
  SET_CAMP_DATE_RANGE,
  SET_DEFAULT_STATE,
  SET_EVENT_NAMES,
  SET_USER_PROP_NAME,
  SET_EVENT_PROP_NAME,
} from './actions';
import {
  SHOW_ANALYTICS_RESULT,
  INITIALIZE_MTA_STATE,
  INITIALIZE_CAMPAIGN_STATE,
  INITIALIZE_TOUCHPOINT_DIMENSIONS,
} from '../types';
import { DefaultDateRangeFormat } from '../../Views/CoreQuery/utils';

const DEFAULT_TOUCHPOINTS = [
  'Campaign',
  'Source',
  'AdGroup',
  'Keyword',
  'Channel',
];

const defaultState = {
  eventOptions: [],
  eventProperties: {},
  userProperties: [],
  groupBy: {
    global: [],
    event: [],
  },
  touchpointOptions: [
    {
      label: 'Digital',
      icon: 'fav',
      values: DEFAULT_TOUCHPOINTS.map((v) => [v]),
    },
  ],
  show_analytics_result: false,
  eventGoal: {},
  touchpoint: '',
  touchpoint_filters: [],
  attr_query_type: 'EngagementBased',
  tacticOfferType: '',
  attr_dateRange: {
    ...DefaultDateRangeFormat,
    dateStr: '',
  },
  attr_dimensions: [],
  models: [],
  window: null,
  linkedEvents: [],
  campaign_config: {
    metrics: [],
    properties: [],
  },
  camp_channels: 'google_ads',
  camp_measures: [],
  camp_filters: [],
  camp_groupBy: [],
  camp_dateRange: {
    ...DefaultDateRangeFormat,
    frequency: 'date',
    dateStr: '',
  },
  eventNames: [],
  userPropNames: [],
  eventPropNames: [],
};

export default function (state = defaultState, action) {
  switch (action.type) {
    case FETCH_EVENTS:
      return { ...state, eventOptions: action.payload };
    case SET_EVENT_NAMES:
      return { ...state, eventNames: action.payload };
    case SET_USER_PROP_NAME:
      return { ...state, userPropNames: action.payload };
    case FETCH_USER_PROPERTIES:
      return { ...state, userProperties: action.payload };
    case FETCH_EVENT_PROPERTIES:
      const eventPropState = Object.assign({}, state.eventProperties);
      eventPropState[action.eventName] = action.payload;
      return { ...state, eventProperties: eventPropState };
    case SET_EVENT_PROP_NAME:
      const evnPropNames = { ...state.eventPropNames, ...action.payload };
      return { ...state, eventPropNames: evnPropNames };
    case INITIALIZE_GROUPBY: {
      return {
        ...state,
        groupBy: action.payload,
      };
    }
    case DEL_GROUPBY: {
      return {
        ...state,
        groupBy: {
          ...state.groupBy,
          [action.groupByType]: state.groupBy[action.groupByType]
            .filter((gb) => {
              return gb.overAllIndex !== action.payload.overAllIndex;
            })
            .map((gb) => {
              if (gb.overAllIndex > action.payload.overAllIndex) {
                return {
                  ...gb,
                  overAllIndex: gb.overAllIndex - 1,
                };
              }
              return gb;
            }),
        },
      };
    }

    case SET_GROUPBY:
      let groupByState = Object.assign({}, state.groupBy);
      if (
        groupByState[action.groupByType] &&
        groupByState[action.groupByType][action.index]
      ) {
        groupByState[action.groupByType][action.index] = action.payload;
      } else if (
        groupByState[action.groupByType] &&
        action.index === groupByState[action.groupByType].length
      ) {
        groupByState[action.groupByType].push({
          ...action.payload,
          overAllIndex: groupByState[action.groupByType].length,
        });
      }
      return { ...state, groupBy: groupByState };
    case DEL_GROUPBY_EVENT: {
      return {
        ...state,
        groupBy: {
          ...state.groupBy,
          event: state.groupBy.event
            .filter((gb) => {
              return gb.eventIndex - 1 !== action.index;
            })
            .map((gb) => {
              if (gb.eventIndex > action.index) {
                return {
                  ...gb,
                  eventIndex: gb.eventIndex - 1,
                };
              }
              return gb;
            }),
        },
      };
    }
    case SHOW_ANALYTICS_RESULT: {
      return {
        ...state,
        show_analytics_result: action.payload,
      };
    }
    case SET_EVENT_GOAL: {
      return {
        ...state,
        eventGoal: action.payload,
      };
    }
    case SET_TOUCHPOINTS: {
      return {
        ...state,
        touchpoint: action.payload,
      };
    }
    case INITIALIZE_TOUCHPOINT_DIMENSIONS: {
      return {
        ...state,
        attr_dimensions: action.payload,
      };
    }
    case SET_TOUCHPOINT_FILTERS: {
      return {
        ...state,
        touchpoint_filters: action.payload,
      };
    }
    case SET_ATTR_QUERY_TYPE: {
      return {
        ...state,
        attr_query_type: action.payload,
      };
    }
    case SET_TACTIC_OFFER_TYPE: {
      return {
        ...state,
        tacticOfferType: action.payload,
      };
    }
    case SET_ATTR_DATE_RANGE: {
      return {
        ...state,
        attr_dateRange: action.payload,
      };
    }
    case SET_ATTRIBUTION_MODEL: {
      if (action.payload.length > 1) {
        return {
          ...state,
          models: action.payload,
          linkedEvents: [], // clear linked events if comparison model is added
        };
      }
      return {
        ...state,
        models: action.payload,
      };
    }
    case SET_ATTRIBUTION_WINDOW: {
      return {
        ...state,
        window: action.payload,
      };
    }
    case SET_ATTR_LINK_EVENTS: {
      return {
        ...state,
        linkedEvents: action.payload,
      };
    }
    case INITIALIZE_MTA_STATE: {
      return {
        ...state,
        ...action.payload,
      };
    }
    case INITIALIZE_CAMPAIGN_STATE: {
      return {
        ...state,
        ...action.payload,
      };
    }
    case SET_CAMP_CHANNEL: {
      return {
        ...state,
        camp_channels: action.payload,
      };
    }
    case SET_CAMP_MEASURES: {
      return {
        ...state,
        camp_measures: action.payload,
      };
    }
    case FETCH_CAMP_CONFIG: {
      return {
        ...state,
        campaign_config: action.payload,
      };
    }
    case SET_CAMP_FILTERS: {
      return {
        ...state,
        camp_filters: action.payload,
      };
    }
    case SET_CAMP_GROUBY: {
      return {
        ...state,
        camp_groupBy: action.payload,
      };
    }
    case SET_CAMP_DATE_RANGE: {
      return {
        ...state,
        camp_dateRange: action.payload,
      };
    }
    case SET_DEFAULT_STATE: {
      return {
        ...defaultState,
      };
    }
    default:
      return state;
  }
}
