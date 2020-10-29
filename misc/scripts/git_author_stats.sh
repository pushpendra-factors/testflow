#!/bin/bash
#Script for extracting lines of code per author for a given time interval.
#USAGE: ./git_author_stats.sh 1.weeks
#optionally send channel_id and app_token of a bot to post it into a slack channel.
#Ref:https://stackoverflow.com/a/7010890

path_to_repo=$1
stats_since=$2
channel_id=$3
app_token=$4

if [[ "${path_to_repo}" != "" ]]; then
  cd "${path_to_repo}"
fi

IFS=$'\n'
authors_arr=($(git shortlog --summary --numbered --since="$stats_since" | cut -f2))
IFS=' '

payload="Author stats for the last $stats_since"
for author in "${authors_arr[@]}"
do
   :
  payload="${payload}\n\n${author}"
  author_lines=$(git log --author="$author" --pretty=tformat: --numstat --since="$stats_since" | awk '{ add += $1; subs += $2; loc += $1 - $2 } END { printf "added_lines: %s,     removed_lines: %s,    total_lines: %s\n", add, subs, loc }' -)
  payload="${payload}\n${author_lines}"
done
payload=`echo "${payload}" | sed 's/"/\\\"/g'`
echo $payload

if [[ "${channel_id}" != "" && ${app_token} != "" ]]; then
  curl -X POST -H "Authorization: Bearer ${app_token}" -H "Content-type: application/json" --data '{"text":"'"${payload}"'", "channel": "'"${channel_id}"'" }' https://slack.com/api/chat.postMessage
fi
