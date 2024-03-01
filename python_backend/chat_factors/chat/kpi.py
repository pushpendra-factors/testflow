import logging as log
import os
import json
from datetime import datetime, timedelta
import pytz

DEFAULT_TIMEZONE = "Asia/Kolkata"
DEFAULT_CATEGORY = "events"
DEFAULT_DISPLAY_CATEGORY = "website_session"


def get_transformed_kpi_query(gpt_response, kpi_config,
                              raw_data_path=os.path.join('chat_factors/chatgpt_poc', 'data.json')):
    try:
        data = json.load(open(raw_data_path, 'r'))
    except Exception as e:
        log.error("Error processing request: %s", str(e))
    from_time_stamp, to_time_stamp = get_start_end_timestamps(DEFAULT_TIMEZONE, gpt_response["time"])
    kpi_to_search = gpt_response["qe"]
    kpi_info = get_kpi_info(kpi_to_search, kpi_config)
    if kpi_info:
        category = kpi_info['category']
        display_category = kpi_info['display_category']
        query_type = kpi_info['kpi_query_type']
        log.info("done step 3 \n kpi_info from kpi_config :%s", kpi_info)
        query_payload = {
            "cl": gpt_response["qt"],
            "qG": [
                {
                    "ca": category,
                    "pgUrl": "",
                    "dc": display_category,
                    "me": [gpt_response["qe"]],
                    "fil": [],
                    "gBy": [],
                    "fr": from_time_stamp,
                    "to": to_time_stamp,
                    "tz": DEFAULT_TIMEZONE,
                    "qt": query_type,
                    "an": ""
                },
                {
                    "ca": category,
                    "pgUrl": "",
                    "dc": display_category,
                    "me": [gpt_response["qe"]],
                    "fil": [],
                    "gBy": [],
                    "gbt": "date",
                    "fr": from_time_stamp,
                    "to": to_time_stamp,
                    "tz": DEFAULT_TIMEZONE,
                    "qt": query_type,
                    "an": ""
                }
            ],
            "gGBy": [],
            "gFil": []
        }
    else:
        log.info("done step 3 \n No information found in kpi_config for kpi :%s", kpi_to_search)
        query_payload = {}

    return query_payload


def get_kpi_info(name, kpi_config):
    for category_config in kpi_config:
        for metric in category_config.get("metrics", []):
            if metric.get("name") == name:
                return {
                    "category": category_config.get("category"),
                    "display_category": category_config.get("display_category"),
                    "kpi_query_type": metric.get("kpi_query_type"),
                }

    log.error("kpi not found in the kpi_config")
    raise KpiNotFoundError(f"KPI with name '{name}' not found in the kpi_config")


class KpiNotFoundError(Exception):
    pass


# returns default time range as "this_week"
def get_start_end_timestamps(timezone, duration):
    # Get the timezone object
    tz = pytz.timezone(timezone)

    # Get the current time in the specified timezone
    current_time = datetime.now(tz)

    # Calculate the start and end times based on the provided duration
    if duration == "this_week":
        # Calculate the most recent Sunday
        # current_time.weekday() : Monday-0, Sunday-6
        start_time = current_time - timedelta(days=current_time.weekday() + 1)
        start_time = start_time.replace(hour=0, minute=0, second=0, microsecond=0)
        end_time = current_time
    elif duration == "last_week":
        # Calculate the Sunday before last
        start_time = current_time - timedelta(days=current_time.weekday() + 7 + 1)
        start_time = start_time.replace(hour=0, minute=0, second=0, microsecond=0)
        # Calculate the Saturday before last midnight
        end_time = start_time + timedelta(days=6, hours=23, minutes=59, seconds=59)
    elif duration == "this_month":
        # Calculate the first day of the current month
        start_time = current_time.replace(day=1, hour=0, minute=0, second=0, microsecond=0)
        end_time = current_time
    elif duration == "last_month":
        # Calculate the first day of the previous month
        first_day_of_current_month = current_time.replace(day=1, hour=0, minute=0, second=0, microsecond=0)
        start_time = first_day_of_current_month - timedelta(days=1)
        start_time = start_time.replace(day=1)
        end_time = first_day_of_current_month - timedelta(days=1)
        end_time = end_time.replace(hour=23, minute=59, second=59, microsecond=999999)
    elif duration == "today":
        start_time = current_time
        start_time = start_time.replace(hour=0, minute=0, second=0, microsecond=0)
        end_time = start_time + timedelta(days=0, hours=23, minutes=59, seconds=59)
    elif duration == "yesterday":
        start_time = current_time - timedelta(days=1)
        start_time = start_time.replace(hour=0, minute=0, second=0, microsecond=0)
        end_time = start_time + timedelta(days=0, hours=23, minutes=59, seconds=59)
    else:
        # default time range : "last_week"
        # Calculate the Sunday before last
        start_time = current_time - timedelta(days=current_time.weekday() + 7 + 1)
        start_time = start_time.replace(hour=0, minute=0, second=0, microsecond=0)
        # Calculate the Saturday before last midnight
        end_time = start_time + timedelta(days=6, hours=23, minutes=59, seconds=59)

    # Convert datetime objects to timestamps
    start_timestamp = int(start_time.timestamp())
    end_timestamp = int(end_time.timestamp())

    return start_timestamp, end_timestamp
