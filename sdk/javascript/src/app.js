"use strict";

var Cookie = require("./utils/cookie");
const logger = require("./utils/logger");
const Capture = require("./utils/capture");
const util = require("./utils/util");
var APIClient = require("./api-client");
const constant = require("./constant");
const Properties = require("./properties");
const Cache = require("./cache");

const SDK_NOT_INIT_ERROR = new Error("Factors SDK is not initialized.");
const SDK_NO_USER_ERROR = new Error("No user.");

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

function doesUserIdCookieExist() {
    return Cookie.isExist(constant.cookie.USER_ID);
}

function updatePayloadWithUserIdFromCookie(payload) {
    if (doesUserIdCookieExist()) 
        payload.user_id = Cookie.getDecoded(constant.cookie.USER_ID);
    
    return payload;
}

function getAutoTrackURL() {
    return window.location.host + window.location.pathname + util.getCleanHash(window.location.hash);
}

function addCurrentPageAutoTrackEventIdToStore(eventId, eventNamePageURL, originalPageURL) {
    if (!eventId || eventId == "") return;

    Cache.setFactorsCache(Cache.currentPageOriginalURL, originalPageURL);
    Cache.setFactorsCache(Cache.currentPageURLEventName, eventNamePageURL);
    Cache.setFactorsCache(Cache.currentPageTrackEventId, eventId);
}

function setLastActivityTime() {
    Cache.setFactorsCache(Cache.lastActivityTime, util.getCurrentUnixTimestampInMs());
}

function getLastActivityTime() {
    var lastActivityTime = Cache.getFactorsCache(Cache.lastActivityTime);
    return lastActivityTime ? lastActivityTime : 0;
}

function setPrevActivityTime(t) {
    Cache.setFactorsCache(Cache.prevActivityTime, t ? t : 0)
}

function isObject(obj) {
  return Object.prototype.toString.call(obj) === '[object Object]';
}

function  waitForGlobalKey(key, callback, timer = 0, subkey = null, waitTime = 10000, totalTimerCount = 10) {
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
  if (timer <= totalTimerCount) {
    logger.debug('Checking for key: times ' + timer, false);
    setTimeout(function () {
      waitForGlobalKey(key, callback, timer + 1, subkey, waitTime, totalTimerCount);
    }, waitTime);
  }
}

function triggerQueueInitialisedEvent() {
    window.dispatchEvent(new CustomEvent('FACTORS_INITIALISED_EVENT'));
}

const FACTORS_WINDOW_TIMEOUT_KEY_PREFIX = 'lastTimeoutId_';

function setLastTimeoutIdByPeriod(timeoutIn=0, id=0) {
    if (timeoutIn == 0 || id == 0) return;

    var key = FACTORS_WINDOW_TIMEOUT_KEY_PREFIX + timeoutIn;
    Cache.getFactorsCacheObject()[key] = id;
}

function getLastTimeoutIdByPeriod(timeoutIn=0) {
    if (timeoutIn == 0) return;

    var key = FACTORS_WINDOW_TIMEOUT_KEY_PREFIX + timeoutIn;
    return Cache.getFactorsCacheObject()[key];
}

function clearTimeoutByPeriod(timeoutInPeriod) {
    var lastTimeoutId = getLastTimeoutIdByPeriod(timeoutInPeriod);
    if (!lastTimeoutId) return;

    clearTimeout(lastTimeoutId);
    logger.debug("Cleared timeout of "+timeoutInPeriod+"ms :"+lastTimeoutId, false);
}


function getPrevActivityTime() {
    var prevActivityTime = Cache.getFactorsCache(Cache.prevActivityTime);
    return prevActivityTime ? prevActivityTime : 0;
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

    if (opts.track_page_on_spa === true) {
        Cache.setFactorsCache(Cache.trackPageOnSPA, true); 
    }
    
    // Gets info using temp client with given token, if succeeds, 
    // set temp client as app client and set response as app config 
    // or else app stays unintialized.
    var payload = {};
    updatePayloadWithUserIdFromCookie(payload);
    return _client.getInfo(payload)
        .then(function(response) {
            if (response.status < 200 || response.status > 308) {
                logger.errorLine("Get project settings failed with code : ", response.status); 
                return Promise.reject(new Error("FactorsRequestError: Init failed. App configuration failed."));
            }
            return response;
        })
        .then(function(response) {
            // Initialisation.
            _this.config = response.body;
            _this.client = _client;

            // Check if client has given cookie access
            checkCookiesConsentAndProcess(_this);

            return response;
        });
        
}

