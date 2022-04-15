GCSBUCKET=factors-production-v2
S3BUCKET=data-driven-attribution

PROJECTID=$1
MODE=$2
STDATE=$3

GCSKEY=projects/$PROJECTID/events/$MODE/$STDATE
S3KEY=data/cloud_storage/projects/$PROJECTID/events/$MODE/$STDATE
FILENAME=events.txt

echo "Creating a folder on S3..."
aws s3api put-object --bucket $S3BUCKET --key $S3KEY/

echo "Copying ${FILENAME} from GCS to local..."
gsutil cp gs://$GCSBUCKET/$GCSKEY/$FILENAME /tmp/$FILENAME

echo "Counting lines..."
NLINES=`wc -l /tmp/$FILENAME`
echo "# Lines: ${NLINES}"

echo "Copying ${FILENAME} from local to S3..."
aws s3 cp /tmp/$FILENAME s3://$S3BUCKET/$S3KEY/$FILENAME

echo "Removing ${FILENAME} from local..."
rm /tmp/$FILENAME
