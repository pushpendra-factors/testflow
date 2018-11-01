var Request = require("./utils/request");
const config = require("./config");

const URI_TRACK = "/sdk/event/track";
const URI_IDENTIFY = "/sdk/user/identify";

function APIClient(token) {
   this.token = token;
}

APIClient.prototype.getURL = function(uri) {
    return config.api.host+uri;
}

APIClient.prototype.setToken = function(token) {
    this.token = token;
}

APIClient.prototype.isInitialized = function() {
    return this.token && (this.token.trim().length > 0);
}

APIClient.prototype.track = function(payload) {
    // Mandatory fields check. Other fields are passed as given.
    if (!payload || !payload.event_name) 
        return Promise.reject("Track failed. API Client invalid payload. Missing event_name.");

    let customHeaders = { "Authorization": this.token };
    return Request.post(
        this.getURL(URI_TRACK),
        payload,
        customHeaders
    );
}

APIClient.prototype.identify = function(payload) {
    // Mandatory fields check. Other fields are passed as given.
    if (!payload || !payload.c_uid)
        return Promise.reject("Identify failed. API Client invalid payload. Missing customer_user_id.");

    let customHeaders = { "Authorization": this.token };
    return Request.post(
        this.getURL(URI_IDENTIFY),
        payload,
        customHeaders
    );
}

module.exports = exports = APIClient;
