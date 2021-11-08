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
    if (eventName == Properties.EV_FORM_SUBMITTED) return true;

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

function setPrevActivityTime(t) {
    factorsWindow().prevActivityTime = t ? t : 0;
}

function setLastPollerId(id) {
    factorsWindow().lastPollerId = id;
}

function getLastPollerId() {
    return factorsWindow().lastPollerId;
}

function isObject(obj) {
  return Object.prototype.toString.call(obj) === '[object Object]';
}

function waitForGlobalKey(key, callback, timer = 0, subkey = null, waitTime = 10000) {
  if (window[key]) {
    if (subkey) {
      if (Array.isArray(window[key])) {
        const isPresent = window[key].find(function (elem) {
          return isObject(elem) && Object.keys(elem).indexOf(subkey) > -1;
        });
        if (isPresent) {
          callback();
          return;
        }
      }
    } else {
      callback();
      return;
    }
  }
  if (timer <= 10) {
    logger.debug('Checking for key: times ' + timer, false);
    setTimeout(function () {
      waitForGlobalKey(key, callback, timer + 1, subkey, waitTime);
    }, waitTime);
  }
}

const FACTORS_WINDOW_TIMEOUT_KEY_PREFIX = 'lastTimeoutId_';

function setLastTimeoutIdByPeriod(timeoutIn=0, id=0) {
    if (timeoutIn == 0 || id == 0) return;

    var key = FACTORS_WINDOW_TIMEOUT_KEY_PREFIX + timeoutIn;
    factorsWindow()[key] = id;
}

function getLastTimeoutIdByPeriod(timeoutIn=0) {
    if (timeoutIn == 0) return;

    var key = FACTORS_WINDOW_TIMEOUT_KEY_PREFIX + timeoutIn;
    return factorsWindow()[key];
}

function clearTimeoutByPeriod(timeoutInPeriod) {
    var lastTimeoutId = getLastTimeoutIdByPeriod(timeoutInPeriod);
    if (!lastTimeoutId) return;

    clearTimeout(lastTimeoutId);
    logger.debug("Cleared timeout of "+timeoutInPeriod+"ms :"+lastTimeoutId, false);
}


function getPrevActivityTime() {
    var prevActivityTime = factorsWindow().prevActivityTime;
    if (!prevActivityTime) prevActivityTime = 0;
    return prevActivityTime
}

function getCurrentPageSpentTimeInMs(pageLandingTimeInMs, lastSpentTimeInMs) {
    var prevActivityTime = getPrevActivityTime();
    if (prevActivityTime == 0) prevActivityTime = pageLandingTimeInMs;

    var lastActivityTime = getLastActivityTime();
    if (lastActivityTime == 0) return 0;

    // init with last spent time.
    var totalSpentTimeInMs = lastSpentTimeInMs;
    
    // Add to total spent time only if diff is lesser than
    // inactivity threshold (5 mins).
    var diffTimeInMs = lastActivityTime - prevActivityTime;
    if (diffTimeInMs < 300000) {
        totalSpentTimeInMs = totalSpentTimeInMs + diffTimeInMs;
    }

    setPrevActivityTime(lastActivityTime);

    return totalSpentTimeInMs;
}

/**
 * App prototype.
 */
function App() {
    // lazy initialization with .init
    this.client = null; 
    this.config = {};
}

App.prototype.init = function(token, opts={}, afterPageTrackCallback) {
    token = util.validatedStringArg("token", token);

    // Doesn't allow initialize with different token as it needs _fuid reset.
    if (this.isInitialized() && !this.isSameToken(token))
        return Promise.reject(new Error("FactorsInitError: Initialized already. Use reset() and init(), if you really want to do this."));

    if (!token) return Promise.reject(new Error("FactorsArgumentError: Invalid token."));

    let _this = this; // Remove arrows;

    if (!opts) opts = {};
    
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
            return trackOnInit ? _this.autoTrack(_this.getConfig("auto_track"), false, afterPageTrackCallback) : null;
        })
        .then(function() {
            return _this.autoFormCapture(_this.getConfig("auto_form_capture"));
        })
        .then(function() {
            return _this.autoDriftEventsCapture(_this, _this.getConfig("int_drift"));
        })
        .then(function() {
            return _this.autoClearbitRevealCapture(_this, _this.getConfig("int_clear_bit"));
        })
        .catch(function(err) {
            logger.errorLine(err);
            return Promise.reject(err.stack + " during get_settings on init.");
        });
}

