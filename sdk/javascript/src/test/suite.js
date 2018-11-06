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
        { properties: { mobile: true } }
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
    return setupNewProject()
        .then((r) => {
            factors.reset(); // Clear existing environment.
            factors.init(r.body.token, {});
            assert.equal(factors.app.client.token, r.body.token, "App initialization failed.");
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

// Test: Init

Suite.testInit = function() {
    setupNewProject()
        .then((r) => {
            factors.reset();
            assert.isTrue(r.body.hasOwnProperty("token"), "Token should be in the response.");
            assert.isTrue(r.body.token.trim().length > 0, "Token should not be empty.")
            
            factors.reset();
            factors.init(r.body.token, { mobile: true });
            assert.isTrue(factors.app.client.token === r.body.token, "Token should be set as api client token for the app.");

            // init without props.
            factors.reset();
            factors.init(r.body.token);
            assert.isTrue(factors.app.client.token === r.body.token, "Should be able to init without properties");

            // init with empty props.
            factors.reset();
            factors.init(r.body.token, {});
            assert.isTrue(factors.app.client.token === r.body.token, "Should be able to init with empty properties");
        })
        .catch(assertOnCall);
}

Suite.testInitWithBadInput = function() {
    factors.reset();

    // Bad input. Invalidated on sdk.
    assert.throws(() => factors.init(" "), Error, "FactorsError: Initialization failed. Invalid Token.");
    assert.equal(factors.app.client.token, null, "Bad input token should not be allowed.");
}

// Test: Track

Suite.testTrackBeforeInit = function() {
    factors.reset();
    // Should throw exception.
    assert.throws(factors.track, Error, "FactorsArgumentError: Invalid type for event_name");
}

Suite.testTrackAfterInit = function() {
    return setupNewProjectAndInit()
        .then((r) => {
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

    // Bad input. Invalidated on backend.
    let eventName = randomAlphaNumeric(10);
    factors.init("BAD_TOKEN", {});
    assert.equal(factors.app.client.token, "BAD_TOKEN", "App initialization failed.");
    return factors.track(eventName)
        .then(assertIfHttpSuccess)
        .catch(suppressExpectedError);
}

Suite.testTrackWithoutEventName = function() {
    factors.reset();

    // Fail if no eventName.
    assert.throws(factors.track, Error, "FactorsArgumentError: Invalid type for event_name");
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
Suite.testTrackWithoutUserCookie = function() {
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
        })
        .catch(assertOnCall);
}

// Track as existing user. Track with user cookie.
Suite.testTrackWithUserCookie = function() {
    setupNewProjectWithUser()
        .then((r) => {
            factors.reset(); // Clears existing env.
            factors.init(r.project.body.token, {});
            Cookie.set(constant.cookie.USER_ID, r.user.body.id);

            factors.track(randomAlphaNumeric(10), {})
                .then(assertIfHttpFailure)
                .then(assertIfUserIdOnResponse) // user_id shouldn't be there on response.
        })
        .catch(assertOnCall);
}

// Test: identify

Suite.testIdentifyBeforeInit = function() {
    factors.reset();

    // Throws error, if not initialized.
    assert.throws(factors.identify, Error, "FactorsArgumentError: Invalid type for customer_user_id");
}

Suite.testIdentifyAfterInit = function() {
    return setupNewProjectAndInit()
        .then((r) => {
            let customerUserId = randomAlphaNumeric(15);
            return factors.identify(customerUserId)
                .then(assertIfHttpFailure)
                .then(assertOnUserIdMapFailure) // It has to set cookie.
        })
        .catch(assertOnCall);
}

Suite.testIdentifyWithBadToken = function() {
    factors.reset();

    // Bad input. Invalidated on backend.
    let customerUserId = randomAlphaNumeric(10);
    factors.init("BAD_TOKEN", {});
    return factors.identify(customerUserId)
        .then(assertIfHttpSuccess)
        .catch(suppressExpectedError);
}

Suite.testIdentifyWithoutCustomerUserId = function() {
    factors.reset();

    return setupNewProjectAndInit()
        .then((r) => {
            assert.throws(factors.identify, Error, "FactorsArgumentError: Invalid type for customer_user_id");
            assert.throws(() => factors.identify(" "), Error, "FactorsArgumentError: customer_user_id cannot be empty.");
        })
        .catch(assertOnCall);
}

// New user => Without user cookie.
Suite.testIdentifyWithoutUserCookie = function() {
    factors.reset();

    return setupNewProjectAndInit()
        .then((r) => {
            let customerUserId = randomAlphaNumeric(15);
            return factors.identify(customerUserId)
                .then(assertIfHttpFailure)
                .then(assertOnUserIdMapFailure);
        })
        .catch(assertOnCall);
}

// With user cookie => Identify existing unidentified user.
Suite.testIdentifyWithUserCookie = function() {
    return setupNewProjectWithUser()
        .then((r) => {
            factors.reset(); // Clear existing env.
            factors.init(r.project.body.token, {});
            assert.equal(factors.app.client.token, r.project.body.token, "App initialization failed.");

            Cookie.set(constant.cookie.USER_ID, r.user.body.id); // Setting new user.

            let customerUserId = randomAlphaNumeric(15);
            return factors.identify(customerUserId)
                .then(assertIfHttpFailure)
                .then(assertIfUserIdOnResponse) // Should not respond new user.
        })
        .catch(assertOnCall);
}

Suite.testIdentifyWithIdentifiedCustomerUserWithSameUserCookie = function() {
    return setupNewProjectAndInit()
        .then((r) => {
            let customerUserId = randomAlphaNumeric(15);
            let _token = factors.app.client.token; // Fix: Project context copy assign.
            // Identified as customer user and user cookie set here.
            return factors.identify(customerUserId)
                .then(assertIfHttpFailure)
                .then(assertOnUserIdMapFailure) 
                .then((r1) => {
                    factors.init(_token); // Fix: Project context copy assign.
                    // Same user cookie used to identify with same customer user id.
                    // No user_id changes required. user_id should not exist on response.
                    return factors.identify(customerUserId)
                        .then(assertIfHttpFailure)
                        .then(assertIfUserIdOnResponse); // Should not response new user as identified with same user already.
                });
        })
        .catch(assertOnCall);
}

Suite.testIdentifyWithIdentifiedCustomerUserWithDifferentUserCookie = function() {
    return setupNewProjectWithUser()
        .then((r1) => {
            factors.reset();
            factors.init(r1.project.body.token, {});
            assert.equal(factors.app.client.token, r1.project.body.token, "App initialization failed.");

            Cookie.set(constant.cookie.USER_ID, r1.user.body.id);
            let customerUserId = randomAlphaNumeric(15);
            let _token = factors.app.client.token; // Fix: Project context copy assign.

            return factors.identify(customerUserId)
                .then(assertIfHttpFailure)
                .then(assertIfUserIdOnResponse) 
                .then((r2) => {
                    let _customerUserId = customerUserId;

                    return setupNewUser(r1.project.body.id)
                        .then((rUser) => {
                            // Setting new user cookie.
                            Cookie.set(constant.cookie.USER_ID, rUser.body.id);
                            factors.init(_token);
                            
                            // Re-identify with same customer_user and diff user.
                            // The new user_id will be identified as customer_user.
                            // No changes on user_id required.
                            factors.identify(_customerUserId)
                                .then(assertIfHttpFailure)
                                .then(assertIfUserIdOnResponse) 
                        })
                })
        })
        .catch(assertOnCall);
}

// Identify as an identified user without user cookie should respond
// with latest user of the customer_user. Reusing same user session.
Suite.testIdentifyWithIdentifiedCustomerUserWithoutUserCookie = function() {
    return setupNewProjectWithUser()
        .then((r1) => {
            factors.reset();
            factors.init(r1.project.body.token, {});
            assert.equal(factors.app.client.token, r1.project.body.token, "App initialization failed.");

            Cookie.set(constant.cookie.USER_ID, r1.user.body.id);

            let customerUserId = randomAlphaNumeric(15);
            return factors.identify(customerUserId)
                .then(assertIfHttpFailure)
                .then(assertIfUserIdOnResponse) // User cookie already set.
                .then(() => {
                    Cookie.remove(constant.cookie.USER_ID); // Remove cookie.
                    factors.init(r1.project.body.token); // Set same project context.

                    return factors.identify(customerUserId) // Re-identify as same customer.
                        .then(assertIfHttpFailure)
                        .then(assertOnUserIdMapFailure)
                        .then((r2) => {
                            // Check latest user_id of customer returned or not.
                            assert.equal(r2.body.user_id, r1.user.body.id);
                            return r2;
                        });
                });
        })
        .catch(assertOnCall);
}

// Test: addUserProperties

Suite.testAddUserPropertiesWithEmptyProperties = function() {
    factors.reset();

    return factors.addUserProperties({})
        .then(assertOnCall) // Should not be resolved.
        .catch((m) => {
            // To be rejected with message.
            assert.equal(m, "No changes. Empty properties.");
        });
}

Suite.testAddUserPropertiesBadProperties = function() {
    factors.reset();

    // Properties argument type mismatch. Using string as properties.
    assert.throws(() => factors.addUserProperties("STRING_INPUT"), "FactorsArgumentError: Properties should be an Object(key/values).")
}

Suite.testAddUserPropertiesWithoutUserCookie = function() {
    return setupNewProjectAndInit()
        .then((r) => {
            let properties = { userHandle: randomAlphaNumeric(15) };
            factors.addUserProperties(properties)
                .then(assertIfHttpFailure)
                .then(assertOnUserIdMapFailure) // UserId on response expected.
        })
        .catch(assertOnCall)
}

Suite.testAddUserPropertiesWithUserCookie = function() {
    return setupNewProjectWithUser()
        .then((r) => {
            factors.reset();
            factors.init(r.project.body.token, {});
            assert.equal(factors.app.client.token, r.project.body.token, "App initialization failed.");
            
            Cookie.set(constant.cookie.USER_ID, r.user.body.id);

            let properties = { userHandle: randomAlphaNumeric(15) };
            factors.addUserProperties(properties)
                .then(assertIfHttpFailure)
                .then(assertOnUserIdMapFailure) // User should not be created.
        })
        .catch(assertOnCall)
}

function run() {
    // Runs individual test in the test_suite.
    for (let test in Suite) {
        console.log('%c Running '+test+'..', 'color: green');
        Suite[test]();
    }
    return true;
}

module.exports = exports = { Suite, run };

