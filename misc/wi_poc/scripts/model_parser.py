import sys
import os
import subprocess
import json
from datetime import datetime 

def extract_basename(x):
    return os.path.basename(os.path.normpath(x))

def extract_project_id(x):
    return os.path.normpath(x).split(os.sep)[-3]

def parse_utc(t):
    return datetime.utcfromtimestamp(int(t))

project_id = sys.argv[1]
names = [x.strip() for x in sys.stdin.readlines()]
# print('{} files found.'.format(len(names)))
p_id = extract_project_id(names[0])
mids = [extract_basename(x) for x in names]
mids.reverse()
st_time = parse_utc(sys.argv[2])

def get_line_from_events(line_num=1):
    command = 'head'
    if line_num < 0:
        command = 'tail'
        line_num = -1 * line_num
    p1 = subprocess.Popen(['gsutil', 'cat', events_file_path], stdout=subprocess.PIPE)
    p2 = subprocess.Popen([command, '-{}'.format(line_num)], stdin=p1.stdout, stdout=subprocess.PIPE)
    js, _ = p2.communicate()
    return js

def check_line(line_num=1):
    js = get_line_from_events(line_num)
    event_time = parse_utc(json.loads(js)['et'])
    diff_time = (event_time - st_time).days
    return diff_time >= 0 and diff_time < 7

for m in mids[:10]:
    events_file_path = 'gs://factors-production-v2/projects/{p}/models/{m}/events_{m}.txt'.format(p=p_id, m=m)
    if check_line(1) and check_line(-1):
        print(m)
        break
