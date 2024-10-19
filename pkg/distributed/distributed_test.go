package distributed

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/memberlist"
)

func TestNewDistributedCache(t *testing.T) {
	port := 7946
	node_name := "node1"
	dc, err := NewDistributedCache(port, node_name)
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
	_, err = NewDistributedCache(port, node_name)
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

	dc1, err := NewDistributedCache(config1.BindPort, config1.Name)
	if err != nil {
		t.Fatalf("Failed to create first distributed cache: %v", err)
	}

	config2 := memberlist.DefaultLocalConfig()
	config2.Name = "node2"
	config2.BindAddr = "127.0.0.1" // Using a different loopback address
	config2.BindPort = 7948
	config2.AdvertiseAddr = "127.0.0.1" // Using a different loopback address
	config2.AdvertisePort = 7948

	dc2, err := NewDistributedCache(config2.BindPort, config2.Name)
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

func TestHTTPHandlerSetAndGet(t *testing.T) {
	dc, _ := NewDistributedCache(7949, "node1")

	// Testing HTTP PUT
	req := httptest.NewRequest(http.MethodPut, "/cache/key1", strings.NewReader("value=hello&duration=5000000000"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	dc.HTTPHandler(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 OK for PUT, got %v", resp.Status)
	}

	// Testing HTTP GET
	req = httptest.NewRequest(http.MethodGet, "/cache/key1", nil)
	w = httptest.NewRecorder()
	dc.HTTPHandler(w, req)
	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 OK for GET, got %v", resp.Status)
	}
	body := w.Body.String()
	if body != "hello" {
		t.Errorf("Expected body 'hello', got %v", body)
	}

	// Testing GET for non-existent key
	req = httptest.NewRequest(http.MethodGet, "/cache/nonexistent", nil)
	w = httptest.NewRecorder()
	dc.HTTPHandler(w, req)
	resp = w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404 Not Found for non-existent key, got %v", resp.Status)
	}
}

func TestHTTPHandlerDelete(t *testing.T) {
	dc, _ := NewDistributedCache(7950, "node1")

	// Set a value first
	dc.Cache.Set("key1", "value1", 5*time.Second)

	// Test HTTP DELETE
	req := httptest.NewRequest(http.MethodDelete, "/cache/key1", nil)
	w := httptest.NewRecorder()
	dc.HTTPHandler(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 OK for DELETE, got %v", resp.Status)
	}

	// Check if the value was deleted
	_, found := dc.Cache.Get("key1")
	if found {
		t.Errorf("Expected key1 to be deleted")
	}
}

func TestHTTPHandlerInvalidMethod(t *testing.T) {
	dc, _ := NewDistributedCache(7951, "node1")

	req := httptest.NewRequest(http.MethodPost, "/cache/key1", nil)
	w := httptest.NewRecorder()
	dc.HTTPHandler(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405 Method Not Allowed for POST, got %v", resp.Status)
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
