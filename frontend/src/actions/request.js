function request(dispatch, method, url, headers, data){
    
    let options = {
        method: method,
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json'
        },
    };

    if(data && data != undefined){
        options.body= JSON.stringify(data)
    }

    if(headers && headers != undefined ) {
        options.headers = headers;
        options.headers["Content-Type"] = "application/json";
    }

    return fetch(url, options)
    .then((response)=>{ 
        if (response.status === 401){
            if (dispatch && dispatch != undefined) 
                dispatch({type: "AGENT_LOGIN_FORCE"});
            return Promise.reject(response);
        }
        return response.json()
            .then((r) => { return { data: r, status: r.status }; });
    });
}

export function get(dispatch, url, headers={}) { return request(dispatch, "GET", url, headers); }

export  function post(dispatch, url, data, headers={}) { return request(dispatch, "POST", url, headers, data); }

export  function put(dispatch, url, data, headers={}) { return request(dispatch, "PUT", url, headers, data); }

export function del(dispatch, url, headers={}) { return request(dispatch, "DELETE", url, headers); }