deployer_email=`gcloud config list account --format "value(core.account)" 2> /dev/null`

# TODO(prateek): Make alert more rich in terms of tagging and blocks.
echo "Sending alert on slack.."
payload="-------------------------------------------------------------
*Deployment initiated for production frontend. By ${deployer_email}*.
"

# Escape double quotes from payload.
payload=`echo "${payload}" | sed 's/"/\\\"/g'`

curl -X POST -H 'Content-type: application/json' --data '{"text":"'"${payload}"'", "type": "mrkdwn"}' https://hooks.slack.com/services/${CHANNEL_TOKEN}
