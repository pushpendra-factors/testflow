const { App } = require('@slack/bolt');
const { WebClient } = require('@slack/web-api');
const { createEventAdapter } = require('@slack/events-api');
const { config } = require('dotenv');

config();
// const authorizeFn = async (installation) => {
//     return {
//       botToken: process.env.SLACK_BOT_TOKEN,
//     //   botId: "xxx",
//     //   botUserId: "xxx",
//     };
//   };

  const app=new App({
      token: process.env.SLACK_BOT_TOKEN,
      signingSecret:process.env.SIGNING_SECRET
  });

// const app = new App({
//     //authorize:authorizeFn,
//     signingSecret: process.env.SIGNING_SECRET,
//     clientId: process.env.CLIENT_ID,
//     clientSecret: process.env.CLIENT_SECRET,
//     //token:process.env.SLACK_BOT_TOKEN,
//     appToken:process.env.SLACK_APP_TOKEN,
    
//     stateSecret: '',
//     scopes: ['channels:read', 'groups:read', 'channels:manage', 'chat:write', 'incoming-webhook','app_mentions:read','channels:history','commands','files:read','files:write','groups:history','im:history','im:read','incoming-webhook','mpim:history','reactions:read','reactions:write'],
//     installationStore: {
//       storeInstallation: async (installation) => {
//         // change the line below so it saves to your database
//         if (installation.isEnterpriseInstall && installation.enterprise !== undefined) {
//           // support for org wide app installation
//           return await database.set(installation.enterprise.id, installation);
//         }
//         if (installation.team !== undefined) {
//           // single team app installation
//           return await database.set(installation.team.id, installation);
//         }
//         throw new Error('Failed saving installation data to installationStore');
//       },
//       fetchInstallation: async (installQuery) => {
//         // change the line below so it fetches from your database
//         if (installQuery.isEnterpriseInstall && installQuery.enterpriseId !== undefined) {
//           // org wide app installation lookup
//           return await database.get(installQuery.enterpriseId);
//         }
//         if (installQuery.teamId !== undefined) {
//           // single team app installation lookup
//           return await database.get(installQuery.teamId);
//         }
//         throw new Error('Failed fetching installation');
//       },
//       deleteInstallation: async (installQuery) => {
//         // change the line below so it deletes from your database
//         if (installQuery.isEnterpriseInstall && installQuery.enterpriseId !== undefined) {
//           // org wide app installation deletion
//           return await database.delete(installQuery.enterpriseId);
//         }
//         if (installQuery.teamId !== undefined) {
//           // single team app installation deletion
//           return await database.delete(installQuery.teamId);
//         }
//         throw new Error('Failed to delete installation');
//       },
//     },
//   });

  const slackClient = new WebClient(process.env.SLACK_BOT_TOKEN);
  const channelId = process.env.CHANNEL_ID;

  app.message('hello', async ({ message, say }) => {
    // say() sends a message to the channel where the event was triggered
    var arr = [];

    var obj = {
        "type": "section",
        "text": {
            "type": "mrkdwn",
            "text": `id: ${json1[0].id}\n Type: ${json1[0].type}\n Title: ${json1[0].title}\n Created By: ${json1[0].created_by_name}\n date: ${json1[0].created_at}`
        },
        "accessory": {
            "type": "button",
            "style": "danger",
            "text": {
                "type": "plain_text",
                "text": "Click Me"
            },
            "action_id": "button_click"
        }
    }
    arr.push(obj);

    const cTable = require('console.table');
    const table = cTable.getTable([
        {
            Title: json1[0].title,
            Date: json1[0].created_at,
            Id: json1[0].id,
            CreatedBy: json1[0].created_by_name
        }, {
            Title: json1[1].title,
            Date: json1[1].created_at,
            Id: json1[1].id,
            CreatedBy: json1[0].created_by_name
        },
        {
            Title: json1[2].title,
            Date: json1[2].created_at,
            Id: json1[2].id,
            CreatedBy: json1[0].created_by_name
        }
    ]);



    await say({
        // blocks: arr,
        // text: `Hey there <@${message.user}>!`,
        "attachments": [
            {
                "mrkdwn_in": ["text"],
                "color": "#36a64f",
                // "pretext": "Optional pre-text that appears above the attachment block",
                "text": 'HERE IS YOUR TABLE: : \n ```' + table + '```',

                //"Optional *`text`* that `appears within` the \n>attachment\n> \n ```Hiiii``` ```Hello```",

            }
        ]
    });


});
app.action('button_click', async ({ body, ack, say }) => {
    // Acknowledge the action
    await ack();
    // await say(`Hey there`);
});




//const channelId = process.env.CHANNEL_ID;

const fetch = require("node-fetch");
var myHeaders = new fetch.Headers();
myHeaders.append("authority", "api.factors.ai");
myHeaders.append("sec-ch-ua", "\"Google Chrome\";v=\"89\", \"Chromium\";v=\"89\", \";Not A Brand\";v=\"99\"");
myHeaders.append("sec-ch-ua-mobile", "?0");
myHeaders.append("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_2_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.114 Safari/537.36");
myHeaders.append("content-type", "application/json");
myHeaders.append("accept", "*/*");
myHeaders.append("origin", "https://app.factors.ai");
myHeaders.append("sec-fetch-site", "same-site");
myHeaders.append("sec-fetch-mode", "cors");
myHeaders.append("sec-fetch-dest", "empty");
myHeaders.append("referer", "https://app.factors.ai/");
myHeaders.append("accept-language", "en-GB,en-US;q=0.9,en;q=0.8");
myHeaders.append("cookie", "_fuid=ZTgzMGNhY2MtZGZjZS00ODAwLWFmMjgtYzY3NDBiZDY1ZTBh; intercom-id-rvffkuu7=42563bf0-0a8f-4d7f-9206-c78d12da8b9f; __insp_uid=3315439817; __insp_nv=false; __insp_wid=1994835818; __insp_targlpu=aHR0cHM6Ly90dWZ0ZS1wcm9kLmZhY3RvcnMuYWkvYW5hbHlzZQ%3D%3D; __insp_targlpt=RmFjdG9yc0FJ; __insp_sid=3637986842; __insp_pad=3; __insp_slim=1613625805760; _ga=GA1.1.1006357273.1615287660; hubspotutk=fd020dcaaca8a927cdcadddbf096f31a; messagesUtk=cd5a6425c4d54856bbe0f2aae6e8f809; _ga_ZM08VH2CGN=GS1.1.1617693918.6.0.1617693918.0; __hstc=3500975.fd020dcaaca8a927cdcadddbf096f31a.1616149361639.1616168812248.1617693920508.3; __hssrc=1; intercom-session-rvffkuu7=OU5nY0lGZWF3QzgvTUx4QS9Fb01Ec3o4T20ycVBpMjN3b2hscnYySXdTVGtEd1lyZXpRc1NPWHZYMlhEeUtyMi0tZ0xyeGxpUWM0U2F4VzNTc2gwbG52UT09--79701c268966f6db11c99f64aa12e40816cb9e76; factors-sid=eyJhdSI6IjFiOTkzYTFiLThhNzYtNGRkZC1hODI3LWZhM2M1NzliYThiOSIsInBmIjoiTVRZeE9UazBNemN3TjN4aGRVWnFVVkJ4V0RVNGJqTkNhWFJ2WnpWMGQxaDJZbUZEYVdaV05VNXFXR2g0VlcxaWNrdEhjMXB2YVRCVWNIVTVVMU5KUkcxWmExRnFOM00yY1ZFNU5qYzJSak5NYkRaTGNGVnNOMUJNZEZKT2F6MTh6VElUSmtiY3VMMk9rUUxUb2pLQ1lGc0F6X1Rwa05nSEkyYmwyNm11NHhZPSJ9");

var raw = JSON.stringify({
    "query_group": [
        {
            "cl": "events",
            "ty": "events_occurrence",
            "fr": 1619375400,
            "to": 1619461799,
            "ewp": [
                {
                    "na": "$session",
                    "pr": []
                }
            ],
            "gbt": "hour",
            "gbp": [],
            "ec": "each_given_event",
            "tz": "Asia/Kolkata"
        },
        {
            "cl": "events",
            "ty": "events_occurrence",
            "fr": 1619375400,
            "to": 1619461799,
            "ewp": [
                {
                    "na": "$session",
                    "pr": []
                }
            ],
            "gbt": "",
            "gbp": [],
            "ec": "each_given_event",
            "tz": "Asia/Kolkata"
        }
    ]
});

var requestOptions = {
    method: 'POST',
    headers: myHeaders,
    body: raw,
    redirect: 'follow'
};


