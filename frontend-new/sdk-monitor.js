const https = require('https');

const checkSDK = () => {
    let exitCode = 0;
    console.log("!!! Checking SDK on " + deployEnv)
    return https.get(SDK_URL, (response) => {
        if(!response.statusCode === 200) {
            console.log("!!! Response was not 200 Ok. Please check the deployment of the sdk")
            exitCode = 1;
            return exitCode;
        }
        if(!response.headers["content-length"]) {
            console.log("!!! Couldn't access content length, please check the deployment of sdk")
            exitCode = 1;
            return exitCode;
        }
        const contentLength = Number(response.headers["content-length"]);
        if(contentLength == NaN || contentLength <=2000) {
            console.log("!!! Something wrong with the content, please check the deployment of sdk")
            exitCode = 1;
            return exitCode;
        }
    
        console.log("!!! SDK was fetched and had a size greater than 2000B. Looks Good.")
        return exitCode;
    })
}

const main = () => {
    let exitCodes = [];
    if(checkTimes && parseInt(checkTimes) && parseInt(checkTimes) > 1) {
        checkSDK().on('response', (val) => exitCodes.push(val))
        setInterval(() => {
            if(exitCodes.length < parseInt(checkTimes)) {
                checkSDK().on('response', (val) => exitCodes.push(val));
            } else {
                process.exit(exitCodes[exitCodes.length -1]);
            }

        }, 5000)
    } else {
        checkSDK().on('response', (val) => process.exit(val))
    }

}

const deployEnv = process.argv[2];
const checkTimes = process.argv[3] ? process.argv[3] : 1;

const SDK_URL = deployEnv === 'prod' ? 
'https://app.factors.ai/assets/factors.js' : 'https://staging-app.factors.ai/assets/factors.js'


console.log("!!! Starting process, will check in intervals of 5 secs X" + checkTimes + "times");

main();
