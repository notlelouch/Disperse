package distributed

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hashicorp/memberlist"
	"github.com/notlelouch/Distributed-Cache/pkg/cache"
)

type DistributedCache struct {
	Cache    *cache.Cache
	List     *memberlist.Memberlist
	Config   *memberlist.Config
	mu       sync.RWMutex
	Meta     []byte
	HTTPPort int
}

type Member struct {
	Name     string `json:"name"`
	Addr     string `json:"addr"`
	Port     int    `json:"port"`
	HTTPPort int    `json:"http_port"`
}

var UpdatedMembersList []*memberlist.Node

// Defining a delegate to handle metadata
// Each node in the cluster will have its own delegate
// The delegate stores that node's HTTP port
// Memberlist uses the delegate to share metadata (including HTTP port) with other nodes

type cacheDelegate struct {
	httpPort int
}
type NodeMetadata struct {
	HTTPPort int `json:"http_port"`
}

// NodeMeta is required by the Delegate interface
func (d *cacheDelegate) NodeMeta(limit int) []byte {
	// Create metadata with HTTP port
	meta := NodeMetadata{
		HTTPPort: d.httpPort,
	}

	metaBytes, _ := json.Marshal(meta)
	if len(metaBytes) > limit {
		return metaBytes[:limit]
	}
	return metaBytes
}

// These methods are required by the Delegate interface but we won't use them
func (d *cacheDelegate) NotifyMsg([]byte)                           {}
func (d *cacheDelegate) GetBroadcasts(overhead, limit int) [][]byte { return nil }
func (d *cacheDelegate) LocalState(join bool) []byte                { return nil }
func (d *cacheDelegate) MergeRemoteState(buf []byte, join bool)     {}

func NewDistributedCache(memberlistPort int, httpPort int, node_name string) (*DistributedCache, error) {
	// Initialize the local cache
	cacheInstance := cache.NewCache()
	config := memberlist.DefaultLocalConfig()
	config.Name = node_name
	config.BindAddr = "127.0.0.1"

	// config.BindPort = port
	config.BindPort = memberlistPort // Use different port for memberlist
	config.AdvertiseAddr = "127.0.0.1"
	// config.

	// config.AdvertisePort = port
	config.AdvertisePort = memberlistPort
	// Configure memberlist with default LAN settings and custom port
	// config := memberlist.DefaultLANConfig()
	// config.BindPort = port
	// config.AdvertisePort = port

	meta := NodeMetadata{
		HTTPPort: httpPort,
	}

	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %v", err)
	}

	// Create and set the delegate
	delegate := &cacheDelegate{
		httpPort: httpPort,
	}
	config.Delegate = delegate

	// Create a memberlist instance
	list, err := memberlist.Create(config)
	if err != nil {
		return nil, err
	}

	// Create the DistributedCache instance
	dc := &DistributedCache{
		Cache:    cacheInstance,
		List:     list,
		Config:   config,
		HTTPPort: httpPort,
		Meta:     metaBytes,
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
	log.Printf("################ %s", peer)
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

// SyncPayload represents the structure for synchronization requests
type SyncPayload struct {
	Method   string `json:"method"`
	Key      string `json:"key"`
	Value    string `json:"value"`
	Duration string `json:"duration"`
	IsSync   bool   `json:"is_sync"` // Flag to prevent infinite loops
}

// FiberHandler handles the main cache operations
func (dc *DistributedCache) FiberHandler(c *fiber.Ctx) error {
	fmt.Println("################   FiberHandler   ##################")

	key := c.Params("key")
	// Check if this is a sync request by looking at the headers
	isSync := c.Get("X-Is-Sync") == "true"

	switch c.Method() {
	case "PUT":
		log.Println("METHODEPUT#####")

		var requestBody struct {
			Value    string `json:"value"`
			Duration string `json:"duration"`
		}

		if !c.Is("json") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Content-Type must be application/json",
			})
		}

		if err := json.Unmarshal(c.Body(), &requestBody); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid JSON format",
			})
		}

		value := requestBody.Value
		durationStr := requestBody.Duration

		if value == "" || durationStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Missing required fields",
			})
		}

		// Only broadcast to other nodes if this is not a sync request
		if !isSync {
			payload := SyncPayload{
				Method:   c.Method(),
				Key:      key,
				Value:    value,
				Duration: durationStr,
				IsSync:   true,
			}

			if err := dc.broadcastToOtherNodes(payload); err != nil {
				log.Printf("Failed to broadcast: %v", err)
				// Continue with local operation even if broadcast fails
			}
		}
		log.Printf("##### broadcastToOtherNodes called #####")

		log.Printf("value: %s, duration: %s", value, durationStr)

		duration, err := strconv.ParseInt(durationStr, 10, 64)
		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		log.Printf("##### Preparing to set value in cahce #####")
		dc.Cache.Set(key, value, time.Duration(duration))
		log.Printf("##### Successfully set value in cahce #####")

		return c.SendStatus(fiber.StatusOK)

	case "GET":
		log.Printf("METHODGET#####")

		if !isSync {
			payload := SyncPayload{
				Method: c.Method(),
				Key:    key,
				IsSync: true,
			}

			if err := dc.broadcastToOtherNodes(payload); err != nil {
				log.Printf("Failed to broadcast: %v", err)
				// Continue with local operation even if broadcast fails
			}
		}
		log.Printf("##### broadcastToOtherNodes called #####")

		value, found := dc.Cache.Get(key)
		if !found {
			return c.SendStatus(fiber.StatusNotFound)
		}
		log.Printf("value of %s is %s", key, value)
		return c.SendString(fmt.Sprintf("%v", value))

	case "DELETE":
		log.Printf("METHODEDELETE####")
		if !isSync {
			payload := SyncPayload{
				Method: c.Method(),
				Key:    key,
				IsSync: true,
			}

			if err := dc.broadcastToOtherNodes(payload); err != nil {
				log.Printf("Failed to broadcast: %v", err)
				// Continue with local operation even if broadcast fails
			}
		}
		log.Printf("##### broadcastToOtherNodes called #####")

		dc.Cache.Delete(key)

		log.Printf("Successfully deleted %s", key)
		return c.SendStatus(fiber.StatusOK)

	default:
		return c.Status(fiber.StatusMethodNotAllowed).SendString("Method not allowed")
	}
}

