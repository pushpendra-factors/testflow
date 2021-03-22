import json
import re
from datetime import datetime, timedelta

BUCKET_NAME = 'factors-staging'
START_DATE = 20210315
END_DATE = 20210315
FILE_NAMES = ["campaign_performance_report.csv", "ad_group_performance_report.csv", "ad_performance_report.csv",
              "keyword_performance_report.csv", "click_performance_report.csv", "search_performance_report.csv"]
DRY = True

def get_extract_blobs():
    from google.cloud import storage

    client = storage.Client()
    bucket = client.bucket(BUCKET_NAME)
    blobs = []
    for blob in bucket.list_blobs(prefix='adwords_extract'):
        blobs.append(blob)
    return blobs


def filter_blobs_by_dates(blobs, start_date, end_date):
    date_range = get_dates_between_ranges(start_date, end_date)
    result_regex = ""
    for current_date in date_range:
        current_regex = "(.*{0})|".format(current_date)
        result_regex += current_regex

    result_blobs = []
    for blob in blobs:
        result = re.match(result_regex, blob.name)
        if len(result.group()) > 0:
            result_blobs.append(blob)

    return result_blobs


def filter_blobs_by_type(blobs, file_names):
    result_regex = ""
    for file_name in file_names:
        current_regex = ".*{0}|".format(file_name)
        result_regex += current_regex

    result_blobs = []
    for blob in blobs:
        result = re.match(result_regex, blob.name)
        if len(result.group()) > 0:
            result_blobs.append(blob)

    return result_blobs

def print_blobs(blobs):
    result_names = []
    for blob in blobs:
        result_names.append(blob.name)
    processed_blob_names = json.dumps(result_names, indent=1)
    with open("/var/tmp/blobs.txt", "w+") as writer:
        writer.write(processed_blob_names)


def delete_blobs(blobs):
    for blob in blobs:
        blob.delete()
    return

def get_dates_between_ranges(from_timestamp, to_timestamp):
    date_range = []
    start_timestamp = from_timestamp
    while start_timestamp <= to_timestamp:
        date_range.append(start_timestamp)
        start_timestamp = get_next_day_timestamp(start_timestamp)

    return date_range

def get_next_day_timestamp(timestamp):
    start_datetime = datetime.strptime(str(timestamp), "%Y%m%d")
    return get_timestamp_from_datetime(start_datetime + timedelta(days=1))


def get_timestamp_from_datetime(dt):
    return int(dt.strftime('%Y%m%d'))


if __name__ == "__main__":
    blobs = get_extract_blobs()
    blobs = filter_blobs_by_dates(blobs, START_DATE, END_DATE)
    blobs = filter_blobs_by_type(blobs, FILE_NAMES)
    print_blobs(blobs)
    if not DRY:
        delete_blobs(blobs)
