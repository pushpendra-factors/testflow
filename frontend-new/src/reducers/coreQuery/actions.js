/* eslint-disable */

export const FETCH_EVENTS = 'FETCH_EVENTS';
export const FETCH_EVENT_PROPERTIES = 'FETCH_EVENT_PROPERTIES';
export const FETCH_USER_PROPERTIES = 'FETCH_USER_PROPERTIES';
export const SET_USER_PROP_NAME = 'SET_USER_PROP_NAME';
export const SET_EVENT_PROP_NAME = 'SET_EVENT_PROP_NAME';
export const INITIALIZE_GROUPBY = 'INITIALIZE_GROUPBY';
export const SET_GROUPBY = 'SET_GROUPBY';
export const RESET_GROUPBY = 'RESET_GROUPBY';
export const DEL_GROUPBY = 'DEL_GROUPBY';
export const DEL_GROUPBY_EVENT = 'DEL_GROUPBY_EVENT';
export const SET_EVENT_GOAL = 'SET_EVENT_GOAL';
export const SET_TOUCHPOINTS = 'SET_TOUCHPOINTS';
export const SET_TOUCHPOINT_FILTERS = 'SET_TOUCHPOINT_FILTERS';
export const SET_ATTR_QUERY_TYPE = 'SET_ATTR_QUERY_TYPE';
export const SET_TACTIC_OFFER_TYPE = 'SET_TACTIC_OFFER_TYPE';
export const SET_ATTRIBUTION_MODEL = 'SET_ATTRIBUTION_MODEL';
export const SET_ATTRIBUTION_WINDOW = 'SET_ATTRIBUTION_WINDOW';
export const SET_ATTR_LINK_EVENTS = 'SET_ATTR_LINK_EVENTS';
export const SET_ATTR_DATE_RANGE = 'SET_ATTR_DATE_RANGE';
export const FETCH_CAMP_CONFIG = 'FETCH_CAMP_CONFIG';
export const SET_CAMP_CHANNEL = 'SET_CAMP_CHANNEL';
export const SET_CAMP_MEASURES = 'SET_CAMP_MEASURES';
export const SET_CAMP_FILTERS = 'SET_CAMP_FILTERS';
export const SET_CAMP_GROUBY = 'SET_CAMP_GROUBY';
export const SET_CAMP_DATE_RANGE = 'SET_CAMP_DATE_RANGE';
export const SET_DEFAULT_STATE = 'SET_DEFAULT_STATE';
export const SET_EVENT_NAMES = 'SET_EVENT_NAMES';
export const SET_ATTR_QUERIES = 'SET_ATTR_QUERIES';

// Action creators
export const fetchEventsAction = (events, status = 'started') => {
  return { type: FETCH_EVENTS, payload: events };
};

export const setEventsDisplayAction = (displayNames, status = 'started') => {
  return { type: SET_EVENT_NAMES, payload: displayNames };
};

export const fetchUserPropertiesAction = (userProps) => {
  return { type: FETCH_USER_PROPERTIES, payload: userProps};
}

export const setUserPropertiesNamesAction = (userPropsDisplayNames) => {
  return { type: SET_USER_PROP_NAME, payload: userPropsDisplayNames};
}

export const fetchEventPropertiesAction = (eventProps, name) => {
  return { type: FETCH_EVENT_PROPERTIES, payload: eventProps, eventName: name };
}

export const setEventPropertiesNamesAction = (eventPropDisplayNames) => {
  return { type: SET_EVENT_PROP_NAME, payload: eventPropDisplayNames};
}

export const delGroupByAction = (type, payload, index) => {
  return { type: DEL_GROUPBY, payload: payload, index: index, groupByType: type };
}

export const deleteGroupByEventAction = (ev, index) => {
  return { type: DEL_GROUPBY_EVENT, payload: ev, index: index};
}

export const setGroupByAction = (type, groupBy, index) => {
  return { type: SET_GROUPBY, payload: groupBy, index: index, groupByType: type };
}
export const resetGroupByAction = () => {
  return { type: RESET_GROUPBY};
}

export const setEventGoalAction = (goal) => {
  return { type: SET_EVENT_GOAL, payload: goal};
}

export const setMarketingTouchpointsAction = (touchpoints) => {
  return { type: SET_TOUCHPOINTS, payload: touchpoints};
}

export const setTouchPointFiltersAction = (touchpointFilters) => {
  return { type: SET_TOUCHPOINT_FILTERS, payload: touchpointFilters};
}

export const setAttributionQueryTypeAction = (attrQueryType) => {
  return { type: SET_ATTR_QUERY_TYPE, payload: attrQueryType};
}

export const setTacticOfferTypeAction = (tacticOfferType) => {
  return { type: SET_TACTIC_OFFER_TYPE, payload: tacticOfferType};
}

export const setAttributionModelsAction = (models) => {
  return { type: SET_ATTRIBUTION_MODEL, payload: models};
}

export const setAttributionWindowAction = (window) => {
  return { type: SET_ATTRIBUTION_WINDOW, payload: window};
}

export const setAttrLinkEventsAction = (linkedEvents) => {
  return { type: SET_ATTR_LINK_EVENTS, payload: linkedEvents};
}

export const setAttrDateRangeAction = (dateRange) => {
  return { type: SET_ATTR_DATE_RANGE, payload: dateRange};
}

export const getCampaignConfigAction = (config) => {
  return { type: FETCH_CAMP_CONFIG, payload: config};
}

export const setCampChannelAction = (channel) => {
  return { type: SET_CAMP_CHANNEL, payload: channel};
}

export const setMeasuresAction = (measures) => {
  return { type: SET_CAMP_MEASURES, payload: measures};
}

export const setCampFiltersAction = (filters) => {
  return { type: SET_CAMP_FILTERS, payload: filters};
}

export const setCampGroupByAction = (groupBy) => {
  return { type: SET_CAMP_GROUBY, payload: groupBy};
}

export const setCampDateRangeAction = (dateRange) => {
  return { type: SET_CAMP_DATE_RANGE, payload: dateRange};
}

export const setDefaultStateAction = () => {
  return { type: SET_DEFAULT_STATE};
}


