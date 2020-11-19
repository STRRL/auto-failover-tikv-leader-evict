# auto-failover-tikv-leader-evict

## Overview

This project provides `evictor`. `evictor` pulls network latency metrics (provided by blackbox_exporter) from prometheus every 15 seconds(by default). If a tikv node trapped into network latency for a certain time, it will execute `pd-ctl scheduler add evict-leader-scheduler <storeId>` for evicting leaders on this tikv.

Caution: This tool is only compatible with TiDB v3.x, not play well with TiDB v4.x.

## Prerequisites

- Here must a metric named `probe_duration_seconds` exist in your prometheus. This metric is provided by `blackbox_exporter`, if you deploy tidb by `tidb-ansible`, it should be exists.
- `evictor` should runs on a node which contains `pd-ctl` in its `$PATH`, as a daemon service.

## Compile

```shell
make
```

## Usage

Example:  

```shell
./bin/evictor --prometheus=http://10.108.242.231:9090 --pd=10.99.183.247:2379 --debug
```

Systemd service example:

```service
[Unit]
Description=Auto failover tikv-leader-evict Service

[Service]
Type=simple
User=tidb
Restart=on-failure
RestartSec=5s
ExecStart=/usr/local/bin/evictor --prometheus=http://10.96.206.21:9090 --pd=10.98.225.221:2379 --interval 10s --threshold 1s --pending-for-evict=60s --pending-for-recover=30s --debug

[Install]
WantedBy=multi-user.target
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

## Important Logs

When a tikv store is evicted/recovered, it will print some logs like:

```json
{"level":"info","ts":1605602637.155664,"msg":"tikv node evicted","store":{"id":4,"address":"basic-tikv-2.basic-tikv-peer.tidb-cluster.svc:20160"}}
...
{"level":"info","ts":1605602685.4265118,"msg":"tikv node recovered","store":{"id":4,"address":"basic-tikv-2.basic-tikv-peer.tidb-cluster.svc:20160"}}
```
