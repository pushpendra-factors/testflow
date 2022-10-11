// Gets org level domain.
// e.g., www.example.co.uk -> example.co.uk 
// This method tries to set cookies with hostname split by dot
// from backwards to find the org level domain.
function getOrgLevelDomain() {
    var try_cookie='try_get_old_cookie=try_get_old_value';
    var hostname = window.location.hostname.split('.');

    var i, h;
    for (i=hostname.length-1; i>=0; i--) {
      h = hostname.slice(i).join('.');
      document.cookie = try_cookie + ';domain=.' + h + ';';
      if (document.cookie.indexOf(try_cookie) > -1) {
        // we are able to store cookie.
        document.cookie = try_cookie.split('=')[0] + '=;domain=.' + h + ';expires=Thu, 01 Jan 1970 00:00:01 GMT;';
        return h;
      }
    }

    // return complete domain name try fails.
    return window.location.hostname;
}

// Sets the cookie. Replaces if exists already.
function set(name, value, days) {
    let expires = "";
    if (days) {
        let date = new Date();
        date.setTime(date.getTime() + (days*24*60*60*1000));
        expires = "; expires=" + date.toUTCString();
    }
    
    let orgDomain = getOrgLevelDomain();
    document.cookie = name + "=" + (value || "")  + expires + "; domain=" + orgDomain + "; path=/";
}

// Gets cookie by its name.
function get(name) {
    let nameEQ = name + "=";
    let ca = document.cookie.split(";");
    for(let i=0;i < ca.length;i++) {
        let c = ca[i];
        while (c.charAt(0)==' ') c = c.substring(1,c.length);
        if (c.indexOf(nameEQ) == 0) return c.substring(nameEQ.length,c.length);
    }
    return null;
}

// Removes the cookie.
function remove(name) {   
    document.cookie = name+"=; Max-Age=-99999999;";  
}

// Checks for existence of the cookie by name.
function isExist(name) {
    return ((document.cookie.indexOf(name+"=") == 0) 
        || (document.cookie.indexOf("; "+name+"=") >= 0));
}

// Set cookie with encoded value. Base64.
function setEncoded(name, value, days) {
    set(name, btoa(value), days);
}

// Get decoded cookie value. Base64.
function getDecoded(name) {
    let value = get(name);
    if (value) value = atob(value);
    return value;
}

// Check Cookie consent
function isEnabled() {
    var cookieEnabled = false;
    var testCookieString = 'fa-testcookie';
    if (!cookieEnabled){ 
        set(testCookieString, null);
        cookieEnabled = isExist(testCookieString);
        cookieEnabled && remove(testCookieString);
    }
    return cookieEnabled;
}

module.exports = exports =  { set, get, setEncoded, getDecoded, remove, isExist, isEnabled };