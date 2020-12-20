RoundNS
=======

Simplest DNS server for health-checking and load-balancing.


## How to use

### 1. Make config file

``` yaml
#address: :53

#log:
#  level: warn

#metrics:
#  namespace: roundns
#  address: :8553

#ttl: 5m  # default TTL

services:
  - fqdn: a.test.local.
    type: A
    ttl: 300
    strategy: round-robin
    health:
      command:
        - "/bin/sh"
        - "-c"
        - "curl -fs http://$HOST"
      interval: 1m
    hosts:
      - 127.0.0.1
      - 127.0.0.2

  - fqdn: b.test.local.
    type: A
    ttl: 5m
    health:
      command: curl -fs http://$HOST  # It's shorthand of ["sh", "-c", "curl -fs http://$HOST"]
      interval: 1h
    hosts:
      - 127.0.1.1
      - 127.0.1.2

  - fqdn: c.test.local.
    type: A
    health: curl -fs http://$HOST  # It's shorthand more
    hosts:
      - 127.0.1.1
      - 127.0.1.2

  # load balancing without health checking
  - fqdn: cname.test.local.
    type: CNAME
    hosts:
      - a.test.local.
      - b.test.local.
```


### 2. Run server

``` shell
$ roundns -config ./config.yml
```

Now server is running on 0.0.0.0:53/udp, and you can get metrics on http://127.0.0.1:8553/metrics
