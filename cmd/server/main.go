package main

import (
	"encoding/json"
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
	go GetMembers(httpPort)
	log.Fatal(app.Listen(fmt.Sprintf(":%d", httpPort)))
}

// type Member struct {
// 	Name string `json:"name"`
// 	Addr string `json:"addr"`
// 	Port int    `json:"port"`
// }

func GetMembers(httpPort int) *[]distributed.Member {
	app1 := fiber.AcquireAgent()
	defer fiber.ReleaseAgent(app1)

	req := app1.Request()
	req.Header.SetMethod(fiber.MethodGet)
	req.SetRequestURI(fmt.Sprintf("http://127.0.0.1:%s/cache/members", strconv.Itoa(httpPort)))

	if err := app1.Parse(); err != nil {
		log.Fatalf("Failed to parse request: %v", err)
	}

	// Send request then Get response
	statusCode, body, errs := app1.Bytes()
	if len(errs) > 0 {
		log.Fatalf("Failed to make request: %v", errs[0])
	}

	if statusCode != fiber.StatusOK {
		log.Fatalf("Request failed with status code: %d", statusCode)
	}

	// Parse JSON response
	// var members []Member
	var members []distributed.Member
	if err := json.Unmarshal(body, &members); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	// Log the members
	fmt.Println("\nCluster Members:")
	fmt.Println("----------------")
	for _, member := range members {
		fmt.Printf("Node: %s\n", member.Name)
		fmt.Printf("Address: %s:%d\n", member.Addr, member.Port)
		fmt.Println("----------------")
	}

	distributed.Members = members
	return &members
}
