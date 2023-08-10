window.faitracker = window.faitracker||(function(){
    this.q = [];
    var qEvent = new CustomEvent('FAITRACKER_QUEUED_EVENT');
    this.init = function(at, op, ac) {
        this.TOKEN = at;
        this.INIT_PARAMS = op;
        this.INIT_CALLBACK = ac;
        window.dispatchEvent(new CustomEvent('FAITRACKER_INIT_EVENT'))
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
        window.dispatchEvent(qEvent);
    }
    this.message = function(){
        window.addEventListener('message', function(e) {
            if(e.data.origin === 'faitracker'){
                this.call('message', e.data.type, e.data.message);
            }
        });
    }
    this.message();
    this.init("${projectToken}", {host: "${apiURL}"});
    return this;
})(),
(function() {
    var s = document.createElement("script");
    s.type = "text/javascript";
    s.src = "${assetURL}";
    s.async = true;
    d = document.getElementsByTagName('script')[0];
    d.parentNode.insertBefore(s, d);
})();