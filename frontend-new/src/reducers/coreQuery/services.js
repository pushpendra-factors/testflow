/* eslint-disable */

import { get, getHostUrl, post } from '../../utils/request';
const host = getHostUrl();

export const getEventNames = (dispatch, projectId) => {
  return get(dispatch, host + 'projects/' + projectId + '/v1/event_names', {});
}

export const runQuery = (projectId, query_group) => {
  const url = host + "projects/" + projectId + "/v1/query";
  return post(null, url, { query_group });
}

export function fetchEventProperties(projectId, eventName) {
  const url = host + "projects/" + projectId + "/event_names/" + btoa(eventName) + "/properties";
  return get(null, url);
}

export function fetchEventPropertyValues(projectId, eventName, propertyName) {
  const url = host + "projects/" + projectId + "/event_names/" + btoa(eventName)
    + "/properties/" + propertyName + "/values";
  return get(null, url);
}

export function fetchUserProperties(projectId, queryType) {
  const url = host + "projects/" + projectId + "/user_properties?query_type=" + queryType;
  return get(null, url);
}

export const getFinalData = (projectId, query) => {
  const url = host + "projects/" + projectId + "/query";
  return post(null, url, { query });
}