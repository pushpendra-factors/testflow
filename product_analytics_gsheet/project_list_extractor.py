import requests
from signin_req import sign_in
from unix_date import get_previous_day_date


def project_list_extractor():
    factors_key = sign_in()
    yesterday_date = get_previous_day_date()

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

    response_list = requests.get(
        'https://api.factors.ai/projectanalytics',
        cookies=cookies,
        headers=headers,
    )

# print(response_list.json())

    project_dump = response_list.json()
    project_list = []
    # print(project_dump)
    for projects in project_dump['analytics'][yesterday_date]:
        proj_hold = [projects['project_id'], projects['project_name']]
        project_list.append(proj_hold)

    return (project_list)
