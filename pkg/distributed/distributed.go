package distributed

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/notlelouch/Distributed-Cache/pkg/cache"
)

type DistributedCache struct {
	Cache  *cache.Cache
	List   *memberlist.Memberlist
	Config *memberlist.Config
}

// func NewDistributedCache(bindAddr string, port int) (*DistributedCache, error) {
// 	// Initialize the local cache
// 	cacheInstance := cache.NewCache()
//
// 	// Configure memberlist with default LAN settings and custom port
// 	config := memberlist.DefaultLANConfig()
// 	config.BindAddr = bindAddr
// 	config.BindPort = port
// 	config.AdvertiseAddr = bindAddr
// 	config.AdvertisePort = port
//
// 	// Create a memberlist instance
// 	list, err := memberlist.Create(config)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	// Create the DistributedCache instance
// 	dc := &DistributedCache{
// 		Cache:  cacheInstance,
// 		List:   list,
// 		Config: config,
// 	}
//
// 	return dc, nil
// }

func NewDistributedCache(port int) (*DistributedCache, error) {
	// Initialize the local cache
	cacheInstance := cache.NewCache()

	// Configure memberlist with default LAN settings and custom port
	config := memberlist.DefaultLANConfig()
	config.BindPort = port
	config.AdvertisePort = port

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
	_, err := dc.List.Join([]string{peer})
	return err
}

func (dc *DistributedCache) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[len("/cache/"):]
	switch r.Method {
	case http.MethodGet:
		value, found := dc.Cache.Get(key)
		if !found {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		fmt.Fprintf(w, "%v", value)

	case http.MethodPut:
		value := r.PostFormValue("value")
		durationStr := r.PostFormValue("duration")
		duration, err := strconv.ParseInt(durationStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		dc.Cache.Set(key, value, time.Duration(duration))

	case http.MethodDelete:
		dc.Cache.Delete(key)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Method not allowed")
	}
}
