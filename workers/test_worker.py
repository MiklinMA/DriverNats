import worker

@worker.api_method("api.sync.test.get")
def test_get(data):
    print("GET:", data)
    return "TEST GET".encode()

@worker.api_method("api.sync.test.post")
def test_get(data):
    print("POST:", data)

worker.start()
