import requests
import gspread
from date_check import *
from signin_req import sign_in
import numpy as np
import logging
import os
import schedule
import time
import datetime
import pytz
import sys
from signin_req import sign_in
from unix_date import get_previous_day_date
from project_list_extractor import project_list_extractor

script_dir = os.path.dirname(os.path.abspath(__file__))


logging.basicConfig(stream=sys.stdout, level=logging.INFO)
logging.info("Starting script...")


def job1(t):
    logging.info("Job 1 function is called.")
    factors_key = sign_in()
    fr_timestamp, to_timestamp = date_range()
    cookies = {"factors-sid": factors_key}

    headers = {
        "authority": "api.factors.ai",
        "accept": "*/*",
        "accept-language": "en-US,en;q=0.9",
        "content-type": "application/json",
        # 'cookie': '_ga=GA1.1.1601406471.1651752988; _fbp=fb.1.1651752988247.1278804336; _fuid=MDA0N2M4ODMtMjg2OS00MmFkLWE1YmMtMGIxMDc1NDZmMTBj; hubspotutk=f046b5607e193341f0ed88c30a8e3ee4; insent-user-id=4YLePmp3LZY9rxhNr1657085174920; _lfa=LF1.1.c4b6630b03d3bd97.1658719707839; intercom-device-id-rvffkuu7=36f85754-62c4-4c1e-a604-c9bf0390d967; _hjSessionUser_2443020=eyJpZCI6IjFlYzRmZjI1LTJlYzQtNTVkOC05YTgyLTRiMWNkMjI1YmM0MyIsImNyZWF0ZWQiOjE2NTY0MTEyMDc2NTgsImV4aXN0aW5nIjp0cnVlfQ==; _gcl_au=1.1.1724467748.1675138461; _gcl_aw=GCL.1676668380.CjwKCAiA85efBhBbEiwAD7oLQGfDACLojBMAmyEiTgFuCZia1NeOXPqj2EHWfxnfaR7UJPp16vBrphoCFr8QAvD_BwE; _vwo_uuid_v2=D9DC126A5F4E48EFE668CC4EEC12111DB|30177c01a82bb88b6625a87299663a25; _vis_opt_s=1%7C; _vis_opt_test_cookie=1; _vwo_uuid=D9DC126A5F4E48EFE668CC4EEC12111DB; _vwo_ds=3%241681393539%3A43.88865128%3A%3A; __hssrc=1; factors-sid=eyJhdSI6IjFiOTkzYTFiLThhNzYtNGRkZC1hODI3LWZhM2M1NzliYThiOSIsInBmIjoiTVRZNE1UZ3dNek14TTN4dlNYRXpWV0ZFYVRkV1JIaEZhSEJQTFZGZlFXVnhSazl3VlZjNGNVWjRNMkpPYVhsQ1JXbHVPV1JwYzFaTFMxQllTVTFKZVRWRWNtZ3RjbGg0YVVNeFozWjJablJGUW1weVdtcHhkbVl6UTFRMWJ6MThtQ0s4TXFrekdpc20tVUFMME1vTzF0LVRvTkJweHA5ekNnTENLQThYUzdRPSJ9; _clck=zgrpep|1|faw|0; __hstc=3500975.f046b5607e193341f0ed88c30a8e3ee4.1654147042599.1681699724568.1681896923140.271; _vwo_sn=503382%3A1; _uetsid=65a35510dc4511edb5aee1a5b60ef0c9; intercom-session-rvffkuu7=QkFYZFczU2JmaVhRUzN4T3F2aDlUb0d2TnE2TFY5d2dNUm1CR1lLMWdVc3pOS1JHOENTRFFzMXBFTDZYcmxxVi0tb1d4cDhKQUR6WTN2U1BqZmdFTXVrQT09--1bb721871152e233dd1fb266ad02d36323f3b38c; _ga_ZM08VH2CGN=GS1.1.1681923001.1125.1.1681923119.0.0.0',
        "origin": "https://app.factors.ai",
        "referer": "https://app.factors.ai/",
        "sec-ch-ua": '"Chromium";v="112", "Google Chrome";v="112", "Not:A-Brand";v="99"',
        "sec-ch-ua-mobile": "?0",
        "sec-ch-ua-platform": '"macOS"',
        "sec-fetch-dest": "empty",
        "sec-fetch-mode": "cors",
        "sec-fetch-site": "same-site",
        "user-agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36",
    }

    sa = gspread.service_account(filename="/usr/local/var/product/dependency.json")
    sh = sa.open_by_key("1srY93iXEtPhC0f1ghXTHB7LhQX5JdogpuVcXnl32IxM")
    login_sheet = sh.worksheet("Logins Activity-Usage")
    ran_sheet = sh.worksheet("Reports Ran-Usage")
    viewed_sheet = sh.worksheet("Reports Viewed-Usage")
    logging.info("login 1 completed.")

    json_data_reports_viewed = {
        "query_group": [
            {
                "cl": "events",
                "ty": "events_occurrence",
                "grpa": "users",
                "fr": fr_timestamp,
                "to": to_timestamp,
                "ewp": [
                    {
                        "an": "",
                        "na": "VIEW_QUERY",
                        "pr": [
                            {
                                "en": "user",
                                "lop": "AND",
                                "op": "notContains",
                                "pr": "$email",
                                "ty": "categorical",
                                "va": "factors.ai",
                            },
                            {
                                "en": "user",
                                "lop": "AND",
                                "op": "notContains",
                                "pr": "$email",
                                "ty": "categorical",
                                "va": "demo",
                            },
                            {
                                "en": "user",
                                "lop": "AND",
                                "op": "notEqual",
                                "pr": "$email",
                                "ty": "categorical",
                                "va": "swamikrish2001@wadog.com",
                            },
                            {
                                "en": "user",
                                "lop": "AND",
                                "op": "notEqual",
                                "pr": "$email",
                                "ty": "categorical",
                                "va": "rahul-test@ebarg.net",
                            },
                        ],
                    },
                ],
                "gup": [
                    {
                        "en": "user_g",
                        "lop": "AND",
                        "op": "notContains",
                        "pr": "$user_id",
                        "ty": "categorical",
                        "va": "factors.ai",
                    },
                ],
                "gbt": "",
                "gbp": [
                    {
                        "pr": "$email",
                        "en": "user",
                        "pty": "categorical",
                        "ena": "VIEW_QUERY",
                        "eni": 1,
                    },
                    {
                        "pr": "$timestamp",
                        "en": "event",
                        "pty": "datetime",
                        "ena": "VIEW_QUERY",
                        "eni": 1,
                        "grn": "day",
                    },
                    {
                        "pr": "project_name",
                        "en": "event",
                        "pty": "categorical",
                        "ena": "VIEW_QUERY",
                        "eni": 1,
                    },
                    {
                        "pr": "$timestamp",
                        "en": "event",
                        "pty": "datetime",
                        "ena": "VIEW_QUERY",
                        "eni": 1,
                        "grn": "day",
                    },
                    {
                        "pr": "project_id",
                        "en": "event",
                        "pty": "categorical",
                        "ena": "VIEW_QUERY",
                        "eni": 1,
                    },
                ],
                "ec": "each_given_event",
                "tz": "Asia/Kolkata",
            },
        ],
    }

    json_data_reports_ran = {
        "query_group": [
            {
                "cl": "events",
                "ty": "events_occurrence",
                "grpa": "users",
                "fr": fr_timestamp,
                "to": to_timestamp,
                "ewp": [
                    {
                        "an": "",
                        "na": "RUN-QUERY",
                        "pr": [
                            {
                                "en": "user",
                                "lop": "AND",
                                "op": "notContains",
                                "pr": "$email",
                                "ty": "categorical",
                                "va": "factors.ai",
                            },
                            {
                                "en": "user",
                                "lop": "AND",
                                "op": "notEqual",
                                "pr": "$email",
                                "ty": "categorical",
                                "va": "swamikrish2001@wadog.com",
                            },
                            {
                                "en": "user",
                                "lop": "AND",
                                "op": "notEqual",
                                "pr": "$email",
                                "ty": "categorical",
                                "va": "rahul-test@ebarg.net",
                            },
                            {
                                "en": "event",
                                "lop": "AND",
                                "op": "notContains",
                                "pr": "project_name",
                                "ty": "categorical",
                                "va": "demo",
                            },
                        ],
                    },
                ],
                "gup": [],
                "gbt": "",
                "gbp": [
                    {
                        "pr": "$email",
                        "en": "user",
                        "pty": "categorical",
                        "ena": "RUN-QUERY",
                        "eni": 1,
                    },
                    {
                        "pr": "$timestamp",
                        "en": "event",
                        "pty": "datetime",
                        "ena": "RUN-QUERY",
                        "eni": 1,
                        "grn": "day",
                    },
                    {
                        "pr": "project_name",
                        "en": "event",
                        "pty": "categorical",
                        "ena": "RUN-QUERY",
                        "eni": 1,
                    },
                    {
                        "pr": "$timestamp",
                        "en": "event",
                        "pty": "datetime",
                        "ena": "RUN-QUERY",
                        "eni": 1,
                        "grn": "day",
                    },
                    {
                        "pr": "project_id",
                        "en": "event",
                        "pty": "categorical",
                        "ena": "RUN-QUERY",
                        "eni": 1,
                    },
                ],
                "ec": "each_given_event",
                "tz": "Asia/Kolkata",
            },
        ],
    }

    json_data_login = {
        "query_group": [
            {
                "cl": "events",
                "ty": "events_occurrence",
                "grpa": "users",
                "fr": fr_timestamp,
                "to": to_timestamp,
                "ewp": [
                    {
                        "an": "",
                        "na": "VIEW_DASHBOARD",
                        "pr": [
                            {
                                "en": "user",
                                "lop": "AND",
                                "op": "notContains",
                                "pr": "$email",
                                "ty": "categorical",
                                "va": "factors.ai",
                            },
                            {
                                "en": "event",
                                "lop": "AND",
                                "op": "notContains",
                                "pr": "project_name",
                                "ty": "categorical",
                                "va": "demo",
                            },
                            {
                                "en": "user",
                                "lop": "AND",
                                "op": "notContains",
                                "pr": "$email",
                                "ty": "categorical",
                                "va": "swamikrish2001@wadog.com",
                            },
                            {
                                "en": "user",
                                "lop": "AND",
                                "op": "notContains",
                                "pr": "$email",
                                "ty": "categorical",
                                "va": "rahul-test@ebarg.net",
                            },
                        ],
                    },
                ],
                "gup": [
                    {
                        "en": "user_g",
                        "lop": "AND",
                        "op": "notContains",
                        "pr": "$user_id",
                        "ty": "categorical",
                        "va": "factors.ai",
                    },
                ],
                "gbt": "",
                "gbp": [
                    {
                        "pr": "$email",
                        "en": "user",
                        "pty": "categorical",
                        "ena": "VIEW_DASHBOARD",
                        "eni": 1,
                    },
                    {
                        "pr": "$timestamp",
                        "en": "event",
                        "pty": "datetime",
                        "ena": "VIEW_DASHBOARD",
                        "eni": 1,
                        "grn": "day",
                    },
                    {
                        "pr": "project_name",
                        "en": "event",
                        "pty": "categorical",
                        "ena": "VIEW_DASHBOARD",
                        "eni": 1,
                    },
                    {
                        "pr": "$timestamp",
                        "en": "event",
                        "pty": "datetime",
                        "ena": "VIEW_DASHBOARD",
                        "eni": 1,
                        "grn": "day",
                    },
                    {
                        "pr": "project_id",
                        "en": "event",
                        "pty": "numerical",
                        "ena": "VIEW_DASHBOARD",
                        "eni": 1,
                        "gbty": "raw_values",
                    },
                ],
                "ec": "each_given_event",
                "tz": "Asia/Kolkata",
            },
        ],
    }

    response_viewed = requests.post(
        "https://api.factors.ai/projects/2/v1/query",
        cookies=cookies,
        headers=headers,
        json=json_data_reports_viewed,
    )
    response_ran = requests.post(
        "https://api.factors.ai/projects/2/v1/query",
        cookies=cookies,
        headers=headers,
        json=json_data_reports_ran,
    )
    response_login = requests.post(
        "https://api.factors.ai/projects/2/v1/query",
        cookies=cookies,
        headers=headers,
        json=json_data_login,
    )
    logging.info("api calls completed.")

    viewed_data = response_viewed.json()
    ran_data = response_ran.json()
    login_data = response_login.json()

    write_sheet_viewed = []
    write_sheet_ran = []
    write_sheet_login = []

    for values in viewed_data["result_group"][0]["rows"]:
        val = values[2:]
        val[1] = val[1][:10]
        val[3] = val[3][:10]
        element = val.pop()
        val.insert(0, element)
        write_sheet_viewed.append(val)

    for values in ran_data["result_group"][0]["rows"][2:]:
        val = values[2:]
        val[1] = val[1][:10]
        val[3] = val[3][:10]
        element = val.pop()
        val.insert(0, element)
        write_sheet_ran.append(val)

    for values in login_data["result_group"][0]["rows"][2:]:
        val = values[2:]
        val[1] = val[1][:10]
        val[3] = val[3][:10]
        element = val.pop()
        write_sheet_login.append(val)

    logging.info("data appended.")

    sh.values_append(
        "Logins Activity-Usage!A:E",
        params={"valueInputOption": "USER_ENTERED"},
        body={"values": write_sheet_login},
    )

    sh.values_append(
        "Reports Ran-Usage!A:F",
        params={"valueInputOption": "USER_ENTERED"},
        body={"values": write_sheet_ran},
    )

    sh.values_append(
        "Reports Viewed-Usage!A:F",
        params={"valueInputOption": "USER_ENTERED"},
        body={"values": write_sheet_viewed},
    )
    logging.info("job completed.")