var raw1;
var requestOptions1;

var res;
//var txt1;

//var fs = require('fs');
const { query, json } = require('express');

let j = 0;
let k = 1;

//.result_group.length
async function convertToCSV(json, txt1, type) {
    console.log(json);
    console.log(json.length);

    if (json.result_group || json.length >= 1 || json.result) {

        var n;
        if (json.result_group) {
            n = json.result_group.length;
            for (let i = 0; i < n; i++) {

                if (type == "funnel") {
                    var header = [];
                    header = funnelMethods.generateFunnelHeaders(json.result_group[i]);
                    console.log(header);

                    var csv = [];
                    csv = funnelMethods.generateFunnelData(json.result_group[i]);
                    console.log(csv);

                    const replacer = (key, value) => value === null ? '' : value

                    let cs = csv.map(row => head.map(fieldName =>
                        JSON.stringify(row[fieldName], replacer)).join(','))
                    cs.unshift(header.join(','));
                    cs = cs.join('\r\n');
                    j = i + 1;
                    var fileName = txt1 + `${j}.csv`;

                    await tabularForm1(json.result_group[i], txt1, "");

                    // fs.writeFile(fileName, cs, async function (err) {
                    //     if (err) throw err;
                    //     console.log('Saved!');
                    //     try {
                    //         // Call the files.upload method using the WebClient
                    //         k = k + 1;
                    //         console.log(k);
                    //         const result = await slackClient.files.upload({
                    //             // channels can be a list of one to many strings
                    //             channels: channelId,
                    //             initial_comment: `Result for Query - "${txt1}"`,
                    //             file: fs.readFileSync(txt1 + `${k}.csv`),
                    //             filename:  txt1 + `${k}.csv`,
                    //         });

                    //     }
                    //     catch (error) {
                    //         console.error(error);
                    //     }
                    // });


                }
                else {
                    const header = json.result_group[i].headers;
                    let csv = json.result_group[i].rows;
                    //  let csv1 = json.result_group[i].rows[0][0];
                    console.log(csv);
                    console.log(header);

                    const head = Object.keys(header);

                    const replacer = (key, value) => value === null ? '' : value

                    let cs = csv.map(row => head.map(fieldName =>
                        JSON.stringify(row[fieldName], replacer)).join(','))
                    cs.unshift(header.join(','));
                    cs = cs.join('\r\n');
                    j = i + 1;
                    var fileName = txt1 + `${j}.csv`;
                    await tabularForm1(json.result_group[i], txt1, "");

                    // fs.writeFile(fileName, cs, async function (err) {
                    //     if (err) throw err;
                    //     console.log('Saved!');
                    //     try {
                    //         // Call the files.upload method using the WebClient
                    //         k = k + 1;
                    //         console.log(k);
                    //         const result = await slackClient.files.upload({
                    //             // channels can be a list of one to many strings
                    //             channels: channelId,
                    //             initial_comment: `Result for Query - "${txt1}"`,
                    //             file: fs.readFileSync(txt1 + `${k}.csv`),
                    //             filename: txt1 + `${k}.csv`,
                    //         });

                    //     }
                    //     catch (error) {
                    //         console.error(error);
                    //     }
                    // })


                }

            }

        }

        else if (json.result.result_group) {
            n = json.result.result_group.length;
            for (let i = 0; i < n; i++) {

                const header = json.result.result_group[i].headers;
                let csv = json.result.result_group[i].rows;
                //  let csv1 = json.result_group[i].rows[0][0];
                console.log(csv);
                console.log(header);

                const head = Object.keys(header);

                const replacer = (key, value) => value === null ? '' : value

                let cs = csv.map(row => head.map(fieldName =>
                    JSON.stringify(row[fieldName], replacer)).join(','))
                cs.unshift(header.join(','));
                cs = cs.join('\r\n');
                j = i + 1;
                var fileName = txt1 + `${j}.csv`;
                await tabularForm1(json.result.result_group[i], txt1, "");

                // fs.writeFile(fileName, cs, async function (err) {
                //     if (err) throw err;
                //     console.log('Saved!');
                //     try {
                //         // Call the files.upload method using the WebClient
                //         k = k + 1;
                //         console.log(k);
                //         const result = await slackClient.files.upload({
                //             // channels can be a list of one to many strings
                //             channels: channelId,
                //             initial_comment: `Result for Query - "${txt1}"`,
                //             file: fs.readFileSync(txt1 + `${k}.csv`),
                //             filename: txt1 + `${k}.csv`,
                //         });

                //     }
                //     catch (error) {
                //         console.error(error);
                //     }
                // })
            }
        }

        else
            n = json.length;

    }
    else {
        if (type == "funnel") {
            var header = [];
            header = funnelMethods.generateFunnelHeaders(json);
            console.log(header);

            var csv = [];
            csv = funnelMethods.generateFunnelData(json);
            console.log(csv);

            const head = Object.keys(header);
            const replacer = (key, value) => value === null ? '' : value

            let cs = csv.map(row => head.map(fieldName =>
                JSON.stringify(row[fieldName], replacer)).join(','))
            cs.unshift(header.join(','));
            cs = cs.join('\r\n');
            // j = i + 1;
            var fileName = txt1 + `.csv`;
            await tabularForm1(json, txt1, "funnel");

            // fs.writeFile(fileName, cs, async function (err) {
            //     if (err) throw err;
            //     console.log('Saved!');
            //     try {
            //         // Call the files.upload method using the WebClient
            //         k = k + 1;
            //         console.log(k);
            //         const result = await slackClient.files.upload({
            //             // channels can be a list of one to many strings
            //             channels: channelId,
            //             initial_comment: `Result for Query - "${txt1}"`,
            //             file: fs.readFileSync(txt1 + `.csv`),
            //             filename: txt1 + `.csv`,
            //         });

            //     }
            //     catch (error) {
            //         console.error(error);
            //     }
            // });


        }
        else {
            tabularForm1(json, txt1, "");
            const header = json.headers;
            let csv = json.rows;
            // let csv1 = json.rows[0][0];
            console.log(csv);
            console.log(header);

            const head = Object.keys(header);

            const replacer = (key, value) => value === null ? '' : value

            let cs = csv.map(row => head.map(fieldName =>
                JSON.stringify(row[fieldName], replacer)).join(','))
            cs.unshift(header.join(','));
            cs = cs.join('\r\n');


            var fileName = txt1 + `.csv`;
            // await fs.writeFile(fileName, cs, async function (err) {
            //     if (err) throw err;
            //     console.log('Saved!');
            //     try {
            //         // Call the files.upload method using the WebClient

            //         console.log(k);
            //         const result = await slackClient.files.upload({
            //             // channels can be a list of one to many strings
            //             channels: channelId,
            //             initial_comment: `Result for Query - "${txt1}"`,
            //             file: fs.readFileSync(txt1 + `.csv`),
            //             filename: txt1 + `.csv`,
            //         });

            //     }
            //     catch (error) {
            //         console.error(error);
            //     }
            // });
        }
    }
}


async function tabularForm1(json, txt1, type) {
    var header, csv;
    if (type == "funnel") {
        header = [];
        header = funnelMethods.generateFunnelHeaders(json);
        csv = [];
        csv = funnelMethods.generateFunnelData(json);
    }
    else {
        header = json.headers;
        console.log(header.length);
        csv = json.rows;
    }
    var len = csv.length < 5 ? csv.length : 5;


    const cTable = require('console.table');

    var arr = [];
    for (var i = 0; i < len; i++) {
        var obj = {};
        for (var j = 0; j < header.length; j++) {
            var key = header[j];
            obj[key] = csv[i][j];
        }

        arr.push(obj);
    }

    const table = cTable.getTable(arr);
    console.log(table);

    var text = table.replace(/ /g, '\u00A0');
    // text = table.replace(/ /t, '\u00A0');

    const textToImage = require('text-to-image');

    num = json.headers.length * 290;

    const dataUri = await textToImage.generate(text, {
        // debug: true,
        maxWidth: num,
        fontSize: 18,
        fontFamily: 'monospace',
        lineHeight: 30,
        margin: 5,

        // bgColor: "blue",
        //  textColor: "red"
    });

    let imgAsBase64 = dataUri.substring(dataUri.indexOf(',') + 1);

    //console.log(imgAsBase64);
    // require('fs').writeFileSync(txt1 + '.png', imgAsBase64, 'base64', (err) => {
    //     console.log(err);
    // });
    var fileName = txt1 + '.png';
    try {
        // Call the files.upload method using the WebClient
        const result = await slackClient.files.upload({
            // channels can be a list of one to many strings
            channels: channelId,
            // initial_comment: "Here\'s my file :smile:",
            // Include your filename in a ReadStream here
           // file: fs.createReadStream(fileName)
        });

        console.log(result);
    }
    catch (error) {
        console.error(error);
    }


}


