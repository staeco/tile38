import sys
import httplib
import json
import time
import subprocess

url2 = "127.0.0.1"
port = 9851

self_not_ready = True

def make_request(url):
    conn = httplib.HTTPConnection(url, port)
    conn.request("GET", "/server")
    res = conn.getresponse()
    body = res.read().decode('utf-8')
    return json.loads(body)

def wait():
    sys.stdout.write('.')
    sys.stdout.flush()
    time.sleep(1)
    return

while self_not_ready:
    try:
        obj = make_request(url2)
        if obj['stats']['num_objects'] > 10:
            self_not_ready = False
            break
    except Exception as err:
        print(err)
        wait()