def job2(t):
    logging.info("Job 2 function is called.")
    factors_key = sign_in()
    yesterday_date = get_previous_day_date()
    project_list = project_list_extractor()

    headers = {
        'authority': 'api.factors.ai',
        'accept': '*/*',
        'accept-language': 'en-US,en;q=0.9',
        'content-type': 'application/json',
        # 'cookie': '_ga=GA1.1.1601406471.1651752988; _fbp=fb.1.1651752988247.1278804336; _fuid=MDA0N2M4ODMtMjg2OS00MmFkLWE1YmMtMGIxMDc1NDZmMTBj; intercom-id-rvffkuu7=adc2e6dd-f134-492b-99d1-72d298526358; hubspotutk=f046b5607e193341f0ed88c30a8e3ee4; insent-user-id=4YLePmp3LZY9rxhNr1657085174920; _lfa=LF1.1.c4b6630b03d3bd97.1658719707839; intercom-device-id-rvffkuu7=36f85754-62c4-4c1e-a604-c9bf0390d967; __hssrc=1; _hjSessionUser_2443020=eyJpZCI6IjFlYzRmZjI1LTJlYzQtNTVkOC05YTgyLTRiMWNkMjI1YmM0MyIsImNyZWF0ZWQiOjE2NTY0MTEyMDc2NTgsImV4aXN0aW5nIjp0cnVlfQ==; _gcl_au=1.1.1724467748.1675138461; _gcl_aw=GCL.1675227264.CjwKCAiAleOeBhBdEiwAfgmXf5enXMU_sZjzS0iGxXWSG_hAEGs9hO5QRjJ7SxixbueBG5BrdsExRhoCLQYQAvD_BwE; __hstc=3500975.f046b5607e193341f0ed88c30a8e3ee4.1654147042599.1675402252449.1675442342435.201; _clck=zgrpep|1|f8v|0; _uetsid=1e956fb0a2cc11ed8c21bd3601427fbb; _clsk=1s43j6y|1675579001298|1|1|l.clarity.ms/collect; intercom-session-rvffkuu7=ZVF5RTJoZk9RVUt6bkZtSWJ4NWpDTlhPektVVkRmK2JFbloyTUhVaGNmaHlSaFJBOVB0SzhsR1NaUS9GeURPSC0tV1hFZ2N3elkrdFgweVNZYjM3RER1QT09--d9dbac12695d6d49afc38d82b5df868d15bb9872; _ga_ZM08VH2CGN=GS1.1.1675580139.782.1.1675580145.0.0.0',
        'origin': 'https://app.factors.ai',
        'referer': 'https://app.factors.ai/',
        'sec-ch-ua': '"Not?A_Brand";v="8", "Chromium";v="108", "Google Chrome";v="108"',
        'sec-ch-ua-mobile': '?0',
        'sec-ch-ua-platform': '"macOS"',
        'sec-fetch-dest': 'empty',
        'sec-fetch-mode': 'cors',
        'sec-fetch-site': 'same-site',
        'user-agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36',
    }

    cookies = {
        'factors-sid': factors_key
    }

    params = {
        'days': '1',
    }

    gc = gspread.service_account(
        filename="/usr/local/var/product/dependency.json")
    sh = gc.open_by_key('1C3ZYwZ1QAicqrcYZZNx4NT6z0YT3Rsk2QkSNNJEIhbw')
    wks1 = sh.worksheet("Integrations Table - Master")
    logging.info("login 2 completed.")

    for project in project_list:
        try:
            time.sleep(0.5)
            metrics_list = [yesterday_date, project[0], project[1]]
            connected_list = [yesterday_date, project[0], project[1]]

            response = requests.get(
                'https://api.factors.ai/'+project[0]+'/projectanalytics',
                params=params,
                cookies=cookies,
                headers=headers,
            )
            project_data = response.json()
            master_connect_list = ["Factors 6Signal", "Adwords", "Linkedin", "Google Organic", "Segment", "Drift",	"Rudderstack",
                                "Clinet 6Signal",	"Hubspot",	"Salesforce",	"Bing",	"Facebook",	"Marketo", "G2", "Clearbit", "Lead Squared",	"Slack", "Teams"]

            index_list = ["No", "No", "No", "No", "No", "No", "No", "No",
                        "No", "No", "No", "No", "No", "No", "No", "No", "No", "No"]
            for index, metrics in enumerate(master_connect_list):
                if metrics in project_data['metrics'][1]['connected']:
                    index_list[index] = "Yes"

            connected_list = connected_list+index_list

            sh.values_append(
                'Integrations Table - Master!A:Z',
                params={
                    'valueInputOption': 'USER_ENTERED'
                },
                body={
                    'values': [connected_list]
                }
            )

            for metrics in list(project_data['metrics'][0].values()):
                metrics_list.append(metrics)

            connected_list = project_data['metrics'][1]['connected']
            disconnected_list = project_data['metrics'][2]['disconnected']

            sh.values_append(
                'Metrics Table- Master!A:Z',
                params={
                    'valueInputOption': 'USER_ENTERED'
                },
                body={
                    'values': [metrics_list]
                }
            )
            logging.info("job 2 completed for project: %s", project)
        except Exception as e:
            logging.error("Error processing project %s: %s", project, str(e))
                # Continue to the next project even if an exception occurs
        continue


ist = pytz.timezone("Asia/Kolkata")

# time defined as (13:15:00 PM) for cron job
scheduled_time = datetime.time(16, 30)

scheduled_datetime = datetime.datetime.combine(datetime.date.today(), scheduled_time)
scheduled_datetime = ist.localize(scheduled_datetime)

schedule.every().day.at(scheduled_datetime.strftime("%H:%M:%S")).do(
    job1, "job 1 scheduled at 00:01:15 PM"
)

schedule.every().day.at(scheduled_datetime.strftime("%H:%M:%S")).do(
    job2, "job 2 scheduled at 00:01:15 PM"
)
while True:
    schedule.run_pending()
    time.sleep(60)
