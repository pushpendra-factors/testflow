function validatedStringArg(name, value) {
    if (typeof(value) != "string")
        throw new Error("FAITrackerArgumentError: Invalid "+name);
    
    value = value.trim();
    if (!value) throw new Error("FAITrackerArgumentError: "+name+" cannot be empty.");
    
    return value;
}

function convertIfNumber(nString) {
    if (!nString.match(/^[+-]?\d+(\.\d+)?$/)) return nString;
    n = Number(nString); // Supports float.
    if (isNaN(n)) return nString;
    return n;
}

function getCleanHash(hash) {
    // excludes query params on hash if any.
    return hash.split("?")[0];
}

function parseURLString(urlString="") {
    if (urlString == "" || !urlString) return {
        host: "", path: "", hash: "",
    };
    
    var tempAnchor = document.createElement("a");
    tempAnchor.setAttribute("href", urlString); 
    return { 
        host: tempAnchor.host,
        path: tempAnchor.pathname,
        hash: tempAnchor.hash,
    }
}

function getCurrentUnixTimestampInMs() {
    return new Date().getTime(); 
}

function isLocalStorageAvailable() {
    if (window.localStorage == undefined) return false;

    try {
        var key = "faitracker-test";
        var value = "test";
        window.localStorage.setItem(key, value);
        var isAvailable = window.localStorage.getItem(key) == value;
        window.localStorage.removeItem(key);
        return isAvailable;
    } catch(e) {
        return false;
    }
}

function encode(str, shift=4) {
    var estr = "";
    for (var i=0; i<str.length; i++) {
        var cat = str[i].charCodeAt();
        var last = 126 - shift;

        if (cat >= 33 && cat <= last) {
            dat = cat + shift
            estr = estr + String.fromCharCode(dat);
        } else if (cat > last && cat <= 126) {
            dat = 32 + (cat % last)
            estr = estr + String.fromCharCode(dat);
        } else {
            estr = estr + str[i];
        }
    }

    return estr;
}

function decode(str, shift=4) {
    var estr = "";
    for (var i=0; i<str.length; i++) {
        var cat = str[i].charCodeAt();
        var shift = 4;
        var first = 33 + shift;

        if (cat >= first && cat <= 126) {
            var dat = cat - shift
            estr = estr + String.fromCharCode(dat);
        } else if (cat < first && cat >= 33) {
            var dat = (cat % 33) + (126 - shift) + 1
            estr = estr + String.fromCharCode(dat);
        } else {
            estr = estr + str[i];
        }
    }

    return estr;
}

module.exports = exports =  {
    validatedStringArg: validatedStringArg,
    convertIfNumber: convertIfNumber,
    getCleanHash: getCleanHash,
    parseURLString: parseURLString,
    getCurrentUnixTimestampInMs: getCurrentUnixTimestampInMs,
    isLocalStorageAvailable: isLocalStorageAvailable,
    encode: encode,
    decode: decode,
};