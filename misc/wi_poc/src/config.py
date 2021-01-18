# Change the following as per the system on which this script is to be run.

system = 'mymac' # My local.
# system='emr' # An EMR cluster.

if system == 'emr':
    DEFAULT_CLOUD_PATH = "s3://weekly-insights/data/cloud_storage/"
    DEFAULT_HOME_PATH = "/home/hadoop/repos/factors/misc/wi_poc/"
elif system == 'mymac':
    DEFAULT_CLOUD_PATH = "/usr/local/var/factors/cloud_storage/"
    DEFAULT_HOME_PATH = "/Users/govindjsk/repos/factors.ai/factors/misc/wi_poc/"
else:
    DEFAULT_CLOUD_PATH = "./"
    DEFAULT_HOME_PATH = "./"
    
DEFAULT_FEAT_SCHEMA_FILENAME = DEFAULT_HOME_PATH + "resources/feat_schema.csv"
