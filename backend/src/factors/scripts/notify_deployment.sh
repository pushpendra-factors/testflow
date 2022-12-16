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
#     CHANNEL_TOKEN has to be set in ~/.profile or ~/.bashrc or passed explicitly to the 'make' command.
# 
# Sample Makefile target to set IMAGE_NAME and call this script before deployment:
#     export CHANNEL_TOKEN
#     pack-dashboard-caching: export IMAGE_NAME=dashboard-caching-job
#     pack-dashboard-caching: notify-deployment
#         docker build -t us.gcr.io/factors-$(ENV)/dashboard-caching-job:$(TAG) -f Dockerfile.dashboard_caching_job .
#
#     notify-deployment:
#         $(GOPATH)/src/factors/scripts/notify_deployment.sh

if [[ "${ENV}" == "staging" ]]; then
    CHANNEL_TOKEN="TUD3M48AV/B01J5U9TT2P/TdUBfGLSD3OVUdxA25l7s5Bh"
    PROJECT_ID="factors-staging"
elif [[ "${ENV}" == "production" ]]; then
    CHANNEL_TOKEN="TUD3M48AV/B01AH1YR5JP/9UXlvfv511KdEI5mOsQkMWYi"
    PROJECT_ID="factors-production"
else
    echo "ERROR: Invalid environment ${ENV}."
    exit -1
fi

if [[ "${IMAGE_NAME}" == "" || ${CHANNEL_TOKEN} == "" ]]; then
    echo "ERROR: Value for IMAGE_NAME or CHANNEL_TOKEN can not be empty."
    exit -1
fi

echo "Pulling all tags from repository ..."
all_image_tags=`gcloud container images list-tags us.gcr.io/${PROJECT_ID}/${IMAGE_NAME}`

echo "Fetching the latest tag ..."
latest_tag=`echo "${all_image_tags}" | head -2 | tail -1 | tr -s "[:blank:]" | cut -d' ' -f2`
if [[ "${latest_tag}" == "" ]]; then
    echo "No existing image tag found. Skipping notification."
    exit # Non error exit.
fi

# Get the commit id from the tag. Works when separated with - or _.
# Will work even if the tag is suffixed with PR number or any other suffix.
commit_id=`echo "${latest_tag}" | cut -d'-' -f2 | cut -d'_' -f2`

# Get the commit history.
commit_history=`git log | grep -B10000 "commit ${commit_id}" | sed '$d' | grep -v '^[[:space:]]*$' | grep -e "^Author" -e "^Date" -e "^  " | sed 's/^  /      /g'`

# For an old job being deployed after a long time, git log might not have data for last tag commit id.
# So pick history from the latest commits.
if [[ "${commit_history}" == "" ]]; then
    commit_history=`git log | head -100 | sed '$d' | grep -v '^[[:space:]]*$' | grep -e "^Author" -e "^Date" -e "^  " | sed 's/^  /      /g'`
fi

# Commits after the most recent PR. To capture any hotfixes getting deployed added after PR.
recent_non_pr_commits=`echo "${commit_history}" | grep -m1 -B1000 -e "(#[0-9]\+)$" | sed '$d' | sed '$d' | sed '$d'`

# Highlights captures only pull requests information instead of entire commit history.
# With grep -m3, takes only recent 5 pull requests, otherwise for old image, it will be flooded with PRs.
pr_highlights=`echo "${commit_history}" | grep -m3 -B2 -e "(#[0-9]\+)$"`
highlights="${recent_non_pr_commits}
${pr_highlights}"
if [[ "${highlights}" == "" ]]; then
    # If no pull request information available, use entire commit history.
    highlights="${commit_history}"
    if [[ "${highlights}" == "" ]]; then
        echo "Found no new changes since last deployment."
        exit # Non error exit.
    fi
fi

# Remove any commits of type 'Merge branch ...' to avoid clutterring.
lines_to_delete=`echo "${highlights}" | grep -n -B2 "Merge branch" | cut -d'-' -f1 | cut -d':' -f1`
if [[ "${lines_to_delete}" != "" ]]; then
    lines_to_delete=`echo ${lines_to_delete} | sed 's/ /d;/g'`
    highlights=`echo "${highlights}" | sed "${lines_to_delete}d"`
fi

deployer_email=`git config --list | grep user.email | cut -d'=' -f2`
if [ -z "${deployer_email}" ]; then
    deployer_email=`gcloud config list account --format "value(core.account)" 2> /dev/null`
fi
branch_name=`git rev-parse --abbrev-ref HEAD`

# TODO(prateek): Make alert more rich in terms of tagging and blocks.
echo "Sending alert on slack"
payload="-------------------------------------------------------------
*Deployment initiated for '${IMAGE_NAME}' with tag '${TAG}'. By ${deployer_email} from branch '${branch_name}'*."


# Pick DEV_NAME when running github actions
if [ ! -z "${DEV_NAME}" ]; then
    payload="------------------------------------------------------------- 
    *Deployment initiated for '${IMAGE_NAME}' with tag '${TAG}'. By '${DEV_NAME}' from branch '${branch_name}'*."
fi

# If production, add commit hightlights for the deployment.
if [[ "${ENV}" == "production" ]]; then
    payload=`echo "${payload}\n${highlights}"`
fi

# Escape double quotes from payload.
payload=`echo "${payload}" | sed 's/"/\\\"/g'`

curl -X POST -H 'Content-type: application/json' --data '{"text":"'"${payload}"'", "type": "mrkdwn"}' https://hooks.slack.com/services/${CHANNEL_TOKEN}
echo ""
