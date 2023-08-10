"use strict";

var Cookie = require("./utils/cookie");
const logger = require("./utils/logger");
const Capture = require("./utils/capture");
const util = require("./utils/util");
var APIClient = require("./api-client");
const constant = require("./constant");
const Properties = require("./properties");
const Cache = require("./cache");
const { isLocalStorageAvailable } = require("./utils/util");
const { processAllLocalStorageBacklogRequests } = require("./utils/request");
const properties = require("./properties");
const { lastActivityTime, getFAITrackerCache } = require("./cache");

const SDK_NOT_INIT_ERROR = new Error("FAITracker SDK is not initialized.");
const SDK_NO_USER_ERROR = new Error("No user.");

const FAITRACKER_INPUT_ID_ATTRIBUTE = "data-faitracker-input-id";

function isAllowedEventName(eventName) {
    // whitelisted $ event_name.
    if (eventName == Properties.EV_FORM_SUBMITTED) return true;

    // Don't allow event_name starts with '$'.
    if (eventName.indexOf("$")  == 0) return false; 
    return true;
}

function updateCookieIfUserIdInResponse(response){
    logger.debug("Setting Cookie with response: ", false);
    if (response && response.body && response.body.user_id) {
        let cleanUserId = response.body.user_id.trim();
        if (cleanUserId) 
            Cookie.setEncoded(constant.cookie.USER_ID, cleanUserId, constant.cookie.EXPIRY);
    }
    return response; // To continue chaining.
}

function doesUserIdCookieExist() {
    logger.debug("Checking if cookies have been stored: ", false);
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

function setCurrentPageAttributesToStore(eventId, eventNamePageURL, originalPageURL) {
    if (!eventId || eventId == "") return;

    Cache.setFAITrackerCache(Cache.currentPageOriginalURL, originalPageURL);
    Cache.setFAITrackerCache(Cache.currentPageURLEventName, eventNamePageURL);
    Cache.setFAITrackerCache(Cache.currentPageTrackEventId, eventId);

    // Reset all fields related to timer.
    Cache.setFAITrackerCache(Cache.lastPageProperties, {});
    Cache.setFAITrackerCache(Cache.lastActivityTime, 0);
    Cache.setFAITrackerCache(Cache.currentPageSpentTimeInMs, 0);
}

function setLastActivityTime() {
    var currentTime = util.getCurrentUnixTimestampInMs();

    var lastActivityTime = getLastActivityTime();
    if (lastActivityTime == 0) {
        Cache.setFAITrackerCache(Cache.lastActivityTime, currentTime);
        return
    }

    // The gap between last activity and current activity cannot be more than 
    // 100ms as the activity captured is continious.
    // Anything higher won't be considered as continious.
    if ((currentTime - lastActivityTime) < 100) {
        var diff = currentTime - lastActivityTime;
        var spentTimeInMs = getCurrentPageSpentTimeInMs();
        var newSpentTimeInMs = spentTimeInMs + diff;
        Cache.setFAITrackerCache(Cache.currentPageSpentTimeInMs, newSpentTimeInMs);
        logger.debug("Page spent time: " + newSpentTimeInMs + "ms");
    }

    Cache.setFAITrackerCache(Cache.lastActivityTime, currentTime);
}

function getLastActivityTime() {
    var lastActivityTime = Cache.getFAITrackerCache(Cache.lastActivityTime);
    return lastActivityTime ? lastActivityTime : 0;
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
    window.dispatchEvent(new CustomEvent('FAITRACKER_INITIALISED_EVENT'));

    // For backward compatibility.
    window.dispatchEvent(new CustomEvent('FACTORS_INITIALISED_EVENT'));
}

const FAITRACKER_WINDOW_TIMEOUT_KEY_PREFIX = 'lastTimeoutId_';

function setLastTimeoutIdByPeriod(timeoutIn=0, id=0) {
    if (timeoutIn == 0 || id == 0) return;

    var key = FAITRACKER_WINDOW_TIMEOUT_KEY_PREFIX + timeoutIn;
    Cache.getFAITrackerCacheObject()[key] = id;
}

function getLastTimeoutIdByPeriod(timeoutIn=0) {
    if (timeoutIn == 0) return;

    var key = FAITRACKER_WINDOW_TIMEOUT_KEY_PREFIX + timeoutIn;
    return Cache.getFAITrackerCacheObject()[key];
}

function clearTimeoutByPeriod(timeoutInPeriod) {
    var lastTimeoutId = getLastTimeoutIdByPeriod(timeoutInPeriod);
    if (!lastTimeoutId) return;

    clearTimeout(lastTimeoutId);
    logger.debug("Cleared timeout of "+timeoutInPeriod+"ms :"+lastTimeoutId, false);
}

function getCurrentPageSpentTimeInMs() {
    var currentPageSpentTimeInMs = Cache.getFAITrackerCache(Cache.currentPageSpentTimeInMs);
    return currentPageSpentTimeInMs ? currentPageSpentTimeInMs : 0;
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
        return Promise.reject(new Error("FAITrackerInitError: Initialized already. Use reset() and init(), if you really want to do this."));

    if (!token) return Promise.reject(new Error("FAITrackerArgumentError: Invalid token."));

    

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
        Cache.setFAITrackerCache(Cache.trackPageOnSPA, true); 
    }

    // Enable localstorage for use after checking.
    if (isLocalStorageAvailable()) window.FAITRACKER_LS_AVAILABLE = true;
    
    // Gets info using temp client with given token, if succeeds, 
    // set temp client as app client and set response as app config 
    // or else app stays unintialized.
    var payload = {};

    var _this = this; // Remove arrows;
    updatePayloadWithUserIdFromCookie(payload);
    return _client.getInfo(payload)
        .then(function(response) {
            if (response.status < 200 || response.status > 308) {
                logger.errorLine("Get project settings failed with code : ", response.status); 
                return Promise.reject(new Error("FAITrackerRequestError: Init failed. App configuration failed."));
            }
            return response;
        })
        .then(function(response) {
            // Initialisation.
            _this.config = response.body;
            _this.client = _client;

            // Check if client has given cookie access and process queue, else keep checking
            checkCookiesConsentAndProcess(_this, response, trackOnInit);

            // Process localstorage backlog requests.
            processAllLocalStorageBacklogRequests();

            return response;
        })
        
}

