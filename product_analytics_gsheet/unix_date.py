from datetime import datetime, timedelta


def get_previous_day_date():
    # Get today's date
    today = datetime.now()

    # Calculate the previous day
    previous_day = today - timedelta(days=1)

    # Format the previous day's date as YYYYMMDD
    previous_day_formatted = previous_day.strftime('%Y%m%d')

    return previous_day_formatted


# Call the function and print the result
previous_day_date = get_previous_day_date()
print("Previous day's date:", previous_day_date)
