/* eslint-disable */

export function isRequestSuccess(status) {
  return status >= 200 && status <= 399;
} 

export const getHostUrl = () => {
    let host = BUILD_CONFIG.backend_host;
    host = (host[host.length - 1] === '/') ? host : host + '/';
    return host;
}

function request(dispatch, method, url, headers, data) {
    const options = {
      method,
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json'
      }
    };
  
    if (data && data !== undefined) {
      options.body = JSON.stringify(data);
    }
  
    if (headers && headers !== undefined) {
      options.headers = headers;
      options.headers['Content-Type'] = 'application/json';
    }
  
    if (window.FACTORS_AI_LOGIN_TOKEN && window.FACTORS_AI_LOGIN_TOKEN !== '') {
      options.headers.Authorization = window.FACTORS_AI_LOGIN_TOKEN;
    }
  
    return fetch(url, options)
      .then((response) => {
 
        // CORS 401 for certain apis, ex:fetching projects
        if (response.status === 401){
            if (dispatch && dispatch != undefined) dispatch({type: "AGENT_LOGIN_FORCE"});
            return Promise.reject(response);
        }
        // validates response string before JSON unmarshal,
        // for handling no JSON response.
        return response.text()
          .then((text) => {
            const responsePayload = { status: response.status, ok: isRequestSuccess(response.status) };
            if (text === '') responsePayload.data = {};
            else responsePayload.data = JSON.parse(text);
  
            return responsePayload;
          });
      });
  }
  
  export function get(dispatch, url, headers = {}) { return request(dispatch,'GET', url, headers); }
  
  export function post(dispatch, url, data, headers = {}) { return request(dispatch,'POST', url, headers, data); }
  
  export function put(dispatch, url, data, headers = {}) { return request(dispatch,'PUT', url, headers, data); }
  
  export function del(dispatch, url, headers = {}) { return request(dispatch,'DELETE', url, headers); }
  