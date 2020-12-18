/* eslint-disable */

import { get, getHostUrl, post, del } from "../../utils/request";
import {
  QUERIES_LOADING,
  QUERIES_LOADED,
  QUERIES_LOADING_FAILED,
  QUERY_DELETED,
} from "../types";
const host = getHostUrl();

export const getEventNames = (dispatch, projectId) => {
  return get(dispatch, host + "projects/" + projectId + "/v1/event_names", {});
};

export const runQuery = (
  projectId,
  query_group,
  dashboard = { refresh: true }
) => {
  let url;
  if (dashboard.refresh) {
    url = host + "projects/" + projectId + "/v1/query";
  } else {
    url =
      host +
      "projects/" +
      projectId +
      "/v1/query?refresh=false&dashboard_id=" +
      dashboard.id +
      "&dashboard_unit_id=" +
      dashboard.unit_id;
  }
  return post(null, url, { query_group });
};

export const runAttrQuery = (projectId) => {
  const url = host + "projects/" + projectId + "/attribution/query";
  return post(null, url, sampleReq);
};

export function fetchEventProperties(projectId, eventName) {
  const url =
    host +
    "projects/" +
    projectId +
    "/event_names/" +
    btoa(eventName) +
    "/properties";
  return get(null, url);
}

export function fetchEventPropertyValues(projectId, eventName, propertyName) {
  const url =
    host +
    "projects/" +
    projectId +
    "/event_names/" +
    btoa(eventName) +
    "/properties/" +
    propertyName +
    "/values";
  return get(null, url);
}

export function fetchUserPropertyValues(projectId, propertyName) {
  const url =
    host +
    "projects/" +
    projectId +
    "/user_properties/" +
    propertyName +
    "/values";
  return get(null, url);
}

export function fetchUserProperties(projectId, queryType) {
  const url =
    host + "projects/" + projectId + "/user_properties?query_type=" + queryType;
  return get(null, url);
}

export const getFunnelData = (
  projectId,
  query,
  dashboard = { refresh: true }
) => {
  let url;
  if (dashboard.refresh) {
    url = host + "projects/" + projectId + "/query";
  } else {
    url =
      host +
      "projects/" +
      projectId +
      "/query?refresh=false&dashboard_id=" +
      dashboard.id +
      "&dashboard_unit_id=" +
      dashboard.unit_id;
  }
  return post(null, url, { query });
};

export const saveQuery = (projectId, title, query, type) => {
  const url = host + "projects/" + projectId + "/queries";
  return post(null, url, { query, title, type });
};

export const deleteQuery = async (dispatch, query) => {
  try {
    dispatch({ type: QUERIES_LOADING });
    const url = host + "projects/" + query.project_id + "/queries/" + query.id;
    await del(null, url);
    dispatch({ type: QUERY_DELETED, payload: query.id });
  } catch (err) {
    console.log(err);
  }
};

export const fetchQueries = async (dispatch, projectId) => {
  try {
    dispatch({ type: QUERIES_LOADING });
    const url = host + "projects/" + projectId + "/queries";
    const res = await get(null, url);
    dispatch({ type: QUERIES_LOADED, payload: res.data });
  } catch (err) {
    console.log(err);
    dispatch({ type: QUERIES_LOADING_FAILED });
  }
};

export const getAttributionsData = (
  projectId,
  query,
  dashboard = { refresh: true }
) => {
  let url;
  if (dashboard.refresh) {
    url = host + "projects/" + projectId + "/attribution/query";
  } else {
    url =
      host +
      "projects/" +
      projectId +
      "/attribution/query?refresh=false&dashboard_id=" +
      dashboard.id +
      "&dashboard_unit_id=" +
      dashboard.unit_id;
  }
  return post(null, `http://localhost:8000/query`, { query });
};
