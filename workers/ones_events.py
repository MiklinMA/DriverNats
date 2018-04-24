import worker

@worker.method("api.user.event.get", send_found=False)
def user_event(data):
    u_ref = data['Data']['u_ref']
    jdata = data['Data']['jdata']

    subject = "ones.user." + u_ref + ".event"

    response = yield subject, jdata, 2.0
    response = response and getattr(response, 'data', response)
    response = response or b"not found"

    print(u_ref, response.decode())

    yield response

worker.start()
