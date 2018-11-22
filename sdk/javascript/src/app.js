"use strict";

var Cookie = require("./utils/cookie");
const logger = require("./utils/logger");
const util = require("./utils/util");

var APIClient = require("./api-client");
const constant = require("./constant");

const SDK_NOT_INIT_ERROR = new Error("FactorsError: SDK is not initialized with token.");

function isAllowedEventName(eventName) {
    // Don't allow event_name starts with '$'.
    if (eventName.indexOf('$') == 0) return false; 
    return true;
}

function updateCookieIfUserIdInResponse(response){
    if (response && response.body && response.body.user_id) {
        let cleanUserId = response.body.user_id.trim();
        
        if (cleanUserId) 
            Cookie.setEncoded(constant.cookie.USER_ID, cleanUserId, constant.cookie.EXPIRY);
    }
    return response; // To continue chaining.
}

function throwErrorOnFailureResponse(response, message="Request failed.") {
    if (response.status < 200 || response.status > 308) 
        throw new Error("FactorsRequestError: "+message);
    
    return response;
}

function updatePayloadWithUserIdFromCookie(payload) {
    if (Cookie.isExist(constant.cookie.USER_ID))
        payload.user_id = Cookie.getDecoded(constant.cookie.USER_ID);
    
    return payload;
}

/**
 * Parse query string.
 * @param {*} qString 
 * ----- Cases -----
 * window.search = "" -> {}
 * window.search = "?" -> {}
 * window.search = "?a" -> {a: null}
 * window.search = "?a=" -> {a: null}
 * window.search = "?a=10" -> {a: 10}
 * window.search = "?a=10&" -> {a: 10}
 * window.search = "?a=10&b" -> {a: 10, b: null}
 * window.search = "?a=10&b=20" -> {a: 10, b: 20}
 */
function parseQueryString(qString) {
    var ep = {};
    var t = null;
    // Remove & at the end.
    var ambPos = qString.indexOf("&");
    if (ambPos === qString.length-1) qString = qString.slice(0, qString.length-1);
    if (ambPos >= 0) t = qString.split("&");
    else t = [qString];
    for (var i=0; i<t.length; i++){
        var kv = null;
        if (t[i].indexOf("=") >= 0) kv = t[i].split("=");
        else kv = [t[i], null];
        // Remove ? on first query param.
        if (i == 0 && kv[0].indexOf("?") === 0) kv[0] = kv[0].slice(1);
        // No value, assign null.
        if (kv[1] === "") kv[1] = null;
        ep[kv[0]] = kv[1];
    }
    return ep;
}

/**
 * App prototype.
 */
function App() {
    // lazy initialization with .init
    this.client = null; 
    this.config = {};
}

App.prototype.init = function(token) {
    token = util.validatedStringArg("token", token);

    // Doesn't allow initialize with different token as it needs _fuid reset.
    if (this.isInitialized() && !this.isSameToken(token))
        throw new Error("FactorsInitError: Initialized already. Use reset() and init(), if you really want to do this.");

    if (!token) throw new Error("FactorsArgumentError: Invalid token.");

    let _this = this; // Remove arrows;
    let _client = new APIClient(token);
    // Gets settings using temp client with given token, if succeeds, 
    // set temp client as app client and set response as app config 
    // or else app is stays unintialized.
    return _client.getProjectSettings()
        .then((response) => {
            if (response.status < 200 || response.status > 308) {
                throw new Error("FactorsRequestError: Init failed. App configuration failed.");
            }
            return response;
        })
        .then((response) => {
            _this.config = response.body;
            _this.client = _client;
            return response;
        })
        .then(function() {
            return _this.autoTrack(_this.getConfig("auto_track"));
        })
        .catch(logger.error);
}

App.prototype.track = function(eventName, eventProperties, auto=false) {
    if (!this.isInitialized()) throw SDK_NOT_INIT_ERROR;

    eventName = util.validatedStringArg("event_name", eventName) // Clean event name.
    if (!isAllowedEventName(eventName)) 
        throw new Error("FactorsError: Invalid event name.");

    let payload = {};
    updatePayloadWithUserIdFromCookie(payload);
    payload.event_name = eventName;
    payload.event_properties = eventProperties;
    payload.auto = auto;

    return this.client.track(payload)
        .then(updateCookieIfUserIdInResponse)
        .catch(logger.error);
}

App.prototype.autoTrack = function(enabled=false) {
    if (!enabled) return false; // not enabled.
    var en = window.location.host+window.location.pathname;
    var search =  window.location.search;
    // no query params.
    if (search.length === 0) return this.track(en, {}, true);
    return this.track(en, parseQueryString(search), true);
}

App.prototype.identify = function(customerUserId) {
    if (!this.isInitialized()) throw SDK_NOT_INIT_ERROR;
    
    customerUserId = util.validatedStringArg("customer_user_id", customerUserId);
    
    let payload = {};
    updatePayloadWithUserIdFromCookie(payload);
    payload.c_uid = customerUserId;
    
    return this.client.identify(payload)
        .then(updateCookieIfUserIdInResponse)
        .catch(logger.error);
}

App.prototype.addUserProperties = function (properties={}) {
    if (!this.isInitialized()) throw SDK_NOT_INIT_ERROR;

    if (typeof(properties) != "object")
        throw new Error("FactorsArgumentError: Properties should be an Object(key/values).");
    
    if (Object.keys(properties).length == 0)
        return Promise.reject("No changes. Empty properties.");
    
    let payload = {};
    updatePayloadWithUserIdFromCookie(payload);
    payload.properties = properties;

    return this.client.addUserProperties(payload)
        .then(updateCookieIfUserIdInResponse)
        .catch(logger.error);
}

// Clears the state of the app.
App.prototype.reset = function() {
    this.client = null;
    this.config = {};
    Cookie.remove(constant.cookie.USER_ID);
}

App.prototype.getClient = function() {
    return this.client;
}

App.prototype.getConfig = function(name) {
    if (this.config[name] == undefined)
        throw new Error("FactorsConfigError: Config not present.");

    return this.config[name];
}

App.prototype.isInitialized = function() {
    return !!this.client && !!this.config && (Object.keys(this.config).length > 0);
}

App.prototype.isSameToken = function(token) {
    return this.client && this.client.token && this.client.token === token;
}

module.exports = exports = App;
