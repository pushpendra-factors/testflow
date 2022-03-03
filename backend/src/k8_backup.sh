#!/usr/bin/env bash

# get the directory name in CCYYMMDD-HHMMSS format
export directoryName=$(date +%Y_%m_%d)-$(date +%H_%M_%S)
mkdir $directoryName
cd $directoryName
fileExtension=".yaml"
export GCLOUD_PROJECT=$(gcloud config list --format 'value(core.project)' 2>/dev/null)

if [[ "$GCLOUD_PROJECT" == "factors-staging" ]]; then
    kubectl config set-cluster factors-staging --server=http://104.198.3.193:443
elif [[ "$GCLOUD_PROJECT" == "factors-production" ]]; then
    kubectl config set-cluster factors-production-1 --server=http://35.233.227.213:443
else
    echo "ERROR: Invalid GCLOUD PROJECT $GCLOUD_PROJECT."
    exit -1
fi

# following function takes two arguments, complete name of the job (e.g. deployment.apps/appserver)
# and name of workload (e.g. appserver). It then writes the YAML configuration of workload to the 
# respective text file 
function write_to_file() {
    fileFullName="$2$fileExtension"
    kubectl get $1 -o YAML > $fileFullName
}

# strPosition is the index position where the name of workload starts in the job name.
# e.g. for job_name = deployment.apps/appserver, workload_name = appserver starts from index 16.
strPosition=16
for d in $(kubectl get deployments -o name); do
    workload=${d:strPosition:50}
    write $d $workload
done

# strPosition is the index position where the name of workload starts in the job name.
# e.g. for job_name = cronjob.batch/add-session-beam-job, workload_name = add-session-beam-job starts from index 14.
strPosition=14
for d in $(kubectl get cronjobs -o name); do
    workload=${d:strPosition:50}
    write_to_file $d $workload
done

cd ..

# upload the folder to google cloud storage.
if [[ "$GCLOUD_PROJECT" == "factors-staging" ]]; then
    gsutil -m cp -R $directoryName gs://factors-staging-k8-backup
elif [[ "$GCLOUD_PROJECT" == "factors-production" ]]; then
    gsutil -m cp -R $directoryName gs://factors-production-k8-backup
else
    echo "ERROR: Invalid GCLOUD PROJECT $GCLOUD_PROJECT."
    exit -1
fi

# delete the temporary directory after uploading it to cloud storage
rm -r $directoryName

echo "Backup files generated successfully in folder $directoryName."
