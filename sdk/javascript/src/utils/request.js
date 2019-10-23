var logger = require("./logger");

function request(method, url, headers, data) {
    let options = { method: method };

    if(data && data != undefined) 
        options["body"] = JSON.stringify(data);

    if(headers && headers != undefined ) {
        options.headers = headers;

        // Default headers.
        options.headers["Content-Type"] = "application/json";
    }

    return fetch(url, options)
        .then(function(response) {
            var _response = response;
            return _response.json()
                .then(function(responseJSON) {
                    if (!_response.ok) return Promise.reject("Failed to fetch.");
                    return { status: _response.status, body: responseJSON };
                })
        });
}

function get(url, headers={}) { return request("get", url, headers); }

function post(url, data, headers={}) { return request("post", url, headers, data); }

module.exports = exports = { get, post };