export function isStaging() {
    return ENV === "staging";
}

export function getHostURL() {
    // Use window.origin as host for staging to support subdomain login.
    let host = isStaging() ? window.origin : BUILD_CONFIG.backend_host;
    // Usable URL.
    return (host[host.length-1] === "/") ? host : host+"/";
}

export function getSDKAssetURL() {
    return isStaging() ? (getHostURL() + "assets/factors.js") : BUILD_CONFIG.sdk_asset_url;
}

