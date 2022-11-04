import { InfoCircleFilled, InfoCircleOutlined } from "@ant-design/icons";
import { notification } from "antd";
import React, { useEffect, useState } from "react";
const AdBlockerDetector = (props)=>{

    let [isAdBlocker,setIsAdBlocker] = useState(false);
    useEffect(()=>{


        // Detects Crystal Adblocker
        // Method 3
        (function detectAdblockWithInvalidURL(callback) { 
            var flaggedURL = 'https://pagead2.googlesyndication.com/pagead/js/adsbygoogle.js';

            if (window.fetch) {
                var request = new Request(flaggedURL, {
                    method: 'HEAD',
                    mode: 'no-cors',
                });
                fetch(request)
                .then(function(response) {
                    if (response.status === 404) {
                        callback(false,response);
                    }
                })
                .catch(function(error) {
                    callback(true,error);
                });
            } else {
                var http = new XMLHttpRequest();
                http.open('HEAD', flaggedURL, false);

                try {
                    http.send();
                } catch (err) {
                    callback(true,err);
                }

                if (http.status === 404) {
                    callback(false,http);
                }
            }
        })(function(usingAdblock,res) {
            // returns if adblocker is there or not

            if(isAdBlocker === false){
                console.log("METHOD 3", usingAdblock,res)
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