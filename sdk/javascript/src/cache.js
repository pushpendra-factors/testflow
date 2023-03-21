"use strict";

const cacheWindowKey = "FACTORS_CACHE";

function getFactorsCache(key) { 
    if (!window[cacheWindowKey]) window[cacheWindowKey]={}; 
    return window[cacheWindowKey][key];
}

function setFactorsCache(key, value) {
    if (!window[cacheWindowKey]) window[cacheWindowKey]={};
    window[cacheWindowKey][key] = value;
}

function getFactorsCacheObject() {
    if (!window[cacheWindowKey]) window[cacheWindowKey]={};
    return window[cacheWindowKey];
}

module.exports = {
    getFactorsCache: getFactorsCache,
    setFactorsCache: setFactorsCache,
    getFactorsCacheObject: getFactorsCacheObject,

    // List of factors cache keys.
    currentPageURLEventName: "currentPageURLEventName",
    currentPageTrackEventId: "currentPageTrackEventId",
    currentPageOriginalURL: "currentPageOriginalURL",
    currentPageSpentTimeInMs: "currentPageSpentTimeInMs",
    lastPollerId: "lastPollerId",
    trackPageOnSPA: "trackPageOnSPA",
    lastActivityTime: "lastActivityTime",
    lastPageProperties: "lastPageProperties",
}