function checkCookiesConsentAndProcess(_this, response, trackOnInit) {
    // Add user_id from response to cookie.
    updateCookieIfUserIdInResponse(response);
    
    if(!doesUserIdCookieExist()) {
        logger.debug("Checking for cookie consent.", false);
        setTimeout(() => {checkCookiesConsentAndProcess(_this, response, trackOnInit)}, 1000)
    } else {
        logger.debug("Cookie consent is enabled. Continuing process", false);
        
        // Start queue processing.
        triggerQueueInitialisedEvent();
        runPostInitProcess(_this, trackOnInit);
    }
}

function runPostInitProcess(_this, trackOnInit) {
    (function(){
        return Promise.resolve();
    })().then(function() {
        logger.debug("Auto Track call starts", false);
        // Enable auto-track SPA page based on settings or init option.
        var enableTrackSPA = Cache.getFAITrackerCache(Cache.trackPageOnSPA) || _this.getConfig("auto_track_spa_page_view");
        Cache.setFAITrackerCache(Cache.trackPageOnSPA, enableTrackSPA);
        // Auto-track current page on init, if not disabled.
        return trackOnInit ? _this.autoTrack(_this.getConfig("auto_track"), 
            false, afterPageTrackCallback, true) : null;
    })
    .then(function() {
        return _this.autoFormCapture(_this.getConfig("auto_form_capture"));
    })
    .then(function() {
        return _this.autoCaptureFormFills(_this.getConfig("auto_capture_form_fills"));
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
        logger.debug(err);
        return Promise.resolve(err.stack + " during get_settings on init.");
    });
}

