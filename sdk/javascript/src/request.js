import * as logger from "./logger";

function request(method, url, data) {
    let options = { method: method };

    if(data && data != undefined) {
        options["body"] = JSON.stringify(data)
    }

    return fetch(url, options)
        .then((response) =>  {
            return response.json()
                .then((responseJSON) => {
                    let responseBody = { status: response.status, body: responseJSON };
                    return response.ok ? Promise.resolve(responseBody) : Promise.reject(responseBody);
                });
        })
        .catch((error) => {
            logger.error(error.stack);
            return Promise.reject(error.message);
        });
}

function get(url) { return request("get", url); }

function post(url, data) { return request("post", url, data); }

export { get, post };