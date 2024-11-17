package distributed

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/hashicorp/memberlist"
	"github.com/notlelouch/Distributed-Cache/pkg/cache"
)

type DistributedCache struct {
	Cache  *cache.Cache
	List   *memberlist.Memberlist
	Config *memberlist.Config
	mu     sync.RWMutex
}

var UpdatedMembersList []*memberlist.Node

func NewDistributedCache(memberlistPort int, node_name string) (*DistributedCache, error) {
	// Initialize the local cache
	cacheInstance := cache.NewCache()
	config := memberlist.DefaultLocalConfig()
	config.Name = node_name
	config.BindAddr = "127.0.0.1"

	// config.BindPort = port
	config.BindPort = memberlistPort // Use different port for memberlist
	config.AdvertiseAddr = "127.0.0.1"

	// config.AdvertisePort = port
	config.AdvertisePort = memberlistPort
	// Configure memberlist with default LAN settings and custom port
	// config := memberlist.DefaultLANConfig()
	// config.BindPort = port
	// config.AdvertisePort = port

	// Create a memberlist instance
	list, err := memberlist.Create(config)
	if err != nil {
		return nil, err
	}

	// Create the DistributedCache instance
	dc := &DistributedCache{
		Cache:  cacheInstance,
		List:   list,
		Config: config,
	}

	return dc, nil
}

func NewDistributedCacheWithConfig(config *memberlist.Config) (*DistributedCache, error) {
	// Initialize the local cache
	cacheInstance := cache.NewCache()
	// Create a memberlist instance
	list, err := memberlist.Create(config)
	if err != nil {
		return nil, err
	}
	// Create the DistributedCache instance
	dc := &DistributedCache{
		Cache:  cacheInstance,
		List:   list,
		Config: config,
	}

	return dc, nil
}

// JoinCluster allows the current node to join an existing cluster using a peer address.
func (dc *DistributedCache) JoinCluster(peer string) error {
	// Log initial members
	members := dc.List.Members()
	log.Printf("Members before joining: %d", len(members))

	// Join cluster
	_, err := dc.List.Join([]string{peer})

	// Log updated members after joining
	UpdatedMembersList = dc.List.Members()
	log.Printf("Members after joining: %d", len(UpdatedMembersList))
	for _, member := range UpdatedMembersList {
		log.Printf("Node: %s, Address: %s:%d", member.Name, member.Addr, member.Port)
	}

	return err
}

// // net/http Handlers
// func (dc *DistributedCache) HTTPHandler(w http.ResponseWriter, r *http.Request) {
// 	key := r.URL.Path[len("/cache/"):]
// 	switch r.Method {
// 	case http.MethodGet:
// 		log.Printf("METHODGET#####")
// 		value, found := dc.Cache.Get(key)
// 		if !found {
// 			w.WriteHeader(http.StatusNotFound)
// 			return
// 		}
// 		fmt.Fprintf(w, "%v", value)
// 		log.Printf("value of %s is %s", key, value)
//
// 	case http.MethodPut:
// 		log.Printf("METHODEPUT#####")
// 		value := r.PostFormValue("value")
// 		durationStr := r.PostFormValue("duration")
// 		duration, err := strconv.ParseInt(durationStr, 10, 64)
// 		if err != nil {
// 			w.WriteHeader(http.StatusBadRequest)
// 			return
// 		}
// 		dc.Cache.Set(key, value, time.Duration(duration))
// 		log.Printf("Successfully Set %s to %s", key, value)
//
// 	case http.MethodDelete:
// 		log.Printf("METHODEDELETE####")
// 		dc.Cache.Delete(key)
//
// 	default:
// 		w.WriteHeader(http.StatusMethodNotAllowed)
// 		fmt.Fprintf(w, "Method not allowed")
// 		log.Printf("Successfully delete %s", key)
// 	}
// }

