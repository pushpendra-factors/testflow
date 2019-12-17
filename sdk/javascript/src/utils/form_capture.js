function isEmail(email) {
    // ref: https://stackoverflow.com/questions/46155/how-to-validate-an-email-address-in-javascript
    var re = /^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
    return re.test(String(email).toLowerCase());
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

function getPropertiesFromForm(form) {
    var inputs = form.querySelectorAll('input');
    var properties = {};

    for (var i=0; i<inputs.length; i++) {
        // exclude password from any processing.
        if (inputs[i].type == 'password') continue;
        if (!inputs[i].value) continue;
        
        var value = inputs[i].value.trim();
        if (inputs[i].value == "") continue;

        // any input field with a valid email catuptured as email.
        if (isEmail(value) && !properties['$email']) 
            properties['$email'] = value;

        if (inputs[i].type == 'tel') properties['$phone'] = value;

        if (!properties['$company'] && 
            (isFieldByMatch(inputs[i], 'company') || isFieldByMatch(inputs[i], 'org')))
            properties['$company'] = value;

        // name or placeholder as first and name tokens.
        if (isFieldByMatch(inputs[i], 'first', 'name')) {
            if (!properties['$name']) properties['$name'] = '';
            // prepend first name with existing name value.
            properties['$name'] = value + properties['$name'];
        }

        // name or placeholder as last and name tokens.
        if (isFieldByMatch(inputs[i], 'last', 'name')) {
            if (!properties['$name'] || properties['$name'] == "") properties['$name'] = value;
            else properties['$name'] + ' ' + value; // appended with space.
        }

        // only name.
        if (isFieldByMatch(inputs[i], 'name')) {
            // add only if it is not filled by first and last name already.
            if (!properties['$name']) properties['$name'] = value;
        }
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