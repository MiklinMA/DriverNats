import asyncio
from nats.aio.client import Client as NATS
from nats.aio.errors import ErrConnectionClosed, ErrTimeout, ErrNoServers
import signal
import json
import inspect

methods = {}

nc = NATS()

class CallbackData:
    pass

def api_method(method):
    def api_decorator(callback):
        async def wrapper(msg):
            subject = msg.subject
            reply = msg.reply
            data = msg.data.decode()
            try:
                data = json.loads(data)
            except:
                pass

            # print("RUN", reply, 'working')
            # await nc.publish(reply, b'working')
            # print("OK", reply, 'working')

            # res = callback(data)
            # print(res)
            # if res is None:
            #     res = b"DONE"

            # print("RUN", reply, res)
            # await nc.publish(reply, res.encode())
            # print("OK", reply, res)

            async def publish_data(res):
                s = reply
                m = nc.publish

                if res is None:
                    res = b"DONE"

                if isinstance(res, tuple):
                    s = res[0]
                    if len(res) == 2:
                        res = res[1]

                    elif len(res) == 3:
                        res = res[1]
                        m = nc.request

                if isinstance(res, str):
                    res = res.encode()

                print(s, type(res), res)
                return await m(s, res)

            if inspect.isgeneratorfunction(callback):
                gen = callback(data)
                try:
                    res = None
                    while True:
                        res = gen.send(res)
                        res = await publish_data(res)
                except StopIteration as e:
                    pass
                except ErrTimeout as e:
                    res = gen.send('timeout')
                    res = await publish_data(res)
                except Exception as e:
                    print("ERROR", type(e), e)

            else:
                await publish_data(callback(data))

        methods[method] = wrapper

        return wrapper

    return api_decorator

async def run(loop):
    queue = ""

    await nc.connect(io_loop=loop)
    print("Connected to NATS at {}...".format(nc.connected_url.netloc))
  
    for method, callback in methods.items():
        print("Add subscriber:", method)
        await nc.subscribe(method, queue, callback)

def start():
    if not len(methods):
        print("API methods are not defined")
        return
  
    loop = asyncio.get_event_loop()
    loop.run_until_complete(run(loop))
    try:
        loop.run_forever()
    except Exception as e:
        print("ERROR", e)
    finally:
        loop.close()
  
if __name__ == '__main__':
    start()
