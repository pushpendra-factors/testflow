import config from './config';
import * as Request from "./request";
import APIClient from "./api"

function isInstalled() {
    console.log(config);
    return "Factors sdk v0.1 is installed!";
}

function init(token, config) {}

function track(eventName, eventProperties) {}

function identify(userId) {}

function addUserProperties() {}

export { isInstalled, init, track, identify, addUserProperties };