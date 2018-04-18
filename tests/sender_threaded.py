import asyncio
from nats.aio.client import Client as NATS
from nats.aio.errors import ErrConnectionClosed, ErrTimeout, ErrNoServers
from datetime import datetime

maxrange = int(1e5)
print("MAXRANGE:", maxrange)
print()

async def run(loop):
    nc = NATS()

    servers = []
    servers.append('nats://nero:4222')

    await nc.connect(io_loop=loop, servers=servers)

    ts = []

    print('Create tasks', datetime.now())
    dt = datetime.now()
    for n in range(maxrange):
        subject = "test.message.sender.threaded"
        message = "message #%d" % n
        t = loop.create_task(nc.publish(subject, message.encode()))
        ts.append(t)

    d1 = datetime.now()
    print('Await tasks ', datetime.now() - dt)
    dt = datetime.now()
    for t in ts:
        await t

    print('Done tasks  ', datetime.now() - dt)
    dt = datetime.now()
    
    s = (datetime.now() - d1).seconds

    # await asyncio.sleep(1, loop=loop)
    await nc.close()

    print('Exit        ', datetime.now() - dt)

    print()
    print('MPS         ', maxrange, '/', s, '=', s > 0 and int(maxrange / s) or maxrange)

if __name__ == '__main__':
    loop = asyncio.get_event_loop()
    loop.run_until_complete(run(loop))
    loop.close()