function getEventProperties(eventProperties={}) {
    // Other property validations done on backend.
    eventProperties = Properties.getTypeValidated(eventProperties);

    var referrer = document.referrer;
    // Use page event name on cache as referrer for SPA auto tracking.
    var currentPageOriginalURL = Cache.getFAITrackerCache(Cache.currentPageOriginalURL)
    if (Cache.getFAITrackerCache(Cache.trackPageOnSPA) && currentPageOriginalURL) 
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
        return Promise.reject(new Error("FAITrackerError: Invalid event name."));

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
        if (pageLoadTime > 0) payload.event_properties[Properties.PAGE_LOAD_TIME] = pageLoadTime;
    }
    return this.client.track(payload)
        .then(function(response) {
            if (response && response.body) {
                if (!response.body.event_id) {
                    logger.debug("No event_id found on track response.", false);
                    return response;
                }

                if (auto) setCurrentPageAttributesToStore(response.body.event_id, eventName, originalPageURL);
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
        logger.debug("No properties given to update event.", false);
        
    var payload = { event_id: eventId, properties: properties };
    
    return this.client.updateEventProperties(updatePayloadWithUserIdFromCookie(payload));
}

App.prototype.updatePagePropertiesIfChanged = function(defaultPageSpentTimeInMs=0) {

    let lastPageProperties = Cache.getFAITrackerCache(Cache.lastPageProperties);

    let lastPageSpentTimeInMs = lastPageProperties && lastPageProperties[Properties.PAGE_SPENT_TIME] ? 
        lastPageProperties[Properties.PAGE_SPENT_TIME] : 0;

    let lastPageScrollPercentage = lastPageProperties && lastPageProperties[Properties.PAGE_SCROLL_PERCENT] ?
        lastPageProperties[Properties.PAGE_SCROLL_PERCENT] : 0;

    var pageSpentTimeInMs = getCurrentPageSpentTimeInMs();
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
        var eventId = Cache.getFAITrackerCache(Cache.currentPageTrackEventId);
        var payload = { event_id: eventId, properties: properties };
        this.client.updateEventProperties(updatePayloadWithUserIdFromCookie(payload)).catch(logger.debug);
    } else {
        logger.debug("No change on page properties, skipping update for : " + JSON.stringify(lastPageProperties), false);
    }

    Cache.setFAITrackerCache(Cache.lastPageProperties, {
        [Properties.PAGE_SCROLL_PERCENT]: pageScrollPercentage, 
        [Properties.PAGE_SPENT_TIME]: pageSpentTimeInMs
    });
}

function isPageAutoTracked() {
    var pageEventId = Cache.getFAITrackerCache(Cache.currentPageTrackEventId); 
    if (pageEventId && pageEventId != undefined) {
        return getAutoTrackURL() == Cache.getFAITrackerCache(Cache.currentPageURLEventName);
    }

    return false
}

App.prototype.autoTrack = function(enabled=false, force=false, afterCallback, initFAITrackerQueue=false) {
    if (!enabled) return false;

    if (!force && isPageAutoTracked()) {
        logger.debug('Page tracked already as per store : '+JSON.stringify(Cache.getFAITrackerCacheObject()))
        return false;
    }

    if(!doesUserIdCookieExist()) return false;
    
    var _this = this;

    this.track(getAutoTrackURL(), Properties.getFromQueryParams(window.location), true, afterCallback);

    // Todo: Use curried function to remove multiple set timeouts.
    // update page properties after 5s and 10s with default value.
    var fiveSecondsInMs = 5000;
    clearTimeoutByPeriod(fiveSecondsInMs);
    var timoutId5thSecond = setTimeout(function() {
        logger.debug("Triggered properties update after 5s.", false);
        _this.updatePagePropertiesIfChanged(fiveSecondsInMs);
    }, fiveSecondsInMs);
    setLastTimeoutIdByPeriod(fiveSecondsInMs, timoutId5thSecond);

    var tenSecondsInMs = 10000;
    clearTimeoutByPeriod(tenSecondsInMs);
    var timoutId10thSecond = setTimeout(function() {
        logger.debug("Triggered properties update after 10s.", false);
        _this.updatePagePropertiesIfChanged(tenSecondsInMs);
    }, tenSecondsInMs);
    setLastTimeoutIdByPeriod(tenSecondsInMs, timoutId10thSecond);

    // clear the previous poller, if exist.
    var lastPollerId = Cache.getFAITrackerCache(Cache.lastPollerId);
    clearInterval(lastPollerId);
    if (lastPollerId) logger.debug("Cleared previous page poller: "+lastPollerId, false);

    // update page properties every 20s.
    var pollerId = setInterval(function() {
        _this.updatePagePropertiesIfChanged();
    }, 20000);
    
    Cache.setFAITrackerCache(Cache.lastPollerId, pollerId);

    // update page properties before leaving the page.
    window.addEventListener("beforeunload", function() {
        _this.updatePagePropertiesIfChanged();
        return;
    });

    window.addEventListener("scroll", setLastActivityTime);
    window.addEventListener("mouseover", setLastActivityTime);
    window.addEventListener("mousemove", setLastActivityTime);

    // Todo(Dinesh): Find ways to automate tests for SPA support.
    
    // AutoTrack SPA using history.
    // checks support for history and onpopstate listener.
    if ( !Cache.getFAITrackerCache(Cache.trackPageOnSPA) && window.history && window.onpopstate !== undefined) {
        var prevLocation = window.location.href;
        window.addEventListener('popstate', function() {
            logger.debug("Triggered window.onpopstate goto: "+window.location.href+", prev: "+prevLocation);
            if (prevLocation !== window.location.href) {
                // should be called before next page track.
                _this.updatePagePropertiesIfChanged();
                
                _this.track(getAutoTrackURL(), Properties.getFromQueryParams(window.location), true, afterCallback);
                prevLocation = window.location.href; 
            }
        })
    }

    if (Cache.getFAITrackerCache(Cache.trackPageOnSPA)) {
        setInterval(function(){
            if (Cache.getFAITrackerCache(Cache.currentPageURLEventName) && Cache.getFAITrackerCache(Cache.currentPageURLEventName) != getAutoTrackURL()) {
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
    
    var _this = this;
    // bind immediately.
    bindForFormCapture(_this);
    // starts the binder for lazy loaded forms.
    Capture.startBackgroundFormBinder();
    // binds on event triggers.
    document.addEventListener(Capture.TRIGGER_FORM_BINDING_EVENT, function(e) {
        bindForFormCapture(_this);
    });

    return true;
}

function bindForFormCapture(appInstance) {
    logger.debug("Binding for form capture", false);

    // Captures properties from ideal forms which has a submit button. 
    // The fields sumbmitted are processed on callback onSubmit of form.
    Capture.bindAllFormsOnSubmit(appInstance, appInstance.captureAndTrackFormSubmit);

    // Captures properties from input fields, which are not part of any form
    // on click of any button on the page, which is not a submit button of any form.
    // Note: submit button which is not inside a form is also bound.
    Capture.bindAllNonFormButtonOnClick(appInstance, appInstance.captureAndTrackNonFormInput);
}

function captureInputFieldValues(appInstance, input) {
    var newValue = input.value.trim();

    // Below conditions would limit the no.of entries captured.
    if (newValue.length < 4) return; 
    var isCapturable = Capture.isPossibleEmail(newValue);
    if (!isCapturable) return
    
    var inputId = input.getAttribute(FAITRACKER_INPUT_ID_ATTRIBUTE);
    if (window.FAITRACKER_FORM_FILLS == undefined) 
        window.FAITRACKER_FORM_FILLS = {};
    
    var existingValue = window.FAITRACKER_FORM_FILLS[inputId];
    if (newValue != "" && newValue != existingValue) {
        window.FAITRACKER_FORM_FILLS[inputId] = newValue;

        var formId = inputId.split(".")[0];
        var payload = {
            "form_id": formId,
            "field_id": inputId,
            "value": newValue,
        }
        payload.event_properties = getEventProperties();
        updatePayloadWithUserIdFromCookie(payload);
        logger.debug(payload, false);
                
        return appInstance.client.captureFormFill(payload);
    }
}

function captureAllInputFieldsInput(appInstance) {
    var inputs = document.querySelectorAll("input");
    for(var i=0; i<inputs.length; i++) {
        if(properties.DISABLED_INPUT_TYPES.indexOf(inputs[i].type) >= 0) continue;
        captureInputFieldValues(appInstance, inputs[i]);
    }
}

App.prototype.autoCaptureFormFills = function(enabled) {
    if (!enabled) return false;

    bindAndCaptureFormFills(this);

    // starts the binder for lazy loaded  
    // forms, if not started already.
    Capture.startBackgroundFormBinder();

    var _appInstance = this;
    document.addEventListener(Capture.TRIGGER_FORM_BINDING_EVENT, function() {
        bindAndCaptureFormFills(_appInstance);
    });
}

function bindAndCaptureFormFills(appInstance) {
    logger.debug("Binding for form fills capture", false);

    // Assigning incremental id to the form.
    var forms = Capture.getElemsFromTopAndIframes('form');
    const FAITRACKER_FORM_ID_ATTRIBUTE = "data-faitracker-form-id";
    if (!window.FAITRACKER_FORMS_ID) window.FAITRACKER_FORMS_ID = 0;
 
    for (var fi=0; fi<forms.length; fi++) {
        if (forms[fi].getAttribute(FAITRACKER_FORM_ID_ATTRIBUTE)) continue;
        forms[fi].setAttribute(FAITRACKER_FORM_ID_ATTRIBUTE, "form-"+window.FAITRACKER_FORMS_ID);
        window.FAITRACKER_FORMS_ID++;
    }

    // Assigns non-form fields with noform.field-1.
    // Assign form input fields with form-1.field-1.
    var inputs = Capture.getElemsFromTopAndIframes('input');
    if (!window.FAITRACKER_INPUTS_ID) window.FAITRACKER_INPUTS_ID = 0;

    for(var i=0; i<inputs.length; i++) {
        if (inputs[i].getAttribute(FAITRACKER_INPUT_ID_ATTRIBUTE)) continue;
        if(properties.DISABLED_INPUT_TYPES.indexOf(inputs[i].type) >= 0) continue;

        var formId = "noform";
        var hasForm = inputs[i].form && 
            inputs[i].form.getAttribute(FAITRACKER_FORM_ID_ATTRIBUTE) != "";
        
        if (hasForm) formId = inputs[i].form.getAttribute(FAITRACKER_FORM_ID_ATTRIBUTE);

        inputs[i].setAttribute(FAITRACKER_INPUT_ID_ATTRIBUTE, formId+".field-"+window.FAITRACKER_INPUTS_ID);

        // Captures values while typing.
        inputs[i].addEventListener("input", function() { 
            var _input = this;
            captureInputFieldValues(appInstance, _input); 
        });

        // Captures values when the cursor is out of focus and the value is changed.
        inputs[i].addEventListener("change", function() { 
            var _input = this;
            captureInputFieldValues(appInstance, _input); 
        });

        window.FAITRACKER_INPUTS_ID++;
    }

    // Create timeouts for every 5s for 30s.
    var t = 5000;
    while (t < 30000) {
        setTimeout(function() { captureAllInputFieldsInput(appInstance); }, t);
        t = t + 5000;
    }
}

App.prototype.captureClick = function(appInstance, element) {
    if (!appInstance.isInitialized()) return Promise.reject(SDK_NOT_INIT_ERROR);

    if (!element) logger.debug("Element is undefined on capture click.");

    if (!doesUserIdCookieExist()) return Promise.reject(SDK_NO_USER_ERROR);

    var payload = Capture.getClickCapturePayloadFromElement(element);

    // Add event and user_properties for tracking after enabling.
    payload.event_properties = getEventProperties();
    payload.user_properties = Properties.getUserDefault();

    // Add user_id to payload.
    updatePayloadWithUserIdFromCookie(payload);

    logger.debug(element, false);
    logger.debug("Click capture payload: "+JSON.stringify(payload), false);

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
        return Promise.reject(new Error("FAITrackerArgumentError: Properties should be an Object(key/values)."));
    
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
        logger.errorLine(new Error("FAITrackerConfigError: Config not present."));
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

    logger.debug(error);

    return Promise ? Promise.resolve(error) : error;
}

module.exports = exports = App;