App.prototype.track = function(eventName, eventProperties, auto=false, afterCallback) {
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
            if (response && response.body) {
                if (!response.body.event_id) {
                    logger.debug("No event_id found on track response.", false);
                    return response;
                }

                if (auto) addCurrentPageAutoTrackEventIdToStore(response.body.event_id, eventName);
                if (afterCallback) afterCallback(response.body.event_id);
            }            

            return response;
        });
}

App.prototype.updateEventProperties = function(eventId, properties={}) {
    if (!this.isInitialized()) return Promise.reject(SDK_NOT_INIT_ERROR);
    
    if (!eventId || eventId == '') 
        return Promise.reject("No eventId provided for update.");
    
    if (Object.keys(properties).length == 0)
        logger.debug("No properties given to update event.");
        
    var payload = { event_id: eventId, properties: properties };
    return this.client.updateEventProperties(updatePayloadWithUserIdFromCookie(payload));
}

App.prototype.updatePagePropertiesIfChanged = function(pageLandingTimeInMs, 
    lastPageProperties, defaultPageSpentTimeInMs=0) {

    let lastPageSpentTimeInMs = lastPageProperties && lastPageProperties[Properties.PAGE_SPENT_TIME] ? 
        lastPageProperties[Properties.PAGE_SPENT_TIME] : 0;

    let lastPageScrollPercentage = lastPageProperties && lastPageProperties[Properties.PAGE_SCROLL_PERCENT] ?
        lastPageProperties[Properties.PAGE_SCROLL_PERCENT] : 0;

    var pageSpentTimeInMs = getCurrentPageSpentTimeInMs(pageLandingTimeInMs, lastPageSpentTimeInMs);
    var pageScrollPercentage = Properties.getPageScrollPercent();

    if (pageSpentTimeInMs == 0 && defaultPageSpentTimeInMs > 0) {
        pageSpentTimeInMs = defaultPageSpentTimeInMs;
    }

    // add page_load_time to page_spent_time initially and when defaulted.
    if (lastPageSpentTimeInMs == 0 || defaultPageSpentTimeInMs > 0) {
        pageSpentTimeInMs = pageSpentTimeInMs + Properties.getPageLoadTimeInMs();
    }
    
    // add properties if changed.
    var properties = {};
    if (pageSpentTimeInMs > 0 && pageSpentTimeInMs > lastPageSpentTimeInMs) {
        // page spent time added to payload in secs.
        var pageSpentTimeInSecs = pageSpentTimeInMs / 1000;
        pageSpentTimeInSecs = Number(pageSpentTimeInSecs.toFixed(2));
        properties[Properties.PAGE_SPENT_TIME] = pageSpentTimeInSecs;
    } 
    if (pageScrollPercentage > 0 && pageScrollPercentage > lastPageScrollPercentage ) {
        pageScrollPercentage = Number(pageScrollPercentage.toFixed(2));
        properties[Properties.PAGE_SCROLL_PERCENT] = pageScrollPercentage;
    }
        
    // update if any properties given.
    if (Object.keys(properties).length > 0) {
        logger.debug("Updating page properties : " + JSON.stringify(properties), false);
        var eventId = getCurrentPageAutoTrackEventIdFromStore();
        var payload = { event_id: eventId, properties: properties };
        this.client.updateEventProperties(updatePayloadWithUserIdFromCookie(payload));
    } else {
        logger.debug("No change on page properties, skipping update for : " + JSON.stringify(lastPageProperties), false);
    }

    return {
        [Properties.PAGE_SCROLL_PERCENT]: pageScrollPercentage, 
        [Properties.PAGE_SPENT_TIME]: pageSpentTimeInMs
    };
}

function isPageAutoTracked() {
    var pageEventId = getCurrentPageAutoTrackEventIdFromStore();
    if (pageEventId && pageEventId != undefined) {
        return getAutoTrackURL() == getCurrentPageAutoTrackEventPageURLFromStore();
    }

    return false
}