var stopWords = ["a", "about", "above", "after", "again", "against", "all", "am", "an", "and", "any", "are", "aren't", "as", "at", "be", "because", "been", "before", "being", "below", "between", "both", "but", "by", "can't", "cannot", "could", "couldn't", "did", "didn't", "do", "does", "doesn't", "doing", "don't", "down", "during", "each", "few", "for", "from", "further", "had", "hadn't", "has", "hasn't", "have", "haven't", "having", "he", "he'd", "he'll", "he's", "her", "here", "here's", "hers", "herself", "him", "himself", "his", "how", "how's", "i", "i'd", "i'll", "i'm", "i've", "if", "in", "into", "is", "isn't", "it", "it's", "its", "itself", "let's", "me", "more", "most", "mustn't", "my", "myself", "no", "nor", "not", "of", "off", "on", "once", "only", "or", "other", "ought", "our", "ours", "ourselves", "out", "over", "own", "same", "shan't", "she", "she'd", "she'll", "she's", "should", "shouldn't", "so", "some", "such", "than", "that", "that's", "the", "their", "theirs", "them", "themselves", "then", "there", "there's", "these", "they", "they'd", "they'll", "they're", "they've", "this", "those", "through", "to", "too", "under", "until", "up", "very", "was", "wasn't", "we", "we'd", "we'll", "we're", "we've", "were", "weren't", "what", "what's", "when", "when's", "where", "where's", "which", "while", "who", "who's", "whom", "why", "why's", "with", "won't", "would", "wouldn't", "you", "you'd", "you'll", "you're", "you've", "your", "yours", "yourself", "yourselves"];

function remove_stopwords(str) {
    res = []
    words = str.split(' ')
    for (i = 0; i < words.length; i++) {
        word_clean = words[i].split(".").join("")
        if (!stopWords.includes(word_clean)) {
            res.push(word_clean)
        }
    }
    return (res)
}

function removeDuplicates(data) {
    return [...new Set(data)]
}

var myHeaders1 = new fetch.Headers();
myHeaders1.append("authority", "api.factors.ai");
myHeaders1.append("sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\"");
myHeaders1.append("sec-ch-ua-mobile", "?0");
myHeaders1.append("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36");
myHeaders1.append("content-type", "application/json");
myHeaders1.append("accept", "*/*");
myHeaders1.append("origin", "https://app.factors.ai");
myHeaders1.append("sec-fetch-site", "same-site");
myHeaders1.append("sec-fetch-mode", "cors");
myHeaders1.append("sec-fetch-dest", "empty");
myHeaders1.append("referer", "https://app.factors.ai/");
myHeaders1.append("accept-language", "en-GB,en-US;q=0.9,en;q=0.8");
myHeaders1.append("cookie", "_fuid=ZTgzMGNhY2MtZGZjZS00ODAwLWFmMjgtYzY3NDBiZDY1ZTBh; intercom-id-rvffkuu7=42563bf0-0a8f-4d7f-9206-c78d12da8b9f; __insp_uid=3315439817; __insp_nv=false; __insp_wid=1994835818; __insp_targlpu=aHR0cHM6Ly90dWZ0ZS1wcm9kLmZhY3RvcnMuYWkvYW5hbHlzZQ%3D%3D; __insp_targlpt=RmFjdG9yc0FJ; __insp_sid=3637986842; __insp_pad=3; __insp_slim=1613625805760; _ga=GA1.1.1006357273.1615287660; hubspotutk=fd020dcaaca8a927cdcadddbf096f31a; messagesUtk=cd5a6425c4d54856bbe0f2aae6e8f809; _ga_ZM08VH2CGN=GS1.1.1617693918.6.0.1617693918.0; __hstc=3500975.fd020dcaaca8a927cdcadddbf096f31a.1616149361639.1616168812248.1617693920508.3; __hssrc=1; factors-sid=eyJhdSI6IjFiOTkzYTFiLThhNzYtNGRkZC1hODI3LWZhM2M1NzliYThiOSIsInBmIjoiTVRZeE9UazBNemN3TjN4aGRVWnFVVkJ4V0RVNGJqTkNhWFJ2WnpWMGQxaDJZbUZEYVdaV05VNXFXR2g0VlcxaWNrdEhjMXB2YVRCVWNIVTVVMU5KUkcxWmExRnFOM00yY1ZFNU5qYzJSak5NYkRaTGNGVnNOMUJNZEZKT2F6MTh6VElUSmtiY3VMMk9rUUxUb2pLQ1lGc0F6X1Rwa05nSEkyYmwyNm11NHhZPSJ9; intercom-session-rvffkuu7=eHVVVVRUN2JxODA1SUtyZDBQQ0N2TFVvQm9uUmJ3ekZneVZ1a2EzbnFxOFEzZlRvNkk3Mlg2L0x3L2VFdEFTYS0tbDZ2QmNGY3JqbTc4UU5sMkVNZWcrQT09--a83e2de47486b1a463d4401d953eb027898aa645");

