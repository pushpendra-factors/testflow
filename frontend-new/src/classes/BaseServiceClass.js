import { get, post, put, del, getHostUrl } from '../utils/request';

const host = getHostUrl();

export class BaseService {
  baseRoute;
  dispatch;
  projectId;

  constructor(dispatch, projectId) {
    this.baseRoute = getHostUrl() + 'projects/' + projectId;
    this.dispatch = dispatch;
    this.projectId = projectId;
  }

  get = (route) => get(this.dispatch, this.baseRoute + route, {});
  post = (route, data, headers) =>
    post(this.dispatch, this.baseRoute + route, data, headers);
  put = (route, data, headers = {}) =>
    put(this.dispatch, this.baseRoute + route, data, headers);
  del = (route, data, headers = {}) =>
    del(this.dispatch, this.baseRoute + route, data, headers);
}