App.prototype.autoTrack = function(enabled=false, force=false, afterCallback) {
    if (!enabled) return false; // not enabled.

    if (!force && isPageAutoTracked()) {
        logger.debug('Page tracked already as per store : '+JSON.stringify(factorsWindow()))
        return false;
    }
    
    var _this = this;

    this.track(getAutoTrackURL(), Properties.getFromQueryParams(window.location), true, afterCallback);

    var lastPageProperties = {};
    var startOfPageSpentTime = util.getCurrentUnixTimestampInMs();

    // Todo: Use curried function to remove multiple set timeouts.
    // update page properties after 5s and 10s with default value.
    var fiveSecondsInMs = 5000;
    clearTimeoutByPeriod(fiveSecondsInMs);
    var timoutId5thSecond = setTimeout(function() {
        logger.debug("Triggered properties update after 5s.", false);
        lastPageProperties = _this.updatePagePropertiesIfChanged(
            startOfPageSpentTime, lastPageProperties, fiveSecondsInMs);
    }, fiveSecondsInMs);
    setLastTimeoutIdByPeriod(fiveSecondsInMs, timoutId5thSecond);

    var tenSecondsInMs = 10000;
    clearTimeoutByPeriod(tenSecondsInMs);
    var timoutId10thSecond = setTimeout(function() {
        logger.debug("Triggered properties update after 10s.", false);
        lastPageProperties = _this.updatePagePropertiesIfChanged(
            startOfPageSpentTime, lastPageProperties, tenSecondsInMs);
    }, tenSecondsInMs);
    setLastTimeoutIdByPeriod(tenSecondsInMs, timoutId10thSecond);

    // clear the previous poller, if exist.
    var lastPollerId = getLastPollerId();
    clearInterval(lastPollerId);
    if (lastPollerId) logger.debug("Cleared previous page poller: "+lastPollerId, false);

    // update page properties every 20s.
    var pollerId = setInterval(function() {
        lastPageProperties = _this.updatePagePropertiesIfChanged(
            startOfPageSpentTime, lastPageProperties);
    }, 20000);
    
    setLastPollerId(pollerId);

    // update page properties before leaving the page.
    window.addEventListener("beforeunload", function() {
        lastPageProperties = _this.updatePagePropertiesIfChanged(
            startOfPageSpentTime, lastPageProperties);
        return;
    });

    window.addEventListener("scroll", setLastActivityTime);
    window.addEventListener("mouseover", setLastActivityTime);
    window.addEventListener("mousemove", setLastActivityTime);

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
                
                _this.track(getAutoTrackURL(), Properties.getFromQueryParams(window.location), true, afterCallback);
                startOfPageSpentTime = util.getCurrentUnixTimestampInMs();
                prevLocation = window.location.href; 
            }
        })
    }
}

App.prototype.captureAndTrackFormSubmit = function(appInstance, formElement) {
    if (!formElement)
        logger.debug("Form element is undefined on capture form submit.");

    var properties = Properties.getPropertiesFromForm(formElement);
    var formProperties = Properties.getFormMetaAttributes(formElement);

    if(formProperties && Object.keys(formProperties).length > 0) {
        logger.debug("Collecting form meta attributes", false);
        properties = Object.assign(formProperties, properties);
    }
    if (!properties || Object.keys(properties).length == 0)
        logger.debug("No properties captured from form.", false);

    // do not track if email and phone is not there on captured properties.
    if (!properties[Properties.EMAIL] && !properties[Properties.PHONE]) {
        logger.debug("No email and phone, skipping form submit.", false);
        return;
    } 

    logger.debug("Capturing form submit with properties: "+JSON.stringify(properties), false);

    appInstance.track(Properties.EV_FORM_SUBMITTED, properties);
}

App.prototype.captureAndTrackNonFormInput = function(appInstance) {
    var properties = Properties.getPropertiesFromAllNonFormInputs();

    // do not track if email and phone is not there on captured properties.
    if (!properties[Properties.EMAIL] && !properties[Properties.PHONE]){
        logger.debug("No email and phone, skipping form submit.", false);
        return; 
    }

    logger.debug("Capturing non-form submit with properties: "+JSON.stringify(properties), false);

    appInstance.track(Properties.EV_FORM_SUBMITTED, properties);
}

App.prototype.autoDriftEventsCapture = function(appInstance, enabled) {
    if (!enabled) return false; // not enabled.
    waitForGlobalKey("drift", function() {
        window.drift.on('phoneCapture', function (e) {
            if(!FormCapture.isPhone(e.phone)) return null;
            var props = {}
            props[Properties.PHONE] = e.phone;
            props[Properties.SOURCE] = 'drift';
            appInstance.track(Properties.EV_FORM_SUBMITTED, props);
        });

        window.drift.on('emailCapture', function (e) {
            if((!e.data || !e.data.email) || !FormCapture.isEmail(e.data.email)) return null;
            var props = {}
            props[Properties.EMAIL] = e.data.email;
            props[Properties.SOURCE] = 'drift';
            appInstance.track(Properties.EV_FORM_SUBMITTED, props);
        });
    });

    return true;
}

