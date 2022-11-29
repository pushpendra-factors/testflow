const logger = require("./logger");

const FACTORS_BIND_ATTRIBUTE = 'data-factors-bind';
const FACTORS_CLICK_BIND_ATTRIBUTE = 'data-factors-click-bind';

function isEmail(email) {
    // ref: https://stackoverflow.com/questions/46155/how-to-validate-an-email-address-in-javascript
    var re = /^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
    return re.test(String(email).toLowerCase());
}

function isPhone(phone) {
    // should not be empty.
    if (!phone || phone.length == 0) return false;
    var numbers = phone.match(/\d/g);
    // atleast 4 numbers.
    if (!numbers || numbers.length < 4) return false;
    return true;
}

// Subtle email check.
function isPossibleEmail(value) {
    return value.indexOf("@") > -1 && value.indexOf(".") > -1
}

// Checks for string on name and placeholder. Say company.
// Cases: 
// isFieldByMatch({name: 'first_name'}, 'name', 'first');
// isFieldByMatch({name: 'no value', placeholder: 'enter your name'}, 'name');
function isFieldByMatch(elem, typeStr1, typeStr2) {
    if (!elem || !typeStr1 || typeStr1 == "") return false;

    var lTypeStr1 = typeStr1.toLowerCase(); 
    var lTypeStr2 = typeStr2 && typeStr2 != "" ? typeStr2.toLowerCase() : null;
    
    if (elem.name && elem.name.toLowerCase().indexOf(lTypeStr1) > -1) {
        // make sure secondary string is also there on name, if given.
        if (lTypeStr2) return elem.name.toLowerCase().indexOf(lTypeStr2) > -1;
        return true;
    }
        
    if (elem.placeholder && elem.placeholder.toLowerCase().indexOf(lTypeStr1) > -1) {
        // make sure secondary string is also there on placeholder, if given.
        if (lTypeStr2) return elem.placeholder.toLowerCase().indexOf(lTypeStr2) > -1;
        return true;
    }

    return false;
}

function getFormsFromIframes() {
    logger.debug("Scanning for iframe forms.", false);
    var frames = document.querySelectorAll('iframe');
    var formsFromFrames = [];
    for(var fri=0; fri<frames.length; fri++) {
        // This only takes care of certain scenarios. Will have to check all iframe scenarios.
        var frms = [];

        if(frames[fri].contentDocument && frames[fri].contentDocument.querySelectorAll) {
            frms = frames[fri].contentDocument.querySelectorAll('form');
        }

        for(var formI=0; formI<frms.length; formI++) {
            formsFromFrames.push(frms[formI])
        }
    }
    return formsFromFrames;
}

function bindAllFormsOnSubmit(appInstance, processCallback) {
    logger.debug("Scanning for unbind forms.", false);

    var iframeForms = getFormsFromIframes();

    // bind processForm to onSubmit event for all forms.
    var formsList = document.querySelectorAll('form');

    // converts to array. Will be empty array if any of the items are empty
    var forms = [].slice.call(formsList);
    
    for (var ind=0; ind<iframeForms.length; ind++) {
        forms.push(iframeForms[ind]);
    }
    
    for (var i=0; i<forms.length; i++) {
        var maxCallCount = 0; // Max callback calls expected.
        var callCount = 0; // Current callback calls.

        // Using unique attribute as flag to avoid binding,
        // multiple times.
        if (!forms[i].getAttribute(FACTORS_BIND_ATTRIBUTE)) {
            maxCallCount = maxCallCount + 1;
            forms[i].addEventListener('submit', function(e) {
                if (callCount > 0) {
                    // Reset the call count after one iteration of possible 
                    // duplicate callbacks, for capturing re-submit.
                    if (callCount == (maxCallCount-1)) callCount = 0;
                    return;
                }
                
                logger.debug("Executing callback on submit of form.", false);

                var _appInstance = appInstance;
                processCallback(_appInstance, e.target);
                callCount = callCount + 1;
            });
            
            forms[i].setAttribute(FACTORS_BIND_ATTRIBUTE, true);
        }
        

        // Bind processCallback to on-click of form's submit element,
        // But call only if not processed on submit.
        var submitElement = forms[i].querySelector('*[type="submit"]');
        if (!submitElement || submitElement.getAttribute(FACTORS_BIND_ATTRIBUTE)) continue;

        maxCallCount = maxCallCount + 1;
        submitElement.addEventListener('click', function(e) {
            if (callCount > 0) {
                if (callCount == (maxCallCount-1)) callCount = 0;
                return;
            }

            logger.debug("Executing callback on click of form submit button.", false);

            var _appInstance = appInstance;
            processCallback(_appInstance, e.target.form);
            callCount = callCount + 1;
        });
        submitElement.setAttribute(FACTORS_BIND_ATTRIBUTE, true);
    }
}

function isPartOfForm(e) {
    var isForm = e ? !!e.form : false;
    if (isForm) {
        // Only forms with submit button should be considered as standard form.
        return e.form.querySelectorAll("[type='submit']") > 0;
    }

    return false;
}

function bindAllNonFormButtonOnClick(appInstance, processCallback) {
    var buttons = document.querySelectorAll('button,input[type="button"],input[type="submit"]');
    for (var i=0; i<buttons.length; i++) {
        // do not bind button part of a form.
        if (isPartOfForm(buttons[i])) continue;

        buttons[i].addEventListener('click', function() {
            logger.debug("Executing callback on click of button.", false);
            
            var _appInstance = appInstance;
            processCallback(_appInstance);
        });
    }
}

