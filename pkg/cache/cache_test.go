package cache

import (
	"testing"
	"time"
)

func TestCacheSetGet(t *testing.T) {
	c := NewCache()
	c.Set("key1", "value1", 2*time.Second)

	value, found := c.Get("key1")
	if !found || value != "value1" {
		t.Errorf("Expected to find key1 with value1, got %v, found: %v", value, found)
	}
}

func TestCacheExpiration(t *testing.T) {
	c := NewCache()
	c.Set("key2", "value2", 1*time.Second)

	time.Sleep(2 * time.Second)

	_, found := c.Get("key2")
	if found {
		t.Errorf("Expected key2 to be expired")
	}
}
