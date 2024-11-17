package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/notlelouch/Distributed-Cache/pkg/distributed"
)

func main() {
	// Separating the HTTP (Fiber) port from the Memberlist port
	// (Fiber is not supporting both Memberlist and HTTP communication on the same port)
	// port, err := strconv.Atoi(os.Getenv("PORT"))
	httpPort, _ := strconv.Atoi(os.Getenv("HTTP_PORT"))
	memberlistPort, _ := strconv.Atoi(os.Getenv("PORT"))
	// if err != nil {
	// 	log.Fatalf("Invalid PORT: %v", port)
	// }

	node_name := os.Getenv("NODE_NAME")

	peer := os.Getenv("PEER")

	// dc, err := distributed.NewDistributedCache(port, node_name)
	dc, err := distributed.NewDistributedCache(memberlistPort, node_name)
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

	// // net/http Handlers
	// http.HandleFunc("/cache/", dc.HTTPHandler)
	// log.Printf("Server is running on port: %d", port)
	// log.Print(dc.Config.Name)
	// log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))

	// Fiber Handler
	app := fiber.New()
	app.Get("/cache/members", dc.HandleGetMembers)
	app.All("/cache/:key", dc.FiberHandler)

	log.Printf("Server is running on port: %d", httpPort)
	log.Print(dc.Config.Name)
	// log.Fatal(app.Listen(fmt.Sprintf(":%d", port)))
	log.Fatal(app.Listen(fmt.Sprintf(":%d", httpPort)))
}
