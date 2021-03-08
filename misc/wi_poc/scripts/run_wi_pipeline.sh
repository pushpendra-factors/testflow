GCSBUCKET=factors-production-v2
S3BUCKET=weekly-insights

PROJECTID=$1
STDATE=$2
MODELID=${3:-}

PROJECT_QUERY="\"pid\":${PROJECTID}"
WK_QUERY="\"mt\":\"w\""
STDATE_QUERY="\"st\":${STDATE}"

if [ -z "$MODELID" ]
then
    echo "Model ID not given as input. Reading metadata..."
    MODELID=$(gsutil cat `gsutil ls gs://$GCSBUCKET/metadata/ | grep -P '[0-9][0-9]+.txt' | tail -1` | grep $PROJECT_QUERY | grep $WK_QUERY | grep $STDATE_QUERY | tail -1 | jq -r '.mid')
fi

if [ -z "$MODELID" ]
then
    echo "Model ID for start-date ${STDATE} not found in metadata."
    echo "Checking the projects directly..."
    MODELID=`gsutil ls gs://$GCSBUCKET/projects/$PROJECTID/models/ | python3 model_parser.py $PROJECTID $STDATE`
fi

if [ -z "$MODELID" ]
then
    echo "Model ID for start-date ${STDATE} not found in projects as well."
else
    echo "Model ID found: ${MODELID}"
    GCSKEY=projects/$PROJECTID/models/$MODELID
    S3KEY=data/cloud_storage/projects/$PROJECTID/models/$MODELID
    FILENAME=events_$MODELID.txt

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
fi
