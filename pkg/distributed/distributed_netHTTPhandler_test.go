package distributed

//
// import (
// 	"net/http"
// 	"net/http/httptest"
// 	"strings"
// 	"testing"
// 	"time"
// )
//
// // ####################################################   Testing net/http HTTP handlers   ##############################################
// func TestHTTPHandlerSetAndGet(t *testing.T) {
// 	dc, _ := NewDistributedCache(7949, "node1")
//
// 	// Testing HTTP PUT
// 	req := httptest.NewRequest(http.MethodPut, "/cache/key1", strings.NewReader("value=hello&duration=5000000000"))
// 	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
// 	w := httptest.NewRecorder()
// 	dc.HTTPHandler(w, req)
// 	resp := w.Result()
// 	if resp.StatusCode != http.StatusOK {
// 		t.Errorf("Expected status 200 OK for PUT, got %v", resp.Status)
// 	}
//
// 	// Testing HTTP GET
// 	req = httptest.NewRequest(http.MethodGet, "/cache/key1", nil)
// 	w = httptest.NewRecorder()
// 	dc.HTTPHandler(w, req)
// 	resp = w.Result()
// 	if resp.StatusCode != http.StatusOK {
// 		t.Errorf("Expected status 200 OK for GET, got %v", resp.Status)
// 	}
// 	body := w.Body.String()
// 	if body != "hello" {
// 		t.Errorf("Expected body 'hello', got %v", body)
// 	}
//
// 	// Testing GET for non-existent key
// 	req = httptest.NewRequest(http.MethodGet, "/cache/nonexistent", nil)
// 	w = httptest.NewRecorder()
// 	dc.HTTPHandler(w, req)
// 	resp = w.Result()
// 	if resp.StatusCode != http.StatusNotFound {
// 		t.Errorf("Expected status 404 Not Found for non-existent key, got %v", resp.Status)
// 	}
// }
//
// func TestHTTPHandlerDelete(t *testing.T) {
// 	dc, _ := NewDistributedCache(7950, "node1")
//
// 	// Set a value first
// 	dc.Cache.Set("key1", "value1", 5*time.Second)
//
// 	// Test HTTP DELETE
// 	req := httptest.NewRequest(http.MethodDelete, "/cache/key1", nil)
// 	w := httptest.NewRecorder()
// 	dc.HTTPHandler(w, req)
// 	resp := w.Result()
// 	if resp.StatusCode != http.StatusOK {
// 		t.Errorf("Expected status 200 OK for DELETE, got %v", resp.Status)
// 	}
//
// 	// Check if the value was deleted
// 	_, found := dc.Cache.Get("key1")
// 	if found {
// 		t.Errorf("Expected key1 to be deleted")
// 	}
// }
//
// func TestHTTPHandlerInvalidMethod(t *testing.T) {
// 	dc, _ := NewDistributedCache(7951, "node1")
//
// 	req := httptest.NewRequest(http.MethodPost, "/cache/key1", nil)
// 	w := httptest.NewRecorder()
// 	dc.HTTPHandler(w, req)
// 	resp := w.Result()
// 	if resp.StatusCode != http.StatusMethodNotAllowed {
// 		t.Errorf("Expected status 405 Method Not Allowed for POST, got %v", resp.Status)
// 	}
// }