// broadcastToOtherNodes broadcasts the incoming cache request(to a single node) to all the other nodes in the cluster
func (dc *DistributedCache) broadcastToOtherNodes(payload SyncPayload) error {
	log.Print("Inside the broadcastToOtherNodes function")

	agent := fiber.AcquireAgent()
	defer fiber.ReleaseAgent(agent)

	httpPort := "8001"

	// ##### Request for getting httpPorts of all the nodes in the cluster #####
	req2 := agent.Request()
	req2.Header.SetMethod(fiber.MethodGet)
	req2.Header.SetContentType("application/json")
	req2.SetRequestURI(fmt.Sprintf("http://127.0.0.1:%s/cache/members", httpPort))

	if err := agent.Parse(); err != nil {
		log.Printf("Failed to parse request for member %s", err)
	}

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 || statusCode != fiber.StatusOK {
		log.Printf("Failed to sync with 8001 :", errs)
	}
	var members []Member

	if err := json.Unmarshal(body, &members); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}
	log.Print(members)
	for _, member := range members {
		// ##### Request for broadcasting a Cache request to a specific httpPort in the cluster #####

		// Setup request
		req := agent.Request()
		req.Header.SetMethod(payload.Method)
		req.Header.SetContentType("application/json")
		req.Header.Set("X-Is-Sync", "true") // Mark this as a sync request
		req.SetRequestURI(fmt.Sprintf("http://127.0.0.1:%s/cache/%s", strconv.Itoa(member.HTTPPort), payload.Key))
		req.SetBody(jsonPayload)

		if err := agent.Parse(); err != nil {
			log.Printf("Failed to parse request for member %s", err)
		}

		log.Printf("########## port is : %s", strconv.Itoa(member.HTTPPort))
		statusCode, _, errs = agent.Bytes()
		if len(errs) > 0 || statusCode != fiber.StatusOK {
			log.Printf("Failed to sync with 8001 :", err)
		}

	}

	return nil
}

var Members []Member

func (dc *DistributedCache) HandleGetMembers(c *fiber.Ctx) error {
	fmt.Print("################   HandleGetMembers   ##################")
	// dc.mu.RLock()
	members := dc.List.Members()
	// dc.mu.RUnlock()

	response := make([]fiber.Map, len(members))

	// Previous Method(not working)
	// httpPort, _ := strconv.Atoi(os.Getenv("HTTP_PORT"))

	// for i, member := range members { response[i] = fiber.Map{
	// 		"name": member.Name,
	// 		"addr": member.Address(),
	// 		"port": httpPort,
	// 	}
	// }
	// log.Printf("response is %s", response)

	// New Method(using delegate meta data), smooooooothh operatorrrrrr!
	for i, member := range members {
		// Parse metadata to get HTTP port
		var meta NodeMetadata
		if err := json.Unmarshal(member.Meta, &meta); err != nil {
			// If metadata parsing fails, use default member port
			// meta.HTTPPort = member.Port
			log.Fatal("aab toh BHANG BHOSDAA hogya hai dimag ka!")
		}

		response[i] = fiber.Map{
			"name":      member.Name,
			"addr":      member.Address(),
			"port":      member.Port,
			"http_port": meta.HTTPPort,
		}
	}
	// log.Printf("response is %s", response)

	return c.JSON(response)
}
