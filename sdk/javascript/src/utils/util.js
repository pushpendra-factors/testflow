function validatedStringArg(name, value) {
    if (typeof(value) != "string")
        throw new Error("FactorsArgumentError: Invalid type for "+name);
    
    value = value.trim();
    if (!value) throw new Error("FactorsArgumentError: "+name+" cannot be empty.");
    
    return value;
}

module.exports = exports =  { validatedStringArg: validatedStringArg };

