"use strict";

const logger = require("./utils/logger");
var App = require("./app");

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

function processQueue() {
    if(factors && factors.q && factors.q.length > 0 && !isQueueBeingProcessed) {
        isQueueBeingProcessed = true;
        logger.debug("Starting Queue", false);
        try{
            while(factors.q.length > 0) {
                logger.debug("Processing Queue", false);
                switch(factors.q[0].k) {
                    case 'track': {
                        track(factors.q[0].a[0], factors.q[0].a[1]);
                        factors.q.shift();
                        break;
                    }
                    case 'reset': {
                        reset();
                        factors.q.shift();
                        break;
                    }
                    case 'page': {
                        page(factors.q[0].a[0], factors.q[0].a[1]);
                        factors.q.shift();
                        break;
                    }
                    case 'updateEventProperties': {
                        updateEventProperties(factors.q[0].a[0], factors.q[0].a[1]);
                        factors.q.shift();
                        break;
                    }
                    case 'identify': {
                        identify(factors.q[0].a[0], factors.q[0].a[1]);
                        factors.q.shift();
                        break;
                    }
                    case 'addUserProperties': {
                        addUserProperties(factors.q[0].a[0]);
                        factors.q.shift();
                        break;
                    }
                    case 'getUserId': {
                        getUserId();
                        factors.q.shift();
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

if(factors && factors.TOKEN) {
    init(factors.TOKEN, factors.INIT_PARAMS, INIT_CALLBACK);
}

window.addEventListener('FACTORS_INIT_EVENT', function(e) {
    init(factors.TOKEN, factors.INIT_PARAMS, INIT_CALLBACK);
})


window.addEventListener('FACTORS_INITIALISED_EVENT', function(e){
    processQueue();
    window.addEventListener('FACTORS_QUEUED_EVENT', function(e) {
            processQueue();
    })
});

module.exports = exports = exposed;