function stopFormBinderTask() {
    window.clearInterval(window._factorsFormBinderTaskId);
    logger.debug('Stopped form binder task.');
}

function stopBackgroundFormBinderTaskLater() {
    // Stop the form binder only after 30 minutes.
    window.setTimeout(stopFormBinderTask, 1800000);
}

function startBackgroundFormBinder(appInstance, processCallback) {
    // binder should start only once for a window.
    if (!!window._factorsFormBinderTaskId) {
        logger.debug("Form binder started already.", true)
        return;
    }
    
    // try to bind on forms every 2secs tilltask gets 
    // cancelled using _factorsFormBinderTaskId.
    var taskId = window.setInterval(function() {
        var _appInstance = appInstance;
        bindAllFormsOnSubmit(_appInstance, processCallback);
    }, 2000);

    window._factorsFormBinderTaskId = taskId;
    stopBackgroundFormBinderTaskLater();
}

function bindAllClickableElements(appInstance, processCallback) {
    var buttons = document.querySelectorAll('button,input[type="button"],input[type="submit"]');
    for (var i=0; i<buttons.length; i++) {
        // TODO: Do we have to exclude click capture for submit buttons of form?
        if (!buttons[i].getAttribute(FACTORS_CLICK_BIND_ATTRIBUTE)) {

            buttons[i].addEventListener('click', function(e) {
                logger.debug("Executing callback on click of button as part of click capture.", false);
                var _appInstance = appInstance;
                processCallback(_appInstance, e.target);
            });
            
            buttons[i].setAttribute(FACTORS_CLICK_BIND_ATTRIBUTE, true);
        }
    }

    var anchors = document.querySelectorAll('a');
    for (var i=0; i<anchors.length; i++) {
        if (!anchors[i].getAttribute(FACTORS_CLICK_BIND_ATTRIBUTE)) {
            anchors[i].addEventListener('click', function(e) {
                logger.debug("Executing callback on click of anchor as part of click capture.", false);

                var anchor = null;
                if (e.target && e.target.nodeName == 'A')
                    anchor = e.target;
                else if (e.path) {
                    // Any element can be encapsulated by anchor, on click of the 
                    // element the target received will be the internal element 
                    // rather than anchor. Anchor will available on the node path stack.
                    for(var p=0; p<e.path; p++) {
                        if (e.path[p].nodeName == 'A') {
                            anchor = e.path[p];
                            break;
                        }
                    }
                } else 
                    logger.errorLine("Unable to get anchor element on click.")
                
                if (anchor) {
                    var _appInstance = appInstance;
                    processCallback(_appInstance, anchor);
                }
            });
            
            anchors[i].setAttribute(FACTORS_CLICK_BIND_ATTRIBUTE, true);
        }
    }
}

function startBackgroundClickBinder(appInstance, processCallback) {
    logger.debug("Scanning for unbound clickable elements.", false);

    // Look for new buttons and bind, every 2 seconds.
    window.setInterval(function() {
        var _appInstance = appInstance;
        bindAllClickableElements(_appInstance, processCallback);
    }, 2000);
}

function addElementAttributeIfExists(element, key, attributes) {
    var value = element.getAttribute(key);
    if (!value) return;

    value = value.trim();
    attributes[key] = value;
}

function cleanupString(s) {
    // Allows only letters, numbers, punctuation, whitespace, symbols.
    return s.replace(/[^\p{L}\p{N}\p{P}\p{Z}^$\n]/gu, '').trim();
}

function getClickCapturePayloadFromElement(element) {
    var payload = {};

    var displayName = element.textContent;
    if (!displayName) displayName = element.value;
    if (displayName != "") displayName = displayName.trim();

    // Possibilities A, BUTTON, INPUT[type=button].
    var elementType = element.nodeName; 

    var nodeName = element.nodeName;
    if (nodeName && nodeName != "") 
        nodeName = nodeName.trim().toLowerCase();

    // Defines the node type by the name of allowed list.
    if (nodeName == "a") elementType = "ANCHOR";
    else if (nodeName == "button") elementType = "BUTTON";
    // For "input" node with type button.
    if (element.getAttribute("type") == "button") 
        elementType = "BUTTON";

    var attributes = {};
    attributes.display_text = cleanupString(displayName);
    attributes.element_type = elementType;

    // default display_name. Still display_text 
    // attribute will be as same as in UI.
    if (displayName == "" || displayName == undefined)
        displayName = "unnamed_button_click";

    addElementAttributeIfExists(element, "class", attributes);
    addElementAttributeIfExists(element, "id", attributes);
    addElementAttributeIfExists(element, "name", attributes);
    addElementAttributeIfExists(element, "rel", attributes);
    addElementAttributeIfExists(element, "role", attributes);
    addElementAttributeIfExists(element, "target", attributes);
    addElementAttributeIfExists(element, "href", attributes);
    addElementAttributeIfExists(element, "media", attributes);
    addElementAttributeIfExists(element, "type", attributes);

    payload.display_name = cleanupString(displayName);
    payload.element_type = elementType;
    payload.element_attributes = attributes;

    return payload
}

module.exports = exports =  { isEmail, isPhone, isPossibleEmail, isFieldByMatch, isPartOfForm,
    bindAllFormsOnSubmit, bindAllNonFormButtonOnClick, startBackgroundFormBinder,
    startBackgroundClickBinder, getClickCapturePayloadFromElement };