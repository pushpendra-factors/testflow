/* eslint-disable */

import { get, getHostUrl, post } from '../../utils/request';
const host = getHostUrl();

export const getEventNames = (dispatch, projectId) => {
  return get(dispatch, host + 'projects/' + projectId + '/v1/event_names', {});
}

export const runQuery = (projectId, query) => {
  const url = host + "projects/" + projectId + "/query";
  return post(null, url, { query: query });
}

export function fetchEventProperties(projectId, eventName) {
  const url = host + "projects/" + projectId + "/event_names/" + btoa(eventName) + "/properties";
  return get(null, url);
}

export function fetchUserProperties(projectId, queryType) {
  const url = host + "projects/" + projectId + "/user_properties?query_type="+queryType;
  return get(null, url);
}


