import asyncio
from nats.aio.client import Client as NATS
from nats.aio.errors import ErrConnectionClosed, ErrTimeout, ErrNoServers
import signal
import json
import inspect
import time
import pdb

methods = {}

nc = NATS()

def method(method, send_found=True):
    def decorator(callback):
        async def wrapper(msg):
            subject = msg.subject
            reply = msg.reply
            data = msg.data.decode()
            try:
                data = json.loads(data)
            except:
                pass

            if send_found:
                await nc.publish(reply, b'found')

            async def publish_data(res):
                s = reply
                m = nc.publish
                kw = {}

                if res is None:
                    res = b"DONE"

                if isinstance(res, tuple):
                    s = res[0]
                    if len(res) == 2:
                        res = res[1]

                    elif len(res) == 3:
                        m = nc.request
                        kw['timeout'] = float(res[2])
                        res = res[1]

                if isinstance(res, str):
                    res = res.encode()

                # print(s, type(res), res, kw)
                return await m(s, res, **kw)

            if inspect.isgeneratorfunction(callback):
                gen = callback(data)
                try:
                    res = None
                    while True:
                        res = gen.send(res)
                        res = await publish_data(res)
                except ErrTimeout as e:
                    res = gen.send(b'timeout')
                    res = await publish_data(res)
                except StopIteration:
                    pass
                except Exception as e:
                    print("ERROR", type(e), e)

            else:
                await publish_data(callback(data))

        methods[method] = wrapper

        return wrapper

    return decorator

# @method(">")
# def test(data):
#     print(data)
#     return "OK"

async def run(loop):
    async def closed_cb():
        print("Connection to NATS is closed.")
        await asyncio.sleep(0.1, loop=loop)
        loop.stop()

    options = {
        "name":"Worker",
        "servers": ["nats://nero:4222"],
        # "servers": ["nats://127.0.0.1:4222"],
        "io_loop": loop,
        "closed_cb": closed_cb
    }

    await nc.connect(**options)
    print("Connected to NATS at {}".format(nc.connected_url.netloc))

    queue = ""
    for method, callback in methods.items():
        print("Add subscriber:", method)
        await nc.subscribe(method, queue, callback)

    def signal_handler():
        if nc.is_closed:
            return
        print("Disconnecting...")
        loop.create_task(nc.close())

    for sig in ('SIGINT', 'SIGTERM'):
        loop.add_signal_handler(getattr(signal, sig), signal_handler)

def start():
    if not len(methods.keys()):
        print("No methods defined")
        test_run()
        return

    loop = asyncio.get_event_loop()
    loop.run_until_complete(run(loop))
    try:
        loop.run_forever()
    finally:
        loop.close()

if __name__ == '__main__':
    start()

