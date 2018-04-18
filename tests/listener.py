import asyncio
from nats.aio.client import Client as NATS
from nats.aio.errors import ErrConnectionClosed, ErrTimeout, ErrNoServers
import time
from random import random

nc = NATS()
loop = None

def logger(msg):
    print("START:", msg.data.decode())
    time.sleep(random())
    print("FINISH:", msg.data.decode())

async def run(loop):

  servers = []
  # servers.append('localhost:4222')

  await nc.connect(io_loop=loop, verbose=True) # , servers=servers)

  await nc.subscribe(">", "pipe", logger)


if __name__ == '__main__':
  loop = asyncio.get_event_loop()
  loop.run_until_complete(run(loop))
  try:
      loop.run_forever()
  finally:
      nc.close()
      loop.close()
