# Distributed Write-Ahead Log

fault-tolerant distributed Write-Ahead Log (WAL) built in Go using the Raft consensus algorithm.

The project demonstrates how a distributed log can provide **strong consistency**, **leader election**, **automatic failover**, and **replicated storage** across multiple servers. 

---

## Deploy to Kubernetes locally

### Requirements

- Go 1.24+
- Docker
- Kubernetes
- kubectl
- Helm
- Make

### Run
```bash
make tidy
```

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

Forward leader pod to localhost:8400
```bash
make forward-port
````

### Produce request

```bash
grpcurl -d '{"record": {"value": "aGVsbG8gd29ybGQ="}}' -plaintext localhost:8400 log.v1.Log/Produce
```

### Consume request
```bash
grpcurl -d '{"offset": 0}' -plaintext localhost:8401 log.v1.Log/Consume
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


## Learning Goals

This project was built to explore:

- distributed systems
- system programming
- consensus algorithms
- fault tolerance
- leader election
- distributed storage
- Kubernetes deployments
- service discovery


---

## License

MIT
