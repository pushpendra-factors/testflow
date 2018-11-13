# Web Form with API Gateway, Lamda and SNS

## Setup SNS
* Open SNS on web console -> Choose topics -> Create new topic -> Give a topic name and display name and save.
* Select your topic check box -> click Actions -> click Subscribe to topic -> Choose protocol as 'Email', fill Endpoint with your email and save.
* Accept through subscription link sent to your email.
* Test: Select your topic and click on 'Publish to topic'.

## Setup Lambda for sending message through SNS
* Open Lambda on web console -> Click create function -> Select 'Aws Serverless Application Repository.'
* Search for 'step-functions-send-to-sns' and click the result.
* Fill the 'Application settings' ->  TopicNameParameter with 'YOUR-TOPIC-NAME' and click deploy. This will add permissions and deploy Lambda with the cloud formation.
* Click on the lambda function you created -> Change the code to below.

```python
# lambda_function.py

from __future__ import print_function

import json
import urllib
import boto3

print('Loading function..')

def send_to_sns(message, context):
    sns = boto3.client('sns')
    sns.publish(
        TopicArn="arn:aws:sns:us-east-1:834234938474:WebsiteSignupNotification",
        Subject="An early bird for factors.ai",
        Message="User Email : "+message["email"]
    )
    
    return json.dumps({"subscribed_email": message["email"]}) # Try greping email from log.

```

* Click 'Save' at the top right corner.
* Select 'Configure test event' -> Create new test -> Name your test event 'SignupTest' -> and set test content as below and save.

```json
{ "email": "user@example.com" }
```
* Test: Choose 'SignupTest' and Click test. You would have got an email.

## Setup API Gateway
* Open Api gateway on web console  -> Click on Create API and Select 'New API'.
* Fill API name and description and keep Endpoint Type as Regional.
* Click Actions -> Select Create Resource -> Fill Name, Path and **Select Enable API Gateway CORS** and proceed.
* Click on newly created resource and Select 'Create Methods' from Actions list -> Choose POST.
* Select 'Integration type' as Lambda and Fill the lambda function name on 'Lambda Function' and save.
* Click on Action and choose 'Deploy API', select 'Deployment Stage' as New Stage and fill stage details (Stage name: v1) and save.
* Copy the Invoke URL on the top centre.
* Test: Use the URL for making API calls. Example: `curl -H "Content-Type: application/json" -i -X POST https://fjnvg9a8wi.execute-api.us-east-1.amazonaws.com/v1/signup -d '{"email": "apiuser@gmail.com"}'`




