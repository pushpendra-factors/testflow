"use strict";

var Cookie = require("./utils/cookie");
const logger = require("./utils/logger");
const FormCapture = require("./utils/form_capture");
const util = require("./utils/util");
var APIClient = require("./api-client");
const constant = require("./constant");
const Properties = require("./properties");

const SDK_NOT_INIT_ERROR = new Error("Factors SDK is not initialized.");

function isAllowedEventName(eventName) {
    // whitelisted $ event_name.
    if (eventName == "$form_submitted") return true;

    // Don't allow event_name starts with '$'.
    if (eventName.indexOf("$")  == 0) return false; 
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

function updatePayloadWithUserIdFromCookie(payload) {
    if (Cookie.isExist(constant.cookie.USER_ID))
        payload.user_id = Cookie.getDecoded(constant.cookie.USER_ID);
    
    return payload;
}

function getAutoTrackURL() {
    return window.location.host + window.location.pathname + util.getCleanHash(window.location.hash);
}

function factorsWindow() { 
    if (!window._FactorsCache) window._FactorsCache={}; 
    return window._FactorsCache; 
}

function addCurrentPageAutoTrackEventIdToStore(eventId) {
    if (!eventId || eventId == "") return;
    factorsWindow().currentPageTrackEventId = eventId;
}

function getCurrentPageAutoTrackEventIdFromStore() {
    if (!factorsWindow().currentPageTrackEventId) return;
    return factorsWindow().currentPageTrackEventId;    
}

function setLastActivityTime() {
    factorsWindow().lastActivityTime = util.getCurrentUnixTimestampInMs();
}

function getLastActivityTime() {
    var lastActivityTime = factorsWindow().lastActivityTime;
    if (!lastActivityTime) lastActivityTime = 0;
    return lastActivityTime
}

function getCurrentPageSpentTime(startTime) {
    var lastActivityTime = getLastActivityTime();
    if (lastActivityTime == 0) return lastActivityTime;

    return (lastActivityTime - startTime) / 1000;
}

/**
 * App prototype.
 */
function App() {
    // lazy initialization with .init
    this.client = null; 
    this.config = {};
}

App.prototype.init = function(token, opts={}) {
    token = util.validatedStringArg("token", token);

    // Doesn't allow initialize with different token as it needs _fuid reset.
    if (this.isInitialized() && !this.isSameToken(token))
        return Promise.reject(new Error("FactorsInitError: Initialized already. Use reset() and init(), if you really want to do this."));

    if (!token) return Promise.reject(new Error("FactorsArgumentError: Invalid token."));

    let _this = this; // Remove arrows;
    
    let _client = null;
    if (opts.host && opts.host !== "")
        _client = new APIClient(token, opts.host);
    else 
        _client = new APIClient(token);

    // Turn off auto_track on init with additional opts.
    var trackOnInit = true;
    if (opts.track_on_init === false) {
        trackOnInit = false;
    }
    
    // Gets settings using temp client with given token, if succeeds, 
    // set temp client as app client and set response as app config 
    // or else app stays unintialized.
    return _client.getProjectSettings()
        .then(function(response) {
            if (response.status < 200 || response.status > 308) {
                return Promise.reject(new Error("FactorsRequestError: Init failed. App configuration failed."));
            }
            return response;
        })
        .then(function(response) {
            _this.config = response.body;
            _this.client = _client;
            return response;
        })
        .then(function() {
            return trackOnInit ? _this.autoTrack(_this.getConfig("auto_track")) : null;
        })
        .then(function() {
            return _this.autoFormCapture(_this.getConfig("auto_form_capture"));
        })
        .catch(function() {
            return Promise.reject(new Error("FactorsRequestError: Init failed. App configuration failed."));
        });
}

App.prototype.track = function(eventName, eventProperties, auto=false) {
    if (!this.isInitialized()) return Promise.reject(SDK_NOT_INIT_ERROR);

    eventName = util.validatedStringArg("event_name", eventName) // Clean event name.
    if (!isAllowedEventName(eventName)) 
        return Promise.reject(new Error("FactorsError: Invalid event name."));
        
    // Other property validations done on backend.
    eventProperties = Properties.getTypeValidated(eventProperties);
    // Merge default properties.
    eventProperties = Object.assign(eventProperties, Properties.getEventDefault())

    let payload = {};
    updatePayloadWithUserIdFromCookie(payload);
    payload.event_name = eventName;
    payload.event_properties = eventProperties;
    payload.user_properties = Properties.getUserDefault();
    payload.auto = auto;

    if (auto) {
        var pageLoadTime = Properties.getPageLoadTime();
        if (pageLoadTime > 0) eventProperties[Properties.PAGE_LOAD_TIME] = pageLoadTime;
    }

    return this.client.track(payload)
        .then(updateCookieIfUserIdInResponse)
        .then(function(response) {
            if (auto && response.body && response.body) 
                addCurrentPageAutoTrackEventIdToStore(response.body.event_id);
            return response;
        });
}

App.prototype.updatePageTimeProperties = function(startOfPageSpentTime) {
    var eventId = getCurrentPageAutoTrackEventIdFromStore();
    var pageSpentTime = getCurrentPageSpentTime(startOfPageSpentTime);

    let properties = {};
    if (pageSpentTime > 0) properties[Properties.PAGE_SPENT_TIME] = pageSpentTime;
    this.client.updateEventProperties({event_id: eventId, properties: properties});
}

App.prototype.autoTrack = function(enabled=false) {
    if (!enabled) return false; // not enabled.
    var _this = this;
    
    var startOfPageSpentTime = util.getCurrentUnixTimestampInMs();
    this.track(getAutoTrackURL(), Properties.getFromQueryParams(window.location), true);
    window.addEventListener("beforeunload", function() {
        _this.updatePageTimeProperties(startOfPageSpentTime);
        return;
    });
    window.addEventListener("scroll", setLastActivityTime);
    window.addEventListener("mouseover", setLastActivityTime);
    
    // Todo(Dinesh): Find ways to automate tests for SPA support.
    
    // AutoTrack SPA
    // checks support for history and onpopstate listener.
    if (window.history && window.onpopstate !== undefined) {
        var prevLocation = window.location.href;
        window.addEventListener('popstate', function() {
            logger.debug("Triggered window.onpopstate goto: "+window.location.href+", prev: "+prevLocation);
            if (prevLocation !== window.location.href) {
                // should be called before next page track.
                _this.updatePageTimeProperties(startOfPageSpentTime);
                
                _this.track(getAutoTrackURL(), Properties.getFromQueryParams(window.location), true);
                startOfPageSpentTime = util.getCurrentUnixTimestampInMs();
                prevLocation = window.location.href;
            }
        })
    }
}

// captureTrackFormSubmit - would be attached to 
// form's onSubmit.
App.prototype.captureAndTrackFormSubmit = function(appInstance, e) {
    if (!e || !e.target)
        logger.debug("Form event or event.target is undefined on capture.");

    var properties = FormCapture.getPropertiesFromForm(e.target);
    appInstance.track("$form_submitted", properties);
}

// autoFormCapture - Captures properties from ideal forms which 
// has a submit button. The fields sumbmitted are processed 
// on callback onSubmit(form).
App.prototype.autoFormCapture = function(enabled=false) {
    if (!enabled) return false; // not enabled.
    FormCapture.bindAllFormsOnSubmit(this, this.captureAndTrackFormSubmit);
    return true;
}

App.prototype.page = function() {
    if (!this.isInitialized()) return Promise.reject(SDK_NOT_INIT_ERROR);
    
    return Promise.resolve(this.autoTrack(this.getConfig("auto_track")));
}

App.prototype.identify = function(customerUserId) {
    if (!this.isInitialized()) return Promise.reject(SDK_NOT_INIT_ERROR);
    
    customerUserId = util.validatedStringArg("customer_user_id", customerUserId);
    
    let payload = {};
    updatePayloadWithUserIdFromCookie(payload);
    payload.c_uid = customerUserId;
    
    return this.client.identify(payload)
        .then(updateCookieIfUserIdInResponse);
}

App.prototype.addUserProperties = function (properties={}) {
    if (!this.isInitialized()) 
        return Promise.reject(SDK_NOT_INIT_ERROR);

    if (typeof(properties) != "object")
        return Promise.reject(new Error("FactorsArgumentError: Properties should be an Object(key/values)."));
    
    if (Object.keys(properties).length == 0)
        return Promise.reject("No changes. Empty properties.");

    // Other property validations done on backend.
    properties = Properties.getTypeValidated(properties);

    // Adds default user properties.
    properties = Object.assign(properties, 
            Properties.getUserDefault());
    
    let payload = {};
    updatePayloadWithUserIdFromCookie(payload);
    payload.properties = properties

    return this.client.addUserProperties(payload)
        .then(updateCookieIfUserIdInResponse);
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
    if (this.config[name] == undefined) {
        logger.errorLine(new Error("FactorsConfigError: Config not present."));
        return
    }

    return this.config[name];
}

App.prototype.isInitialized = function() {
    return !!this.client && !!this.config && (Object.keys(this.config).length > 0);
}

App.prototype.isSameToken = function(token) {
    return this.client && this.client.token && this.client.token === token;
}

module.exports = exports = App;
