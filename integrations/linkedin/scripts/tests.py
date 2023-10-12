from util import Util
from transformations import DataTransformation
from datetime import datetime
from _datetime import timedelta

def test_get_timestamp_range():
    doc_type = "campaign"
    sync_info_with_type = {
        doc_type: "2023-05-31"
    }
    end_timestamp = "20230610"
    # case when end timestamp given
    timestamps, errMsg = Util.get_timestamp_range(doc_type, sync_info_with_type, end_timestamp)
    assert errMsg == ""
    assert len(timestamps) == 10
    assert timestamps[0] == "20230601"
    assert timestamps[len(timestamps)-1] == "20230610"

    # case when range exceeding MAX_LOOKBACK and number of timestamps sliced to MAX_LOOKBACK
    end_timestamp = "20230810"
    timestamps, errMsg = Util.get_timestamp_range(doc_type, sync_info_with_type, end_timestamp)
    assert errMsg != ""
    assert len(timestamps) == 30
    assert timestamps[0] == "20230712"
    assert timestamps[len(timestamps)-1] == "20230810"

    # case when end timestamp not given
    last_sync_timestamp = (datetime.now() - timedelta(days=5)).date()
    sync_info_with_type[doc_type] = str(last_sync_timestamp)
    timestamps, errMsg = Util.get_timestamp_range(doc_type, sync_info_with_type, None)
    assert errMsg == ""
    assert len(timestamps) == 4

def test_get_timestamp_chunks_to_be_backfilled():
    last_backfill_timestamp = (datetime.now() - timedelta(days=35)).date().strftime("%Y%m%d")
    timestamp_chunks = Util.get_timestamp_chunks_to_be_backfilled(last_backfill_timestamp)
    for i in range(len(timestamp_chunks)):
        assert len(timestamp_chunks[i]) <= 7

        # for verifying that the end timestamp is always a sunday
        if i == len(timestamp_chunks)-1:
            end_timestamp = timestamp_chunks[i][len(timestamp_chunks[i])-1]
            end_datetime = datetime.strptime(str(end_timestamp), '%Y%m%d')
            end_day_of_week = end_datetime.isoweekday()
            assert end_day_of_week == 7 # 7 denotes sunday here

def test_distribute_data_across_timerange():
    records = [
        {
            'impressions': 7, 'clicks': 6, 'costInUsd': 91, 'costInLocalCurrency': 7500, 'org_id': '1234', 'name': 'abc'
        }
    ]
    timerange = ['20230801', '20230802']
    distributed_records = DataTransformation.distribute_data_across_timerange(records, timerange)

    for timestamp, records in distributed_records.items():
        assert len(records) == 1

        # for impression = 7, it will be divided in 2 parts 4 and 3, the initial date will get the bigger value
        if timestamp == '20230801':
            assert records[0]['impressions'] == 4 
        else:
            assert records[0]['impressions'] == 3
        
        assert records[0]['clicks'] == 3
        assert records[0]['costInUsd'] == 45.5
        assert records[0]['costInLocalCurrency'] == 3750.0

        # this is check for non changing fields
        assert records[0]['org_id'] == '1234'
        assert records[0]['name'] == 'abc'