function handleRevealData(appInstance) {
  const revealData = window.dataLayer.find(function (d) {
    return isObject(d) && Object.keys(d).indexOf('reveal') > -1;
  }).reveal;
  const availableProperties = {};
  if (revealData) {
    if (revealData.company && isObject(revealData.company)) {
      const companyData = revealData.company;
      const companyPrefix = '$clr_company';

      if (companyData.name) {
        availableProperties[companyPrefix + '_name'] = companyData.name;
      }

      if (companyData.foundedYear) {
        availableProperties[companyPrefix + '_foundedYear'] =
          companyData.foundedYear;
      }

      if (companyData.type) {
        availableProperties[companyPrefix + '_type'] = companyData.type;
      }

      if (companyData.geo && isObject(companyData.geo)) {
        const companyGeographicalData = companyData.geo;
        const requiredGeographicalKeys = [
          'city',
          'country',
          'postalCode',
          'state',
          'stateCode',
          'countryCode',
          'lat',
          'lng',
        ];
        const availableGeographicalKeys = requiredGeographicalKeys.filter(
          function (key) {
            return !!companyGeographicalData[key];
          }
        );
        for (let i = 0; i < availableGeographicalKeys.length; i++) {
          const key = availableGeographicalKeys[i];
          availableProperties[companyPrefix + '_geo_' + key] =
            companyGeographicalData[key];
        }
      }

      if (companyData.category && isObject(companyData.category)) {
        const companyCategoryData = companyData.category;
        const requiredCategoryKeys = [
          'sector',
          'industryGroup',
          'industry',
          'subIndustry',
          'sicCode',
          'naicsCode',
        ];
        const availableCategoryKeys = requiredCategoryKeys.filter(function (
          key
        ) {
          return !!companyCategoryData[key];
        });
        for (let i = 0; i < availableCategoryKeys.length; i++) {
          const key = availableCategoryKeys[i];
          availableProperties[companyPrefix + '_category_' + key] =
            companyCategoryData[key];
        }
      }

      if (companyData.metrics && isObject(companyData.metrics)) {
        const companyMetricsData = companyData.metrics;
        const requiredMetricsKeys = [
          'alexaUsRank',
          'alexaGlobalRank',
          'employees',
          'employeesRange',
          'marketCap',
          'raised',
          'annualRevenue',
          'estimatedAnnualRevenue',
          'fiscalYearEnd',
        ];
        const availableMetricsKeys = requiredMetricsKeys.filter(function (key) {
          return !!companyMetricsData[key];
        });
        for (let i = 0; i < availableMetricsKeys.length; i++) {
          const key = availableMetricsKeys[i];
          availableProperties[companyPrefix + '_metrics_' + key] =
            companyMetricsData[key];
        }
      }

      if (companyData.parent && isObject(companyData.parent)) {
        const companyParentData = companyData.parent;
        const requiredParentKeys = ['domain'];
        const availableParentKeys = requiredParentKeys.filter(function (key) {
          return !!companyParentData[key];
        });
        for (let i = 0; i < availableParentKeys.length; i++) {
          const key = availableParentKeys[i];
          availableProperties[companyPrefix + '_parent_' + key] =
            companyParentData[key];
        }
      }
    }
  }
  if (Object.keys(availableProperties).length) {
    appInstance.addUserProperties(availableProperties);
  }
}

App.prototype.autoClearbitRevealCapture = function (appInstance, enabled) {
  if (!enabled) return false; // not enabled.
  waitForGlobalKey(
    'dataLayer',
    handleRevealData.bind(null, appInstance),
    0,
    'reveal',
    5000
  );
  return true;
};

App.prototype.autoFormCapture = function(enabled=false) {
    if (!enabled) return false; // not enabled.

    // Captures properties from ideal forms which has a submit button. 
    // The fields sumbmitted are processed on callback onSubmit of form.
    FormCapture.startBackgroundFormBinder(this, this.captureAndTrackFormSubmit)

    // Captures properties from input fields, which are not part of any form
    // on click of any button on the page, which is not a submit button of any form.
    // Note: submit button which is not inside a form is also bound.
    FormCapture.bindAllNonFormButtonOnClick(this, this.captureAndTrackNonFormInput)


    return true;
}

App.prototype.page = function(afterCallback, force=false) {
    if (!this.isInitialized()) return Promise.reject(SDK_NOT_INIT_ERROR);
    
    return Promise.resolve(this.autoTrack(this.getConfig("auto_track"), force, afterCallback));
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
    payload.auto_track_url = getAutoTrackURL();
    updatePayloadWithUserIdFromCookie(payload);

    // Using new client without token, 
    // as this is generic call not specific to a project.
    var client = new APIClient('', '');
    client.sendError(payload); 

    logger.errorLine(error);

    return Promise ? Promise.reject(error) : error;
}

module.exports = exports = App;
