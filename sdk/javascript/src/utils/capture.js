const logger = require("./logger");

const FAITRACKER_FORM_BIND_ATTRIBUTE = 'data-faitracker-form-bind';
const FAITRACKER_CLICK_BIND_ATTRIBUTE = 'data-faitracker-click-bind';

const TRIGGER_FORM_BINDING_EVENT = "trigger-form-binding";

function isEmail(email) {
    // ref: https://stackoverflow.com/questions/46155/how-to-validate-an-email-address-in-javascript
    var re = /^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
    return re.test(String(email).toLowerCase());
}

function isPhone(phone) {
    // should not be empty.
    if (!phone || phone.length == 0 || phone.length > 20) return false;

    // should not contain any alphabets.
    if (/[a-zA-Z]/g.test(phone)) return false;

    // atleast 4 numbers.
    var numbers = phone.match(/\d/g);
    if (!numbers || numbers.length < 4 || numbers.length > 20) return false;
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

function getElemsFromIframes(elemType) {
    logger.debug("Scanning for iframe forms.", false);
    var frames = document.querySelectorAll('iframe');
    var elemsFromFrames = [];
    for(var fri=0; fri<frames.length; fri++) {
        // This only takes care of certain scenarios. Will have to check all iframe scenarios.
        var elems = [];

        if(frames[fri].contentDocument && frames[fri].contentDocument.querySelectorAll) {
            elems = frames[fri].contentDocument.querySelectorAll(elemType);
        }

        for(var formI=0; formI<elems.length; formI++) {
            elemsFromFrames.push(elems[formI])
        }
    }
    return elemsFromFrames;
}

function getFormsFromIframes() {
    return getElemsFromIframes('form');
}

function getElemsFromTopAndIframes(elemType) {
    var topElems = document.querySelectorAll(elemType);
    var elems = [].slice.call(topElems);
    
    // Add iframe elems to the top list.
    var iframeElems = getElemsFromIframes(elemType);
    for (var ind=0; ind<iframeElems.length; ind++) {
        elems.push(iframeElems[ind]);
    }

    return elems;
}

function bindAllFormsOnSubmit(appInstance, processCallback) {
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
        if (!forms[i].getAttribute(FAITRACKER_FORM_BIND_ATTRIBUTE)) {
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
            
            forms[i].setAttribute(FAITRACKER_FORM_BIND_ATTRIBUTE, true);
        }
        

        // Bind processCallback to on-click of form's submit element,
        // But call only if not processed on submit.
        var submitElement = forms[i].querySelector('*[type="submit"]');
        if (!submitElement || submitElement.getAttribute(FAITRACKER_FORM_BIND_ATTRIBUTE)) continue;

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
        submitElement.setAttribute(FAITRACKER_FORM_BIND_ATTRIBUTE, true);
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
        var _button = buttons[i];

        // Do not bind button part of a form or bound already.
        if (isPartOfForm(_button) || _button.getAttribute(FAITRACKER_FORM_BIND_ATTRIBUTE)) continue;

        _button.addEventListener('click', function() {
            logger.debug("Executing callback on click of button.", false);
            
            var _appInstance = appInstance;
            processCallback(_appInstance);
        });

        _button.setAttribute(FAITRACKER_FORM_BIND_ATTRIBUTE, true);
    }
}

function stopFormBinderTask() {
    window.clearInterval(window.FAITRACKER_FORM_BINDER_ID);
    logger.debug('Stopped form binder task.');
}

function stopBackgroundFormBinderTaskLater() {
    // Stop the form binder only after 1 hour.
    window.setTimeout(stopFormBinderTask, 3600000);
}

function startBackgroundFormBinder(appInstance, processCallback) {
    // Binder should start only once for a window.
    if (!!window.FAITRACKER_FORM_BINDER_ID) {
        logger.debug("Form binder started already.", true)
        return;
    }
    
    // Triggers event EVENT_TRIGGER_FORM_BIND which is 
    // listened by all form captures to bind.
    var taskId = window.setInterval(function() {
        // We can add a mechanism to trigger only if new forms/inputs 
        // added to page. But that would require 2 pass of document tree.
        // Hence not added.
        logger.debug("Triggering form binding event.", false);
        document.dispatchEvent(new Event(TRIGGER_FORM_BINDING_EVENT));
    }, 2000);

    window.FAITRACKER_FORM_BINDER_ID = taskId;
    stopBackgroundFormBinderTaskLater();
}

function bindAllClickableElements(appInstance, processCallback) {
    var buttons = document.querySelectorAll('button,input[type="button"],input[type="submit"]');
    for (var i=0; i<buttons.length; i++) {
        // TODO: Do we have to exclude click capture for submit buttons of form?
        if (!buttons[i].getAttribute(FAITRACKER_CLICK_BIND_ATTRIBUTE)) {

            buttons[i].addEventListener('click', function(e) {
                logger.debug("Executing callback on click of button as part of click capture.", false);
                var _appInstance = appInstance;
                processCallback(_appInstance, e.target);
            });
            
            buttons[i].setAttribute(FAITRACKER_CLICK_BIND_ATTRIBUTE, true);
        }
    }

    var anchors = document.querySelectorAll('a');
    for (var i=0; i<anchors.length; i++) {
        if (!anchors[i].getAttribute(FAITRACKER_CLICK_BIND_ATTRIBUTE)) {
            anchors[i].addEventListener('click', function(e) {
                logger.debug("Executing callback on click of anchor as part of click capture.", false);

                var anchor = null;
                // Path is not a standard, composedPath is the standard and path can be backup
                var path = e.composedPath ? e.composedPath() : e.path;
                if (e.target && e.target.nodeName == 'A')
                    anchor = e.target;
                else if (path) {
                    // Any element can be encapsulated by anchor, on click of the 
                    // element the target received will be the internal element 
                    // rather than anchor. Anchor will available on the node path stack.
                    for(var p=0; p<path; p++) {
                        if (path[p].nodeName == 'A') {
                            anchor = path[p];
                            break;
                        }
                    }
                } else 
                logger.debug("Unable to get anchor element on click.", false);
                
                if (anchor) {
                    var _appInstance = appInstance;
                    processCallback(_appInstance, anchor);
                }
            });
            
            anchors[i].setAttribute(FAITRACKER_CLICK_BIND_ATTRIBUTE, true);
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

function cleanupString(s = "") {
    // Allows only letters, numbers, punctuation, whitespace, symbols.
    return s.replace(/[^\p{L}\p{N}\p{P}\p{Z}^$\n]/gu, '').trim();
}

function getClickCapturePayloadFromElement(element) {
    var payload = {};

    var displayName = element.textContent;
    if (!displayName) displayName = element.value;
    if (displayName != "" && displayName !== undefined) displayName = displayName.trim();

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
    startBackgroundClickBinder, getClickCapturePayloadFromElement, getElemsFromTopAndIframes, 
    getFormsFromIframes, TRIGGER_FORM_BINDING_EVENT };