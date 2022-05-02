# Generators

These generators leverage the [Locust](https://locust.io) python framework for generating load.

The generators can be run via the command line, or a web interface.


Provided generators do the following:

- _authentication_server.py_: mimics player signup and player retrieval by UUID. Login is not handled currently due to the necessity to track password creation.

Run on the CLI:
```
locust -H http://127.0.0.1:8080 -f authentication_server.py --headless -u=2 -r=2 -t=10s
```

Run on port 8090:
```
locust --web-port 8090 -f authentication_server.py
```

- _match_server.py_: mimics game servers matching players together, and closing games out.

Run on the CLI:
```
locust -H http://127.0.0.1:8081 -f match_server.py --headless -u=1 -r=1 -t=10s
```

Run on port 8091:
```
locust --web-port 8091 -f match_server.py
```
