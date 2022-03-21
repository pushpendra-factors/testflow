window.factors = window.factors||(function(){
    this.q = [];
    var factorsQueuedEvent = new CustomEvent('FACTORS_QUEUED_EVENT');
    var addToQueue = function(k,a) {
        this.q.push({k,a});
        window.dispatchEvent(factorsQueuedEvent);   
    }
    this.track = function(ev, ep, ac) {
        addToQueue('track', arguments);  
    }
    this.init = function(at, op, ac) {
        this.TOKEN = at;
        this.INIT_PARAMS = op;
        this.INIT_CALLBACK = ac;
    }
    this.reset = function() {
        addToQueue('reset', arguments);
    }
    this.page = function(ac, frc) {
        addToQueue('page', arguments);
    }
    this.updateEventProperties = function(evI, pr) {
        addToQueue('updateEventProperties', arguments);
    }
    this.identify = function(cus, uP) {
        addToQueue('identify', arguments);
    }
    this.addUserProperties = function(pr){
        addToQueue('addUserProperties', arguments);
    }
    this.getUserId = function(){
        addToQueue('getUserId', arguments);
    }
    this.call = function(){
        var callMap = {k: '', a: []}
        if(arguments && arguments.length >= 1){
            for(var i=1;i<arguments.length;i++) {
                callMap.a.push(arguments[i]);
            }
            callMap.k = arguments[0];
        }
        this.q.push(callMap);
        window.dispatchEvent(factorsQueuedEvent);
    }
    this.init("${projectToken}", {'param': 'asd'}, function(){console.log("Init callback")});
    return this;
})();

(function() {
    var s = document.createElement("script");
    s.type = "text/javascript";
    s.src = "${assetURL}";
    s.async = true;
    d = document.getElementsByTagName('script')[0];
    document.head.insertBefore(s, d);
})()