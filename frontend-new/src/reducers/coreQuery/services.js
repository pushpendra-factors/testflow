import { get, getHostUrl, post, del } from '../../utils/request';
import {
  QUERIES_LOADING,
  QUERIES_LOADED,
  QUERIES_LOADING_FAILED,
  QUERY_DELETED,
  QUERIES_LOADING_STOPPED,
  INITIALIZE_TOUCHPOINT_DIMENSIONS,
} from '../types';
import { notification } from 'antd';
import { getErrorMessage } from '../../utils/dataFormatter';
import {
  MARKETING_TOUCHPOINTS_ALIAS,
  KEY_TOUCH_POINT_DIMENSIONS,
} from '../../utils/constants';
const host = getHostUrl();

export const getEventNames = (dispatch, projectId) => {
  return get(
    dispatch,
    host +
      'projects/' +
      projectId +
      '/v1/event_names?is_display_name_enabled=true',
    {}
  );
};

export const getEventsData = (projectId, query_group, dashboard) => {
  let url;
  if (!dashboard) {
    url = host + 'projects/' + projectId + '/v1/query';
  } else {
    url =
      host +
      'projects/' +
      projectId +
      `/v1/query?refresh=${dashboard.refresh}&dashboard_id=` +
      dashboard.id +
      '&dashboard_unit_id=' +
      dashboard.unit_id;
  }
  return post(null, url, { query_group });
};

export function fetchEventProperties(projectId, eventName) {
  const url =
    host +
    'projects/' +
    projectId +
    '/event_names/' +
    btoa(eventName) +
    '/properties?is_display_name_enabled=true';
  return get(null, url);
}

export function fetchEventPropertyValues(projectId, eventName, propertyName) {
  const url =
    host +
    'projects/' +
    projectId +
    '/event_names/' +
    btoa(eventName) +
    '/properties/' +
    propertyName +
    '/values';
  return get(null, url);
}

export const fetchChannelObjPropertyValues = (
  projectId,
  channel = 'all_channels',
  filterObj,
  property
) => {
  const url =
    host +
    'projects/' +
    projectId +
    '/v1/channels/filter_values?channel=' +
    channel +
    '&filter_object=' +
    filterObj +
    '&filter_property=' +
    property;
  // const url =
  //   filterObj === "campaign"
  //     ? `http://localhost:8000/getChannelFilters`
  //     : `http://localhost:8000/adGroupFilters`;
  return get(null, url);
};

export function fetchUserPropertyValues(projectId, propertyName) {
  const url =
    host +
    'projects/' +
    projectId +
    '/user_properties/' +
    propertyName +
    '/values';
  return get(null, url);
}

export function fetchUserProperties(projectId, queryType) {
  const url =
    host +
    'projects/' +
    projectId +
    '/user_properties?is_display_name_enabled=true';
  return get(null, url);
}

export const getFunnelData = (projectId, query, dashboard) => {
  let url;
  if (!dashboard) {
    url = host + 'projects/' + projectId + '/query';
  } else {
    url =
      host +
      'projects/' +
      projectId +
      `/query?refresh=${dashboard.refresh}&dashboard_id=` +
      dashboard.id +
      '&dashboard_unit_id=' +
      dashboard.unit_id;
  }
  return post(null, url, { query });
};

export const saveQuery = (projectId, title, query, type) => {
  const url = host + 'projects/' + projectId + '/queries';
  return post(null, url, { query, title, type });
};

export const deleteQuery = async (dispatch, query) => {
  try {
    dispatch({ type: QUERIES_LOADING });
    const url = host + 'projects/' + query.project_id + '/queries/' + query.id;
    await del(null, url);
    dispatch({ type: QUERY_DELETED, payload: query.id });
  } catch (err) {
    console.log(err);
    dispatch({ type: QUERIES_LOADING_STOPPED });
    notification.error({
      message: 'Something went wrong!',
      description: getErrorMessage(err),
      duration: 5,
    });
  }
};

export const fetchQueries = async (dispatch, projectId) => {
  try {
    dispatch({ type: QUERIES_LOADING });
    const url = host + 'projects/' + projectId + '/queries';
    const res = await get(null, url);
    dispatch({ type: QUERIES_LOADED, payload: res.data });
  } catch (err) {
    console.log(err);
    dispatch({ type: QUERIES_LOADING_FAILED });
  }
};

export const getAttributionsData = (projectId, reqBody, dashboard) => {
  let url;
  if (!dashboard) {
    url = host + 'projects/' + projectId + '/attribution/query';
  } else {
    url =
      host +
      'projects/' +
      projectId +
      `/attribution/query?refresh=${dashboard.refresh}&dashboard_id=` +
      dashboard.id +
      '&dashboard_unit_id=' +
      dashboard.unit_id;
  }
  return post(null, url, reqBody);
};

export const fetchCampaignConfig = (projectId, channel) => {
  const url =
    host + 'projects/' + projectId + '/v1/channels/config?channel=' + channel;
  return get(null, url);
};

export const getCampaignsData = (projectId, reqBody, dashboard) => {
  let url;
  if (!dashboard) {
    url = host + 'projects/' + projectId + '/v1/channels/query';
  } else {
    url =
      host +
      'projects/' +
      projectId +
      `/v1/channels/query?refresh=${dashboard.refresh}&dashboard_id=` +
      dashboard.id +
      '&dashboard_unit_id=' +
      dashboard.unit_id;
  }
  return post(null, url, reqBody);
};

export const getWebAnalyticsData = (
  projectId,
  reqBody,
  dashboardId,
  refresh
) => {
  const url = `${host}projects/${projectId}/dashboard/${dashboardId}/units/query/web_analytics?refresh=${refresh}`;
  return post(null, url, reqBody);
};

export const fetchSmartPropertyRules = async (dispatch, projectId) => {
  try {
    const url = host + 'projects/' + projectId + '/v1/smart_properties/rules';
    const res = await get(null, url);
    const customDimensions = res.data.map((elem) => {
      return {
        title: elem.name,
        header: elem.name,
        responseHeader: elem.name,
        enabled: false,
        type: 'custom',
        touchPoint: MARKETING_TOUCHPOINTS_ALIAS[elem.type_alias],
        defaultValue: false,
      };
    });
    dispatch({
      type: INITIALIZE_TOUCHPOINT_DIMENSIONS,
      payload: [...KEY_TOUCH_POINT_DIMENSIONS, ...customDimensions],
    });
  } catch (err) {
    console.log(err);
  }
};
