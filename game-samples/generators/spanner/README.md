# Spanner load generation

This is a WIP to generate load directly to Spanner, instead of having a REST api.

## TODO
- handles responses better for Locust framework

## Known issues
- Locust uses gevent, which [patches several standard python libraries](https://www.gevent.org/api/gevent.monkey.html), to work for events. One of those is the ssl library. Whatever patch is done here, it inteferes with the spanner library's ability to create sessions. A workaround is to patch locust's __init.py__:

```
# From
monkey_patch.all()

# To
monkey_patch.all(thread=False)
```
