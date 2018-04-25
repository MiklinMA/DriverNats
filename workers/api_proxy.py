import worker
import requests
import urllib.parse
import json

server = "http://localhost:2030/"

@worker.method("api.>")
def api_proxy(data):
    method = data.get('Method')
    method = method.split('.')

    payload = data.get('Data')

    uri = server + "_".join(method[:-1])
    method = method[-1]
    print(uri, method)

    s = requests.Session()
    s.headers = data.get('Header', {})

    try:
        if method == 'get':
            payload = urllib.parse.urlencode(payload)
            resp = s.get(uri, params=payload)
        elif method == 'post':
            resp = s.post(uri, data=payload)

        if resp.status_code == 200:
            return resp.content

        return int(resp.status_code)
    except Exception as e:
        print(e)


worker.start()