var arrOp1 = [];
app.command("/search-queries", async ({ command,
    ack }) => {
    console.log(command)
    await ack()

    channel_id = command.channel_id
    user_id = command.user_id
    trigger_id = command.trigger_id
    var text1 = command.text.toLowerCase();
    console.log(text1);

    var res = remove_stopwords(text1);
    var res1 = removeDuplicates(res);
    console.log(res1);

    let map = new Map();

    for (var i = 0; i < json1.length; i++) {
        for (var j = 0; j < res1.length; j++) {
            if (json1[i].title.toLowerCase().includes(res1[j])) {
                if (map.has(i)) {
                    map.set(i, map.get(i) + 1);
                }
                else
                    map.set(i, 1);
            }
        }

    }

    const mapSort1 = new Map([...map.entries()].sort((a, b) => b[1] - a[1]));
    console.log(mapSort1.size);

    arr2 = [];
    arrOp1 = [];
    ///arrOp=[];
    var y = 0;

    for (const [i, value] of mapSort1.entries()) {


        console.log(arr2.length);
        if (arr2.length == 10) {
            console.log("yes");
            arrOp1.push(arr2);
            arr2 = [];
            console.log(arrOp1);
            
            y = y + 10
        }
        else {

            var obj = {
                "value": `${i}`,
                "text": {
                    "type": "mrkdwn",
                    "text": `*${json1[i].title.substring(0, 70)}*\nid: ${json1[i].id} | Type: ${json1[i].type} | Created By: ${json1[i].created_by_name} | ${json1[i].created_at.substring(0, 10)}`
                }

            }

            arr2.push(obj);
            //y++;
        }


    }
    console.log(arrOp1);
    console.log(arr2);
    if (arr2 != []) {
        arrOp1.push(arr2);
        y = y + arr2.length;
    }
    console.log(arrOp1.length);
    console.log(arr2);

    if (mapSort1.size = 0 || res1 == '') {

        const { channel, ts } = await slackClient.chat.postEphemeral({
            channel: channelId,
            user: user_id,
            text: "Please try with a valid query name"
        });

    }

    else {

        var arrBlocks = [];
        var z = 0;
        for (var i = 0; i < arrOp1.length; i++) {
            var cur = z + 1;
            if (y > (z + 10))
                z = z + 10;
            else
                z = y;
            var obj = {
                "block_id": `blk${i}`,

                "type": "input",
                "element": {

                    "type": "checkboxes",
                    "options": arrOp1[i],
                    "action_id": "input",
                },
                "label": {
                    "type": "plain_text",
                    "text": `Queries ${cur} to ${z}`,
                    "emoji": true
                },
                "optional": true


            }
            arrBlocks.push(obj);
        }

        await slackClient.views.open({
            trigger_id,
            view: {
                "type": "modal",
                "callback_id": "form_modal5",
                "title": {
                    "type": "plain_text",
                    "text": "List of Queries",
                    "emoji": true
                },
                "submit": {
                    "type": "plain_text",
                    "text": "Fetch",
                    "emoji": true
                },
                "close": {
                    "type": "plain_text",
                    "text": "Cancel",
                    "emoji": true
                },
                "blocks":
                    arrBlocks,
            }
        })
    }
});
app.view('form_modal5', async ({ ack, body, view, client }) => {
    await ack();
    const { user } = body;

    console.log(arrOp1.length);
    var vals = [];
    for (var i = 0; i < arrOp1.length; i++) {
        var blk = `blk${i}`;

        console.log(view.state.values[blk].input.selected_options.length);
        for (var j = 0; j < view.state.values[blk].input.selected_options.length; j++) {
            vals.push(view.state.values[blk].input.selected_options[j].value);
        }
    }
    console.log(vals);

    for (var i = 0; i < vals.length; i++) {
        console.log(json1[vals[i]]);
        if ((json1[vals[i]].query.cl) && (json1[vals[i]].query.query_group)) {
            console.log("QGrp with Channel");
            var myHeaders = new fetch.Headers();
            myHeaders.append("authority", "staging-api.factors.ai");
            myHeaders.append("sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\"");
            myHeaders.append("sec-ch-ua-mobile", "?0");
            myHeaders.append("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36");
            myHeaders.append("content-type", "application/json");
            myHeaders.append("accept", "*/*");
            myHeaders.append("origin", "https://staging-app.factors.ai");
            myHeaders.append("sec-fetch-site", "same-site");
            myHeaders.append("sec-fetch-mode", "cors");
            myHeaders.append("sec-fetch-dest", "empty");
            myHeaders.append("referer", "https://staging-app.factors.ai/");
            myHeaders.append("accept-language", "en-GB,en-US;q=0.9,en;q=0.8");
            myHeaders.append("cookie", "_fuid=OGJjZTliODAtNTFhNy00ZGVlLWJjY2ItZjgyYWVhMDQ0MGRj; intercom-id-rvffkuu7=6b713a2d-ba5a-42e6-aa4a-fc423fc72328; factors-sids=eyJhdSI6ImY3MmIyODc4LTJhYjAtNGZjZC05ODNhLWZjOWEwN2E3OTBhOCIsInBmIjoiTVRZeU1USXpOVFV5TTN4Rk0zRkxPVjl5WW5GcU1uaFdhMFZJYTJ4TFozQnJaR3BZZW1GbmJsQmZia0ZOVkZCSFVURlBRVTV2U0VKa2VYcFJaVlpFWTI5WE9UUjFVVTVQUmtGTGVGTTBNRTFZZFZGdE5sQlBjMFl0V0djek5EMThPOW9yejRxbHhMdnZEVkp3Z2QwemQyYTZGVFNSTGRqRUhOblJRWDd6MTRJPSJ9; intercom-session-rvffkuu7=MExxdFhIR2RHRkM0emJsRnRXSEswOE40MlRBNUtscmFoTVJscmVBaklSOCtOcExOT3lDalJuQlM5R084MXhlcC0tTzhLN2FrKzNKdVVPRXliaHdlejlJZz09--3747bb11c8fbd01c5ee52991b87b69b610f22248");
            query_group = json1[vals[i]].query.query_group;
            var cl = json1[vals[i]].query.cl;
            console.log(query_group);
            raw1 = JSON.stringify({ "query_group": query_group, "cl": cl });
            console.log(raw1);
            requestOptions1 = {
                method: 'POST',
                headers: myHeaders,
                body: raw1,
                redirect: 'follow'
            };
            console.log(json1[vals[i]].title);
            var txt1 = json1[vals[i]].title;
            fetch("https://staging-api.factors.ai/projects/51/v1/channels/query", requestOptions1)
                .then(response => response.text())
                .then(result => {
                    console.log(result)
                    res = result;
                    k = 0;
                    convertToCSV(JSON.parse(res), txt1, "");


                })
                .catch(error => console.log('error', error));

        }
        else if (json1[vals[i]].query.query_group) {
            var myHeaders = new fetch.Headers();
            myHeaders.append("authority", "staging-api.factors.ai");
            myHeaders.append("sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\"");
            myHeaders.append("sec-ch-ua-mobile", "?0");
            myHeaders.append("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36");
            myHeaders.append("content-type", "application/json");
            myHeaders.append("accept", "*/*");
            myHeaders.append("origin", "https://staging-app.factors.ai");
            myHeaders.append("sec-fetch-site", "same-site");
            myHeaders.append("sec-fetch-mode", "cors");
            myHeaders.append("sec-fetch-dest", "empty");
            myHeaders.append("referer", "https://staging-app.factors.ai/");
            myHeaders.append("accept-language", "en-GB,en-US;q=0.9,en;q=0.8");
            myHeaders.append("cookie", "_fuid=OGJjZTliODAtNTFhNy00ZGVlLWJjY2ItZjgyYWVhMDQ0MGRj; intercom-id-rvffkuu7=6b713a2d-ba5a-42e6-aa4a-fc423fc72328; factors-sids=eyJhdSI6ImY3MmIyODc4LTJhYjAtNGZjZC05ODNhLWZjOWEwN2E3OTBhOCIsInBmIjoiTVRZeU1USXpOVFV5TTN4Rk0zRkxPVjl5WW5GcU1uaFdhMFZJYTJ4TFozQnJaR3BZZW1GbmJsQmZia0ZOVkZCSFVURlBRVTV2U0VKa2VYcFJaVlpFWTI5WE9UUjFVVTVQUmtGTGVGTTBNRTFZZFZGdE5sQlBjMFl0V0djek5EMThPOW9yejRxbHhMdnZEVkp3Z2QwemQyYTZGVFNSTGRqRUhOblJRWDd6MTRJPSJ9; intercom-session-rvffkuu7=MExxdFhIR2RHRkM0emJsRnRXSEswOE40MlRBNUtscmFoTVJscmVBaklSOCtOcExOT3lDalJuQlM5R084MXhlcC0tTzhLN2FrKzNKdVVPRXliaHdlejlJZz09--3747bb11c8fbd01c5ee52991b87b69b610f22248");
            console.log("QGrp");
            query_group = json1[vals[i]].query.query_group;
            console.log(query_group);
            raw1 = JSON.stringify({ "query_group": query_group });
            console.log(raw1);
            requestOptions1 = {
                method: 'POST',
                headers: myHeaders,
                body: raw1,
                redirect: 'follow'
            };

            console.log(json1[vals[i]].title);
           var  txt1 = json1[vals[i]].title;
            fetch("https://staging-api.factors.ai/projects/51/v1/query", requestOptions1)
                .then(response => response.text())
                .then(result => {
                    console.log(result)
                    res = result;
                    k = 0;
                    convertToCSV(JSON.parse(res), txt1, "");


                })
                .catch(error => console.log('error', error));


        }
        else if (json1[vals[i]].query.query) {
            var myHeaders = new fetch.Headers();
            myHeaders.append("authority", "staging-api.factors.ai");
            myHeaders.append("sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\"");
            myHeaders.append("sec-ch-ua-mobile", "?0");
            myHeaders.append("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36");
            myHeaders.append("content-type", "application/json");
            myHeaders.append("accept", "*/*");
            myHeaders.append("origin", "https://staging-app.factors.ai");
            myHeaders.append("sec-fetch-site", "same-site");
            myHeaders.append("sec-fetch-mode", "cors");
            myHeaders.append("sec-fetch-dest", "empty");
            myHeaders.append("referer", "https://staging-app.factors.ai/");
            myHeaders.append("accept-language", "en-GB,en-US;q=0.9,en;q=0.8");
            myHeaders.append("cookie", "_fuid=OGJjZTliODAtNTFhNy00ZGVlLWJjY2ItZjgyYWVhMDQ0MGRj; intercom-id-rvffkuu7=6b713a2d-ba5a-42e6-aa4a-fc423fc72328; factors-sids=eyJhdSI6ImY3MmIyODc4LTJhYjAtNGZjZC05ODNhLWZjOWEwN2E3OTBhOCIsInBmIjoiTVRZeU1USXpOVFV5TTN4Rk0zRkxPVjl5WW5GcU1uaFdhMFZJYTJ4TFozQnJaR3BZZW1GbmJsQmZia0ZOVkZCSFVURlBRVTV2U0VKa2VYcFJaVlpFWTI5WE9UUjFVVTVQUmtGTGVGTTBNRTFZZFZGdE5sQlBjMFl0V0djek5EMThPOW9yejRxbHhMdnZEVkp3Z2QwemQyYTZGVFNSTGRqRUhOblJRWDd6MTRJPSJ9; intercom-session-rvffkuu7=ZnhpdExNeXBJcWdaR3BndzAzekdWaHFRQkJiYTBmUVdqOGtkSHVjcEtDSWNjblBRNURUNDJtU2dWRFR0dXJsYi0tNzREaGdNU1RmdTBHOThLRnA0c3dKdz09--a19ccf403eabc5d34d8d992c9357afbfe32e7fd7");

            query1 = json1[vals[i]].query;
            console.log(query1);
            raw1 = JSON.stringify(query1);
            console.log(raw1);
            requestOptions1 = {
                method: 'POST',
                headers: myHeaders,
                body: raw1,
                redirect: 'follow'
            };
            console.log(json1[vals[i]].title);
            var txt1 = json1[vals[i]].title;

            fetch("https://staging-api.factors.ai/projects/51/attribution/query", requestOptions1)
                .then(response => response.text())
                .then(result => {
                    console.log(result)
                    res = result;
                    k = 0;
                    convertToCSV(JSON.parse(res), txt1, "");


                })
                .catch(error => console.log('error', error));

        }
        else {
            console.log("ONly Query");
            var myHeaders = new fetch.Headers();
            myHeaders.append("authority", "staging-api.factors.ai");
            myHeaders.append("sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\"");
            myHeaders.append("sec-ch-ua-mobile", "?0");
            myHeaders.append("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36");
            myHeaders.append("content-type", "application/json");
            myHeaders.append("accept", "*/*");
            myHeaders.append("origin", "https://staging-app.factors.ai");
            myHeaders.append("sec-fetch-site", "same-site");
            myHeaders.append("sec-fetch-mode", "cors");
            myHeaders.append("sec-fetch-dest", "empty");
            myHeaders.append("referer", "https://staging-app.factors.ai/");
            myHeaders.append("accept-language", "en-GB,en-US;q=0.9,en;q=0.8");
            myHeaders.append("cookie", "_fuid=OGJjZTliODAtNTFhNy00ZGVlLWJjY2ItZjgyYWVhMDQ0MGRj; intercom-id-rvffkuu7=6b713a2d-ba5a-42e6-aa4a-fc423fc72328; factors-sids=eyJhdSI6ImY3MmIyODc4LTJhYjAtNGZjZC05ODNhLWZjOWEwN2E3OTBhOCIsInBmIjoiTVRZeU1USXpOVFV5TTN4Rk0zRkxPVjl5WW5GcU1uaFdhMFZJYTJ4TFozQnJaR3BZZW1GbmJsQmZia0ZOVkZCSFVURlBRVTV2U0VKa2VYcFJaVlpFWTI5WE9UUjFVVTVQUmtGTGVGTTBNRTFZZFZGdE5sQlBjMFl0V0djek5EMThPOW9yejRxbHhMdnZEVkp3Z2QwemQyYTZGVFNSTGRqRUhOblJRWDd6MTRJPSJ9; intercom-session-rvffkuu7=MExxdFhIR2RHRkM0emJsRnRXSEswOE40MlRBNUtscmFoTVJscmVBaklSOCtOcExOT3lDalJuQlM5R084MXhlcC0tTzhLN2FrKzNKdVVPRXliaHdlejlJZz09--3747bb11c8fbd01c5ee52991b87b69b610f22248");

            query1 = json1[vals[i]].query;//.query;
            console.log(query1);
            raw1 = JSON.stringify({ "query": query1 });

            requestOptions1 = {
                method: 'POST',
                headers: myHeaders,
                body: raw1,
                redirect: 'follow'
            };
            console.log(json1[vals[i]].title);
            var txt1 = json1[vals[i]].title;

            fetch("https://staging-api.factors.ai/projects/51/query", requestOptions1)
                .then(response => response.text())
                .then(result => {
                    console.log(result)
                    res = result;
                    k = 0;
                    convertToCSV(JSON.parse(res), txt1, "");


                })
                .catch(error => console.log('error', error));
        }


    }
});



