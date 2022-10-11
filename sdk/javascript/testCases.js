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
        const testResult = new Promise(fn);
        return testResult.then((res) => {
            console.log(res.msg)
            return res.success;
        })
    });
}

const testCookieConsent = (res, rej) => {

    // stop accepting cookie.
   const cookieCodeGet = document.__lookupGetter__('cookie');
   const cookieCodeSet = document.__lookupSetter__('cookie');
    
    document.__defineGetter__("cookie", function() { return '';} );
    document.__defineSetter__("cookie", function() {} );

    const consentCheckElement = document.createElement('div');

    consentCheckElement.style = "position: fixed; height: 200px; width: 200px; background: white;"
    consentCheckElement.innerHTML = "<div>Accept Cookies?</div><div><button id='consentButtonY'>Yes</button><button id='consentButtonX'>No</button></div>";

    const bodyChildElement = document.body.children[0];
    document.body.insertBefore(consentCheckElement, bodyChildElement);
    document.getElementById('consentButtonY').addEventListener('click', function () {
        const wasQueueEmpty = !window.factors?.q?.length;   
        document.__defineGetter__("cookie", cookieCodeGet );
        document.__defineSetter__("cookie", cookieCodeSet );
        consentCheckElement.style = "display: none;";
        setTimeout(() => {
            if(!wasQueueEmpty && !window.factors?.q?.length) {
                res({
                    success: true,
                    msg: 'Step 0: Cookie consent && queue order: Success'
                });
            }
            
            res({
                success: false,
                msg: 'Step 0'
            });

        }, 10)
    })
    document.getElementById('consentButtonX').addEventListener('click', () => {      
        consentCheckElement.style = "display: none;";
    });
}


// create model to set consent
// when user accepts, enable cookie
// check if cookie is set

docReady(testRunner)
