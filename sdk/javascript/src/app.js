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

// Todo: Use a prototype for window cache object.
function factorsWindow() { 
    if (!window._FactorsCache) window._FactorsCache={}; 
    return window._FactorsCache; 
}

function addCurrentPageAutoTrackEventIdToStore(eventId, eventNamePageURL) {
    if (!eventId || eventId == "") return;
    factorsWindow().currentPageURLEventName = eventNamePageURL;
    factorsWindow().currentPageTrackEventId = eventId;
}

function getCurrentPageAutoTrackEventIdFromStore() {
    if (!factorsWindow().currentPageTrackEventId) return;
    return factorsWindow().currentPageTrackEventId;    
}

function getCurrentPageAutoTrackEventPageURLFromStore() {
    if (!factorsWindow().currentPageTrackEventId) return;
    return factorsWindow().currentPageURLEventName; 
}

function setLastActivityTime() {
    factorsWindow().lastActivityTime = util.getCurrentUnixTimestampInMs();
}

function getLastActivityTime() {
    var lastActivityTime = factorsWindow().lastActivityTime;
    if (!lastActivityTime) lastActivityTime = 0;
    return lastActivityTime
}

function getCurrentPageSpentTimeInSecs(startTimeInMs) {
    var lastActivityTime = getLastActivityTime();
    if (lastActivityTime == 0) return lastActivityTime;

    return (lastActivityTime - startTimeInMs) / 1000;
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
                logger.errorLine("Get project settings failed with code : ", response.status); 
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
        .catch(function(err) {
            logger.errorLine(err);
            return Promise.reject(err.stack + " during get_settings on init.");
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
                addCurrentPageAutoTrackEventIdToStore(response.body.event_id, eventName);
            return response;
        });
}

App.prototype.updatePagePropertiesIfChanged = function(startOfPageSpentTimeInMs, lastPageProperties) {
    let lastPageSpentTimeInSecs = lastPageProperties && lastPageProperties[Properties.PAGE_SPENT_TIME] ? 
        lastPageProperties[Properties.PAGE_SPENT_TIME] : 0;
    
    let lastPageScrollPercentage = lastPageProperties && lastPageProperties[Properties.PAGE_SCROLL_PERCENT] ?
        lastPageProperties[Properties.PAGE_SCROLL_PERCENT] : 0;

    var pageSpentTimeInSecs = getCurrentPageSpentTimeInSecs(startOfPageSpentTimeInMs);
    var pageScrollPercentage = Properties.getPageScrollPercent();

    // add properties if changed.
    let properties = {};
    if (pageSpentTimeInSecs > 0 && pageSpentTimeInSecs > lastPageSpentTimeInSecs) 
        properties[Properties.PAGE_SPENT_TIME] = pageSpentTimeInSecs;

    if (pageScrollPercentage > 0 && pageScrollPercentage > lastPageScrollPercentage )
        properties[Properties.PAGE_SCROLL_PERCENT] = pageScrollPercentage;

    // update if any properties given.
    if (Object.keys(properties).length > 0) {
        logger.debug("Updating page properties : " + JSON.stringify(properties), false);
        var eventId = getCurrentPageAutoTrackEventIdFromStore();
        this.client.updateEventProperties({ event_id: eventId, properties: properties });
    } else {
        logger.debug("No change on page properties, skipping update for : " + JSON.stringify(lastPageProperties), false);
    }

    return {
        [Properties.PAGE_SCROLL_PERCENT]: pageScrollPercentage, 
        [Properties.PAGE_SPENT_TIME]: pageSpentTimeInSecs
    };
}

function isPageAutoTracked() {
    var pageEventId = getCurrentPageAutoTrackEventIdFromStore();
    if (pageEventId && pageEventId != undefined) {
        return getAutoTrackURL() == getCurrentPageAutoTrackEventPageURLFromStore();
    }

    return false
}

App.prototype.autoTrack = function(enabled=false) {
    if (!enabled) return false; // not enabled.

    if (isPageAutoTracked()) {
        logger.debug('Page tracked already as per store : '+JSON.stringify(factorsWindow()))
        return false;
    }
    
    var _this = this;

    this.track(getAutoTrackURL(), Properties.getFromQueryParams(window.location), true);

    var lastPageProperties = {};
    var startOfPageSpentTime = util.getCurrentUnixTimestampInMs();
    // update page properties every 20s.
    setInterval(function() {
        lastPageProperties = _this.updatePagePropertiesIfChanged(
            startOfPageSpentTime, lastPageProperties);
    }, 20000);

    // update page properties before leaving the page.
    window.addEventListener("beforeunload", function() {
        lastPageProperties = _this.updatePagePropertiesIfChanged(
            startOfPageSpentTime, lastPageProperties);
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
                lastPageProperties = _this.updatePagePropertiesIfChanged(
                    startOfPageSpentTime, lastPageProperties);
                
                _this.track(getAutoTrackURL(), Properties.getFromQueryParams(window.location), true);
                startOfPageSpentTime = util.getCurrentUnixTimestampInMs();
                prevLocation = window.location.href; 
            }
        })
    }
}

App.prototype.captureAndTrackFormSubmit = function(appInstance, e) {
    if (!e || !e.target)
        logger.debug("Form event or event.target is undefined on capture.");

    var properties = Properties.getPropertiesFromForm(e.target ? e.target : e);
    if (!properties || Object.keys(properties).length)
        logger.debug("No properties captured from form.", false);

    // do not track if email and phone is not there on captured properties.
    if (!properties[Properties.EMAIL] && !properties[Properties.PHONE]) {
        logger.debug("No email and phone, skipping form submit.", false);
        return;
    } 

    logger.debug("Capturing form submit with properties: "+JSON.stringify(properties), false);

    appInstance.track("$form_submitted", properties);
}

App.prototype.captureAndTrackNonFormInput = function(appInstance) {
    var properties = Properties.getPropertiesFromAllNonFormInputs();

    // do not track if email and phone is not there on captured properties.
    if (!properties[Properties.EMAIL] && !properties[Properties.PHONE]){
        logger.debug("No email and phone, skipping form submit.", false);
        return; 
    }

    logger.debug("Capturing non-form submit with properties: "+JSON.stringify(properties), false);

    appInstance.track("$form_submitted", properties);
}

App.prototype.autoFormCapture = function(enabled=false) {
    if (!enabled) return false; // not enabled.

    // Captures properties from ideal forms which has a submit button. 
    // The fields sumbmitted are processed on callback onSubmit of form.
    FormCapture.bindAllFormsOnSubmit(this, this.captureAndTrackFormSubmit);

    // Captures properties from input fields, which are not part of any form
    // on click of any button on the page, which is not a submit button of any form.
    // Note: submit button which is not inside a form is also bound.
    FormCapture.bindAllNonFormButtonOnClick(this, this.captureAndTrackNonFormInput)
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

App.prototype.getUserId = function() {
    return Cookie.getDecoded(constant.cookie.USER_ID);
}

App.prototype.handleError = function(error) {
    var errMsg = "";
    if (typeof(error) == "string") errMsg = error;
    if (error instanceof Error && error.message) errMsg = error.stack;

    let payload = {};
    payload.domain = window.location.host;
    payload.url = window.location.href;
    payload.error = errMsg;
    updatePayloadWithUserIdFromCookie(payload);

    var client = new APIClient('', '');
    client.sendError(payload);

    logger.errorLine(error);

    return Promise ? Promise.reject(error) : error;
}

module.exports = exports = App;
