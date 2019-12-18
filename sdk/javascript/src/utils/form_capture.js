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

module.exports = exports =  { isEmail, isFieldByMatch, bindAllFormsOnSubmit }; 