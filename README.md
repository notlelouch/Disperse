# Distributed Cache System

The Distributed Cache System is a scalable, high-performance, and fault-tolerant caching solution built using Go and HashiCorp's Memberlist library. It allows data to be cached across multiple nodes in a distributed cluster, enabling robust data storage, efficient retrieval, and automatic failure detection using a gossip-based protocol.

This project extends the functionality of a basic cache by implementing distributed caching across multiple nodes, supporting cluster membership and dynamic peer discovery.

## Key Features

- **Distributed Cache:** Seamless caching across multiple nodes, enabling high availability and fault tolerance.
- **Cluster Membership:** Automatic peer discovery and cluster management using a gossip-based protocol, powered by HashiCorp’s Memberlist.
- **Failure Detection:** Real-time detection and handling of node failures, ensuring uninterrupted service.
- **HTTP API:** Simple REST API for interacting with the cache, supporting basic CRUD operations (Get, Put, Delete).
- **Configurable Gossip Parameters:** Fine-tune gossip and failure detection settings for optimized performance.

## Tech Stack

- **Language:** Go
- **Library:** HashiCorp Memberlist
- **Network Communication:** HTTP for client interaction, Gossip protocol for node-to-node communication

## Installation

Go 1.18 or later
Git


- ***Clone the Repository:***

   ```bash
    git clone https://github.com/notlelouch/distributed-cache.git
    cd distributed-cache
   ```
   
- ***Install Dependencies:***
  ```bash
  go mod tidy
  ```
  
## Usage

- ### Running a Distributed Cache Node
  Each instance of the cache runs on a specific port, and nodes can join a cluster by connecting to existing peers. To start a node(be in the root directory):
  - Start the first node(in one terminal):
   ```bash
   # Terminal 1
   export PORT=<memberlist_port>
   export HTTP_PORT=<fiber_port>
   export NODE_NAME=<node_name>
   make run
   ```
    - Join the cluster(in another terminal instance):
   ```bash
   # Terminal 2
   export PORT=<memberlist_port>       
   export HTTP_PORT=<fiber_port    
   export PEER=127.0.0.1:<memberlist_port>  
   export NODE_NAME=<node_name>
   make run
   ```

  - Example:
  ```bash
   export PORT=7947         # Memberlist port
   export HTTP_PORT=8001    # Fiber port that is different from Memberlist port
   export PEER=127.0.0.1:7946  # Connect to first node's Memberlist port
   export NODE_NAME=beta
   make run
   ``` 
- ### Interacting with the Cache
  The cache can be accessed via simple HTTP requests. Each node in the cluster can handle HTTP requests to interact with the distributed cache.
  #### By default all the requests are being broadcasted in the cluster, but you can optionally set the X-Is-Sync flag to true for any request, then the request will only be a sync request(i.e limited to that particular node) and will not be broadcasted to other nodes in the cluster
  1. #### Get a Value:
      Retrieve a cached value by sending a `GET` request to `/cache/{key}`.
  
     ```bash
      curl -X GET \
     -H "Content-Type: application/json" \
     http://localhost:8000/cache/John10
     ```
     ***Response:*** Returns the cached value if found, or a 404 if the key is not in the cache.

  2. #### Put a Value:
      Store a value in the cache using a `PUT` request with `value` and `duration` as parameters. `duration` is the expiration time (in seconds) for the cached value
  
     ```bash
     curl -X PUT \
     -H "Content-Type: application/json" \
     -d '{"value": "test", "duration": "9000000000000"}' \
     http://localhost:8002/cache/John10
     ```
     ***Response:*** Returns the cached value if found, or a 404 if the key is not in the cache.
     ***Parameters:***
      - `value`: The value to store in the cache.
      - `duration`: How long (in nanoseconds) the value should be stored.

  4. #### Delete a Value:
      Remove a cached value by sending a DELETE request to /cache/{key}.
     ```bash
      curl -X DELETE \
        -H "Content-Type: application/json" \
        http://localhost:8001/cache/John10
     ```
     ***Response:*** Deletes the key if found, no output on success.
  

## Project Structure

```
├── cmd/
│   └── server/
│       └── main.go               # Entry point, starts the cache and joins the cluster
├── pkg/
│   ├── cache/
│   │   ├── cache.go              # Core cache logic for managing data storage and expiration
│   │   └── cache_test.go         # Test file for cache.go
│   └── distributed/
│       ├── distributed.go        # Implementation of the distributed cache, cluster management, HTTP API handlers
│       └── distributed_test.go   # Test file for distributed.go
├── go.mod                        # Go module dependencies
├── go.sum                        # Go module versions
├── README.md                     # Project documentation
└── LICENSE                       # License file for the project
```
### Important Files

- **cache.go:** Contains the basic cache functionality (get, set, delete, expiration).
- **distributed.go:** Handles cluster membership, peer discovery, failure detection, and HTTP request handling.
- **main.go:** Entry point for running the distributed cache node.

## Configuration
The cache uses default settings for the gossip-based protocol and cluster management. However, you can customize the following parameters in `distributed.go` for tuning performance:
- **GossipInterval:** Time interval between gossip messages.
- **GossipNodes:** Number of nodes to gossip with in each interval.
- **ProbeInterval:** Frequency of checking for failed nodes.
- **ProbeTimeout:** Timeout for failure detection after missing heartbeats.

```bash
  config.GossipInterval = 300 * time.Millisecond
  config.GossipNodes = 3
  config.ProbeInterval = 1 * time.Second
  config.ProbeTimeout = 5 * time.Second
```

## How It Works
The Distributed Cache System ensures scalability, fault tolerance, and high availability through a robust architecture:

- **Cluster Membership:** Nodes form a dynamic cluster using HashiCorp's Memberlist, exchanging state via a gossip protocol for seamless peer discovery and consistency.
- **Failure Detection:** Periodic heartbeats detect node failures in real-time, with automatic adjustments to maintain cluster integrity.
- **Caching Operations:** A RESTful API enables efficient data storage, retrieval, and deletion. Requests broadcast across the cluster by default, with optional local-only operations.
- **Scalability & Resilience:** Nodes join or leave seamlessly, maintaining service availability and enabling horizontal scaling.

This design combines simplicity and power, ensuring efficient distributed caching under real-world conditions.
## Contributing

Contributions are welcome! If you have ideas for improvements, new features, or bug fixes, please open an issue or submit a pull request.
