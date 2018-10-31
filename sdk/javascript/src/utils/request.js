import * as logger from "./logger";


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
        .then((response) =>  {
            return response.json()
                .then((responseJSON) => {
                    let responseBody = { status: response.status, body: responseJSON };
                    return response.ok ? Promise.resolve(responseBody) : Promise.reject(responseBody);
                })
                .catch((JSONError) => {
                    logger.error(JSONError);
                    return Promise.reject(JSONError);
                });
        })
        .catch((error) => {
            logger.error(JSON.stringify(error));
            return Promise.reject(error);
        });
}

function get(url, headers={}) { return request("get", url, headers); }

function post(url, data, headers={}) { return request("post", url, headers, data); }

export { get, post };