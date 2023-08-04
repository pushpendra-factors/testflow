"use strict";

// Required chai with relative path for enable method suggestions to work.
const chai = require("chai");

const Cookie = require("../utils/cookie");
const Request = require("../utils/request");
const APIClient = require("../api-client");
const Properties = require("../properties");
const util = require("../utils/util");

const config = require("../config");
const constant = require("../constant");

// Enable full stacktrace for chai.
// chai.config.includeStack = true;

// Assertion with chai.assert
const assert = chai.assert;
var COMMON_CUST_ID = randomAlphaNumeric(10);

/**
 * Test Utils
 */

function randomAlphaNumeric(len) {
    var p = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
    return [...Array(len)].reduce(a=>a+p[~~(Math.random()*p.length)],'');
}

function suppressExpectedError(e) {
    if (e instanceof chai.AssertionError){
        console.trace(e);
        throw new Error(e);
    }
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

function setupNewUser(token) {
    if (!token) throw new Error("Setup new user failed. Invalid project.");

    let client = new APIClient(token);
    return client.identify({c_uid: COMMON_CUST_ID});
}

function setupNewProjectWithUser() {
    return setupNewProject()
        .then((r1) => {
            return setupNewUser(r1.body.token)
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
            app.reset(); // Clear existing environment.
            var _r = r;
            return app.init(r.body.token, {})
                .then(() => {
                    assert.equal(app.client.token, _r.body.token, "App initialization failed.");
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

SuitePrivateMethod.testGetUserDefaultProperties = function() {
    assert.isNotEmpty(Properties.getUserDefault())
    let props =  Properties.getUserDefault();
    // No empty values.
    for (let k in props) assert.isNotEmpty(props[k].toString(), "Empty: "+k);

    // Check individual keys needed.
    assert.containsAllKeys(props, 
        ["$platform", "$screen_width", "$screen_height"]);
    
    props = Properties.getUserDefault();
    if (props.$device) assert.isTrue(props.$device != "");
}

SuitePrivateMethod.testGetEventDefaultProperties = function() {
    assert.isNotEmpty(Properties.getEventDefault())
    let props =  Properties.getEventDefault();
    let emptyAllowed = ["$page_title", "$referrer", "$referrer_domain", "$referrer_url"];
    // No empty values.
    for (let k in props)
        // properties can be empty.
        if(emptyAllowed.indexOf(k) == -1) 
            assert.isNotEmpty(props[k].toString(), "Empty: "+k);
    
    // Check individual keys needed.
    assert.containsAllKeys(props, ["$page_domain", "$page_raw_url", "$page_title", "$page_url", 
    "$referrer", "$referrer_domain", "$referrer_url"]);
}

SuitePrivateMethod.testGetTypeValidatedProperties = function() {
    assert.isEmpty(Properties.getTypeValidated()); // no arg.
    assert.isEmpty(Properties.getTypeValidated({})); // empty props.

    let vprops = Properties.getTypeValidated({prop_1: "value_1"})
    assert.containsAllKeys(vprops, ["prop_1"]);
    assert.equal(vprops["prop_1"], "value_1");

    // Property value validation.
    vprops = Properties.getTypeValidated({"int_prop": 10, "float_prop": 10.2, "string_prop": "somevalue", "obj_prop": {"obj_num_prop": 10}});
    assert.containsAllKeys(vprops, ["int_prop", "float_prop", "string_prop"], "Should allow number and string.");
    assert.isFalse(!!vprops["obj_prop"], "Should not allow anything other than number or string.");
    assert.isFalse(!!vprops["obj_num_prop"], "Should not allow object prop.");
}

SuitePrivateMethod.testGetPropertiesFromQueryParams = function() {
    // use window.location mock object.
    let qprops = Properties.getFromQueryParams({hash: "#/activate?token=xxx", search: "?a=10"});
    assert.containsAllKeys(qprops, ["$qp_token", "$qp_a"])
    assert.equal(qprops["$qp_token"], "xxx")
    assert.equal(qprops["$qp_a"], "10")

    qprops = Properties.getFromQueryParams({hash: "", search: "?a=10"});
    assert.containsAllDeepKeys(qprops, ["$qp_a"])
    assert.equal(qprops["$qp_a"], "10")

    qprops = Properties.getFromQueryParams({hash: "#/activate?token=xxx", search: ""});
    assert.containsAllKeys(qprops, ["$qp_token"])
    assert.equal(qprops["$qp_token"], "xxx")
}

SuitePrivateMethod.testGetCleanHash = function() {
    assert.equal(util.getCleanHash("#/faitracker?q=10"), "#/faitracker")
    assert.equal(util.getCleanHash("#/faitracker?"), "#/faitracker")
    assert.equal(util.getCleanHash("#/faitracker"), "#/faitracker")
    assert.equal(util.getCleanHash("#/?q=10"), "#/")
    assert.equal(util.getCleanHash(""), "")
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
    app.reset();
    // Bad input. Invalidated on sdk.
    assert.throws(() => app.init(" "), Error, "FAITrackerArgumentError: token cannot be empty.");
    assert.equal(app.client, null, "Bad input token should not be allowed.");
}

// Test: Track

SuitePublicMethod.testTrackBeforeInit = function() {
    app.reset();

    try {
        app.track();
    } catch(e) {
        assert.equal(e.message, "FAITrackerError: SDK is not initialized with token.")
    }
}

SuitePublicMethod.testTrackAfterInit = function() {
    return setupNewProjectAndInit()
        .then((r) => {
            // Validate track response.
            let eventName = randomAlphaNumeric(10);
            return app.track(eventName, {})
                .then(assertIfHttpFailure)
                .then(assertOnUserIdMapFailure)
        })
}

SuitePublicMethod.testTrackWithBadToken = function() {
    app.reset();

    // Bad input. Invalidated on backend.
    return app.init("BAD_TOKEN", {}) // Should fail on get settings.
        .then(assertOnCall)
        .catch((r) => {
            assert.equal(r, "Failed on fetch.", "Should fail fetch settings on bad token");
        });
}

SuitePublicMethod.testTrackWithoutEventName = function() {
    app.reset();

    return setupNewProjectAndInit()
        .then((r) => {
            // Fail if no eventName.
            app.track().then(assertOnCall);
        });
}

SuitePublicMethod.testTrackWithoutEventProperites = function() {
    return setupNewProjectAndInit()
        .then((r) => {
            let eventName = randomAlphaNumeric(10);
            return app.track(eventName, {})
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
            return app.track(eventName, {})
                .then(assertIfHttpFailure)
                .then(assertOnUserIdMapFailure)
        })
        .catch(assertOnCall);
}

// Track as existing user. Track with user cookie.
SuitePublicMethod.testTrackWithUserCookie = function() {
    return setupNewProjectWithUser()
        .then((r) => {
            app.reset(); // Clears existing env.
            return app.init(r.project.body.token, {})
                .then(() => {
                    Cookie.setEncoded(constant.cookie.USER_ID, r.user.body.user_id);
                    app.track(randomAlphaNumeric(10), {})
                        .then(assertIfHttpFailure)
                        .then(assertIfUserIdOnResponse); // user_id shouldn't be there on response.
            }); 
        })
        .catch(assertOnCall);
}

// Test: identify

SuitePublicMethod.testIdentifyBeforeInit = function() {
    app.reset();

    try {
        app.identify();
    } catch(e) {
        assert.equal(e.message, "FAITrackerError: SDK is not initialized with token.")
    }
}

SuitePublicMethod.testIdentifyAfterInit = function() {
    return setupNewProjectAndInit()
        .then((r) => {
            let customerUserId = randomAlphaNumeric(15);
            Cookie.remove(constant.cookie.USER_ID);
            return app.identify(customerUserId)
                .then(assertIfHttpFailure)
                .then(assertOnUserIdMapFailure) // It has to set cookie.
        })
        .catch(assertOnCall);
}

SuitePublicMethod.testIdentifyWithoutCustomerUserId = function() {
    app.reset();

    return setupNewProjectAndInit()
        .then((r) => {
            try {
                app.identify(" ");
            } catch(e) {
                assert.equal(e.message, "FAITrackerArgumentError: customer_user_id cannot be empty.")
            }
        });
}

// New user => Without user cookie.
SuitePublicMethod.testIdentifyWithoutUserCookie = function() {
    app.reset();

    return setupNewProjectAndInit()
        .then((r) => {
            let customerUserId = randomAlphaNumeric(15);
            Cookie.remove(constant.cookie.USER_ID);
            return app.identify(customerUserId)
                .then(assertIfHttpFailure)
                .then(assertOnUserIdMapFailure);
        });
}

SuitePublicMethod.testIdentifyWithUserCookie = function() {
    return setupNewProjectWithUser()
        .then((r) => {
            app.reset(); // Clear existing env.
            return app.init(r.project.body.token, {})
            .then(() => {
                assert.equal(app.client.token, r.project.body.token, "App initialization failed.");
                Cookie.setEncoded(constant.cookie.USER_ID, r.user.body.user_id); // Setting new user.
                // Workaround as we use identify for setup new user.
                // which makes the user already identified.
                return app.identify(COMMON_CUST_ID) 
                    .then(assertIfHttpFailure)
                    .then(assertIfUserIdOnResponse) // Should not respond new user.
            });
        });
}

SuitePublicMethod.testIdentifyWithIdentifiedCustomerUserWithSameUserCookie = function() {
    return setupNewProjectAndInit()
        .then((r) => {
            var customerUserId = randomAlphaNumeric(15);
            var _token = app.client.token; // Fix: Project context copy assign.
            // Identified as customer user and user cookie set here.
            Cookie.remove(constant.cookie.USER_ID);
            return app.identify(customerUserId)
                .then(assertIfHttpFailure)
                .then(assertOnUserIdMapFailure) 
                .then((r1) => {
                    // Fix: Project context copy assign.
                    return app.init(_token)
                        .then(() => {
                            // Same user cookie used to identify with same customer user id.
                            // No user_id changes required. user_id should not exist on response.
                            return app.identify(customerUserId)
                                .then(assertIfHttpFailure)
                                .then(assertIfUserIdOnResponse); // Should not response new user as identified with same user already.
                        });
                });
        });
}

SuitePublicMethod.testIdentifyWithIdentifiedCustomerUserWithDifferentUserCookie = function() {
    return setupNewProjectWithUser()
        .then((r1) => {
            app.reset();
            return app.init(r1.project.body.token, {})
                .then(() => {
                    assert.equal(app.client.token, r1.project.body.token, "App initialization failed.");

                    Cookie.setEncoded(constant.cookie.USER_ID, r1.user.body.id);
                    let customerUserId = randomAlphaNumeric(15);
                    let _token = app.client.token; // Fix: Project context copy assign.

                    return app.identify(customerUserId)
                        .then(assertIfHttpFailure)
                        .then(assertIfUserIdOnResponse) 
                        .then((r2) => {
                            let _customerUserId = customerUserId;

                            return setupNewUser(r1.project.body.id)
                                .then((rUser) => {
                                    // Setting new user cookie.
                                    Cookie.setEncoded(constant.cookie.USER_ID, rUser.body.id);
                                    return app.init(_token)
                                        .then(() => {
                                            // Re-identify with same customer_user and diff user.
                                            // The new user_id will be identified as customer_user.
                                            // No changes on user_id required.
                                            app.identify(_customerUserId)
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
    app.reset();

    return setupNewProjectAndInit()
        .then(() => {
            return app.addUserProperties({})
                .then(assertOnCall) // Should not be resolved.
                .catch((m) => {
                    // To be rejected with message.
                    assert.equal(m, "No changes. Empty properties.");
                });
        });
}

SuitePublicMethod.testAddUserPropertiesBadProperties = function() {
    app.reset();

    return setupNewProjectAndInit()
        .then(() => {
            // Properties argument type mismatch. Using string as properties.
            return app.addUserProperties("STRING_INPUT")
                .then(assertOnCall)
                .catch(suppressExpectedError);
        });
}

SuitePublicMethod.testAddUserPropertiesWithoutUserCookie = function() {}

SuitePublicMethod.testAddUserPropertiesWithUserCookie = function() {
    return setupNewProjectWithUser()
        .then((r) => {
            app.reset();
            return app.init(r.project.body.token, {})
                .then(() => {
                    assert.equal(app.client.token, r.project.body.token, "App initialization failed.");
            
                    Cookie.setEncoded(constant.cookie.USER_ID, r.user.body.id);

                    let properties = { userHandle: randomAlphaNumeric(15) };
                    return app.addUserProperties(properties)
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

// Todo(Dinesh): Inconsistent test: Fixes - Create a mock object for document.cookie,
// Use app instance instead of faitracker like testInit.
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

module.exports = exports = { SuitePublicMethod, SuitePrivateMethod, runPrivateMethodsSuite, runPublicMethodsSuite, run };

