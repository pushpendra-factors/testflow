#!/bin/bash
# Script to notify of the changes that are going to release.
#
# How it works:
#     1. Pull all the latest tags for the given IMAGE_NAME from Goggle container repository.
#     2. Get the latest tag for the image. Breaks it and extracts commit_id from it.
#     3. Use git log to get the all the commits since commit id used in latest tag.
#     4. Use slack App webhook to post a message from curl to a particular channel.
# 
# CHANNEL_TOKEN for the app is available at https://api.slack.com/apps/A018CF323HS/incoming-webhooks?
#     To enable post to a new channel use 'Add New Webhook to Workspace' on above link.
# 
# Sample Makefile target to set IMAGE_NAME and call this script before deployment:
#     pack-dashboard-caching: export IMAGE_NAME=dashboard-caching-job
#     pack-dashboard-caching: notify-deployment
#         docker build -t us.gcr.io/factors-$(ENV)/dashboard-caching-job:$(TAG) -f Dockerfile.dashboard_caching_job .
#
#     notify-deployment:
#         $(GOPATH)/src/factors/scripts/notify_deployment.sh

if [[ "${ENV}" == "staging" ]]; then
    exit
elif [[ "${IMAGE_NAME}" == "" || ${CHANNEL_TOKEN} == "" ]]; then
    echo "ERROR: Value for IMAGE_NAME or CHANNEL_TOKEN can not be empty."
    exit -1
fi

PROJECT_ID="factors-production"

echo "Pulling all tags from repository ..."
all_image_tags=`gcloud container images list-tags us.gcr.io/${PROJECT_ID}/${IMAGE_NAME}`

echo "Fetching the latest tag ..."
latest_tag=`echo "${all_image_tags}" | head -2 | tail -1 | tr -s "[:blank:]" | cut -d' ' -f2`
if [[ "${latest_tag}" == "" ]]; then
    echo "No existing image tag found. Skipping notification."
    exit # Non error exit.
fi

# Get the commit id from the tag.
# Will work even if the tag is suffixed with PR number or any other suffix.
commit_id=`echo "${latest_tag}" | cut -d'-' -f2`

# Get the commit history.
commit_history=`git log | grep -B10000 "commit ${commit_id}" | sed '$d' | grep -v '^[[:space:]]*$' | grep -e "^Author" -e "^Date" -e "^  " | sed 's/^  /      /g'`

# Highlights captures only pull requests information instead of entire commit history.
highlights=`echo "${commit_history}" | grep -B2 -e "(#[0-9]\+)$"`
if [[ "${highlights}" == "" ]]; then
    # If no pull request information available, use entire commit history.
    highlights="${commit_history}"
    if [[ "${highlights}" == "" ]]; then
        echo "Found no new changes since last deployment."
        exit # Non error exit.
    fi
fi

# TODO(prateek): Make alert more rich in terms of tagging and blocks.
echo "Sending alert on slack"
payload="-------------------------------------------------------------
*Deploying ${IMAGE_NAME}*.
${highlights}"

# Escape double quotes from payload.
payload=`echo "${payload}" | sed 's/"/\\\"/g'`

curl -X POST -H 'Content-type: application/json' --data '{"text":"'"${payload}"'", "type": "mrkdwn"}' https://hooks.slack.com/services/${CHANNEL_TOKEN}
