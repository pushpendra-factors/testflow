function validatedStringArg(name, value) {
    if (typeof(value) != "string")
        throw new Error("FactorsArgumentError: Invalid type for "+name);
    
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

module.exports = exports =  { 
    validatedStringArg: validatedStringArg,
    convertIfNumber: convertIfNumber,
    getCleanHash: getCleanHash,
};

