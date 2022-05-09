"use strict";

const util = require("./utils/util");
const FormCapture = require("./utils/form_capture");
const logger = require("./utils/logger");

const PLATFORM_WEB = "web";

// properties
const PREFIX = "$";

const PAGE_SPENT_TIME = PREFIX+"page_spent_time";
const PAGE_LOAD_TIME = PREFIX+"page_load_time";
const PAGE_SCROLL_PERCENT = PREFIX+"page_scroll_percent";
const COMPANY = PREFIX+"company";
const EMAIL = PREFIX+"email";
const PHONE = PREFIX+"phone";
const NAME = PREFIX+"name";
const FIRST_NAME = PREFIX+"first_name";
const LAST_NAME = PREFIX+"last_name";

//Form properties
const FORM_ID = PREFIX + "form_id";
const FORM_NAME = PREFIX + "form_name";
const FORM_CLASS = PREFIX + "form_class";
const FORM_TYPE = PREFIX + "form_type";
const FORM_METHOD = PREFIX + "form_method";
const FORM_TARGET = PREFIX + "form_target";
const FORM_ACTION = PREFIX + "form_action";

// No $ since it's supposed to be tracked for marketing purposes
const SOURCE = "source";

// Events
const EV_FORM_SUBMITTED = PREFIX + 'form_submitted';


// Input related
const DISABLED_INPUT_TYPES = ['password', 'hidden'];

const isBotUserAgent = function(nAgt) {
    let lcaseAgt = nAgt.toLowerCase();
    // ref: https://stackoverflow.com/a/15047834, https://webmasters.stackexchange.com/a/64805
    return lcaseAgt.indexOf('bot') > -1 || lcaseAgt.indexOf('spider') > -1 || lcaseAgt.indexOf('crawl') > -1 || 
        lcaseAgt.indexOf('slurp') > -1 || lcaseAgt.indexOf('mediapartners') > -1;
}

const BrowserInfo = {
    getDevice: function () {
        var
            i,
            nVer = navigator.appVersion,
            nAgt = navigator.userAgent,
            tabletStrings = [
                    { s: 'iPad', r: /iPad/ },
                    { s: 'Samsung Galaxy', r: /SCH-I800/ },
                    { s: 'Motorola', r: /xoom/ },
                    { s: 'Kindle', r: /kindle/ }
            ],
            phoneStrings = [
                    { s: 'iPhone', r: /iPhone/ },
                    { s: 'iPod', r: /iPod/ },
                    { s: 'blackberry', r: /blackberry/ },
                    { s: 'android 0.5', r: /android 0.5/ },
                    { s: 'htc', r: /htc/ },
                    { s: 'lg', r: /lg/ },
                    { s: 'midp', r: /midp/ },
                    { s: 'mmp', r: /mmp/ },
                    { s: 'mobile', r: /mobile/ },
                    { s: 'nokia', r: /nokia/ },
                    { s: 'opera mini', r: /opera mini/ },
                    { s: 'palm', r: /palm|PalmSource/ },
                    { s: 'pocket', r: /pocket/ },
                    { s: 'psp', r: /psp|Playstation Portable/ },
                    { s: 'sgh', r: /sgh/ },
                    { s: 'smartphone', r: /smartphone/ },
                    { s: 'symbian', r: /symbian/ },
                    { s: 'treo mini', r: /treo mini/ },
                    { s: 'SonyEricsson', r: /SonyEricsson/ },
                    { s: 'Samsung', r: /Samsung/ },
                    { s: 'MobileExplorer', r: /MobileExplorer/ },
                    { s: 'Benq', r: /Benq/ },
                    { s: 'Windows Phone', r: /Windows Phone/ },
                    { s: 'Windows Mobile', r: /Windows Mobile/ },
                    { s: 'IEMobile', r: /IEMobile/ },
                    { s: 'Windows CE', r: /Windows CE/ },
                    { s: 'Nintendo Wii', r: /Nintendo Wii/ }
            ]
        ;


        var is_tablet= false,
            is_phone = false,
            device = "";

        for (var i=0; i<tabletStrings.length; i++) {
            if (tabletStrings[i].r.test(nAgt)) { 
                device = tabletStrings[i].s;
                is_tablet = true;
                break;
            }
        }
        if (device === "") {
            for (var i=0; i<phoneStrings.length; i++) {
                if (phoneStrings[i].r.test(nAgt)) {
                    device = phoneStrings[i].s;
                    is_phone = true;
                    break;
                }
            }
        }


        // if they are on tablet or phone
        var is_mobile = is_tablet || is_phone;
        if (!is_mobile) {
            is_mobile = /Mobile|mini|Fennec|Android/.test(nVer);
        }

        return {
            screen: {
                width: screen.width,
                height: screen.height
            },
            device: device,
            isMobile: is_phone || is_mobile || is_tablet
        };
    },
    getBrowserFeatures: function () {
        // cookie support
        var cookieEnabled = false;
        try {
            // Create cookie
            document.cookie = 'cookietest=1';
            var ret = document.cookie.indexOf('cookietest=') != -1;
            // Delete cookie
            document.cookie = 'cookietest=1; expires=Thu, 01-Jan-1970 00:00:01 GMT';
            cookieEnabled = ret;
        }
        catch (e) {
        }

        return {
            window: {
                width: window.innerWidth,
                height : window.innerHeight
            },
            allowsCookies:cookieEnabled
        }
    }
}

