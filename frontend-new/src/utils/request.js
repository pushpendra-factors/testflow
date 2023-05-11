/* eslint-disable */

export function isRequestSuccess(status) {
  return status >= 200 && status <= 399;
} 

export const getHostUrl = () => {
    let host = BUILD_CONFIG.backend_host;
    host = (host[host.length - 1] === '/') ? host : host + '/';
    let isSlothApp = window.location && window.location.host && window.location.host.indexOf("sloth") == 0;
    // TODO: Remove isFlashApp after deployment of sloth on production.
    let isFlashApp = window.location && window.location.host && window.location.host.indexOf("flash") == 0;
    if (isSlothApp || isFlashApp) {
        host = host + "mql/"
    }
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

    if (window.INVALIDATE_CACHE && window.INVALIDATE_CACHE != "") {
      options.headers["Invalidate-Cache"] = true;
    }

    if (window.USE_FILTER_OPT_PROFILES && window.USE_FILTER_OPT_PROFILES != "") {
      options.headers["Use-Filter-Opt-Profiles"] = true; 
    }

    if (window.USE_FILTER_OPT_EVENTS_USERS && window.USE_FILTER_OPT_EVENTS_USERS != "") {
      options.headers["Use-Filter-Opt-Events-Users"] = true; 
    }

    if (window.FUNNEL_V2 && window.FUNNEL_V2 != "") {
      options.headers["Funnel-V2"] = true;
    }

    if (window.SCORE && window.SCORE != ""){
      options.headers["Score"] = true
    }

    if (window.SCORE_DEBUG && window.SCORE_DEBUG != ""){
      options.headers["Score-Debug"] = true
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
            if(responsePayload.status >= 400) {
              return Promise.reject(responsePayload);
            } else {
              return Promise.resolve(responsePayload);
            }
          });
      });
  }
  
  export function get(dispatch, url, headers = {}) { return request(dispatch,'GET', url, headers); }
  
  export function post(dispatch, url, data, headers = {}) { return request(dispatch,'POST', url, headers, data); }
  
  export function put(dispatch, url, data, headers = {}) { return request(dispatch,'PUT', url, headers, data); }
  
  export function del(dispatch, url, data, headers = {}) { return request(dispatch,'DELETE', url, headers, data); }
  