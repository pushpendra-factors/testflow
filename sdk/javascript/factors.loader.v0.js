(function (c) 
    { 
        var s = document.createElement("script"); 
        s.type = "text/javascript"; 
        if (s.readyState) { 
            s.onreadystatechange = function () { 
                        if (s.readyState == "loaded" || s.readyState == "complete") { 
                            s.onreadystatechange = null; c() 
                        } 
                    } 
                } 
        else { 
            s.onload = function () { 
                c() 
            } 
        } 
        s.src = "/dist/factors.prod.js"; 
        s.async = true; 
        d = !!document.body ? document.body : document.head; 
        d.appendChild(s) 
    }
)(function () { factors.init("dummy") })