// Fiber Handler
func (dc *DistributedCache) FiberHandler(c *fiber.Ctx) error {
	fmt.Println("################   FiberHandler   ##################")
	fmt.Println(c.Request())

	// Create sync request payload
	type SyncPayload struct {
		Method   string `json:"method"`
		Key      string `json:"key"`
		Value    string `json:"value"`
		Duration string `json:"duration"`
	}

	key := c.Params("key")
	switch c.Method() {
	case "PUT":
		log.Println("METHODEPUT#####")
		log.Println(string(c.Body()))
		// value := c.FormValue("value")
		// durationStr := c.FormValue("duration")
		// log.Printf("value: %s, duration: %s", value, durationStr)
		//
		// duration, err := strconv.ParseInt(durationStr, 10, 64)
		// if err != nil {
		// 	return c.SendStatus(fiber.StatusBadRequest)
		// }
		// log.Printf("##### Preparing to set value in cahce #####")
		// dc.Cache.Set(key, value, time.Duration(duration))
		// log.Printf("##### Successfully set value in cahce #####")

		var requestBody struct {
			Value    string `json:"value"`
			Duration string `json:"duration"`
		}

		// Check if the Content-Type is JSON
		if !c.Is("json") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Content-Type must be application/json",
			})
		}

		log.Println("##### Unmarshalling #####")
		// Unmarshal and handle any errors
		if err := json.Unmarshal(c.Body(), &requestBody); err != nil {
			fmt.Printf("Unmarshal error: %v\n", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid JSON format",
			})
		}
		log.Println("##### Unmarshalling Complete#####")

		// Print the unmarshalled data for debugging
		fmt.Printf("Unmarshalled data: %+v\n", requestBody)

		// Use the JSON values directly
		value := requestBody.Value
		durationStr := requestBody.Duration

		// Verify we got the values
		if value == "" || durationStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Missing required fields",
			})
		}

		log.Printf("value: %s, duration: %s", value, durationStr)

		payload := SyncPayload{
			Method:   c.Method(),
			Key:      key,
			Value:    value,
			Duration: durationStr,
		}

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Failed to marshal sync payload: %v", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		// Create agent for HTTP request
		httpPort := "8000"
		agent := fiber.AcquireAgent()
		defer fiber.ReleaseAgent(agent)

		// Setup request
		req := agent.Request()
		req.Header.SetMethod(fiber.MethodPost)
		req.Header.SetContentType("application/json")
		req.SetRequestURI(fmt.Sprintf("http://127.0.0.1:%s/cache/sync", httpPort))
		req.SetBody(jsonPayload)

		log.Printf("##### Setting up the request #####")
		if err := agent.Parse(); err != nil {
			log.Printf("Failed to parse request: %v", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		log.Printf("##### agent parsed the request #####")
		log.Printf(req.String())

		// Send request
		statusCode, body, errs := agent.Bytes()
		if len(errs) > 0 {
			log.Printf("Failed to make sync request: %v", errs[0])
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		log.Printf("##### Sent the request #####")

		if statusCode != fiber.StatusOK {
			log.Printf("Sync request failed with status code %d: %s", statusCode, string(body))
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		log.Printf("Successfully Set %s to %s and broadcasted", key, value)
		return c.SendStatus(fiber.StatusOK)

	case "GET":
		log.Printf("METHODGET#####")
		fmt.Printf("############lALALALALA############ %s", c.Method())

		value, found := dc.Cache.Get(key)
		if !found {
			return c.SendStatus(fiber.StatusNotFound)
		}

		payload := SyncPayload{
			Method: c.Method(),
			Key:    key,
			Value:  fmt.Sprintf("%v", value),
		}

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Failed to marshal sync payload: %v", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		// Create agent for HTTP request
		httpPort := "8000"
		agent := fiber.AcquireAgent()
		defer fiber.ReleaseAgent(agent)

		// Setup request
		req := agent.Request()
		req.Header.SetMethod(fiber.MethodPost)
		req.Header.SetContentType("application/json")
		req.SetRequestURI(fmt.Sprintf("http://127.0.0.1:%s/cache/sync", httpPort))
		req.SetBody(jsonPayload)

		if err := agent.Parse(); err != nil {
			log.Printf("Failed to parse request: %v", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		// Send request
		statusCode, body, errs := agent.Bytes()
		if len(errs) > 0 {
			log.Printf("Failed to make sync request: %v", errs[0])
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		if statusCode != fiber.StatusOK {
			log.Printf("Sync request failed with status code %d: %s", statusCode, string(body))
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		log.Printf("value of %s is %s", key, value)
		return c.SendString(fmt.Sprintf("%v", value))

	case "DELETE":
		log.Printf("METHODEDELETE####")
		dc.Cache.Delete(key)

		payload := SyncPayload{
			Method: c.Method(),
			Key:    key,
		}

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Failed to marshal sync payload: %v", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		// Create agent for HTTP request
		httpPort := "8000"
		agent := fiber.AcquireAgent()
		defer fiber.ReleaseAgent(agent)

		// Setup request
		req := agent.Request()
		req.Header.SetMethod(fiber.MethodPost)
		req.Header.SetContentType("application/json")
		req.SetRequestURI(fmt.Sprintf("http://127.0.0.1:%s/cache/sync", httpPort))
		req.SetBody(jsonPayload)

		if err := agent.Parse(); err != nil {
			log.Printf("Failed to parse request: %v", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		// Send request
		statusCode, body, errs := agent.Bytes()
		if len(errs) > 0 {
			log.Printf("Failed to make sync request: %v", errs[0])
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		if statusCode != fiber.StatusOK {
			log.Printf("Sync request failed with status code %d: %s", statusCode, string(body))
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		log.Printf("Successfully deleted %s and broadcasted", key)
		return c.SendStatus(fiber.StatusOK)

	default:
		return c.Status(fiber.StatusMethodNotAllowed).SendString("Method not allowed")
	}
}

func (dc *DistributedCache) HandleGetMembers(c *fiber.Ctx) error {
	fmt.Print("################   HandleGetMembers   ##################")
	// dc.mu.RLock()
	members := dc.List.Members()
	// dc.mu.RUnlock()

	response := make([]fiber.Map, len(members))

	for i, member := range members {
		response[i] = fiber.Map{
			"name": member.Name,
			"addr": member.Address(),
			"port": member.Port,
		}
	}
	log.Printf("response is %s", response)
	return c.JSON(response)
}

func (dc *DistributedCache) HandleReqBroadcast(c *fiber.Ctx) error {
	fmt.Println("################  HandleReqBroadcast  ##################")

	response := c.Body()
	var request map[string]interface{}
	if err := json.Unmarshal(response, &request); err != nil {
		log.Printf("Failed to unmarshal response: %v", err)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	fmt.Println("##### response unmarshalled into request ######")
	fmt.Println(request)
	// type SyncPayload struct {
	// 	value    string `json:"value"`
	// 	duration string `json:"duration"`
	// }
	//
	// payload := SyncPayload{
	// 	value:    fmt.Sprintf("%v", request["value"]),
	// 	duration: fmt.Sprintf("%v", request["duration"]),
	// }

	// Convert float64 values to strings when creating the payload
	payload := fiber.Map{
		// "key":      fmt.Sprintf("%v", request["key"]),
		"value":    fmt.Sprintf("%v", request["value"]),
		"duration": fmt.Sprintf("%v", request["duration"]),
	}

	fmt.Print("##### This is the payload #####")
	fmt.Println(payload)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal sync payload: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	httpPort := "8001"
	agent := fiber.AcquireAgent()
	defer fiber.ReleaseAgent(agent)

	req := agent.Request()
	req.Header.SetMethod(request["method"].(string))
	req.Header.SetContentType("application/json")
	// req.Header.SetContentType("application/x-www-form-urlencoded") // Changed to match original request
	req.SetRequestURI(fmt.Sprintf("http://127.0.0.1:%s/cache/%s", httpPort, request["key"].(string)))
	req.SetBody(jsonPayload)

	if err := agent.Parse(); err != nil {
		log.Printf("Failed to parse request: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	fmt.Println(req)

	// Send request
	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		log.Printf("Failed to make sync request: %v", errs[0])
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	fmt.Println("#####  Request Sent  #####")
	if statusCode != fiber.StatusOK {
		log.Printf("Sync request failed with status code %d: %s", statusCode, string(body))
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	log.Printf("Response body is %s", string(response))
	return c.JSON(response)
}
