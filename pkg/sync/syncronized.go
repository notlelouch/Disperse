// Using Fiber to send memberlist request
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
)

type Member struct {
	Name string `json:"name"`
	Addr string `json:"addr"`
	Port int    `json:"port"`
}

func main() {
	fmt.Println("############   MEMBERLIST SYNC   ##########")

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "3000"
	}

	app := fiber.AcquireAgent()
	defer fiber.ReleaseAgent(app)

	req := app.Request()
	req.Header.SetMethod(fiber.MethodGet)
	req.SetRequestURI(fmt.Sprintf("http://127.0.0.1:%s/cache/members", httpPort))

	if err := app.Parse(); err != nil {
		log.Fatalf("Failed to parse request: %v", err)
	}

	// Send request then Get response
	statusCode, body, errs := app.Bytes()
	if len(errs) > 0 {
		log.Fatalf("Failed to make request: %v", errs[0])
	}

	if statusCode != fiber.StatusOK {
		log.Fatalf("Request failed with status code: %d", statusCode)
	}

	// Parse JSON response
	var members []Member
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
}

// // Using standard Http to send memberlist request
// package main
//
// import (
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"log"
// 	"net/http"
// 	"os"
// )
//
// type Member struct {
// 	Name string `json:"name"`
// 	Addr string `json:"addr"`
// 	Port int    `json:"port"`
// }
//
// func main() {
// 	fmt.Println("######################################   MEMBERLIST SYNC   #################################")
//
// 	// Get HTTP port from environment variable, default to 3000 if not set
// 	httpPort := os.Getenv("HTTP_PORT")
// 	if httpPort == "" {
// 		httpPort = "3000"
// 	}
//
// 	// Make request to the members endpoint
// 	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%s/cache/members", httpPort))
// 	if err != nil {
// 		log.Fatalf("Failed to make request: %v", err)
// 	}
// 	defer resp.Body.Close()
//
// 	// Read response body
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Fatalf("Failed to read response: %v", err)
// 	}
//
// 	fmt.Print(body)
//
// 	// Parse JSON response
// 	var members []Member
// 	if err := json.Unmarshal(body, &members); err != nil {
// 		log.Fatalf("Failed to parse JSON: %v", err)
// 	}
//
// 	// Log the members
// 	fmt.Println("\nCluster Members:")
// 	fmt.Println("----------------")
// 	for _, member := range members {
// 		fmt.Printf("Node: %s\n", member.Name)
// 		fmt.Printf("Address: %s:%d\n", member.Addr, member.Port)
// 		fmt.Println("----------------")
// 	}
// }
