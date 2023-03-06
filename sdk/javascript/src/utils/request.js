var logger = require("./logger");
const config = require("../config");

const LOCALSTORAGE_PREFIX = "_faireq_";
const URI_SERVICE_ERROR = "/sdk/service/error";

// window.FACTORS_LS_AVAILABLE is set as true after 
// checking accessibility of localstorage during initialisation.
function useLocalStorage(url) {
    var allowedAPI = url.indexOf("/sdk/event/track") > 0 || 
        url.indexOf("/sdk/user/identify") > 0 ||
        url.indexOf("/sdk/capture_click") > 0;
    return  !!window.FACTORS_LS_AVAILABLE && allowedAPI;
}

function getRandomUID() {
    return String(Date.now()) + Math.random();
}

function getLocalStorageKey(rid) {
    return LOCALSTORAGE_PREFIX+rid;
}

function setLocalStorage(rid, request) {
    window.localStorage.setItem(getLocalStorageKey(rid), JSON.stringify(request));
}

function removeLocalStorage(rid, withPrefix=true) {
    var key = withPrefix ? getLocalStorageKey(rid) : rid;
    logger.debug("Removed req from LS: "+key, false);

    // Ways to disable LS removal and post-processing on next load.
    if (window.FACTORS_DISABLE_LS_REMOVE) return;

    window.localStorage.removeItem(key);
}

// Filters factors request keys from localstorage.
function getAllRequestKeysFromLS() {
    var allKeys = Object.keys(window.localStorage);
    var reqKeys = [];

    for(var i=0; i<allKeys.length; i++) 
        if (allKeys[i].indexOf(LOCALSTORAGE_PREFIX) == 0) 
            reqKeys.push(allKeys[i]);
    
    return reqKeys;
}

// Filters all factors request key and value from localstorage.
function getAllRequestsFromLS() {
    var keys = getAllRequestKeysFromLS();
    var requests = {};

    for(var i=0; i<keys.length; i++)
        requests[keys[i]] = window.localStorage[keys[i]];

    return requests;
}

function isFetchAbortedError(e) {
    if (!e) return false;
    return e.toString().indexOf('abort') > -1 || e.toString().indexOf('cancel') > -1;
}

function processAllLocalStorageBacklogRequests() {
    if (!window.FACTORS_LS_AVAILABLE) return;

    var requests = getAllRequestsFromLS();
    var reqKeys = Object.keys(requests);

    if (reqKeys.length > 0) {
        // Not an error. This is to measure the usage of LS based backlog processing.
        sendErrorWithInfo("[STATUS] Processing localstorage.")
    }

    for (var i=0; i<reqKeys.length; i++) {
        var key = reqKeys[i];

        var payload = null;
        try { payload = JSON.parse(requests[key]); } 
        catch(e){ sendErrorWithInfo("Error parsing json from LS: "+requests[key]); }

        // Remove from LS to avoid loop.
        removeLocalStorage(key, false);

        if (payload == null) continue;
        // Skip payload with age greater than 30 mins.
        if (payload.timestamp > 0) {
            var currentTimestamp = (new Date()).getTime();
            var age = (currentTimestamp - payload.timestamp);
            if (age > 1800000) {
                logger.debug("Payload older than 30 minutes skipped.", false);
                continue;
            }
        }

        var errString = "Request failed on LS backlog processing:";

        fetch(payload.url, payload.options)
            .then(function(response) {
                var _response = response;
                _response.json()
                    .then(function() {
                        if (!_response.ok) 
                            sendErrorWithInfo(errString+" "+payload.url+" status "+_response.status);
                    })
            })
            .catch(function(e) {
                sendErrorWithInfo(errString+" "+payload.url+" with error "+JSON.stringify(e));
            });
    }

    return;
}

function sendErrorAPI(payload) {
    // Mandaotry field check.
    if (!payload || !payload.domain || !payload.error) 
        return Promise.reject("Sending error failed. Invalid payload.");

    return post(
        config.api.host+URI_SERVICE_ERROR,
        payload
    );
}

function sendErrorWithInfo(error) {
    var errMsg = "";
    if (typeof(error) == "string") errMsg = error;
    if (error instanceof Error && error.message) errMsg = error.stack;

    let payload = {};
    payload.domain = window.location.host;
    payload.url = window.location.href;
    payload.error = errMsg;

    sendErrorAPI(payload);
}

function requestWithLocalStorage(method, url, headers, data) {
    let options = { method: method, keepalive: true };

    if(data && data != undefined) 
        options["body"] = JSON.stringify(data);

    if(headers && headers != undefined ) {
        options.headers = headers;

        // Default headers.
        options.headers["Content-Type"] = "application/json";
    }

    // Allow only 10 keys into localstorage at max.
    var keysWithinLimit = getAllRequestKeysFromLS().length <= 10;
    var shouldUseLocalStorage = useLocalStorage(url) && keysWithinLimit;

    // Create and set request on localstorage.
    var uid = getRandomUID();
    var currentTimestamp = (new Date()).getTime();
    if (shouldUseLocalStorage) setLocalStorage(uid, {url: url, options: options, timestamp: currentTimestamp});

    return fetch(url, options)
        .then(function(response) {
            // Remove from localstorage once processed successfully.
            if (shouldUseLocalStorage) removeLocalStorage(uid);

            var _response = response;
            return  _response.json()
                .then(function(responseJSON) {
                    if (!_response.ok) {
                        logger.debug(_response);
                        return Promise.resolve(_response);
                    }
                    return { status: _response.status, body: responseJSON };
                })
        })
        .catch(function(e){
            // Remove from local storage, if not failed with cancelled error. 
            // else keep it on localstorage for next time processing.
            if (shouldUseLocalStorage && !isFetchAbortedError(e)) removeLocalStorage(uid);
            
            logger.debug(e);
            // Forward error to next catch.
            return Promise.resolve(e);
        });
}

function get(url, headers={}) { return requestWithLocalStorage("get", url, headers); }

function post(url, data, headers={}) { return requestWithLocalStorage("post", url, headers, data); }

module.exports = exports = { get, post, processAllLocalStorageBacklogRequests, sendErrorAPI };