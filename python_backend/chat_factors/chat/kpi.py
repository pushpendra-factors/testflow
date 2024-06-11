import logging as log
import os
import json
from datetime import datetime, timedelta
import pytz
import Levenshtein

DEFAULT_TIMEZONE = "Asia/Kolkata"
DEFAULT_CATEGORY = "events"
DEFAULT_DISPLAY_CATEGORY = "website_session"

# todo : Add error handling for matching not found
def get_transformed_kpi_query(gpt_response, kpi_config,
                              raw_data_path=os.path.join('chat_factors/chatgpt_poc', 'data.json')):
    query_payload = {}
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
                    "me": [kpi_info["matching_kpi_name"]],
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
                    "me": [kpi_info["matching_kpi_name"]],
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

    break_down_info = add_break_down_info_kpi(display_category, gpt_response["qb"], kpi_config)

    filter_info = add_filter_info_kpi(display_category, gpt_response["qf"], kpi_config)

    query_payload["gGBy"] = break_down_info
    query_payload["gFil"] = filter_info


    return query_payload


def get_kpi_info(name, kpi_config):
    lowest_edit_distance = 100
    kpi_info = {}
    for category_config in kpi_config:
        for metric in category_config.get("metrics", []):
            current_string = metric.get("name")
            current_distance = Levenshtein.distance(name, current_string)
            normalized_distance = current_distance / max(len(name), len(current_string))
            if normalized_distance < 0.3 and current_distance < lowest_edit_distance:
                lowest_edit_distance = current_distance
                best_match = current_string
                kpi_info = {
                    "matching_kpi_name": best_match,
                    "category": category_config.get("category"),
                    "display_category": category_config.get("display_category"),
                    "kpi_query_type": metric.get("kpi_query_type"),
                }
        return kpi_info

    log.error("kpi not found in the kpi_config")
    raise KPIOrPropertyNotFoundError(f"KPI with name '{name}' not found in the kpi_config")


class KPIOrPropertyNotFoundError(Exception):
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


def add_break_down_info_kpi(display_category, breakdowns, kpi_config):
    query_breakdowns = []

    if breakdowns != '':
        breakdown_list = breakdowns.split(', ')
        for breakdown_property in breakdown_list:
            break_down_info = get_break_down_info(display_category, breakdown_property, kpi_config)
            log.info("breakdown info for breakdown_property: %s", break_down_info)

            query_breakdown = {
                "gr": "",
                "prNa": break_down_info["breakdown_property"],
                "prDaTy": break_down_info["data_type"],
                "en": break_down_info["entity"],
                "objTy": "",
                "dpNa": break_down_info["display_name"],
                "isPrMa": False
            }

            query_breakdowns.append(query_breakdown)

    return query_breakdowns


def get_break_down_info(display_category, breakdown_name, kpi_config):
    break_down_info = {}
    lowest_edit_distance = 100
    for category_config in kpi_config:
        if category_config.get("display_category") == display_category:
            for metric in category_config.get("properties", []):
                current_string = metric.get("name")
                current_distance = Levenshtein.distance(breakdown_name, current_string)
                if current_distance < lowest_edit_distance:
                    lowest_edit_distance = current_distance
                    best_match = current_string
                    break_down_info = {
                        "breakdown_property": best_match,
                        "data_type": metric.get("data_type"),
                        "entity": metric.get("entity"),
                        "display_name": metric.get("display_name"),
                    }
            if lowest_edit_distance / len(breakdown_name) < 0.2:
                return break_down_info

    log.error("breakdown property not found in the kpi_config")
    raise KPIOrPropertyNotFoundError(f"breakdown property '{breakdown_name}' not found in the kpi_config")


def add_filter_info_kpi(display_category, filters, kpi_config):
    query_filters = []

    if filters != '':
        filter_list = filters
        for filter_property in filter_list:
            filter_info = get_filter_info(display_category, filter_property, kpi_config)
            log.info("breakdown info for breakdown_property: %s", filter_info)

            query_filter = {
                "extra": [
                    filter_info['display_name'],
                    filter_info["filter_property"],
                    filter_info["data_type"],
                    filter_info["entity"]
                ],
                "objTy": "",
                "prNa": filter_info["filter_property"],
                "prDaTy": filter_info["data_type"],
                "isPrMa": False,
                "en": filter_info["entity"],
                "co": filter_property["co"],
                "va": filter_info["filter_value"],
                "lOp": "AND"
            }


            query_filters.append(query_filter)

    return query_filters


def get_filter_info(display_category, filter_name_val, kpi_config):
    filter_info = {}
    filter_name= filter_name_val["na"]
    lowest_edit_distance = 100
    for category_config in kpi_config:
        if category_config.get("display_category") == display_category:
            for metric in category_config.get("properties", []):
                current_string = metric.get("name")
                current_distance = Levenshtein.distance(filter_name, current_string)
                if current_distance < lowest_edit_distance:
                    lowest_edit_distance = current_distance
                    best_match = current_string
                    filter_info = {
                        "filter_property": best_match,
                        "data_type": metric.get("data_type"),
                        "entity": metric.get("entity"),
                        "display_name": metric.get("display_name"),
                        "filter_value" : filter_name_val["val"]
                    }
            if lowest_edit_distance / len(filter_name) < 0.2:
                return filter_info

    log.error("filter property not found in the kpi_config")
    raise KPIOrPropertyNotFoundError(f"filter property '{filter_name}' not found in the kpi_config")