app.command("/wl", async ({
    command,
    ack
}) => {
    console.log(command)
    await ack()

    channel_id = command.channel_id
    user_id = command.user_id

    await app.client.chat.postEphemeral({
        token: process.env.SLACK_BOT_TOKEN,
        channel: channel_id,
        user: user_id,
        text: "hi"
    });
});



app.shortcut('form_modal', async ({ ack, shortcut, client }) => {
    await ack();
    const { user, trigger_id } = shortcut;

    await client.views.open({
        trigger_id,
        view: {
            "type": "modal",
            "callback_id": "form_modal",
            "title": {
                "type": "plain_text",
                "text": "My App",
                "emoji": true
            },
            "submit": {
                "type": "plain_text",
                "text": "Submit",
                "emoji": true
            },
            "close": {
                "type": "plain_text",
                "text": "Cancel",
                "emoji": true
            },
            "blocks": [
                {
                    "block_id": "project_name",
                    "type": "input",
                    "element": {
                        "type": "plain_text_input",
                        "action_id": "input"
                    },
                    "label": {
                        "type": "plain_text",
                        "text": "Project name",
                        "emoji": true
                    }
                },
                {
                    "block_id": "email",
                    "type": "input",
                    "element": {
                        "type": "plain_text_input",
                        "action_id": "input"
                    },
                    "label": {
                        "type": "plain_text",
                        "text": "Email",
                        "emoji": true
                    }
                },
                {
                    "block_id": "mob_num",
                    "type": "input",
                    "element": {
                        "type": "plain_text_input",
                        "action_id": "input"
                    },
                    "label": {
                        "type": "plain_text",
                        "text": "Mobile Number",
                        "emoji": true
                    }
                }
            ]


        }
    })

});


app.view('form_modal', async ({ ack, body, view, client }) => {
    await ack();
    const { user } = body;
    const {
        project_name,
        email,
        mob_num,
    } = view.state.values;
    console.log(project_name);

    const { channel, ts } = await client.chat.postMessage({
        channel: channelId,
        blocks: [
            {
                type: "section",
                text: {
                    type: "mrkdwn",
                    text: `<@${user.id}>'s project is ${project_name.input.value} `,
                }
            },
            {
                type: "section",
                text: {
                    type: "mrkdwn",
                    text: `<@${user.id}>'s email is ${email.input.value} `,

                }
            },
            {
                type: "section",
                text: {
                    type: "mrkdwn",
                    text: `<@${user.id}>'s Mobile Number is ${mob_num.input.value} `,

                }
            }
        ]
    })
    await client.reactions.add({
        channel,
        name: "one",
        timestamp: ts
    });

});


async function retreiveQueries(json) {
    json1 = json;

}

app.action(action1 = { action_id: `1` }, async ({ body, ack, say }) => {
    await ack();


    //console.log(body.view.blocks[i]);

    console.log(action1);
    //console.log(action_id);
    //const value=body.view['state']['values']['']['button_click1'];
    const value = body.view.blocks[action1.action_id];
    console.log(value);

    const { channel, ts } = await slackClient.chat.postMessage({
        channel: channelId,
        blocks: [
            {
                type: "section",
                text: {
                    type: "mrkdwn",
                    text: `Hi! ${json1[action1.action_id].query.query_group[0].cl} `
                }
            },

        ]
    })
});


app.shortcut('form_modal1', async ({ ack, shortcut, client }) => {
    await ack();
    const { user, trigger_id } = shortcut;
    var arr = [];
    for (var i = 0; i < 50; i++) {
        var obj = {
            "block_id": `blk${i}`,
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": `${json1[i].title}\nid: ${json1[i].id} | Type: ${json1[i].type} | Created By: ${json1[i].created_by_name} | ${json1[i].created_at}`
            },
            "accessory": {
                "type": "button",
                "text": {
                    "type": "plain_text",
                    "text": "Fetch"
                },
                "action_id": `${i}`,
                "value": `${i}`,
            },
        }
        arr.push(obj);
    }
    await client.views.open({
        trigger_id,
        view: {
            "type": "modal",
            "callback_id": "form_modal1",
            "title": {
                "type": "plain_text",
                "text": "Queries",
                "emoji": true
            },

            "blocks": arr,


        }
    })

});


var arrOp = [];
var x = 0;
var y = 0;
var arr2 = [];

