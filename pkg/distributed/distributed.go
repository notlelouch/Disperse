package distributed

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hashicorp/memberlist"
	"github.com/notlelouch/Distributed-Cache/pkg/cache"
)

type DistributedCache struct {
	Cache  *cache.Cache
	List   *memberlist.Memberlist
	Config *memberlist.Config
}

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
	number_of_nodes, err := dc.List.Join([]string{peer})
	log.Printf("total nodes that are in the cluster: %d", number_of_nodes)
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
	key := c.Params("key")
	switch c.Method() {
	case "GET":
		log.Printf("METHODGET#####")
		value, found := dc.Cache.Get(key)
		if !found {
			return c.SendStatus(fiber.StatusNotFound)
		}
		log.Printf("value of %s is %s", key, value)
		return c.SendString(fmt.Sprintf("%v", value))

	case "PUT":
		log.Printf("METHODEPUT#####")
		value := c.FormValue("value")
		durationStr := c.FormValue("duration")
		duration, err := strconv.ParseInt(durationStr, 10, 64)
		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		dc.Cache.Set(key, value, time.Duration(duration))
		log.Printf("Successfully Set %s to %s", key, value)
		return c.SendStatus(fiber.StatusOK)

	case "DELETE":
		log.Printf("METHODEDELETE####")
		dc.Cache.Delete(key)
		log.Printf("Successfully delete %s", key)
		return c.SendStatus(fiber.StatusOK)

	default:
		return c.Status(fiber.StatusMethodNotAllowed).SendString("Method not allowed")
	}
}
