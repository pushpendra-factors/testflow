import * as Request from './utils/request';
import config from "./config";

const URI_TRACK = "/sdk/event/track";
const URI_IDENTIFY = "/sdk/user/identify";

class APIClient {
    constructor(token) {
        this.token = token;
    }

    getURL(uri) {
        return config.api.host+uri;
    }

    track(userId, eventName, eventProperties={}) {
        eventName = eventName.trim()
        if(!eventName || eventName == "")
            return Promise.reject("Failed Tracking. Event name is missing");

        let payload = { "event_name": eventName, "properties": eventProperties };
        if(userId && userId != null && userId != undefined) payload["user_id"] = userId;

        let customHeaders = { "Authorization": this.token };
        return Request.post(
            this.getURL(URI_TRACK),
            payload,
            customHeaders
        );
    }

    identify(userId, customerUserId) {
        let payload = {};

        if(userId && userId != null && userId != undefined) 
            payload["userId"] = userId;

        if(customerUserId && customerUserId != null && customerUserId != undefined) 
            payload["c_uid"] = customerUserId;

        let customHeaders = { "Authorization": this.token };
        return Request.post(
            this.getURL(URI_IDENTIFY),
            payload,
            customHeaders
        );
    }
}

export default APIClient;