// set - Sets the cookie. Replaces if exists already.
function set(name, value, days) {
    let expires = "";
    if (days) {
        let date = new Date();
        date.setTime(date.getTime() + (days*24*60*60*1000));
        expires = "; expires=" + date.toUTCString();
    }
    document.cookie = name + "=" + (value || "")  + expires + "; path=/";
}

function get(name) {
    let nameEQ = name + "=";
    let ca = document.cookie.split(';');
    for(let i=0;i < ca.length;i++) {
        let c = ca[i];
        while (c.charAt(0)==' ') c = c.substring(1,c.length);
        if (c.indexOf(nameEQ) == 0) return c.substring(nameEQ.length,c.length);
    }
    return null;
}

function remove(name) {   
    document.cookie = name+'=; Max-Age=-99999999;';  
}

function isExist(name) {
    return ((document.cookie.indexOf(name+"=") == 0) 
        || (document.cookie.indexOf("; "+name+"=") >= 0));
}


module.exports = exports =  { set, get, remove, isExist };