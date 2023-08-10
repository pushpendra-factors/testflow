"use strict";

const cacheWindowKey = "FAITRACKER_CACHE";

function getFAITrackerCache(key) { 
    if (!window[cacheWindowKey]) window[cacheWindowKey]={}; 
    return window[cacheWindowKey][key];
}

function setFAITrackerCache(key, value) {
    if (!window[cacheWindowKey]) window[cacheWindowKey]={};
    window[cacheWindowKey][key] = value;
}

function getFAITrackerCacheObject() {
    if (!window[cacheWindowKey]) window[cacheWindowKey]={};
    return window[cacheWindowKey];
}

module.exports = {
    getFAITrackerCache: getFAITrackerCache,
    setFAITrackerCache: setFAITrackerCache,
    getFAITrackerCacheObject: getFAITrackerCacheObject,

    // List of faitracker cache keys.
    currentPageURLEventName: "currentPageURLEventName",
    currentPageTrackEventId: "currentPageTrackEventId",
    currentPageOriginalURL: "currentPageOriginalURL",
    currentPageSpentTimeInMs: "currentPageSpentTimeInMs",
    lastPollerId: "lastPollerId",
    trackPageOnSPA: "trackPageOnSPA",
    lastActivityTime: "lastActivityTime",
    lastPageProperties: "lastPageProperties",
}