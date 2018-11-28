"use strict";

// Required chai with relative path for enable method suggestions to work.
const chai = require("chai");

const Cookie = require("../utils/cookie");
const Request = require("../utils/request");

const config = require("../config");
const constant = require("../constant");

const Properties = require("../properties");

// Enable full stacktrace for chai.
// chai.config.includeStack = true;

// Assertion with chai.assert
const assert = chai.assert;


/**
 * Test Utils
 */

function randomAlphaNumeric(len) {
    var p = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
    return [...Array(len)].reduce(a=>a+p[~~(Math.random()*p.length)],'');
}

function suppressExpectedError(e) {
    if (e instanceof chai.AssertionError)
        throw new Error("Sorry, Test failed!");
    console.log("Suppressed Expected Error: "+ e);
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
            var _r = r;
            return factors.init(r.body.token, {})
                .then(() => {
                    assert.equal(factors.app.client.token, _r.body.token, "App initialization failed.");
                    return _r;
                });
        });
}

// Custom assert methods.

function assertOnUserIdMapFailure(r) {
    assert.isTrue(r.body.hasOwnProperty("user_id"), "user_id missing on response.");
    assert.equal(Cookie.getDecoded(constant.cookie.USER_ID), r.body.user_id, 
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
    if (r.status > 300) console.trace();
    assert.isTrue(r.status <= 299, "Response should be successful. Failure status seen.");
    return r;
}

function assertIfUserIdOnResponse(r) {
    assert.isFalse(r.body.hasOwnProperty("user_id"), "user_id on response when not required.");
    return r;
}

/**
 * Private methods
 */

const SuitePrivateMethod = {};

SuitePrivateMethod.testGetPropertiesFromQueryString = function() {
    assert.isEmpty(Object.keys(Properties.parseFromQueryString("")));
    assert.isEmpty(Object.keys(Properties.parseFromQueryString("?")));

    let props = Properties.parseFromQueryString("?a")
    assert.isEmpty(Object.keys(props));

    props = Properties.parseFromQueryString("?a=")
    assert.isEmpty(Object.keys(props));

    // Type check for numerical properties.
    props = Properties.parseFromQueryString("?a=10")
    assert.isNotEmpty(Object.keys(props));
    assert.isTrue(props && props.$qp_a === 10); // Do we need this as int?

    props = Properties.parseFromQueryString("?a=10abc")
    assert.isNotEmpty(Object.keys(props));
    assert.isTrue(props && props.$qp_a === "10abc");

    props = Properties.parseFromQueryString("?a=ab10")
    assert.isNotEmpty(Object.keys(props));
    assert.isTrue(props && props.$qp_a === "ab10");

    props = Properties.parseFromQueryString("?a=10&")
    assert.isNotEmpty(Object.keys(props));
    assert.isTrue(props && props.$qp_a === 10); 

    props = Properties.parseFromQueryString("?a=10&b")
    assert.isNotEmpty(Object.keys(props));
    assert.isTrue(props && props.$qp_a === 10); 
    assert.isTrue(props && props.$qp_b == null);

    props = Properties.parseFromQueryString("?a=10&utm=medium");
    assert.isNotEmpty(Object.keys(props));
    assert.isTrue(props && props.$qp_a === 10); 
    assert.isTrue(props && props.$qp_utm === "medium");
}

SuitePrivateMethod.testGetDefaultProperties = function() {
    assert.isNotEmpty(Properties.getDefault())
    let props =  Properties.getDefault();
    // No empty values.
    for (let k in props) assert.isNotEmpty(props[k].toString(), "Empty: "+k);

    // Check individual keys needed.
    assert.containsAllKeys(props, 
        ["$referrer", "$browser", "$browser_version", "$os", "$os_version", "$screen_width", "$screen_height"]);
    
    props = Properties.getDefault();
    // If 8 properties, includes device.
    if (props.$device) assert.isTrue(props.$device != "");
}


/**
 * Public methods
 */

const App = require("../app");
const SuitePublicMethod = {};

var app = new App();

// Todo: 

// Test: Init

SuitePublicMethod.testInit = function() {
    return setupNewProject()
        .then((r) => {
            app.reset();
            assert.isTrue(r.body.hasOwnProperty("token"), "Token should be in the response.");
            assert.isTrue(r.body.token.trim().length > 0, "Token should not be empty.")

            app.reset();
            return app.init(r.body.token)
                .then( () => assert.isTrue(app.isInitialized()) );
        });
}

SuitePublicMethod.testInitWithBadInput = function() {
    factors.reset();

    // Bad input. Invalidated on sdk.
    assert.throws(() => factors.init(" "), Error, "FactorsArgumentError: token cannot be empty.");
    assert.equal(factors.app.client, null, "Bad input token should not be allowed.");
}

// Test: Track

SuitePublicMethod.testTrackBeforeInit = function() {
    factors.reset();
    // Should throw exception.
    assert.throws(factors.track, Error, "FactorsError: SDK is not initialized with token.");
}

SuitePublicMethod.testTrackAfterInit = function() {
    return setupNewProjectAndInit()
        .then((r) => {
            // Validate track response.
            let eventName = randomAlphaNumeric(10);
            return factors.track(eventName, {})
                .then(assertIfHttpFailure)
                .then(assertOnUserIdMapFailure)
        })
}

SuitePublicMethod.testTrackWithBadToken = function() {
    factors.reset();

    // Bad input. Invalidated on backend.
    let eventName = randomAlphaNumeric(10);
    return factors.init("BAD_TOKEN", {}) // Should fail on get settings.
        .then(assertOnCall)
        .catch(suppressExpectedError);    
}

SuitePublicMethod.testTrackWithoutEventName = function() {
    factors.reset();

    return setupNewProjectAndInit()
        .then((r) => {
            // Fail if no eventName.
            factors.track()
                .then(assertOnCall)
                .catch(suppressExpectedError);
        });
}

SuitePublicMethod.testTrackWithoutEventProperites = function() {
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
SuitePublicMethod.testTrackWithoutUserCookie = function() {
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
SuitePublicMethod.testTrackWithUserCookie = function() {
    return setupNewProjectWithUser()
        .then((r) => {
            factors.reset(); // Clears existing env.
            return factors.init(r.project.body.token, {})
                .then(() => {
                    Cookie.setEncoded(constant.cookie.USER_ID, r.user.body.id);
                    factors.track(randomAlphaNumeric(10), {})
                        .then(assertIfHttpFailure)
                        .then(assertIfUserIdOnResponse); // user_id shouldn't be there on response.
            }); 
        })
        .catch(assertOnCall);
}

// Test: identify

SuitePublicMethod.testIdentifyBeforeInit = function() {
    factors.reset();

    // Throws error, if not initialized.
    assert.throws(factors.identify, Error, "FactorsError: SDK is not initialized with token.");
}

SuitePublicMethod.testIdentifyAfterInit = function() {
    return setupNewProjectAndInit()
        .then((r) => {
            let customerUserId = randomAlphaNumeric(15);
            Cookie.remove(constant.cookie.USER_ID);
            return factors.identify(customerUserId)
                .then(assertIfHttpFailure)
                .then(assertOnUserIdMapFailure) // It has to set cookie.
        })
        .catch(assertOnCall);
}

SuitePublicMethod.testIdentifyWithoutCustomerUserId = function() {
    factors.reset();

    return setupNewProjectAndInit()
        .then((r) => {
            assert.throws(factors.identify, Error, "FactorsArgumentError: Invalid type for customer_user_id");
            assert.throws(() => factors.identify(" "), Error, "FactorsArgumentError: customer_user_id cannot be empty.");
        })
        .catch(assertOnCall);
}

// New user => Without user cookie.
SuitePublicMethod.testIdentifyWithoutUserCookie = function() {
    factors.reset();

    return setupNewProjectAndInit()
        .then((r) => {
            let customerUserId = randomAlphaNumeric(15);
            Cookie.remove(constant.cookie.USER_ID);
            return factors.identify(customerUserId)
                .then(assertIfHttpFailure)
                .then(assertOnUserIdMapFailure);
        });
}

// With user cookie => Identify existing unidentified user.
SuitePublicMethod.testIdentifyWithUserCookie = function() {
    return setupNewProjectWithUser()
        .then((r) => {
            factors.reset(); // Clear existing env.
            return factors.init(r.project.body.token, {})
            .then(() => {
                assert.equal(factors.app.client.token, r.project.body.token, "App initialization failed.");
                Cookie.setEncoded(constant.cookie.USER_ID, r.user.body.id); // Setting new user.
                let customerUserId = randomAlphaNumeric(15);
                return factors.identify(customerUserId)
                    .then(assertIfHttpFailure)
                    .then(assertIfUserIdOnResponse) // Should not respond new user.
            });
        })
        .catch(assertOnCall);
}

SuitePublicMethod.testIdentifyWithIdentifiedCustomerUserWithSameUserCookie = function() {
    return setupNewProjectAndInit()
        .then((r) => {
            var customerUserId = randomAlphaNumeric(15);
            var _token = factors.app.client.token; // Fix: Project context copy assign.
            // Identified as customer user and user cookie set here.
            Cookie.remove(constant.cookie.USER_ID);
            return factors.identify(customerUserId)
                .then(assertIfHttpFailure)
                .then(assertOnUserIdMapFailure) 
                .then((r1) => {
                    // Fix: Project context copy assign.
                    return factors.init(_token)
                        .then(() => {
                            // Same user cookie used to identify with same customer user id.
                            // No user_id changes required. user_id should not exist on response.
                            return factors.identify(customerUserId)
                                .then(assertIfHttpFailure)
                                .then(assertIfUserIdOnResponse); // Should not response new user as identified with same user already.
                        });
                });
        });
}

SuitePublicMethod.testIdentifyWithIdentifiedCustomerUserWithDifferentUserCookie = function() {
    return setupNewProjectWithUser()
        .then((r1) => {
            factors.reset();
            return factors.init(r1.project.body.token, {})
                .then(() => {
                    assert.equal(factors.app.client.token, r1.project.body.token, "App initialization failed.");

                    Cookie.setEncoded(constant.cookie.USER_ID, r1.user.body.id);
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
                                    Cookie.setEncoded(constant.cookie.USER_ID, rUser.body.id);
                                    return factors.init(_token)
                                        .then(() => {
                                            // Re-identify with same customer_user and diff user.
                                            // The new user_id will be identified as customer_user.
                                            // No changes on user_id required.
                                            factors.identify(_customerUserId)
                                                .then(assertIfHttpFailure)
                                                .then(assertIfUserIdOnResponse);
                                        });  
                                });
                        });
                })
        })
}

