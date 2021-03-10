from wi_poc.src.weekly_insights import generate_weekly_insights
import argparse
from wi_poc.src.defaults import WKS, criteria_map, DEFAULT_WK1_KEY, DEFAULT_WK2_KEY, \
    DEFAULT_BASE, DEFAULT_TARGET

def parse_args():
    """
    Parses arguments for the weekly insights module.
    """
    parser = argparse.ArgumentParser(description="DELTA@Factors: Generate week-over-week insights.")
    parser.add_argument('project_id', type=str, help="Project ID")
    parser.add_argument('-month1', type=int, default=None, help="Month of the first week. For example, for December, say 12")
    parser.add_argument('-week1', type=int, default=None, help="Week number (in the month) of the first week. \
        For example, for 2nd week of December, say 2.")
    parser.add_argument('-month2', type=int, default=None, help="Month of the second week.")
    parser.add_argument('-week2', type=int, default=None, help="Week number (in the month) of the second week.")
    args = parser.parse_args()

    project_id = args.project_id
    month1 = args.month1
    wk1 = args.week1
    month2 = args.month2
    wk2 = args.week2
    try:
        wk1_key = WKS[month1][wk1 - 1]
    except KeyError:
        wk1_key = DEFAULT_WK1_KEY
        print("Week {} of month {} not supported. Using defaults.".format(month1, wk1))
    try:
        wk2_key = WKS[month2][wk2 - 1]
    except KeyError:
        wk2_key = DEFAULT_WK2_KEY
        print("Week {} of month {} not supported. Using defaults.".format(month2, wk2))
    return project_id, wk1_key, wk2_key


def main():
    project_id, wk1_key, wk2_key = parse_args()
    wi_args = {'project_id': project_id,
               'wk1_key': wk1_key,
               'wk2_key': wk2_key,
               'base': criteria_map.get(project_id, {}).get('base', DEFAULT_BASE),
               'target': criteria_map.get(project_id, {}).get('target', DEFAULT_TARGET)}
    generate_weekly_insights(**wi_args, filter_params_mode='bucketed')

if __name__ == '__main__':
    main()
