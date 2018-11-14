"use strict";

var APIClient = require("./api-client");

function App(token, config={}) {
    this.client = new APIClient(token);
    this.config = config;
}

App.prototype.setToken = function(token, config={}) {
    if(token) this.client.setToken(token);
}

App.prototype.setConfig = function(config) {
    if(config) this.config = config;
}

App.prototype.reset = function(token=null, config={}) {
    this.client.setToken(token);
    this.config = config;
}

App.prototype.getClient = function() {
    return this.client;
}

module.exports = exports = App;