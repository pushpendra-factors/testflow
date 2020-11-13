#!/bin/bash
#Script for extracting event_name:count from chunks. Runs on chunks.
#USAGE: ./script.sh gs://factors-production-v2/projects/427/models/1594551106399/chunks/ 1 2 gs://factors-misc/script/hippovideo/

if [ "$#" -ne 4 ]; then
  echo "Error: Illegal number of parameters"
  echo "Usage: ./script.sh gs://factors-production-v2/projects/427/models/1594551106399/chunks/ 1 2 gs://factors-misc/script/hippovideo/ "
fi

cloud_location=$1
chunk_start=$2
chunk_end=$3
upload_loc=$4

chunk_names=()
for i in $(seq "$chunk_start" "$chunk_end");
do
  # shellcheck disable=SC2027
  file_name="chunk_"$i".txt"
  echo "$file_name"
  # shellcheck disable=SC2206
  chunk_names+=($file_name)
done

final_join_file=result_chunk_"$chunk_start"_"$chunk_end".txt

for i in "${chunk_names[@]}"
do
  #echo $i
  gsutil cat "$cloud_location""$i" |grep "\"ouc\"" | grep "\"en\"" | grep -oE "\"ouc\":[0-9]*|\"en\":\[\".[a-zA-Z0-9].*\"\]" | sed -e s/"\"en\":\["//g | sed -e s/"\]"//g | sed -e s/"\"ouc\""//g > en_ouc_report_"$i".txt
  paste -sd "\0\n" en_ouc_report_"$i".txt > good_report_"$i".txt
  gsutil cp good_report_"$i".txt "$upload_loc"
  cat good_report_"$i".txt >> "$final_join_file"
  rm en_ouc_report_"$i".txt
  #rm good_report_"$i".txt
done
# Upload the final
gsutil cp "$final_join_file" "$upload_loc"
