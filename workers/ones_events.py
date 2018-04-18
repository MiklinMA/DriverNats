import worker

@worker.api_method("api.user.event.get")
def user_event(data):
    u_ref = data['Data']['u_ref']
    jdata = data['Data']['jdata']

    subject = "ones.user." + u_ref

    print(u_ref, jdata)

    response = yield subject, jdata, True
    response = response and getattr(response, 'data', response)

    print("RESPONSE: ", response)

    yield response or "not found"

worker.start()
