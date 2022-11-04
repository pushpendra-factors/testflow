import { InfoCircleFilled, InfoCircleOutlined } from "@ant-design/icons";
import { notification } from "antd";
import React, { useEffect, useState } from "react";
const AdBlockerDetector = (props)=>{

    let [isAdBlocker,setIsAdBlocker] = useState(false);
    useEffect(()=>{

            // // Method 1
            // (function detectWithAdsDiv() {
            // var detected = false;

            // const ads = document.createElement('div');
            // ads.innerHTML = '&nbsp;';
            // ads.className = 'adsbox';

            // try {
            //     document.body.appendChild(ads);
            //     var node = document.querySelector('.adsbox');
            //     detected = !node || node.offsetHeight === 0;
            // } finally {
            //     ads.parentNode.removeChild(ads);
            // }

            // console.log('Using divAdblocker: ' + detected);
            // })();




            // Method 2
            // Can detect almost all Ad blockers, but can't detect ghostery  & Adblocker Ultimate are not being detected for now.
            var badURL = 'https://pagead2.googlesyndication.com/pagead/js/adsbygoogle.js';

            (function detectWithScriptTag() {
                var script = window.document.createElement('script');

                script.onload = function() {
                    // if loaded there is no adblocker
                        console.log("METHOD 2 NO AD")
                    script.parentNode.removeChild(script);
                };

                script.onerror = function() {
                    // Adblocker exist
                    setIsAdBlocker(isAdBlocker || true);
                    console.log("METHOD 2")
                        
                    
                }

                script.src = badURL;
                window.document.body.appendChild(script);
            })();




        // Detects Crystal Adblocker
        // Method 3
        (function detectAdblockWithInvalidURL(callback) { 
            var flaggedURL = 'pagead/js/adsbygoogle.js';

            if (window.fetch) {
                var request = new Request(flaggedURL, {
                    method: 'HEAD',
                    mode: 'no-cors',
                });
                fetch(request)
                .then(function(response) {
                    if (response.status === 404) {
                        callback(false);
                    }
                })
                .catch(function(error) {
                    callback(true);
                });
            } else {
                var http = new XMLHttpRequest();
                http.open('HEAD', flaggedURL, false);

                try {
                    http.send();
                } catch (err) {
                    callback(true);
                }

                if (http.status === 404) {
                    callback(false);
                }
            }
        })(function(usingAdblock) {
            // returns if adblocker is there or not

            if(isAdBlocker === false){
                console.log("METHOD 3")
                setIsAdBlocker( isAdBlocker || usingAdblock)
            }
        });

    },[]);


    useEffect(()=>{
        console.log(isAdBlocker,"isAdblocker")
        if(isAdBlocker){
        
            notification.info({
                message: `Hmm, it appears that you're using one or more ad blockers`,
                description:
                  'We understand why! But unfortunately, this makes it tougher to build perfect reports.',
                placement:"topRight",
                style:{
                    height:"155px",
                    width:"452px",
                    display:"grid",
                    placeContent:"center"
                    
                },
                icon:<InfoCircleFilled width={'24px'} style={{color:"red"}} />
            });
        }
    },[isAdBlocker]);
    return <>{props.children}</>;
}
export default AdBlockerDetector;