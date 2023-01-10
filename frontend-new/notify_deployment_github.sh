
if [[ "$1" == "staging" ]]; then
    CHANNEL_TOKEN="TUD3M48AV/B01J5U9TT2P/TdUBfGLSD3OVUdxA25l7s5Bh"
elif [[ "$1" == "production" ]]; then
    CHANNEL_TOKEN="TUD3M48AV/B01AH1YR5JP/9UXlvfv511KdEI5mOsQkMWYi"
else
    echo "ERROR: Environment not passed ."
    exit -1
fi

branch_name=`git branch --show-current`

echo "Sending alert on slack.."
payload="-------------------------------------------------------------
*Deployment initiated for '$1' 'frontend-new' from branch '${branch_name}'. By ${DEV_NAME} Description: '${ACTION_DESCRIPTION}'*.
"

# Escape double quotes from payload.
payload=`echo "${payload}" | sed 's/"/\\\"/g'`

curl -X POST -H 'Content-type: application/json' --data '{"text":"'"${payload}"'", "type": "mrkdwn"}' https://hooks.slack.com/services/${CHANNEL_TOKEN}
