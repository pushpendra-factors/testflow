import { get, post, getHostUrl } from '../../../utils/request';

import {
  KEY_CONTENT_GROUPS,
  MARKETING_TOUCHPOINTS_ALIAS,
  KEY_TOUCH_POINT_DIMENSIONS
} from 'Utils/constants';

import {
  setAttributionDashboardUnitsLoading,
  setAttributionDashboardUnitsLoaded,
  setAttributionDashboardUnitsFailed,
  initializeContentGroups,
  initializeTouchPointDimensions
} from './actions';
const host = getHostUrl();

export const fetchAttributionActiveUnits = (projectId, activeDashboardId) =>
  async function (dispatch) {
    try {
      dispatch(setAttributionDashboardUnitsLoading);
      const url = `${host}projects/${projectId}/dashboards/${activeDashboardId}/units`;
      const res = await get(null, url);
      dispatch(setAttributionDashboardUnitsLoaded(res?.data));
    } catch (err) {
      console.log(err);
      dispatch(setAttributionDashboardUnitsFailed());
    }
  };

export const getEventNames = (dispatch, projectId) =>
  get(
    dispatch,
    `${host}projects/${projectId}/user/event_names?is_display_name_enabled=true`,
    {}
  );

export const fetchAttrContentGroups = (projectId) =>
  async function (dispatch) {
    try {
      const url = `${host}projects/${projectId}/v1/contentgroup`;
      const res = await get(null, url);
      const content_group = res.data.map((elem) => ({
        title: elem.content_group_name,
        header: elem.content_group_name,
        responseHeader: elem.content_group_name,
        enabled: false,
        type: 'content_group',
        touchPoint: 'LandingPage',
        defaultValue: false
      }));
      dispatch(
        initializeContentGroups([...KEY_CONTENT_GROUPS, ...content_group])
      );
    } catch (err) {
      console.log(err);
    }
  };

export const fetchSmartPropertyRules = (projectId) =>
  async function (dispatch) {
    try {
      const url = `${host}projects/${projectId}/v1/smart_properties/rules`;
      const res = await get(null, url);
      const customDimensions = res.data.map((elem) => ({
        title: elem.name,
        header: elem.name,
        responseHeader: elem.name,
        enabled: false,
        type: 'custom',
        touchPoint: MARKETING_TOUCHPOINTS_ALIAS[elem.type_alias],
        defaultValue: false
      }));

      dispatch(
        initializeTouchPointDimensions([
          ...KEY_TOUCH_POINT_DIMENSIONS,
          ...customDimensions
        ])
      );
    } catch (err) {
      console.log(err);
    }
  };

export const getAttributionsData = (
  projectId,
  reqBody,
  dashboard,
  isQuery = false
) => {
  let url;
  if (!dashboard) {
    url = `${host}projects/${projectId}/attribution/query`;
  } else {
    url = `${host}projects/${projectId}/attribution/query?refresh=${dashboard.refresh}&dashboard_id=${dashboard.id}&dashboard_unit_id=${dashboard.unit_id}&is_query=${isQuery}`;
  }
  return post(null, url, reqBody);
};
