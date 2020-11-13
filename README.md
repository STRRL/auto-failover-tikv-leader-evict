# auto-failover-tikv-leader-evict

## Overview

This project provides `evictor`. `evictor` pulls network latency metrics (provided by blackbox_exporter) from prometheus every 15 seconds(by default). If a tikv node trapped into network latency for a certain time, it will execute `pd-ctl scheduler add evict-leader-scheduler <storeId>` for evicting leaders on this tikv.

## Compile

```shell 
make
```

## Usage

Example:  

```shell
./bin/evictor --prometheus=http://10.108.242.231:9090 --pd=10.99.183.247:2379 --debug
```

## Flags

`--prometheus <string>` address of prometheus; required;

`--pd <string>` address of pd; required;

`--max-evicted <uint>` max number of tikv which could be evicted leader by this tool; optional; default: 10

`--interval <duration>` interval for refresh latency metrics; optional; default: 15s

`--threshold <duration>` a node which hold a latency longer than threshold will be treated as unhealthy; optional; default: 1s

`--pending-for-evict <duration>` an unhealthy tikv node will be evicted after this duration; optional; default: 1m

`--pending-for-recover <duration>` an evicted tikv with stable latency will recover at least after this duration; optional; default: 30s

`--debug` print debug logs; optional; default: false
