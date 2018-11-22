(function(c) {
    var s = document.createElement("script");
    s.type = "text/javascript";
    // Load completion check.
    if (s.readyState){ 
        s.onreadystatechange = function(){
            if (s.readyState == "loaded" || s.readyState == "complete") {
                s.onreadystatechange = null;
                c();
            }
        };
    } else {
        s.onload = function(){ c(); };
    }
    s.src = "/dist/factors.prod.js";
    d = !!document.body ? document.body : document.head;
    d.appendChild(s);
})(function() { factors.init("YOUR_TOKEN"); });
