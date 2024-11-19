package distributed

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ####################################################   Testing Fiber HTTP handlers   ##############################################
func TestFiberHandlerSetAndGet(t *testing.T) {
	dc, _ := NewDistributedCache(7949, 8000, "node1")

	// Create a new Fiber app
	app := fiber.New()

	// Register the handler
	app.Put("/cache/:key", dc.FiberHandler)
	app.Get("/cache/:key", dc.FiberHandler)

	// Testing PUT
	req := httptest.NewRequest(fiber.MethodPut, "/cache/key1", strings.NewReader("value=hello&duration=5000000000"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test PUT request: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200 OK for PUT, got %v", resp.StatusCode)
	}

	// Testing GET
	req = httptest.NewRequest(fiber.MethodGet, "/cache/key1", nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test GET request: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200 OK for GET, got %v", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	if string(body) != "hello" {
		t.Errorf("Expected body 'hello', got %v", string(body))
	}

	// Testing GET for non-existent key
	req = httptest.NewRequest(fiber.MethodGet, "/cache/nonexistent", nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test GET request for non-existent key: %v", err)
	}
	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("Expected status 404 Not Found for non-existent key, got %v", resp.StatusCode)
	}
}

func TestFiberHandlerDelete(t *testing.T) {
	dc, _ := NewDistributedCache(7950, 8000, "node1")

	// Create a new Fiber application
	app := fiber.New()

	// Register the handler
	app.Delete("/cache/:key", dc.FiberHandler)

	// Set a value first
	dc.Cache.Set("key1", "value1", 5*time.Second)

	// Create a test request
	req := httptest.NewRequest(fiber.MethodDelete, "/cache/key1", nil)

	// Test the request
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status 200 OK for DELETE, got %v with body: %s", resp.Status, string(body))
	}

	// Check if the value was deleted
	_, found := dc.Cache.Get("key1")
	if found {
		t.Errorf("Expected key1 to be deleted")
	}
}

func TestFiberHandlerInvalidMethod(t *testing.T) {
	dc, _ := NewDistributedCache(7951, 8000, "node1")

	// Create a new Fiber app
	app := fiber.New()

	// Register valid methods for the path
	app.Get("/cache/:key", dc.FiberHandler)
	app.Put("/cache/:key", dc.FiberHandler)
	app.Delete("/cache/:key", dc.FiberHandler)

	// Test POST method which isn't registered
	req := httptest.NewRequest(fiber.MethodPost, "/cache/key1", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test POST request: %v", err)
	}

	if resp.StatusCode != fiber.StatusMethodNotAllowed {
		t.Errorf("Expected status 405 Method Not Allowed for POST, got %v", resp.StatusCode)
	}
}
