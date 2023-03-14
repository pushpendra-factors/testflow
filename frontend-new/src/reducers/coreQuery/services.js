import { notification } from 'antd';
import { get, getHostUrl, post, del, put } from '../../utils/request';
import {
  QUERIES_LOADING,
  QUERIES_LOADED,
  QUERIES_LOADING_FAILED,
  QUERY_DELETED,
  QUERIES_LOADING_STOPPED,
  INITIALIZE_TOUCHPOINT_DIMENSIONS,
  INITIALIZE_CONTENT_GROUPS,
  EVENT_DISPLAY_NAMES_LOADING,
  EVENT_DISPLAY_NAMES_ERROR,
  EVENT_DISPLAY_NAMES_LOADED,
  FETCH_GROUPS_FULFILLED,
  FETCH_GROUPS_REJECTED
} from '../types';
import { getErrorMessage } from '../../utils/dataFormatter';
// import { SAVED_QUERIES } from '../../utils/SampleResponse';
import {
  MARKETING_TOUCHPOINTS_ALIAS,
  KEY_TOUCH_POINT_DIMENSIONS,
  KEY_CONTENT_GROUPS
} from '../../utils/constants';

const host = getHostUrl();

export const getEventNames = (dispatch, projectId) =>
  get(
    dispatch,
    `${host}projects/${projectId}/user/event_names?is_display_name_enabled=true`,
    {}
  );

export const getEventsData = (
  projectId,
  query_group,
  dashboard,
  isQuery = false,
  query_id = null
) => {
  let url;
  if (!dashboard) {
    url = `${host}projects/${projectId}/v1/query${
      query_id ? `?&query_id=${query_id}` : ''
    }`;
  } else {
    url = `${host}projects/${projectId}/v1/query?refresh=${
      dashboard.refresh
    }&dashboard_id=${dashboard.id}&dashboard_unit_id=${
      dashboard.unit_id
    }&is_query=${isQuery}${query_id ? `&query_id=${query_id}` : ''}`;
  }
  return post(null, url, query_group && { query_group });
};

export function fetchEventProperties(projectId, eventName) {
  const url = `${host}projects/${projectId}/event_names/${btoa(
    btoa(eventName)
  )}/properties?is_display_name_enabled=true`;
  return get(null, url);
}

export function fetchEventPropertyValues(projectId, eventName, propertyName) {
  const url = `${host}projects/${projectId}/event_names/${btoa(
    btoa(eventName)
  )}/properties/${propertyName}/values`;
  return get(null, url);
}

export const fetchChannelObjPropertyValues = (
  projectId,
  channel = 'all_channels',
  filterObj,
  property
) => {
  const url = `${host}projects/${projectId}/v1/channels/filter_values?channel=${channel}&filter_object=${filterObj}&filter_property=${property}`;
  // const url =
  //   filterObj === "campaign"
  //     ? `http://localhost:8000/getChannelFilters`
  //     : `http://localhost:8000/adGroupFilters`;
  return get(null, url);
};

export function fetchUserPropertyValues(projectId, propertyName) {
  const url = `${host}projects/${projectId}/user_properties/${propertyName}/values`;
  return get(null, url);
}

export function fetchUserProperties(projectId, queryType) {
  const url = `${host}projects/${projectId}/user_properties?is_display_name_enabled=true`;
  return get(null, url);
}

export const getFunnelData = (
  projectId,
  query,
  dashboard,
  isQuery = false,
  query_id = null
) => {
  let url;
  if (!dashboard) {
    url = `${host}projects/${projectId}/query${
      query_id ? `?&query_id=${query_id}` : ''
    }`;
  } else {
    url = `${host}projects/${projectId}/query?refresh=${
      dashboard.refresh
    }&dashboard_id=${dashboard.id}&dashboard_unit_id=${
      dashboard.unit_id
    }&is_query=${isQuery}${query_id ? `&query_id=${query_id}` : ''}`;
  }
  return post(null, url, { query });
};

export const getProfileData = (
  projectId,
  query,
  dashboard,
  isQuery = false
) => {
  let url;
  if (!dashboard) {
    url = `${host}projects/${projectId}/profiles/query`;
  } else {
    url = `${host}projects/${projectId}/profiles/query?refresh=${dashboard.refresh}&dashboard_id=${dashboard.id}&dashboard_unit_id=${dashboard.unit_id}&is_query=${isQuery}`;
  }
  return post(null, url, query);
};

export const getKPIData = (projectId, query, dashboard, isQuery = false) => {
  let url;
  if (!dashboard) {
    url = `${host}projects/${projectId}/v1/kpi/query`;
  } else {
    url = `${host}projects/${projectId}/v1/kpi/query?refresh=${dashboard.refresh}&dashboard_id=${dashboard.id}&dashboard_unit_id=${dashboard.unit_id}&is_query=${isQuery}`;
  }
  return post(null, url, query);
};

export const saveQuery = (projectId, title, query, type, settings) => {
  const url = `${host}projects/${projectId}/queries`;
  return post(null, url, { query, title, type, settings });
};