app.shortcut('form_modal2', async ({ ack, shortcut, client }) => {
    await ack();
    const { user, trigger_id } = shortcut;
    console.log(json1);

    while (y < json1.length) {
        console.log(x);

        console.log(json1.length);
        var n = y + 10;
        console.log(n);
        if (json1.length < (y + 10)) {
            n = json1.length;
        }
        console.log(n);
        var i = x;
        arr2 = [];

        while (i < n) {
            var str = json1[i].title.substring(0, 70);
            var obj = {
                "value": `${i}`,
                "text": {
                    "type": "mrkdwn",
                    "text": `*${str}*\nid: ${json1[i].id} | Type: ${json1[i].type} | Created By: ${json1[i].created_by_name} | ${json1[i].created_at.substring(0, 10)}`
                }
            }
            i++;
            arr2.push(obj);
        }
        x = x + 10;
        y = y + 10;
        arrOp.push(arr2);
    }

    console.log(arrOp);

    console.log(arrOp.length);// in place of 5

    var arrBlocks = [];
    var z = 0;
    for (var i = 0; i < arrOp.length; i++) {
        var cur = z + 1;
        if (json1.length > (z + 10))
            z = z + 10;
        else
            z = json1.length;
        var obj = {
            "block_id": `blk${i}`,

            "type": "input",
            "element": {
                "type": "checkboxes",
                "options": arrOp[i],
                "action_id": "input",
            },
            "label": {
                "type": "plain_text",
                "text": `Queries ${cur} to ${z}`,
                "emoji": true
            },
            "optional": true

        }
        arrBlocks.push(obj);
    }

    console.log(arrBlocks);
try{
    await client.views.open({
        trigger_id,
        view: {
            "type": "modal",
            "callback_id": "form_modal2",
            "title": {
                "type": "plain_text",
                "text": "List of Queries",
                "emoji": true
            },
            "submit": {
                "type": "plain_text",
                "text": "Fetch",
                "emoji": true
            },
            "close": {
                "type": "plain_text",
                "text": "Cancel",
                "emoji": true
            },
            "blocks": arrBlocks,
        }
    })
}
catch(error) {
    console.error(error);
  }
});