function getPageLoadTimeInMs() {
    if (!window.performance) return 0;
    // ref:https://www.html5rocks.com/en/tutorials/webperformance/basics
    return (window.performance.timing.domContentLoadedEventEnd - window.performance.timing.navigationStart);
}

/**
 * @returns Page load time in seconds.
 */
function getPageLoadTime() {
    var pageLoadTimeInMs = getPageLoadTimeInMs();
    return pageLoadTimeInMs > 0 ? pageLoadTimeInMs/1000 : 0;
}

/**
 * @param {string} pfix Prefix to be added to default properties.
 * Property - Example
 * referrer - https://google.com/search
 * */
function getEventDefault(referrer='') {
    let dp = {};
    dp[PREFIX+"page_title"] = document.title;

    dp[PREFIX+"referrer"] = referrer;
    let referrerURL = util.parseURLString(referrer);
    dp[PREFIX+"referrer_domain"] = referrerURL.host;
    // url domain with path and without query params.
    dp[PREFIX+"referrer_url"] = referrerURL.host + referrerURL.path + util.getCleanHash(referrerURL.hash);
    
    dp[PREFIX+"page_raw_url"] = window.location.href;
    dp[PREFIX+"page_domain"] = window.location.host;
    // url domain with path and without query params.
    dp[PREFIX+"page_url"] = window.location.host + window.location.pathname + util.getCleanHash(window.location.hash);
    
    return dp;
}

/**
 * @param {string} pfix Prefix to be added to default properties.
 * Property - Example
 * referrer - https://google.com/search
 * browser - Chrome
 * browser_version - 7
 * browser_version_string - 70.0.3538.102
 * os - Mac OSX
 * os_version_string - 10_13_6
 * */
function getUserDefault() {
    let dp = {};
    dp[PREFIX+"platform"] = PLATFORM_WEB;

    let device = BrowserInfo.getDevice();
    if (device.screen && device.screen.width > 0) 
        dp[PREFIX+"screen_width"] = device.screen.width;
    if (device.screen && device.screen.height > 0) 
        dp[PREFIX+"screen_height"] = device.screen.height;

    // Device name added, if mobile.
    if (device.device) dp[PREFIX+"device_name"] = device.device;

    // Note: IP and location information added by backend.
    return dp;
}

/**
 * Parse query string.
 * @param {string} qString query string.
 * @param {string} prefix Prefix the keys. optional, defaults to "$qp_".
 * ----- Cases -----
 * window.location.search = "" -> {}
 * window.location.search = "?" -> {}
 * window.location.search = "?a" -> {}
 * window.location.search = "?a=" -> {}
 * window.location.search = "?a=10" -> {a: 10}
 * window.location.search = "?a=10&" -> {a: 10}
 * window.location.search = "?a=10&b" -> {a: 10}
 * window.location.search = "?a=10&b=20" -> {a: 10, b: 20}
 * window.location.search = "?a=10&b=medium" -> {a: 10, b: "medium"}
 */
function parseFromQueryString(qString, prefix="$qp_") {
    // "?" check is not necessary for window.search. Added to stay pure.
    if (typeof(qString) !== "string" 
        || qString.length === 0 
        || (qString.length === 1 && qString.indexOf("?") === 0)) 
        return {};
    let ep = {};
    let t = null;
    // Remove & at the end.
    let ambPos = qString.indexOf("&");
    if (ambPos === qString.length-1) qString = qString.slice(0, qString.length-1);
    if (ambPos >= 0) t = qString.split("&");
    else t = [qString];
    for (var i=0; i<t.length; i++){
        let kv = null;
        if (t[i].indexOf("=") >= 0) kv = t[i].split("=");
        else kv = [t[i], null];
        // Remove ? on first query param.
        if (i == 0 && kv[0].indexOf("?") === 0) kv[0] = kv[0].slice(1);
        // don't allow keys with null value.
        if (kv[1] != "" && kv[1] != null && kv[1] != undefined)
            ep[prefix+kv[0]] = util.convertIfNumber(kv[1]); // converts if value is a number.
    }
    return ep;
}

function getTypeValidated(props={}) {
    let vprops = {}
    for (let k in props) {
        // Value validation: Allows only number or string.
        if (typeof(props[k]) == "string" || typeof(props[k]) == "number")
            vprops[k] = props[k];
    }
    return vprops;
}

// Merges query params on hash with URL query params.
function getAllQueryParamStr(location) {
    let hashStr = location.hash;
    let queryStr = location.search;

    let hashQuery = hashStr.split("?")[1];
    if (hashQuery == undefined || hashQuery == "") return queryStr;
    if (queryStr == "") return "?" + hashQuery;
    return queryStr + "&" + hashQuery
}