// export const deleteQueryTest = async () => {
//   const promises = SAVED_QUERIES.filter(
//     (query) => [].indexOf(query.id) === -1
//   ).map((query) => {
//     const url = host + 'projects/' + query.project_id + '/queries/' + query.id;
//     return del(null, url);
//   });
//   try {
//     const response = await Promise.all(promises);
//   } catch (err) {
//     console.log(err);
//   }
// };

export const deleteQuery = ({ project_id, id }) =>
  async function (dispatch) {
    try {
      dispatch({ type: QUERIES_LOADING });
      await deleteReport({ project_id, queryId: id });
      dispatch({ type: QUERY_DELETED, payload: id });
    } catch (err) {
      console.log(err);
      dispatch({ type: QUERIES_LOADING_STOPPED });
      notification.error({
        message: 'Something went wrong!',
        description: getErrorMessage(err),
        duration: 5
      });
    }
  };

export const fetchQueries = (projectId) => async (dispatch) => {
  try {
    dispatch({ type: QUERIES_LOADING });
    const url = `${host}projects/${projectId}/queries`;
    const res = await get(null, url);
    dispatch({ type: QUERIES_LOADED, payload: res.data });
  } catch (err) {
    console.log(err);
    dispatch({ type: QUERIES_LOADING_FAILED });
  }
};

export const fetchGroups =
  (projectId, isAccount = '') =>
  async (dispatch) => {
    try {
      const url = `${host}projects/${projectId}/groups?is_account=${isAccount}`;
      const res = await get(null, url);
      dispatch({ type: FETCH_GROUPS_FULFILLED, payload: res.data });
    } catch (err) {
      console.log(err);
      dispatch({ type: FETCH_GROUPS_REJECTED });
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

export const getAttributionsDataV1 = (
  projectId,
  reqBody,
  dashboard,
  isQuery = false
) => {
  let url;
  if (!dashboard) {
    url = `${host}projects/${projectId}/v1/attribution/query`;
  } else {
    url = `${host}projects/${projectId}/v1/attribution/query?refresh=${dashboard.refresh}&dashboard_id=${dashboard.id}&dashboard_unit_id=${dashboard.unit_id}&is_query=${isQuery}`;
  }
  return post(null, url, reqBody);
};

export const fetchCampaignConfig = (projectId, channel) => {
  const url = `${host}projects/${projectId}/v1/channels/config?channel=${channel}`;
  return get(null, url);
};

export const getCampaignsData = (
  projectId,
  reqBody,
  dashboard,
  isQuery = false
) => {
  let url;
  if (!dashboard) {
    url = `${host}projects/${projectId}/v1/channels/query`;
  } else {
    url = `${host}projects/${projectId}/v1/channels/query?refresh=${dashboard.refresh}&dashboard_id=${dashboard.id}&dashboard_unit_id=${dashboard.unit_id}&is_query=${isQuery}`;
  }
  return post(null, url, reqBody);
};

export const getWebAnalyticsData = (
  projectId,
  reqBody,
  dashboardId,
  refresh,
  isQuery = false
) => {
  const url =
    `${host}projects/${projectId}/dashboard/${dashboardId}/units/query/web_analytics?refresh=${refresh}` +
    `&is_query=${isQuery}`;
  return post(null, url, reqBody);
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
      dispatch({
        type: INITIALIZE_TOUCHPOINT_DIMENSIONS,
        payload: [...KEY_TOUCH_POINT_DIMENSIONS, ...customDimensions]
      });
    } catch (err) {
      console.log(err);
    }
  };

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
      dispatch({
        type: INITIALIZE_CONTENT_GROUPS,
        payload: [...KEY_CONTENT_GROUPS, ...content_group]
      });
    } catch (err) {
      console.log(err);
    }
  };

export const updateQuery = (projectId, savedQueryId, reqBody) => {
  const url = `${host}projects/${projectId}/queries/${savedQueryId}`;
  return put(null, url, reqBody);
};

export const deleteReport = ({ project_id, queryId }) => {
  const url = `${host}projects/${project_id}/queries/${queryId}`;
  return del(null, url);
};

export function fetchGroupProperties(projectId, groupName) {
  const url = `${host}projects/${projectId}/groups/${btoa(
    btoa(groupName)
  )}/properties`;
  return get(null, url);
}

export function fetchGroupPropertyValues(projectId, groupName, propertyName) {
  const url = `${host}projects/${projectId}/groups/${btoa(
    btoa(groupName)
  )}/properties/${propertyName}/values`;
  return get(null, url);
}

export function fetchEventDisplayNames({ projectId }) {
  return async function (dispatch) {
    try {
      dispatch({ type: EVENT_DISPLAY_NAMES_LOADING });
      const url = `${host}projects/${projectId}/v1/events/displayname`;
      const response = await get(null, url);
      dispatch({ type: EVENT_DISPLAY_NAMES_LOADED, payload: response.data });
    } catch (err) {
      dispatch({ type: EVENT_DISPLAY_NAMES_ERROR });
    }
  };
}
