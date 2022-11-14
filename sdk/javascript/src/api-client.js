var Request = require("./utils/request");
const config = require("./config");

const URI_TRACK = "/sdk/event/track";
const URI_IDENTIFY = "/sdk/user/identify";
const URI_ADD_USER_PROPERTIES = "/sdk/user/add_properties";
const URI_UPDATE_EVENT_PROPERTIES = "/sdk/event/update_properties";
const URI_GET_INFO = "/sdk/get_info";
const URI_CAPTURE_CLICK = "/sdk/capture_click";
const URI_FORM_FILL = "/sdk/form_fill";

function APIClient(token, host="") {
   this.token = token;
   this.host = host;
}

APIClient.prototype.getURL = function(uri) {
    // use given host if available.
    return this.host != "" ? this.host+uri : (config.api.host+uri);
}

APIClient.prototype.setToken = function(token) {
    this.token = token;
}

APIClient.prototype.isInitialized = function() {
    return this.token && (this.token.trim().length > 0);
}

APIClient.prototype.getHeaders = function() {
    return { "Authorization": this.token };
}

APIClient.prototype.track = function(payload) {
    // Mandatory fields check. Other fields are passed as given.
    if (!payload || !payload.event_name) 
        return Promise.reject("Track failed. API Client invalid payload. Missing event_name.");

    return Request.post(
        this.getURL(URI_TRACK),
        payload,
        this.getHeaders()
    );
}

APIClient.prototype.identify = function(payload) {
    // Mandatory fields check. Other fields are passed as given.
    if (!payload || !payload.c_uid)
        return Promise.reject("Identify failed. API Client invalid payload. Missing customer_user_id.");

    return Request.post(
        this.getURL(URI_IDENTIFY),
        payload,
        this.getHeaders()
    );
}

APIClient.prototype.addUserProperties = function(payload) {
    // Mandaotry field check.
    if (!payload || !payload.properties) 
        return Promise.reject("Add properties failed. Missing properties on payload.");

    return Request.post(
        this.getURL(URI_ADD_USER_PROPERTIES),
        payload,
        this.getHeaders()
    );
}

APIClient.prototype.updateEventProperties = function(payload) {
    // Mandaotry field check.
    if (!payload || !payload.event_id || !payload.properties) 
        return Promise.reject("Update event properties failed. Invalid payload.");

    return Request.post(
        this.getURL(URI_UPDATE_EVENT_PROPERTIES),
        payload,
        this.getHeaders()
    );
}

APIClient.prototype.captureClick = function(payload) {
    // Mandaotry field check.
    if (!payload || 
        !payload.display_name || 
        !payload.element_type || 
        !payload.element_attributes ||
        !payload.user_id ||
        !payload.event_properties ||
        !payload.user_properties)
        return Promise.reject("Capture click failed. Invalid payload.");

    return Request.post(
        this.getURL(URI_CAPTURE_CLICK),
        payload,
        this.getHeaders()
    );
}

APIClient.prototype.captureFormFill = function(payload) {
    // Mandaotry field check.
    if (!payload || 
        !payload.form_id ||
        !payload.field_id ||
        !payload.user_id)
        return Promise.reject("Form fill failed. Invalid payload.");

    return Request.post(
        this.getURL(URI_FORM_FILL),
        payload,
        this.getHeaders()
    );
}

APIClient.prototype.getInfo = function(payload) {
    if (!payload) return Promise.reject("GetInfo failed. Invalid payload.");

    return Request.post(
        this.getURL(URI_GET_INFO),
        payload,
        this.getHeaders()
    );
}

APIClient.prototype.sendError = function(payload) {
   return Request.sendErrorAPI(payload);
}

module.exports = exports = APIClient;
