"use strict";

var APIClient = require("./api-client");

function App(token, config={}) {
    this.client = new APIClient(token);
    this.config = config;
}

App.prototype.isInitialized = function() {}

App.prototype.set = function(token, config={}) {
    if(token) this.client.setToken(token);
    this.config = config;
}

App.prototype.reset = function(token=null, config={}) {
    this.client.setToken(token);
    this.config = config;
}

App.prototype.getClient = function() {
    return this.client;
}

module.exports = exports = App;