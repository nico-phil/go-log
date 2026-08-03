# go-log (Distributed Write-Ahead Log)

## what is go-log ?
go-log is fault-tolerant distributed Write-Ahead Log (WAL) built in Go using the Raft consensus algorithm. The system exposes an interface to added and retrive records like producer-consumer format

Write-Ahead Log (WAL) is a safety file where databases record changes to disk before updating the main data. It ensures data is not lost if the server crashes. Systems like PostgreSQL, MYSQL, use it to guarantee reliability. Even Raft uses a WAL to replicate state.

Raft is a distributed consensus protocol that allows a cluster of nodes to agree on the state of a replicated state machine, even in the presence of node failures or temporary network partitions. It is one of the fundamental building blocks of fault-tolerant distributed systems.

# What makes the WAL distriubted?
It's distributed because the log isn't stored on a single node. Every write is replicated across multiple nodes using the Raft consensus algorithm. The leader coordinates replication, entries are committed only after a majority quorum acknowledges them, every node applies the same sequence of log entries to maintain a consistent replicated state machine, and the cluster can tolerate node failures through leader election and automatic recovery.

The project demonstrates how a distributed log can provide **strong consistency**, **leader election**, **automatic failover**, and **replicated storage** across multiple servers. 

---

## Quick Deploy to Kubernetes locally

### Requirements

- Go 1.24+
- Docker
- Kubernetes
- kubectl
- Helm
- Make

### Create cluster
```bash
kind create cluster
```

### Download the project
```bash
git clone https://github.com/nico-phil/go-log.git
cd go-log
make tidy
```

### Deploy
make deploy-local will automatically:
- Build the project
- Contenairized it and load the image into the kubernetes cluster
- Install The project Helm Chart
```bash
make deploy-local
```

Verify that the pods are running:
```bash
kubectl get pods
```

Example output:
```text
NAME        READY   STATUS    RESTARTS   AGE
proglog-0   1/1     Running   0          142m
proglog-1   1/1     Running   0          142m
proglog-2   1/1     Running   0          142m
```

### Forward leader pod to localhost:8400
```bash
make forward-port
```

### Produce request
```bash
grpcurl -d '{"record": {"value": "aGVsbG8gd29ybGQ="}}' -plaintext localhost:8400 log.v1.Log/Produce
```

or run an example request
```bash
make produce 
```

### Consume request
```bash
grpcurl -d '{"offset": 0}' -plaintext localhost:8401 log.v1.Log/Consume
```

or run an example request
```bash
make consume
```

### Produce stream request
```bash
grpcurl -plaintext \
  -d @ \
  localhost:8400 \
  log.v1.Log/ProduceStream <<EOF
{"record":{"value":"Zm9v"}}
{"record":{"value":"YmF6"}}
{"record":{"value":"YmFy"}}
EOF

```

### Consume stream request
```bash
grpcurl \
  -plaintext \
  -d '{"offset":0}' \
  localhost:8401 \
  log.v1.Log/ConsumeStream
```


## Running a Three-Node Cluster (Without Kubernetes)

This example starts a cluster with three nodes. Each node uses its own configuration file:

- Node 0 → `config_0.yaml`
- Node 1 → `config_1.yaml`
- Node 2 → `config_2.yaml`

Start each node in a separate terminal:

```bash
go run cmd/proglog/main.go --config-file=config_0.yaml
go run cmd/proglog/main.go --config-file=config_1.yaml
go run cmd/proglog/main.go --config-file=config_2.yaml
```

## Example Configuration

Below is an example configuration for **Node 0** (`config_0.yaml`):

```yaml
data-dir: data_0
rpc-port: 8400
node-name: proglog-0
bind-addr: 127.0.0.1:7373
bootstrap: true
start-join-addrs: []

server-tls-cert-file: /Users/admin/.proglog/server.pem
server-tls-key-file: /Users/admin/.proglog/server-key.pem
server-tls-ca-file: /Users/admin/.proglog/ca.pem

peer-tls-cert-file: /Users/admin/.proglog/server.pem
peer-tls-key-file: /Users/admin/.proglog/server-key.pem
peer-tls-ca-file: /Users/admin/.proglog/ca.pem

acl-model-file: /Users/admin/.proglog/model.conf
acl-policy-file: /Users/admin/.proglog/policy.csv
```

> **Note:** Each node should have its own `data-dir`, `rpc-port`, `node-name`, and `bind-addr`. The TLS and ACL files can be shared by all nodes. Only the bootstrap node should have `bootstrap: true`; all other nodes should set `bootstrap: false` and configure `start-join-addrs` with the bootstrap node's address.


### Run tests
```bash
make test
```


## Features
- Raft consensus algorithm
- Leader election
- Log replication
- Fault tolerance
- Automatic node discovery with Serf
- Multi-node Kubernetes deployment
- gRPC API: Append and consume record, append and consume stream of records 
- Unit tests
- Integration tests
- End-to-end tests


## Architecture
```

                    +----------------+
                    |     Client     |
                    +--------+-------+
                             |
                          gRPC API
                             |
               +-------------+--------------+
               |                            |
         +-----v------+              +------v-----+
         |   Leader   |<-----------> |  Follower  |
         +-----+------+              +------+-----+
               |                            
               |                 
         +-----v------+              +------v-----+
         |  Follower  |              |    Serf    | 
         +------------+              +------------+
                                    serf instance on each node for Automatic node discovery
              
                      Raft Replication
```


## Applications
Distributed logs are the foundation of many production systems.

Examples include:

- Apache Kafka
- etcd
- Consul
- CockroachDB
- Event sourcing platforms
- Streaming systems
- Distributed databases
- Message queues
---

## License

MIT
