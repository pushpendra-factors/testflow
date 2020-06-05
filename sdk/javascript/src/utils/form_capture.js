const logger = require("./logger");

const FACTORS_BIND_ATTRIBUTE = 'data-factors-bind';

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
    if (!numbers || numbers < 4) return false;
    return true;
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

function bindAllFormsOnSubmit(appInstance, processCallback) {
    logger.debug("Scanning for unbind forms.", false);

    // bind processForm to onSubmit event for all forms.
    var forms = document.querySelectorAll('form');
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
    return e ? !!e.form : false;
}

function bindAllNonFormButtonOnClick(appInstance, processCallback) {
    var buttons = document.querySelectorAll('button');
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
    // The JS loaded last could be the one to add form to DOM, 
    // So stop the form binding 10secs after window load.
    window.addEventListener('load', function() {
        window.setTimeout(stopFormBinderTask, 10000);
    });

    // If window load not triggered even after 2mins,
    // stop the form binder.    
    window.setTimeout(stopFormBinderTask, 120000);
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

module.exports = exports =  { isEmail, isPhone, isFieldByMatch, isPartOfForm,
    bindAllFormsOnSubmit, bindAllNonFormButtonOnClick, startBackgroundFormBinder };