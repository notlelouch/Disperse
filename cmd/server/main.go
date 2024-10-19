package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/notlelouch/Distributed-Cache/pkg/distributed"
)

func main() {
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatalf("Invalid PORT: %v", port)
	}

	node_name := os.Getenv("NODE_NAME")

	peer := os.Getenv("PEER")

	dc, err := distributed.NewDistributedCache(port, node_name)
	if err != nil {
		log.Fatalf("Failed to create distributed cache: %v", err)
	}

	// log.Printf("the peer is %v", peer)

	if peer != "" {
		err = dc.JoinCluster(peer)
		if err != nil {
			log.Fatalf("Failed to Join cluster: %v", err)
		}
	}

	http.HandleFunc("/cache/", dc.HTTPHandler)
	log.Printf("Server is running on port: %d", port)
	log.Print(dc.Config.Name)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