app.view('form_modal2', async ({ ack, body, view, client }) => {
    await ack();
    const { user } = body;

    console.log(arrOp.length);
    var vals = [];
    for (var i = 0; i < arrOp.length; i++) {
        var blk = `blk${i}`;

        console.log(view.state.values[blk].input.selected_options.length);
        for (var j = 0; j < view.state.values[blk].input.selected_options.length; j++) {
            vals.push(view.state.values[blk].input.selected_options[j].value);
        }
    }
    console.log(vals);

    for (var i = 0; i < vals.length; i++) {
        console.log(json1[vals[i]]);
        if ((json1[vals[i]].query.cl) && (json1[vals[i]].query.query_group)) {
            console.log("QGrp with Channel");
            var myHeaders = new fetch.Headers();
            myHeaders.append("authority", "staging-api.factors.ai");
            myHeaders.append("sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\"");
            myHeaders.append("sec-ch-ua-mobile", "?0");
            myHeaders.append("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36");
            myHeaders.append("content-type", "application/json");
            myHeaders.append("accept", "*/*");
            myHeaders.append("origin", "https://staging-app.factors.ai");
            myHeaders.append("sec-fetch-site", "same-site");
            myHeaders.append("sec-fetch-mode", "cors");
            myHeaders.append("sec-fetch-dest", "empty");
            myHeaders.append("referer", "https://staging-app.factors.ai/");
            myHeaders.append("accept-language", "en-GB,en-US;q=0.9,en;q=0.8");
            myHeaders.append("cookie", "_fuid=OGJjZTliODAtNTFhNy00ZGVlLWJjY2ItZjgyYWVhMDQ0MGRj; intercom-id-rvffkuu7=6b713a2d-ba5a-42e6-aa4a-fc423fc72328; factors-sids=eyJhdSI6ImY3MmIyODc4LTJhYjAtNGZjZC05ODNhLWZjOWEwN2E3OTBhOCIsInBmIjoiTVRZeU1USXpOVFV5TTN4Rk0zRkxPVjl5WW5GcU1uaFdhMFZJYTJ4TFozQnJaR3BZZW1GbmJsQmZia0ZOVkZCSFVURlBRVTV2U0VKa2VYcFJaVlpFWTI5WE9UUjFVVTVQUmtGTGVGTTBNRTFZZFZGdE5sQlBjMFl0V0djek5EMThPOW9yejRxbHhMdnZEVkp3Z2QwemQyYTZGVFNSTGRqRUhOblJRWDd6MTRJPSJ9; intercom-session-rvffkuu7=MExxdFhIR2RHRkM0emJsRnRXSEswOE40MlRBNUtscmFoTVJscmVBaklSOCtOcExOT3lDalJuQlM5R084MXhlcC0tTzhLN2FrKzNKdVVPRXliaHdlejlJZz09--3747bb11c8fbd01c5ee52991b87b69b610f22248");
            query_group = json1[vals[i]].query.query_group;
            var cl = json1[vals[i]].query.cl;
            console.log(query_group);
            raw1 = JSON.stringify({ "query_group": query_group, "cl": cl });
            console.log(raw1);
            requestOptions1 = {
                method: 'POST',
                headers: myHeaders,
                body: raw1,
                redirect: 'follow'
            };
            console.log(json1[vals[i]].title);
           var  txt1 = json1[vals[i]].title;
            fetch("https://staging-api.factors.ai/projects/51/v1/channels/query", requestOptions1)
                .then(response => response.text())
                .then(result => {
                    console.log(result)
                    res = result;
                    k = 0;
                    convertToCSV(JSON.parse(res), txt1, "");


                })
                .catch(error => console.log('error', error));

        }
        else if (json1[vals[i]].query.cl == "funnel") {
            console.log("Funnel Query");
            var myHeaders = new fetch.Headers();
            myHeaders.append("authority", "staging-api.factors.ai");
            myHeaders.append("sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\"");
            myHeaders.append("sec-ch-ua-mobile", "?0");
            myHeaders.append("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36");
            myHeaders.append("content-type", "application/json");
            myHeaders.append("accept", "*/*");
            myHeaders.append("origin", "https://staging-app.factors.ai");
            myHeaders.append("sec-fetch-site", "same-site");
            myHeaders.append("sec-fetch-mode", "cors");
            myHeaders.append("sec-fetch-dest", "empty");
            myHeaders.append("referer", "https://staging-app.factors.ai/");
            myHeaders.append("accept-language", "en-GB,en-US;q=0.9,en;q=0.8");
            myHeaders.append("cookie", "_fuid=OGJjZTliODAtNTFhNy00ZGVlLWJjY2ItZjgyYWVhMDQ0MGRj; intercom-id-rvffkuu7=6b713a2d-ba5a-42e6-aa4a-fc423fc72328; factors-sids=eyJhdSI6ImY3MmIyODc4LTJhYjAtNGZjZC05ODNhLWZjOWEwN2E3OTBhOCIsInBmIjoiTVRZeU1USXpOVFV5TTN4Rk0zRkxPVjl5WW5GcU1uaFdhMFZJYTJ4TFozQnJaR3BZZW1GbmJsQmZia0ZOVkZCSFVURlBRVTV2U0VKa2VYcFJaVlpFWTI5WE9UUjFVVTVQUmtGTGVGTTBNRTFZZFZGdE5sQlBjMFl0V0djek5EMThPOW9yejRxbHhMdnZEVkp3Z2QwemQyYTZGVFNSTGRqRUhOblJRWDd6MTRJPSJ9; intercom-session-rvffkuu7=MExxdFhIR2RHRkM0emJsRnRXSEswOE40MlRBNUtscmFoTVJscmVBaklSOCtOcExOT3lDalJuQlM5R084MXhlcC0tTzhLN2FrKzNKdVVPRXliaHdlejlJZz09--3747bb11c8fbd01c5ee52991b87b69b610f22248");

            query1 = json1[vals[i]].query;//.query;
            console.log(query1);
            raw1 = JSON.stringify({ "query": query1 });

            requestOptions1 = {
                method: 'POST',
                headers: myHeaders,
                body: raw1,
                redirect: 'follow'
            };
            console.log(json1[vals[i]].title);
            var txt1 = json1[vals[i]].title;
            var type = "funnel";
            fetch("https://staging-api.factors.ai/projects/51/query", requestOptions1)
                .then(response => response.text())
                .then(result => {
                    console.log(result)
                    res = result;
                    k = 0;
                    convertToCSV(JSON.parse(res), txt1, type);


                })
                .catch(error => console.log('error', error));
        }
        else if (json1[vals[i]].query.query_group) {
            var myHeaders = new fetch.Headers();
            myHeaders.append("authority", "staging-api.factors.ai");
            myHeaders.append("sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\"");
            myHeaders.append("sec-ch-ua-mobile", "?0");
            myHeaders.append("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36");
            myHeaders.append("content-type", "application/json");
            myHeaders.append("accept", "*/*");
            myHeaders.append("origin", "https://staging-app.factors.ai");
            myHeaders.append("sec-fetch-site", "same-site");
            myHeaders.append("sec-fetch-mode", "cors");
            myHeaders.append("sec-fetch-dest", "empty");
            myHeaders.append("referer", "https://staging-app.factors.ai/");
            myHeaders.append("accept-language", "en-GB,en-US;q=0.9,en;q=0.8");
            myHeaders.append("cookie", "_fuid=OGJjZTliODAtNTFhNy00ZGVlLWJjY2ItZjgyYWVhMDQ0MGRj; intercom-id-rvffkuu7=6b713a2d-ba5a-42e6-aa4a-fc423fc72328; factors-sids=eyJhdSI6ImY3MmIyODc4LTJhYjAtNGZjZC05ODNhLWZjOWEwN2E3OTBhOCIsInBmIjoiTVRZeU1USXpOVFV5TTN4Rk0zRkxPVjl5WW5GcU1uaFdhMFZJYTJ4TFozQnJaR3BZZW1GbmJsQmZia0ZOVkZCSFVURlBRVTV2U0VKa2VYcFJaVlpFWTI5WE9UUjFVVTVQUmtGTGVGTTBNRTFZZFZGdE5sQlBjMFl0V0djek5EMThPOW9yejRxbHhMdnZEVkp3Z2QwemQyYTZGVFNSTGRqRUhOblJRWDd6MTRJPSJ9; intercom-session-rvffkuu7=MExxdFhIR2RHRkM0emJsRnRXSEswOE40MlRBNUtscmFoTVJscmVBaklSOCtOcExOT3lDalJuQlM5R084MXhlcC0tTzhLN2FrKzNKdVVPRXliaHdlejlJZz09--3747bb11c8fbd01c5ee52991b87b69b610f22248");
            console.log("QGrp");
            query_group = json1[vals[i]].query.query_group;
            console.log(query_group);
            raw1 = JSON.stringify({ "query_group": query_group });
            console.log(raw1);
            requestOptions1 = {
                method: 'POST',
                headers: myHeaders,
                body: raw1,
                redirect: 'follow'
            };

            console.log(json1[vals[i]].title);
            var txt1 = json1[vals[i]].title;
            fetch("https://staging-api.factors.ai/projects/51/v1/query", requestOptions1)
                .then(response => response.text())
                .then(result => {
                    console.log(result)
                    res = result;
                    k = 0;
                    convertToCSV(JSON.parse(res), txt1, "");


                })
                .catch(error => console.log('error', error));


        }
        else if ((json1[vals[i]].query.query) && (json1[vals[i]].query.query.channel)) {
            console.log("Query with Channel");
            var myHeaders = new fetch.Headers();
            myHeaders.append("authority", "staging-api.factors.ai");
            myHeaders.append("sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\"");
            myHeaders.append("sec-ch-ua-mobile", "?0");
            myHeaders.append("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36");
            myHeaders.append("content-type", "application/json");
            myHeaders.append("accept", "*/*");
            myHeaders.append("origin", "https://staging-app.factors.ai");
            myHeaders.append("sec-fetch-site", "same-site");
            myHeaders.append("sec-fetch-mode", "cors");
            myHeaders.append("sec-fetch-dest", "empty");
            myHeaders.append("referer", "https://staging-app.factors.ai/");
            myHeaders.append("accept-language", "en-GB,en-US;q=0.9,en;q=0.8");
            myHeaders.append("cookie", "_fuid=OGJjZTliODAtNTFhNy00ZGVlLWJjY2ItZjgyYWVhMDQ0MGRj; intercom-id-rvffkuu7=6b713a2d-ba5a-42e6-aa4a-fc423fc72328; factors-sids=eyJhdSI6ImY3MmIyODc4LTJhYjAtNGZjZC05ODNhLWZjOWEwN2E3OTBhOCIsInBmIjoiTVRZeU1USXpOVFV5TTN4Rk0zRkxPVjl5WW5GcU1uaFdhMFZJYTJ4TFozQnJaR3BZZW1GbmJsQmZia0ZOVkZCSFVURlBRVTV2U0VKa2VYcFJaVlpFWTI5WE9UUjFVVTVQUmtGTGVGTTBNRTFZZFZGdE5sQlBjMFl0V0djek5EMThPOW9yejRxbHhMdnZEVkp3Z2QwemQyYTZGVFNSTGRqRUhOblJRWDd6MTRJPSJ9; intercom-session-rvffkuu7=MExxdFhIR2RHRkM0emJsRnRXSEswOE40MlRBNUtscmFoTVJscmVBaklSOCtOcExOT3lDalJuQlM5R084MXhlcC0tTzhLN2FrKzNKdVVPRXliaHdlejlJZz09--3747bb11c8fbd01c5ee52991b87b69b610f22248");
            query1 = json1[vals[i]].query.query;
            var cl = json1[vals[i]].query.cl;
            console.log(query1);
            raw1 = JSON.stringify({ "query": query1, "cl": cl });
            console.log(raw1);
            requestOptions1 = {
                method: 'POST',
                headers: myHeaders,
                body: raw1,
                redirect: 'follow'
            };
            console.log(json1[vals[i]].title);
            var txt1 = json1[vals[i]].title;
            fetch("https://staging-api.factors.ai/projects/51/v1/channels/query", requestOptions1)
                .then(response => response.text())
                .then(result => {
                    console.log(result)
                    res = result;
                    k = 0;
                    convertToCSV(JSON.parse(res), txt1, "");


                })
                .catch(error => console.log('error', error));
        }
        else if (json1[vals[i]].query.query) {
            var myHeaders = new fetch.Headers();
            myHeaders.append("authority", "staging-api.factors.ai");
            myHeaders.append("sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\"");
            myHeaders.append("sec-ch-ua-mobile", "?0");
            myHeaders.append("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36");
            myHeaders.append("content-type", "application/json");
            myHeaders.append("accept", "*/*");
            myHeaders.append("origin", "https://staging-app.factors.ai");
            myHeaders.append("sec-fetch-site", "same-site");
            myHeaders.append("sec-fetch-mode", "cors");
            myHeaders.append("sec-fetch-dest", "empty");
            myHeaders.append("referer", "https://staging-app.factors.ai/");
            myHeaders.append("accept-language", "en-GB,en-US;q=0.9,en;q=0.8");
            myHeaders.append("cookie", "_fuid=OGJjZTliODAtNTFhNy00ZGVlLWJjY2ItZjgyYWVhMDQ0MGRj; intercom-id-rvffkuu7=6b713a2d-ba5a-42e6-aa4a-fc423fc72328; factors-sids=eyJhdSI6ImY3MmIyODc4LTJhYjAtNGZjZC05ODNhLWZjOWEwN2E3OTBhOCIsInBmIjoiTVRZeU1USXpOVFV5TTN4Rk0zRkxPVjl5WW5GcU1uaFdhMFZJYTJ4TFozQnJaR3BZZW1GbmJsQmZia0ZOVkZCSFVURlBRVTV2U0VKa2VYcFJaVlpFWTI5WE9UUjFVVTVQUmtGTGVGTTBNRTFZZFZGdE5sQlBjMFl0V0djek5EMThPOW9yejRxbHhMdnZEVkp3Z2QwemQyYTZGVFNSTGRqRUhOblJRWDd6MTRJPSJ9; intercom-session-rvffkuu7=ZnhpdExNeXBJcWdaR3BndzAzekdWaHFRQkJiYTBmUVdqOGtkSHVjcEtDSWNjblBRNURUNDJtU2dWRFR0dXJsYi0tNzREaGdNU1RmdTBHOThLRnA0c3dKdz09--a19ccf403eabc5d34d8d992c9357afbfe32e7fd7");

            query1 = json1[vals[i]].query;
            console.log(query1);
            raw1 = JSON.stringify(query1);
            console.log(raw1);
            requestOptions1 = {
                method: 'POST',
                headers: myHeaders,
                body: raw1,
                redirect: 'follow'
            };
            console.log(json1[vals[i]].title);
            var txt1 = json1[vals[i]].title;

            fetch("https://staging-api.factors.ai/projects/51/attribution/query", requestOptions1)
                .then(response => response.text())
                .then(result => {
                    console.log(result)
                    res = result;
                    k = 0;
                    convertToCSV(JSON.parse(res), txt1, "");


                })
                .catch(error => console.log('error', error));

        }
        else {
            console.log("Only Query");
            var myHeaders = new fetch.Headers();
            myHeaders.append("authority", "staging-api.factors.ai");
            myHeaders.append("sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\"");
            myHeaders.append("sec-ch-ua-mobile", "?0");
            myHeaders.append("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36");
            myHeaders.append("content-type", "application/json");
            myHeaders.append("accept", "*/*");
            myHeaders.append("origin", "https://staging-app.factors.ai");
            myHeaders.append("sec-fetch-site", "same-site");
            myHeaders.append("sec-fetch-mode", "cors");
            myHeaders.append("sec-fetch-dest", "empty");
            myHeaders.append("referer", "https://staging-app.factors.ai/");
            myHeaders.append("accept-language", "en-GB,en-US;q=0.9,en;q=0.8");
            myHeaders.append("cookie", "_fuid=OGJjZTliODAtNTFhNy00ZGVlLWJjY2ItZjgyYWVhMDQ0MGRj; intercom-id-rvffkuu7=6b713a2d-ba5a-42e6-aa4a-fc423fc72328; factors-sids=eyJhdSI6ImY3MmIyODc4LTJhYjAtNGZjZC05ODNhLWZjOWEwN2E3OTBhOCIsInBmIjoiTVRZeU1USXpOVFV5TTN4Rk0zRkxPVjl5WW5GcU1uaFdhMFZJYTJ4TFozQnJaR3BZZW1GbmJsQmZia0ZOVkZCSFVURlBRVTV2U0VKa2VYcFJaVlpFWTI5WE9UUjFVVTVQUmtGTGVGTTBNRTFZZFZGdE5sQlBjMFl0V0djek5EMThPOW9yejRxbHhMdnZEVkp3Z2QwemQyYTZGVFNSTGRqRUhOblJRWDd6MTRJPSJ9; intercom-session-rvffkuu7=MExxdFhIR2RHRkM0emJsRnRXSEswOE40MlRBNUtscmFoTVJscmVBaklSOCtOcExOT3lDalJuQlM5R084MXhlcC0tTzhLN2FrKzNKdVVPRXliaHdlejlJZz09--3747bb11c8fbd01c5ee52991b87b69b610f22248");

            query1 = json1[vals[i]].query;//.query;
            console.log(query1);
            raw1 = JSON.stringify({ "query": query1 });

            requestOptions1 = {
                method: 'POST',
                headers: myHeaders,
                body: raw1,
                redirect: 'follow'
            };
            console.log(json1[vals[i]].title);
            
            var txt1 = json1[vals[i]].title;

            fetch("https://staging-api.factors.ai/projects/51/query", requestOptions1)
                .then(response => response.text())
                .then(result => async()=>{
                    console.log(result);
                    res = result;
                    k = 0;
                    await convertToCSV(JSON.parse(res), json1[vals[i]].title, "");


                })
                .catch(error => console.log('error', error));
        }


    }

});

