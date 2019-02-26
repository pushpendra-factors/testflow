export function isStaging() {
    return ENV === "staging";
}

export function getHostURL() {
    let host = BUILD_CONFIG.backend_host;
    return (host[host.length-1] === "/") ? host : host+"/";
}

export function deepEqual(x, y) {
    return JSON.stringify(x) === JSON.stringify(y);
}