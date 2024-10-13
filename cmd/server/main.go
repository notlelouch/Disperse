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

	peer := os.Getenv("PEER")

	dc, err := distributed.NewDistributedCache(port)
	if err != nil {
		log.Fatalf("Failed to create distributed cache: %v", err)
	}

	if peer != "" {
		err = dc.JoinCluster(peer)
		if err != nil {
			log.Fatalf("Failed to Join cluster: %v", err)
		}
	}

	http.HandleFunc("/cache/", dc.HTTPHandler)
	log.Printf("Server is running on port: %d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%d", port), nil))
}
