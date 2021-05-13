CHANNEL_TOKEN="TUD3M48AV/B01AH1YR5JP/9UXlvfv511KdEI5mOsQkMWYi"
branch_name=`git branch --show-current`
deployer_email=`git config --list | grep user.email | cut -d'=' -f2`
if [ -z "${deployer_email}" ]; then
    deployer_email=`gcloud config list account --format "value(core.account)" 2> /dev/null`
fi

echo "Sending alert on slack.."
payload="-------------------------------------------------------------
*Deployment initiated for production 'frontend-new' from branch '${branch_name}'. By ${deployer_email}*.
"

# Escape double quotes from payload.
payload=`echo "${payload}" | sed 's/"/\\\"/g'`

curl -X POST -H 'Content-type: application/json' --data '{"text":"'"${payload}"'", "type": "mrkdwn"}' https://hooks.slack.com/services/${CHANNEL_TOKEN}
