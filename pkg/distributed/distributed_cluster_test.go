package distributed

import (
	"testing"
	"time"

	"github.com/hashicorp/memberlist"
)

func TestNewDistributedCache(t *testing.T) {
	port := 7946
	node_name := "node1"
	dc, err := NewDistributedCache(port, 8000, node_name)
	if err != nil {
		t.Fatalf("Failed to create distributed cache: %v", err)
	}

	if dc.Cache == nil {
		t.Error("Cache instance is nil")
	}

	if dc.List == nil {
		t.Error("Memberlist instance is nil")
	}

	if dc.Config == nil {
		t.Error("Config is nil")
	}

	if dc.Config.BindPort != port {
		t.Errorf("Expected BindPort to be %d, got %d", port, dc.Config.BindPort)
	}
	if dc.Config.AdvertisePort != port {
		t.Errorf("Expected AdvertisePort to be %d, got %d", port, dc.Config.AdvertisePort)
	}

	// Check if the memberlist is using the correct port
	if dc.List.LocalNode().Port != uint16(port) {
		t.Errorf("Expected LocalNode port to be %d, got %d", port, dc.List.LocalNode().Port)
	}

	// Try to create another instance with the same port (should fail)
	_, err = NewDistributedCache(port, 8000, node_name)
	if err == nil {
		t.Error("Expected error when creating second instance with same port, but got nil")
	}
}

// func TestJoinCluster(t *testing.T) {
// 	dc1, _ := NewDistributedCache(7947)
// 	dc2, _ := NewDistributedCache(7948)
// 	err := dc2.JoinCluster("127.0.0.1:7947")
// 	if err != nil {
// 		t.Fatalf("Failed to join cluster: %v", err)
// 	}
//
// 	// Allow some time for cluster propagation
// 	time.Sleep(time.Second)
//
// 	if len(dc1.List.Members()) != 2 {
// 		t.Errorf("Expected 2 members in dc1, found %d", len(dc1.List.Members()))
// 	}
// 	if len(dc2.List.Members()) != 2 {
// 		t.Errorf("Expected 2 members in dc2, found %d", len(dc2.List.Members()))
// 	}
// }

// func TestJoinCluster(t *testing.T) {
// 	// Create first node
// 	dc1, err := NewDistributedCache("127.0.0.1", 7947)
// 	if err != nil {
// 		t.Fatalf("Failed to create first distributed cache: %v", err)
// 	}
//
// 	// Create second node
// 	dc2, err := NewDistributedCache("127.0.0.2", 7948)
// 	if err != nil {
// 		t.Fatalf("Failed to create second distributed cache: %v", err)
// 	}
//
// 	// Join the second node to the first
// 	err = dc2.JoinCluster("127.0.0.1:7947")
// 	if err != nil {
// 		t.Fatalf("Failed to join cluster: %v", err)
// 	}
//
// 	// Allow some time for cluster propagation
// 	time.Sleep(2 * time.Second)
//
// 	// Check the number of members in each node
// 	if len(dc1.List.Members()) != 2 {
// 		t.Errorf("Expected 2 members in dc1, found %d", len(dc1.List.Members()))
// 	}
// 	if len(dc2.List.Members()) != 2 {
// 		t.Errorf("Expected 2 members in dc2, found %d", len(dc2.List.Members()))
// 	}
//
// 	// Log the members for debugging
// 	t.Logf("DC1 Members: %v", dc1.List.Members())
// 	t.Logf("DC2 Members: %v", dc2.List.Members())
// }

func TestJoinCluster(t *testing.T) {
	config1 := memberlist.DefaultLocalConfig()
	config1.Name = "node1"
	config1.BindAddr = "127.0.0.1"
	config1.BindPort = 7947
	config1.AdvertiseAddr = "127.0.0.1"
	config1.AdvertisePort = 7947

	dc1, err := NewDistributedCache(config1.BindPort, 8000, config1.Name)
	if err != nil {
		t.Fatalf("Failed to create first distributed cache: %v", err)
	}

	config2 := memberlist.DefaultLocalConfig()
	config2.Name = "node2"
	config2.BindAddr = "127.0.0.1" // Using a different loopback address
	config2.BindPort = 7948
	config2.AdvertiseAddr = "127.0.0.1" // Using a different loopback address
	config2.AdvertisePort = 7948

	dc2, err := NewDistributedCache(config2.BindPort, 8001, config2.Name)
	if err != nil {
		t.Fatalf("Failed to create second distributed cache: %v", err)
	}

	err = dc2.JoinCluster("127.0.0.1:7947")
	if err != nil {
		t.Fatalf("Failed to join cluster: %v", err)
	}

	// Allow some time for cluster propagation
	time.Sleep(2 * time.Second)

	if len(dc1.List.Members()) != 2 {
		t.Errorf("Expected 2 members in dc1, found %d", len(dc1.List.Members()))
	}
	if len(dc2.List.Members()) != 2 {
		t.Errorf("Expected 2 members in dc2, found %d", len(dc2.List.Members()))
	}
}

// func TestConcurrency(t *testing.T) {
// 	dc, _ := NewDistributedCache(7952)
// 	key := "concurrentKey"
// 	iterations := 1000
//
// 	done := make(chan bool)
// 	for i := 0; i < iterations; i++ {
// 		go func(val int) {
// 			req := httptest.NewRequest(http.MethodPut, "/cache/"+key, strings.NewReader("value="+string(val)+"&duration=5000000000"))
// 			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
// 			w := httptest.NewRecorder()
// 			dc.HTTPHandler(w, req)
// 			done <- true
// 		}(i)
// 	}
//
// 	for i := 0; i < iterations; i++ {
// 		<-done
// 	}
//
// 	req := httptest.NewRequest(http.MethodGet, "/cache/"+key, nil)
// 	w := httptest.NewRecorder()
// 	dc.HTTPHandler(w, req)
// 	if w.Code != http.StatusOK {
// 		t.Errorf("Expected status 200 OK after concurrent writes, got %v", w.Code)
// 	}
// }
