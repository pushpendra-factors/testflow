"use strict";

// Required chai with relative path for enable method suggestions to work.
const chai = require("chai");

const Cookie = require("../utils/cookie");
const Request = require("../utils/request");

const config = require("../config");
const constant = require("../constant")

// Enable full stacktrace for chai.
// chai.config.includeStack = true;

// Assertion with chai.assert
const assert = chai.assert;

var Suite = {};

function randomAlphaNumeric(len) {
    var p = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
    return [...Array(len)].reduce(a=>a+p[~~(Math.random()*p.length)],'');
}

function suppressExpectedError(e) {
    console.log("Suppressed expected error:")
    console.log(e);
}

// Setup methods.

function setupNewProject() {
    return Request.post(
        config.api.host+"/projects",
        { name: randomAlphaNumeric(32) }
    )
    .then(assertIfHttpFailure);
}

function setupNewUser(projectId) {
    if (!projectId) throw new Error("Setup new user failed. Invalid project.");

    return Request.post(
        config.api.host+"/projects/"+projectId+"/users",
        { c_uid: randomAlphaNumeric(20), properties: { mobile: true } }
    );
}

function setupNewProjectWithUser() {
    return setupNewProject()
        .then((r1) => {
            return setupNewUser(r1.body.id)
                .then((r2) => {
                    return { project: r1, user: r2 };
                });
        });
}

// Sets up new project and new sdk environment.
// Do not reset once again, as it does already.
function setupNewProjectAndInit() {
    factors.reset();

    return setupNewProject()
        .then((r) => { 
            factors.init(r.body.token);
            return r;
        });
}

// Custom assert methods.

function assertOnUserIdMapFailure(r) {
    assert.isTrue(r.body.hasOwnProperty("user_id"), "user_id missing on response.");
    assert.equal(Cookie.get(constant.cookie.USER_ID), r.body.user_id, 
        constant.cookie.USER_ID+" cookie is not set with user_id on response");
    return r;
}

function assertOnCall(e) {
    assert.isFalse(e || e == undefined || e == null, "Catch call when not expected.");
    return e;
}

// Asserted if status is success. 
// Used in failure expected cases.
function assertIfHttpSuccess(r) {
    // Expected is status <= 299 as false.
    assert.isFalse(r.status <= 299, "Response should not be succeeded. Success status seen.");
    return r;
}

function assertIfHttpFailure(r) {
    assert.isTrue(r.status <= 299, "Response should be successful. Failure status seen.");
    return r;
}

function assertIfUserIdOnResponse(r) {
    assert.isFalse(r.body.hasOwnProperty("user_id"), "user_id on response when not required.");
    return r;
}

// Test: Track

Suite.testTrackBeforeInit = function() {
    factors.reset();
    // Should throw exception.
    assert.throws(factors.track, Error);
}

Suite.testTrackAfterInit = function() {
    factors.reset();

    return setupNewProject()
        .then((r) => {
            factors.init(r.body.token, { "platform": "Google Chrome v10.0" });
            assert.equal(factors.app.client.token, r.body.token, "App initialization failed.");
            
            // Validate track response.
            let eventName = randomAlphaNumeric(10);
            return factors.track(eventName, {})
                .then(assertIfHttpFailure)
                .then(assertOnUserIdMapFailure)
                .catch(assertOnCall);
        })
        .catch(assertOnCall);
}

Suite.testTrackWithBadToken = function() {
    factors.reset();

    // Bad input. Invalidated on sdk.
    assert.throws(() => factors.init(" "), Error);
    assert.equal(factors.app.client.token, null, "Bad input but is token set.")

    // Bad input. Invalidated on backend.
    let eventName = randomAlphaNumeric(10);
    factors.init("BAD_TOKEN", {});
    return factors.track(eventName)
        .then(assertIfHttpSuccess)
        .catch(suppressExpectedError);
}

Suite.testTrackWithoutEventName = function() {
    factors.reset();

    // Fail if no eventName.
    assert.throws(factors.track, Error);
}

Suite.testTrackWithoutEventProperites = function() {
    return setupNewProjectAndInit()
        .then((r) => {
            let eventName = randomAlphaNumeric(10);
            return factors.track(eventName, {})
                .then(assertIfHttpFailure)
                .catch(assertOnCall);
        })
        .catch(assertOnCall);
}

// Track as new user. Track without user cookie.
Suite.testTrackAsNewUser = function() {
    return setupNewProjectAndInit()
        .then((r) => {
            // Setup clears the cookie internally.
            // Making sure by clearing it explicitly.
            Cookie.remove(constant.cookie.USER_ID);
            
            // Should get user_id on response and set cookie.
            let eventName = randomAlphaNumeric(10);
            return factors.track(eventName, {})
                .then(assertIfHttpFailure)
                .then(assertOnUserIdMapFailure)
                .catch(assertOnCall);
        })
        .catch(assertOnCall);
}

// Track as existing user. Track with user cookie.
Suite.testTrackAsExistingUser = function() {
    factors.reset();

    setupNewProjectWithUser()
        .then((r) => {
            factors.init(r.project.body.token, {});
            Cookie.set(constant.cookie.USER_ID, r.user.body.id);

            factors.track(randomAlphaNumeric(10), {})
                .then(assertIfHttpFailure)
                .then(assertIfUserIdOnResponse) // user_id shouldn't be there on response.
                .catch(assertOnCall);
        })
        .catch(assertOnCall);
}

/**
 * 
 * Test: Identify
 * 
 * testIdentifyBeforeInit
 * testIdentifyAfterInit
 * testIdentifyWithBadToken
 * testIdentifyWithoutCustomerUserId
 * testIdentifyAsNewUser
 * testIdentifyAsExistingUser
 * testIdentifyAsIdentifiedUser
 * 
 */

function run() {
    // Runs individual test in the test_suite.
    for (let test in Suite) {
        console.log('%c Running '+test+'..', 'color: green');
        Suite[test]();
    }
    return true;
}

module.exports = exports = { Suite, run };

