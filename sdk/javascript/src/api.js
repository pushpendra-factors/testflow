var Request = require('./utils/request');
const config = require('./config').Config;

const URI_TRACK = "/sdk/event/track";
const URI_IDENTIFY = "/sdk/user/identify";

function Client(token) {
   this.token = token;
}

Client.prototype.getURL = function(uri) {
    return config.api.host+uri;
}

Client.prototype.setToken = function(token) {
    this.token = token;
}

Client.prototype.track = function(userId, eventName, eventProperties={}) {
    eventName = eventName.trim()
    if(!eventName || eventName == "")
        return Promise.reject("Failed Tracking. Event name is missing");

    let payload = { "event_name": eventName, "properties": eventProperties };
    if(userId && userId != null && userId != undefined) payload["user_id"] = userId;

    let customHeaders = { "Authorization": this.token };
    return Request.post(
        this.getURL(URI_TRACK),
        payload,
        customHeaders
    );
}

Client.prototype.identify = function(userId, customerUserId) {
    let payload = {};

    if(userId && userId != null && userId != undefined) 
        payload["userId"] = userId;

    if(customerUserId && customerUserId != null && customerUserId != undefined) 
        payload["c_uid"] = customerUserId;

    let customHeaders = { "Authorization": this.token };
    return Request.post(
        this.getURL(URI_IDENTIFY),
        payload,
        customHeaders
    );
}

module.exports = exports = { Client };
