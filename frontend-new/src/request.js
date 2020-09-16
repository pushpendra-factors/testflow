export function isRequestSuccess(status) {
  return status >= 200 && status <= 399;
}

function request(method, url, headers, data) {
  const options = {
    method,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json'
    }
  };

  if (data && data != undefined) {
    options.body = JSON.stringify(data);
  }

  if (headers && headers != undefined) {
    options.headers = headers;
    options.headers['Content-Type'] = 'application/json';
  }

  if (window.FACTORS_AI_LOGIN_TOKEN && window.FACTORS_AI_LOGIN_TOKEN != '') {
    options.headers.Authorization = window.FACTORS_AI_LOGIN_TOKEN;
  }

  return fetch(url, options)
    .then((response) => {
      // validates response string before JSON unmarshal,
      // for handling no JSON response.
      return response.text()
        .then((text) => {
          const responsePayload = { status: response.status, ok: isRequestSuccess(response.status) };
          if (text == '') responsePayload.data = {};
          else responsePayload.data = JSON.parse(text);

          return responsePayload;
        });
    });
}

export function get(url, headers = {}) { return request('GET', url, headers); }

export function post(url, data, headers = {}) { return request('POST', url, headers, data); }

export function put(url, data, headers = {}) { return request('PUT', url, headers, data); }

export function del(url, headers = {}) { return request('DELETE', url, headers); }
