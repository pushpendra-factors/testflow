"use strict";

var Cookie = require("./utils/cookie");
const logger = require("./utils/logger");
const util = require("./utils/util");

var APIClient = require("./api-client");
const constant = require("./constant");

function App() {
    // lazy initialization with .init
    this.client = null; 
    this.config = {};
}

App.prototype.init = function(token) {
    // Doesn't allow initialize with different token as it needs _fuid reset.
    if (this.isInitialized() && !this.isSameToken(token))
        throw new Error("FactorsInitError: Initialized already. Use reset() and init(), if you really want to do this.")

    if (!token) throw new Error("FactorsArgumentError: Invalid token.");

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
            this.config = response.body;
            this.client = _client;
            return response;
        })
        .catch(logger.error);
}

App.prototype.track = function(eventName, eventProperties, auto=false) {
    if (!this.isInitialized())
        throw new Error("FactorsError: SDK is not initialised with token.");

    eventName = util.validatedStringArg("event_name", eventName)

    let payload = {};
    updatePayloadWithUserIdFromCookie(payload);
    payload.event_name = eventName;
    payload.properties = eventProperties;
    // payload.auto = auto;

    return this.client.track(payload)
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

// Common methods.

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

module.exports = exports = {
    App: App,
    updatePayloadWithUserIdFromCookie: updatePayloadWithUserIdFromCookie,
    throwErrorOnFailureResponse: throwErrorOnFailureResponse,
    updateCookieIfUserIdInResponse: updateCookieIfUserIdInResponse,
    parseQueryString: parseQueryString
};

