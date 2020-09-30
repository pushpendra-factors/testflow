/* eslint-disable */

import { get, getHostUrl, post } from '../../utils/request';

const host = getHostUrl();

export const getEventNames = (projectId) => {
  return get(host + 'projects/' + projectId + '/event_names?type=exact', {});
}

export const runQuery = (projectId, query) => {
  const url = host + "projects/" + projectId + "/query";
  return post(url , {query: query});
}
