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
  SET_EVENT_NAMES
} from "./actions";
import {
  SHOW_ANALYTICS_RESULT,
  INITIALIZE_MTA_STATE,
  INITIALIZE_CAMPAIGN_STATE,
} from "../types";
import { DefaultDateRangeFormat } from "../../Views/CoreQuery/utils";

const DEFAULT_TOUCHPOINTS = ["Campaign", "Source", "AdGroup", "Keyword"];

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
      label: "Paid Marketing",
      icon: "fav",
      values: DEFAULT_TOUCHPOINTS.map((v) => [v]),
    },
  ],
  show_analytics_result: false,
  eventGoal: {},
  touchpoint: "",
  touchpoint_filters: [],
  attr_query_type: 'EngagementBased',
  attr_dateRange: {
    ...DefaultDateRangeFormat,
    dateStr: ""
  },
  models: [],
  window: null,
  linkedEvents: [],
  campaign_config: {
    metrics: [],
    properties: [],
  },
  camp_channels: "google_ads",
  camp_measures: [],
  camp_filters: [],
  camp_groupBy: [],
  camp_dateRange: {
    ...DefaultDateRangeFormat,
    dateStr: "",
  },
  eventNames: []
};

export default function (state = defaultState, action) {
  switch (action.type) {
    case FETCH_EVENTS:
      return { ...state, eventOptions: action.payload };
    case SET_EVENT_NAMES:
      return { ...state, eventNames: action.payload}
    case FETCH_USER_PROPERTIES:
      return { ...state, userProperties: action.payload };
    case FETCH_EVENT_PROPERTIES:
      const eventPropState = Object.assign({}, state.eventProperties);
      eventPropState[action.eventName] = action.payload;
      return { ...state, eventProperties: eventPropState };
    case INITIALIZE_GROUPBY: {
      return {
        ...state,
        groupBy: action.payload,
      };
    }
    case DEL_GROUPBY: {
      let groupByState = Object.assign({}, state.groupBy);
      let gbp;
      if (
        groupByState[action.groupByType] &&
        groupByState[action.groupByType][action.index]
      ) {
        let groupTypeState = [...groupByState[action.groupByType]];
        if (groupTypeState[action.index] === action.payload) {
          groupTypeState.splice(action.index, 1);
          // groupTypeState.length -= 1;
        } else {
          gbp = groupTypeState.findIndex((i) => i === state.payload);
          gbp && groupTypeState.splice(gbp, 1);
        }
        groupByState[action.groupByType] = groupTypeState;
      }
      return { ...state, groupBy: groupByState };
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
        groupByState[action.groupByType].push(action.payload);
      }
      groupByState[action.groupByType].sort((a, b) => {
        return a.prop_category >= b.prop_category ? 1 : -1;
      });
      return { ...state, groupBy: groupByState };
    case DEL_GROUPBY_EVENT: {
      let groupByState = Object.assign({}, state.groupBy);
      let eventGroups = groupByState.event;
      const filteredEventGroups = eventGroups.filter((gbp) => {
        return (
          gbp.eventIndex !== action.index + 1 &&
          gbp.eventName !== action.payload.label
        );
      });
      groupByState.event = filteredEventGroups;
      return { ...state, groupBy: groupByState };
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
    case SET_ATTR_DATE_RANGE: {
      return {
        ...state,
        attr_dateRange: action.payload
      }
    }
    case SET_ATTRIBUTION_MODEL: {
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
        camp_dateRange: action.payload
      }
    }
    case SET_DEFAULT_STATE: {
      return {
        ...defaultState
      }
    }
    default:
      return state;
  }
}
