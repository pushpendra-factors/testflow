/* eslint-disable */

import { get } from '../../request';

export const getEventNames = (projectId) => {
  let host = BUILD_CONFIG.backend_host;
  host = (host[host.length - 1] === '/') ? host : host + '/';
  return get(host + 'projects/' + projectId + '/event_names?type=exact', {});
};