function checkCookiesConsentAndProcess(_this, response) {
    if(!Cookie.isEnabled()) {
        logger.debug("Checking for cookie consent.", false);
        setTimeout(() => {checkCookiesConsentAndProcess(_this, response)}, 1000)
    } else {
        logger.debug("Cookie consent is enabled. Continuing process", false);
        // Add user_id from response to cookie.
        updateCookieIfUserIdInResponse(response);

        // Start queue processing.
        triggerQueueInitialisedEvent();
        runPostInitProcess(_this);
    }
}

function runPostInitProcess(_this) {
    (function(){
        return Promise.resolve();
    })().then(function() {
        // Enable auto-track SPA page based on settings or init option.
        var enableTrackSPA = Cache.getFactorsCache(Cache.trackPageOnSPA) || _this.getConfig("auto_track_spa_page_view");
        Cache.setFactorsCache(Cache.trackPageOnSPA, enableTrackSPA);
        // Auto-track current page on init, if not disabled.
        return trackOnInit ? _this.autoTrack(_this.getConfig("auto_track"), 
            false, afterPageTrackCallback, true) : triggerFactorsStartQueu();
    })
    .then(function() {
        return _this.autoFormCapture(_this.getConfig("auto_form_capture"));
    })
    .then(function() {
        return _this.autoClickCapture(_this.getConfig("auto_click_capture"));
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

function getEventProperties(eventProperties={}) {
    // Other property validations done on backend.
    eventProperties = Properties.getTypeValidated(eventProperties);

    var referrer = document.referrer;
    // Use page event name on cache as referrer for SPA auto tracking.
    var currentPageOriginalURL = Cache.getFactorsCache(Cache.currentPageOriginalURL)
    if (Cache.getFactorsCache(Cache.trackPageOnSPA) && currentPageOriginalURL) 
        referrer = currentPageOriginalURL;

    // Merge default properties.
    eventProperties = Object.assign(eventProperties, Properties.getEventDefault(referrer));


    return eventProperties;
}

App.prototype.track = function(eventName, eventProperties, auto=false, afterCallback) {
    if (!this.isInitialized()) return Promise.reject(SDK_NOT_INIT_ERROR);

    if (!doesUserIdCookieExist()) return Promise.reject(SDK_NO_USER_ERROR);

    eventName = util.validatedStringArg("event_name", eventName) // Clean event name.
    if (!isAllowedEventName(eventName)) 
        return Promise.reject(new Error("FactorsError: Invalid event name."));

    // The original page URL is added to cache for tracking referrer etc.,
    var originalPageURL = window.location.href;
        
    let payload = {};
    updatePayloadWithUserIdFromCookie(payload);
    payload.event_name = eventName;
    payload.event_properties = getEventProperties(eventProperties);
    payload.user_properties = Properties.getUserDefault();
    payload.auto = auto;

    if (auto) {
        var pageLoadTime = Properties.getPageLoadTime();
        if (pageLoadTime > 0) eventProperties[Properties.PAGE_LOAD_TIME] = pageLoadTime;
    }
    return this.client.track(payload)
        .then(function(response) {
            if (response && response.body) {
                if (!response.body.event_id) {
                    logger.debug("No event_id found on track response.", false);
                    return response;
                }

                if (auto) addCurrentPageAutoTrackEventIdToStore(response.body.event_id, eventName, originalPageURL);
                if (afterCallback) afterCallback(response.body.event_id);
            }            

            return response;
        });
}

App.prototype.updateEventProperties = function(eventId, properties={}) {
    if (!this.isInitialized()) return Promise.reject(SDK_NOT_INIT_ERROR);
    
    if (!doesUserIdCookieExist()) return Promise.reject(SDK_NO_USER_ERROR)
    
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
        var eventId = Cache.getFactorsCache(Cache.currentPageTrackEventId);
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
    var pageEventId = Cache.getFactorsCache(Cache.currentPageTrackEventId); 
    if (pageEventId && pageEventId != undefined) {
        return getAutoTrackURL() == Cache.getFactorsCache(Cache.currentPageURLEventName);
    }

    return false
}

App.prototype.autoTrack = function(enabled=false, force=false, afterCallback, initFactorsQueue=false) {
    if (!enabled) return false;

    if (!force && isPageAutoTracked()) {
        logger.debug('Page tracked already as per store : '+JSON.stringify(Cache.getFactorsCacheObject()))
        return false;
    }

    if(!doesUserIdCookieExist()) return false;
    
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
    var lastPollerId = Cache.getFactorsCache(Cache.lastPollerId);
    clearInterval(lastPollerId);
    if (lastPollerId) logger.debug("Cleared previous page poller: "+lastPollerId, false);

    // update page properties every 20s.
    var pollerId = setInterval(function() {
        lastPageProperties = _this.updatePagePropertiesIfChanged(
            startOfPageSpentTime, lastPageProperties);
    }, 20000);
    
    Cache.setFactorsCache(Cache.lastPollerId, pollerId);

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
    
    // AutoTrack SPA using history.
    // checks support for history and onpopstate listener.
    if ( !Cache.getFactorsCache(Cache.trackPageOnSPA) && window.history && window.onpopstate !== undefined) {
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

    if (Cache.getFactorsCache(Cache.trackPageOnSPA)) {
        setInterval(function(){
            if (Cache.getFactorsCache(Cache.currentPageURLEventName) != getAutoTrackURL()) {
                _this.track(
                    getAutoTrackURL(), 
                    Properties.getFromQueryParams(window.location), 
                    true, 
                    afterCallback,
                );
            }
        }, 1000);
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
            if(!Capture.isPhone(e.phone)) return null;
            var props = {}
            props[Properties.PHONE] = e.phone;
            props[Properties.SOURCE] = 'drift';
            appInstance.track(Properties.EV_FORM_SUBMITTED, props);
        });

        window.drift.on('emailCapture', function (e) {
            if((!e.data || !e.data.email) || !Capture.isEmail(e.data.email)) return null;
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
    2000,
    20
  );
  return true;
};

App.prototype.autoFormCapture = function(enabled=false) {
    if (!enabled) return false; // not enabled.

    // Captures properties from ideal forms which has a submit button. 
    // The fields sumbmitted are processed on callback onSubmit of form.
    Capture.startBackgroundFormBinder(this, this.captureAndTrackFormSubmit)

    // Captures properties from input fields, which are not part of any form
    // on click of any button on the page, which is not a submit button of any form.
    // Note: submit button which is not inside a form is also bound.
    Capture.bindAllNonFormButtonOnClick(this, this.captureAndTrackNonFormInput)


    return true;
}

App.prototype.captureClick = function(appInstance, element) {
    if (!this.isInitialized()) return Promise.reject(SDK_NOT_INIT_ERROR);

    if (!element) logger.debug("Element is undefined on capture click.");

    if (!doesUserIdCookieExist()) return Promise.reject(SDK_NO_USER_ERROR);

    var payload = Capture.getClickCapturePayloadFromElement(element);

    // Add event and user_properties for tracking after enabling.
    payload.event_properties = getEventProperties();
    payload.user_properties = Properties.getUserDefault();

    // Add user_id to payload.
    updatePayloadWithUserIdFromCookie(payload);

    return appInstance.client.captureClick(payload);
}

App.prototype.autoClickCapture = function(enabled=false) {
    if (!enabled) return false; // not enabled.

    Capture.startBackgroundClickBinder(this, this.captureClick)
}

App.prototype.page = function(afterCallback, force=false) {
    if (!this.isInitialized()) return Promise.reject(SDK_NOT_INIT_ERROR);
    
    return Promise.resolve(this.autoTrack(this.getConfig("auto_track"), force, afterCallback));
}

App.prototype.identify = function(customerUserId, userProperties={}) {
    if (!this.isInitialized()) return Promise.reject(SDK_NOT_INIT_ERROR);

    if (!doesUserIdCookieExist()) return Promise.reject(SDK_NO_USER_ERROR);
    
    customerUserId = util.validatedStringArg("customer_user_id", customerUserId);
    
    let payload = {};
    updatePayloadWithUserIdFromCookie(payload);
    payload.c_uid = customerUserId;
    if (Object.keys(userProperties).length > 0) payload.user_properties = userProperties;
    
    return this.client.identify(payload);
}

App.prototype.addUserProperties = function (properties={}) {
    if (!this.isInitialized()) 
        return Promise.reject(SDK_NOT_INIT_ERROR);
    
    if (!doesUserIdCookieExist()) return Promise.reject(SDK_NO_USER_ERROR);


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

    return this.client.addUserProperties(payload);
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