function getFromQueryParams(location) {
    return parseFromQueryString(getAllQueryParamStr(location));
}

function getFormMetaAttributes(form) {
    if(!form) return {};
    var properties = {};
    properties[FORM_ID] = form.getAttribute('id');
    properties[FORM_NAME] = form.getAttribute('name');
    properties[FORM_CLASS] = form.getAttribute('class');
    properties[FORM_ACTION] = form.getAttribute('action');
    properties[FORM_METHOD] = form.getAttribute('method');
    properties[FORM_TARGET] = form.getAttribute('target');
    properties[FORM_TYPE] = form.getAttribute('type');
    return properties;
}

function getPropertiesFromForm(form) {
    return form ? getPropertiesFromInputs(form.querySelectorAll('input')) : {}
}

function getPropertiesFromAllNonFormInputs() {
    var inputs = document.querySelectorAll('input');
    var properties = {};
    var formProperties = {};
    
    var nonFormInputs = [];
    for (var i=0; i<inputs.length; i++) {
        //exldude disabled types
        if(DISABLED_INPUT_TYPES.indexOf(inputs[i].type) >= 0) continue;

        if (!FormCapture.isPartOfForm(inputs[i])) {
            nonFormInputs.push(inputs[i]);

            var formElement = inputs[i].form? inputs[i].form : null
            formProperties = getFormMetaAttributes(formElement);
        }
    }

    properties = getPropertiesFromInputs(nonFormInputs);

    if(formProperties && Object.keys(formProperties).length > 0) {
        logger.debug("Collecting form meta attributes", false);
        properties = Object.assign(formProperties, properties);
    }

    return properties;
}

function getPropertiesFromInputs(inputs) {
    var properties = {};

    if (!inputs) return properties;    
    for (var i=0; i<inputs.length; i++) {
        // exclude password from any processing.
        if(DISABLED_INPUT_TYPES.indexOf(inputs[i].type) >= 0) continue;
        if (!inputs[i].value) continue;
        
        var value = inputs[i].value.trim();
        if (inputs[i].value == "") continue;

        // any input field with a valid email catuptured as email.
        if (FormCapture.isEmail(value) && !properties[EMAIL]) // captures only first email input.
            properties[EMAIL] = value;

        if (inputs[i].type == 'tel' && !properties[PHONE] && FormCapture.isPhone(value)) 
            properties[PHONE] = value;

        if (!properties[COMPANY] && 
            (FormCapture.isFieldByMatch(inputs[i], 'company') || FormCapture.isFieldByMatch(inputs[i], 'org')))
            properties[COMPANY] = value;

        // name or placeholder as first and name tokens.
        if (FormCapture.isFieldByMatch(inputs[i], 'first', 'name')) {
            if (!properties[FIRST_NAME]) properties[FIRST_NAME] = value;

            if (!properties[NAME]) properties[NAME] = '';
            // prepend first name with existing name value.
            properties[NAME] = value + properties[NAME];
        }

        // name or placeholder as last and name tokens.
        if (FormCapture.isFieldByMatch(inputs[i], 'last', 'name')) {
            if (!properties[LAST_NAME]) properties[LAST_NAME] = value;

            if (!properties[NAME] || properties[NAME] == '') properties[NAME] = value;
            else properties[NAME] + ' ' + value; // appended with space.
        }

        // only name.
        if (FormCapture.isFieldByMatch(inputs[i], 'name')) {
            // add only if it is not filled by first and last name already.
            if (!properties[NAME]) properties[NAME] = value;
        }
    }

    return properties; 
}

function getPageScrollPercent() {
    var h = document.documentElement, 
        b = document.body,
        st = 'scrollTop',
        sh = 'scrollHeight';

    var top = h[st]||b[st], height = (h[sh]||b[sh]) - h.clientHeight;
    return height == 0 ? 0 : (top / height) * 100;
}

module.exports = {
    getUserDefault: getUserDefault,
    getEventDefault: getEventDefault,
    getFromQueryParams: getFromQueryParams,
    parseFromQueryString: parseFromQueryString,
    getTypeValidated: getTypeValidated,
    getPageLoadTimeInMs: getPageLoadTimeInMs,
    getPageLoadTime: getPageLoadTime,
    getPropertiesFromForm: getPropertiesFromForm,
    getFormMetaAttributes: getFormMetaAttributes,
    getPageScrollPercent: getPageScrollPercent,
    getPropertiesFromAllNonFormInputs: getPropertiesFromAllNonFormInputs,

    PAGE_SPENT_TIME: PAGE_SPENT_TIME,
    PAGE_LOAD_TIME: PAGE_LOAD_TIME,
    PAGE_SCROLL_PERCENT: PAGE_SCROLL_PERCENT,
    EMAIL: EMAIL,
    PHONE, PHONE,
    SOURCE: SOURCE,

    EV_FORM_SUBMITTED: EV_FORM_SUBMITTED
}