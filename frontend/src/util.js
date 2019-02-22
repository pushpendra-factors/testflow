export function isStaging() {
    return ENV === "staging";
}

export function getHostURL() {
    let host = BUILD_CONFIG.backend_host;
    return (host[host.length-1] === "/") ? host : host+"/";
}

export function getSDKAssetURL() {
    return isStaging() ? (getHostURL() + "assets/factors.js") : BUILD_CONFIG.sdk_asset_url;
}

export function deepEqual(x, y) {
    return JSON.stringify(x) === JSON.stringify(y);
}