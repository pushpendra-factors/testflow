function isEmail(email) {
    // ref: https://stackoverflow.com/questions/46155/how-to-validate-an-email-address-in-javascript
    var re = /^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
    return re.test(String(email).toLowerCase());
}

function getPropertiesFromForm(form) {
    var inputs = form.querySelectorAll('input');
    var properties = {};

    for (var i=0; i<inputs.length; i++) {
        // exclude password from any processing.
        if (inputs[i].type == 'password') continue;

        // any input field with a valid email catuptured as email.
        if (isEmail(inputs[i].value) && !properties['$email']) 
            properties['$email'] = inputs[i].value;
    }

    return properties;
}

function bindAllFormsOnSubmit(appInstance, processCallback) {
    // bind processForm to onSubmit event for all forms.
    var forms = document.querySelectorAll('form');
    for (var i=0; i<forms.length; i++) {
        // addEventListener does not prevent other 
        // callbacks bound already.
        forms[i].addEventListener('submit', function(e) {
            var _appInstance = appInstance;
            processCallback(_appInstance, e);
        });
    }
}

module.exports = exports =  { isEmail, getPropertiesFromForm, bindAllFormsOnSubmit }; 