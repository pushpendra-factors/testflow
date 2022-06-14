const sanitizeInputString = (inputString) => { 
    let regex = /[^a-zA-Z0-9-_ ]/g //allow only characters, numbers, space, - hyphens, _ underscores
    return inputString.replace(regex, '') 
}

export default sanitizeInputString