var myHeaders = new fetch.Headers();
myHeaders.append("authority", "staging-api.factors.ai");
myHeaders.append("sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\"");
myHeaders.append("sec-ch-ua-mobile", "?0");
myHeaders.append("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36");
myHeaders.append("content-type", "application/json");
myHeaders.append("accept", "*/*");
myHeaders.append("origin", "https://staging-app.factors.ai");
myHeaders.append("sec-fetch-site", "same-site");
myHeaders.append("sec-fetch-mode", "cors");
myHeaders.append("sec-fetch-dest", "empty");
myHeaders.append("referer", "https://staging-app.factors.ai/");
myHeaders.append("accept-language", "en-GB,en-US;q=0.9,en;q=0.8");
myHeaders.append("cookie", "_fuid=OGJjZTliODAtNTFhNy00ZGVlLWJjY2ItZjgyYWVhMDQ0MGRj; intercom-id-rvffkuu7=6b713a2d-ba5a-42e6-aa4a-fc423fc72328; _ga=GA1.1.238961614.1621781340; messagesUtk=ad0efef354284d9c8fa998e6c0e9804e; hubspotutk=ee403cde5d672d63ade8cec1c544656f; __hssrc=1; __hstc=3500975.ee403cde5d672d63ade8cec1c544656f.1621781343910.1622552954796.1622878517150.6; _ga_ZM08VH2CGN=GS1.1.1623322978.7.0.1623322978.0; factors-sids=eyJhdSI6ImY3MmIyODc4LTJhYjAtNGZjZC05ODNhLWZjOWEwN2E3OTBhOCIsInBmIjoiTVRZeU16Z3lOemM0Tm54NloweFhaR3BGWW1sNVJFZ3llSEZ4TmtFeE5IRXdVWGxWV0MxNGRVNTZSVTVaWTJST1JubHBla3R2UjNkbk1IUk1lSHBsTUdaYWNGSm1WMEZxTVRGSmJFOXlTazlVVERsUVNVWkVielJMTWpWeFNUMTgxZWhKc280WTJiMFZtMkNVcW9GUU5wYm5xSWhoUVNXeHExOEJhQlZ3bHR3PSJ9; intercom-session-rvffkuu7=ZS9uVzVwaFpBdG5WWXNOMjI2SHBadmcvaHE5OHZaS1k0UmsxUzhNdGVFUG5MU05YSEd6d1NtRThIMTNDZjJOVy0tK2F0dTJ5aWFoaVliaWJLaUdsNitkQT09--0158b8a67c246ffd64ab3b38e985e3ca38f075fb");

// var myHeaders = new fetch.Headers();
// myHeaders.append("Connection", "keep-alive");
// myHeaders.append("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36");
// myHeaders.append("Content-Type", "application/json");
// myHeaders.append("Accept", "*/*");
// myHeaders.append("Origin", "http://factors-dev.com:3000");
// myHeaders.append("Referer", "http://factors-dev.com:3000/");
// myHeaders.append("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8");
// myHeaders.append("Authorization",process.env.SLACK_APP_TOKEN); //slack_token is appended here.
// //myHeaders.append("slackAuthToken",process.env.SLACK_APP_TOKEN);
// //myHeaders.append("Cookie", "_fuid=ZTE4YzI3N2ItY2MxNy00OTQ2LThmNTAtZTYwOGJkMGY2Njc3; factors-sidd=eyJhdSI6IjRkOGRiMWFmLWVmODctNDZiZS1iM2Y5LTIwOGM2YjkxYjViNyIsInBmIjoiTVRZeU1qRXlOVEUwTW54aFEyTkNjalZSUlVwRGVVaFpka1JIYVRoSWFrOUtSRTVPU2xGaGVWVkVlalJXUnpSak4yTnBTVXhYT1U5UUxXODBXRVUzUVZVeVJsbEdlRTF2TlVKRlRtSnJiRk15VDJ4eGEyZG5XSGg2WDFkUVdUMThEZWNqaFNYVjBjQ050aExoa010RkpSc1JIM3lYZkExa1l0X1N4a1k5aERVPSJ9");

var requestOptions = {
    method: 'GET',
    headers: myHeaders,
    redirect: 'follow'
};


(async () => {
    // Start your app
    await app.start(process.env.PORT || 3001);

    console.log('⚡️ Bolt app is running!');

    fetch("https://staging-api.factors.ai/projects/51/queries", requestOptions)
        .then(response => response.text())
        .then(result => {

           console.log(result);
            retreiveQueries(JSON.parse(result));

        })
        .catch(error => console.log('error', error));


})();

// // add in .env
// // SLACK_BOT_TOKEN=xoxb-965123144369-2182216118050-ELuDdc8LAN1adZc4gOa5pp3I
// // SLACK_APP_TOKEN=xoxp-965123144369-1951223304055-2021747359424-72c6706b41b826ad3dddf780287d6339
// // SIGNING_SECRET=fbacaf9a6d7da40b06efbfbb7bcdc4d8
// // PORT=3001
// // CHANNEL_ID=C02566VSSLV

// require('dotenv').config();
// const express = require('express');
// const request = require('request');
// const PORT = process.env.PORT || 8090;

// const app = express();
// app.use(express.urlencoded({ extended: false }));
// app.use(express.json());

// app.get('/auth/redirect', (req, res) =>{
//     console.log(process.env.CLIENT_ID);
//     console.log(process.env.CLIENT_SECRET);
//     console.log(process.env.REDIRECT_URI);
//     const options = {
//         uri: 'https://slack.com/api/oauth.access?code='
//             +req.query.code+
//             '&client_id='+process.env.CLIENT_ID+
//             '&client_secret='+process.env.CLIENT_SECRET+
//             '&redirect_uri='+process.env.REDIRECT_URI,
//         method: 'GET'
//     };
    
//     request(options, (error, response, body) => {
//         const JSONresponse = JSON.parse(body);
        
//         if (!JSONresponse.ok){
//             console.log(JSONresponse);
//             res.send("Error encountered: \n" + JSON.stringify(JSONresponse)).status(200).end();
//         } else {
//             console.log(JSONresponse);
//             res.send("Success!");
//         }
//     });
// });

// app.listen(PORT, function() {
//   console.log(`App listening on PORT: ${PORT}`);
// });