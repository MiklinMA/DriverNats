import asyncio
from nats.aio.client import Client as NATS
from nats.aio.errors import ErrConnectionClosed, ErrTimeout, ErrNoServers

async def run(loop):
  nc = NATS()

  servers = []
  # servers.append('localhost:4222')

  await nc.connect(io_loop=loop) # , servers=servers)

  await nc.publish("ones.user.asdfdsa.test1", b'Hello')
  await nc.publish("ones.user.asdfdsa.test2", b'World')
  await nc.publish("ones.user.asdfdsa", b'!!!')

  await asyncio.sleep(1, loop=loop)
  await nc.close()

if __name__ == '__main__':
  loop = asyncio.get_event_loop()
  loop.run_until_complete(run(loop))
  loop.close()
