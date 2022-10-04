const docReady = (fn) => {
    // see if DOM is already available
    if (document.readyState === "complete" || document.readyState === "interactive") {
        // call on next available tick
        setTimeout(fn, 1);
    } else {
        document.addEventListener("DOMContentLoaded", fn);
    }
}

const testRunner = () => {
    [testCookieConsent].every((fn) => {
        const testResult = fn();
        if(!testResult.success) {
            console.log(testResult.msg);
            return false;
        }
        return true;

    });
}

const testCookieConsent = () => {
    const result = {
        success: false,
        msg: 'Step 0'
    }

    let origCookieProp = Object.assign({}, document.cookie);
    
    // stop accepting cookie.
    if(!document.__defineGetter__) {

        Object.defineProperty(document, 'cookie', {
            get: function(){return ''},
            set: function(){console.log("setter called"); return true},
        });
        } else {
            document.__defineGetter__("cookie", function() { return '';} );
            document.__defineSetter__("cookie", function() {} );
    }

    const consentCheckElement = document.createElement('div');

    consentCheckElement.style = "position: fixed; height: 200px; width: 200px; background: white;"
    consentCheckElement.innerHTML = "<div>Accept Cookies?</div><div><button id='consentButtonY'>Yes</button><button id='consentButtonX'>No</button></div>";

    const bodyChildElement = document.body.children[0];
    document.body.insertBefore(consentCheckElement, bodyChildElement);
    document.getElementById('consentButtonY').addEventListener('click', function() {      
        document.cookie = origCookieProp;
        factors.init();
        console.log(document.cookie);
        consentCheckElement.style = "display: none;";
    })
    document.getElementById('consentButtonX').addEventListener('click', () => {      
        consentCheckElement.style = "display: none;";
    });
    return result;
}


// create model to set consent
// when user accepts, enable cookie
// check if cookie is set

docReady(testRunner)

