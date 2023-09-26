import datetime
import pytz

def date_range():
    # Set the timezone to IST
    ist = pytz.timezone('Asia/Kolkata')

    # Get the current date in IST
    current_date = datetime.datetime.now(ist).date()

    # Calculate yesterday's date
    yesterday = current_date - datetime.timedelta(days=1)

    # Create datetime objects for yesterday's 12:00 AM and 11:59:59 PM
    start_of_day = ist.localize(datetime.datetime.combine(yesterday, datetime.time.min))
    end_of_day = ist.localize(datetime.datetime.combine(yesterday, datetime.time.max))

    # Convert datetime objects to Unix timestamps
    start_timestamp = int(start_of_day.timestamp())
    end_timestamp = int(end_of_day.timestamp())

    #start_timestamp = 1693506600
    #end_timestamp = 1694370599

    # Return the Unix timestamps
    return start_timestamp, end_timestamp


# Call the function and print the outputs
start_timestamp, end_timestamp = date_range()
print(f"Yesterday 12:00 AM IST Unix timestamp: {start_timestamp}")
print(f"Yesterday 11:59:59 PM IST Unix timestamp: {end_timestamp}")