// Identify as an identified user without user cookie should respond
// with latest user of the customer_user. Reusing same user session.
SuitePublicMethod.testIdentifyWithIdentifiedCustomerUserWithoutUserCookie = function() {}

// Test: addUserProperties

SuitePublicMethod.testAddUserPropertiesWithEmptyProperties = function() {
    factors.reset();

    return setupNewProjectAndInit()
        .then(() => {
            return factors.addUserProperties({})
                .then(assertOnCall) // Should not be resolved.
                .catch((m) => {
                    // To be rejected with message.
                    assert.equal(m, "No changes. Empty properties.");
                });
        });
}

SuitePublicMethod.testAddUserPropertiesBadProperties = function() {
    factors.reset();

    return setupNewProjectAndInit()
        .then(() => {
            // Properties argument type mismatch. Using string as properties.
            return factors.addUserProperties("STRING_INPUT")
                .then(assertOnCall)
                .catch(suppressExpectedError);
        });
}

SuitePublicMethod.testAddUserPropertiesWithoutUserCookie = function() {}

SuitePublicMethod.testAddUserPropertiesWithUserCookie = function() {
    return setupNewProjectWithUser()
        .then((r) => {
            factors.reset();
            return factors.init(r.project.body.token, {})
                .then(() => {
                    assert.equal(factors.app.client.token, r.project.body.token, "App initialization failed.");
            
                    Cookie.setEncoded(constant.cookie.USER_ID, r.user.body.id);

                    let properties = { userHandle: randomAlphaNumeric(15) };
                    return factors.addUserProperties(properties)
                        .then(assertIfHttpFailure)
                        .then(assertIfUserIdOnResponse) // User should not be created.  
                });
        });
}

/**
 * Test Runners
 */

function runPrivateMethodsSuite() {
    window.FACOTRS_DEBUG=true;

    for (let test in SuitePrivateMethod) {
        console.log('%c Running Private Methods Suite '+test+'..', 'color: green');
        SuitePrivateMethod[test]();
    }
    return true;
}

// Todo(Dinesh): Inconsistent: Make async methods to run synchronously as cookie 
// is a shared object.
function runPublicMethodsSuite() {
    window.FACOTRS_DEBUG=true;

    // Runs individual test in the test_suite.
    for (let test in SuitePublicMethod) {
        console.log('%c Running Public Methods Suite '+test+'..', 'color: green');
        SuitePublicMethod[test]();
    }
    return true;
}

// Runs everything.
function run() {
    runPrivateMethodsSuite();
    runPublicMethodsSuite();
}

module.exports = exports = { SuitePublicMethod, SuitePrivateMethod, runPrivateMethodsSuite, runPrivateMethodsSuite, run };

