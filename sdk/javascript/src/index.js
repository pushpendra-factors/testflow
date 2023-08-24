"use strict";

const logger = require("./utils/logger");
var App = require("./app");
// const { EV_FORM_SUBMITTED } = require("./properties");
const Properties = require("./properties");

// Global reference.
var app = new App();
var isQueueBeingProcessed = false;

/**
 * Initializes sdk environment on user application. Overwrites if initialized already.
 * @param {string} appToken Unique application token.
 * @param {object} opts Additional opts: {track_on_init: false}
 * @param {function(eventId)} afterPageTrackCallback Callback called after tracking the page, with eventId.
 */
function init(appToken, opts={}, afterPageTrackCallback) {
    try {
        return app.init(appToken, opts, afterPageTrackCallback)
        .catch(app.handleError);
    } catch(e) {
        return app.handleError(e);
    }
}

/**
 * Clears existing SDK environment, both API token and cookies. 
 */
function reset() {
    try {
        app.reset();
    } catch(e) {
        app.handleError(e);
    }

    return;
}

/**
 * Track events on user application.
 * @param {string} eventName
 * @param {Object} eventProperties 
 * @param {function(eventId)} afterCallback
 */
function track(eventName, eventProperties={}, afterCallback) {
    try {
        app.track(eventName, eventProperties, false, afterCallback)
            .catch(app.handleError);
    } catch(e) {
        app.handleError(e);
    }

    return;
}

/**
 * Track visit to page as event.
 * @param {function(eventId)} afterCallback 
 * @param {boolean} force Force track page, even if tracked already.
 */
function page(afterCallback, force=false) {
    try {
        app.page(afterCallback, force)
            .catch(app.handleError);
    } catch(e) {
        app.handleError(e);
    }

    return;
}

/**
 * Update properties of given event.
 * @param {string} eventId 
 * @param {Object} properties
 */
function updateEventProperties(eventId, properties={}) {
    try {
        app.updateEventProperties(eventId, properties)
            .catch(app.handleError);
    } catch(e) {
        app.handleError(e);
    }

    return;
}

/**
 * Identify user with original userId from the application.
 * @param {string} customerUserId Actual id of the user from the application.
 * @param {Object} userProperties Optional - Traits of the users.
 */
function identify(customerUserId, userProperties={}) {
    try {
        app.identify(customerUserId, userProperties)
            .catch(app.handleError);
    } catch(e) {
        app.handleError(e);
    }

    return;
}

/**
 * Add additional user properties.
 * @param {Object} properties 
 */
function addUserProperties(properties={}) {
    try {
        app.addUserProperties(properties)
            .catch(app.handleError);
    } catch(e) {
        app.handleError(e);
    }

    return;
}

function getUserId() {
    return app.getUserId();
}


function primaryWindowVar() {
    if (window.faitracker) return window.faitracker;
    // For backward compatibility.
    if (window.factors) return window.factors;
}

function processQueue() {
    if(primaryWindowVar() && primaryWindowVar().q && primaryWindowVar().q.length > 0 && !isQueueBeingProcessed) {
        isQueueBeingProcessed = true;
        logger.debug("Starting Queue", false);
        try{
            while(primaryWindowVar().q.length > 0) {
                logger.debug("Processing Queue", false);
                // q[0] indicates first item of the queue;
                // a indicates list of arguments;
                switch(primaryWindowVar().q[0].k) {
                    case 'message': {
                        if(primaryWindowVar().q[0].a[0] === Properties.EV_FORM_SUBMITTED) {
                            var properties = primaryWindowVar().q[0].a[1];
                            if (!Properties.hasEmailOrPhone(properties)) {
                                logger.debug("No email and phone, skipping form submit.", false);
                                primaryWindowVar().q.shift();
                                break;
                            }
                            track(primaryWindowVar().q[0].a[0], primaryWindowVar().q[0].a[1]);
                            primaryWindowVar().q.shift();
                        }
                        break;
                    }
                    case 'track': {
                        track(primaryWindowVar().q[0].a[0], primaryWindowVar().q[0].a[1], primaryWindowVar().q[0].a[2]);
                        primaryWindowVar().q.shift();
                        break;
                    }
                    case 'reset': {
                        reset();
                        primaryWindowVar().q.shift();
                        break;
                    }
                    case 'page': {
                        page(primaryWindowVar().q[0].a[0], primaryWindowVar().q[0].a[1]);
                        primaryWindowVar().q.shift();
                        break;
                    }
                    case 'updateEventProperties': {
                        updateEventProperties(primaryWindowVar().q[0].a[0], primaryWindowVar().q[0].a[1]);
                        primaryWindowVar().q.shift();
                        break;
                    }
                    case 'identify': {
                        identify(primaryWindowVar().q[0].a[0], primaryWindowVar().q[0].a[1]);
                        primaryWindowVar().q.shift();
                        break;
                    }
                    case 'addUserProperties': {
                        addUserProperties(primaryWindowVar().q[0].a[0]);
                        primaryWindowVar().q.shift();
                        break;
                    }
                    case 'getUserId': {
                        getUserId();
                        primaryWindowVar().q.shift();
                        break;
                    }
                    default:
                        app.handleError("Unknown call parameters");
                }
            }
        } 
        catch (e) {
            app.handleError(e);
            isQueueBeingProcessed = false;
        }
        logger.debug("Queue Processed", false);
        isQueueBeingProcessed = false;
    }
}

let exposed = { init, reset, track, page, updateEventProperties, 
    identify, addUserProperties, getUserId };

if (process.env.NODE_ENV === "development") {
    exposed["test"] = require("./test/suite.js");
    exposed["app"] = app;
}

if(primaryWindowVar() && primaryWindowVar().TOKEN) {
    init(primaryWindowVar().TOKEN, primaryWindowVar().INIT_PARAMS, primaryWindowVar().INIT_CALLBACK);
}

window.addEventListener('FAITRACKER_INIT_EVENT', function(e) {
    init(primaryWindowVar().TOKEN, primaryWindowVar().INIT_PARAMS, primaryWindowVar().INIT_CALLBACK);
});

// For backward compatibility.
window.addEventListener('FACTORS_INIT_EVENT', function(e) {
    init(primaryWindowVar().TOKEN, primaryWindowVar().INIT_PARAMS, primaryWindowVar().INIT_CALLBACK);
});

window.addEventListener('FAITRACKER_INITIALISED_EVENT', function(e){
    processQueue();
    window.addEventListener('FAITRACKER_QUEUED_EVENT', function(e) {
        processQueue();
    });
});

// For backward compatibility.
window.addEventListener('FACTORS_INITIALISED_EVENT', function(e){
    processQueue();
    window.addEventListener('FACTORS_QUEUED_EVENT', function(e) {
        processQueue();
    });
});

module.exports = exports = exposed;

