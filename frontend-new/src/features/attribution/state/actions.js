import {
  ATTRIBUTION_DASHBOARD_UNITS_FAILED,
  ATTRIBUTION_DASHBOARD_UNITS_LOADED,
  ATTRIBUTION_DASHBOARD_UNITS_LOADING,
  SET_EVENT_GOAL,
  SET_TOUCHPOINTS,
  SET_ATTRIBUTION_MODEL,
  SET_ATTRIBUTION_WINDOW,
  SET_ATTR_LINK_EVENTS,
  SET_ATTR_DATE_RANGE,
  SET_ATTR_QUERY_TYPE,
  SET_TOUCHPOINT_FILTERS,
  SET_TACTIC_OFFER_TYPE,
  INITIALIZE_CONTENT_GROUPS,
  INITIALIZE_TOUCHPOINT_DIMENSIONS,
  INITIALIZE_ATTRIBUTION_STATE
} from './action.constants';

export const setAttributionDashboardUnitsLoading = () => ({
  type: ATTRIBUTION_DASHBOARD_UNITS_LOADING
});

export const setAttributionDashboardUnitsLoaded = (payload) => ({
  type: ATTRIBUTION_DASHBOARD_UNITS_LOADED,
  payload
});

export const setAttributionDashboardUnitsFailed = () => ({
  type: ATTRIBUTION_DASHBOARD_UNITS_FAILED
});

export const setGoalEvent = (goal) => ({
  type: SET_EVENT_GOAL,
  payload: goal
});

export const setTouchPoint = (touchpoints) => ({
  type: SET_TOUCHPOINTS,
  payload: touchpoints
});

export const setModels = (models) => ({
  type: SET_ATTRIBUTION_MODEL,
  payload: models
});

export const setWindow = (window) => ({
  type: SET_ATTRIBUTION_WINDOW,
  payload: window
});

export const setLinkedEvents = (linkedEvents) => ({
  type: SET_ATTR_LINK_EVENTS,
  payload: linkedEvents
});

export const setAttrDateRange = (dateRange) => ({
  type: SET_ATTR_DATE_RANGE,
  payload: dateRange
});

export const setattrQueryType = (attrQueryType) => ({
  type: SET_ATTR_QUERY_TYPE,
  payload: attrQueryType
});

export const setTouchPointFilters = (touchpointFilters) => ({
  type: SET_TOUCHPOINT_FILTERS,
  payload: touchpointFilters
});

export const setTacticOfferType = (tacticOfferType) => ({
  type: SET_TACTIC_OFFER_TYPE,
  payload: tacticOfferType
});

export const initializeContentGroups = (payload) => ({
  type: INITIALIZE_CONTENT_GROUPS,
  payload
});

export const initializeTouchPointDimensions = (payload) => ({
  type: INITIALIZE_TOUCHPOINT_DIMENSIONS,
  payload
});

export const initializeAttributionState = (payload) => ({
  type: INITIALIZE_ATTRIBUTION_STATE,
  payload
});
