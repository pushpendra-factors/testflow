function validatedStringArg(name, value) {
    if (typeof(value) != "string")
        throw new Error("FactorsArgumentError: Invalid "+name);
    
    value = value.trim();
    if (!value) throw new Error("FactorsArgumentError: "+name+" cannot be empty.");
    
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
        var key = "factors-test";
        var value = "test";
        window.localStorage.setItem(key, value);
        var isAvailable = window.localStorage.getItem(key) == value;
        window.localStorage.removeItem(key);
        return isAvailable;
    } catch(e) {
        return false;
    }
}

module.exports = exports =  {
    validatedStringArg: validatedStringArg,
    convertIfNumber: convertIfNumber,
    getCleanHash: getCleanHash,
    parseURLString: parseURLString,
    getCurrentUnixTimestampInMs: getCurrentUnixTimestampInMs,
    isLocalStorageAvailable: isLocalStorageAvailable,
};

