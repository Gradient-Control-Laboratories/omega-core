package meshasset

import (
	"fmt"
	"testing"
)

func TestMeshCache_BasicAndLRU(t *testing.T) {
	cache := NewMeshCache(10000, 2)

	a1 := &Asset{ContentHash: "hash-1", Bytes: make([]byte, 100)}
	a2 := &Asset{ContentHash: "hash-2", Bytes: make([]byte, 100)}
	a3 := &Asset{ContentHash: "hash-3", Bytes: make([]byte, 100)}

	cache.Put("hash-1", a1)
	cache.Put("hash-2", a2)

	if cache.Len() != 2 {
		t.Fatalf("expected 2 entries, got %d", cache.Len())
	}

	// Access a1 so a2 becomes LRU
	if _, ok := cache.Get("hash-1"); !ok {
		t.Fatal("expected hash-1 hit")
	}

	// Insert a3 -> should evict a2
	cache.Put("hash-3", a3)

	if cache.Len() != 2 {
		t.Fatalf("expected 2 entries after eviction, got %d", cache.Len())
	}

	if _, ok := cache.Get("hash-2"); ok {
		t.Fatal("expected hash-2 to be evicted")
	}
	if _, ok := cache.Get("hash-1"); !ok {
		t.Fatal("expected hash-1 to remain")
	}
	if _, ok := cache.Get("hash-3"); !ok {
		t.Fatal("expected hash-3 to exist")
	}
}

func TestMeshCache_ByteCap(t *testing.T) {
	// Limit to 300 bytes
	cache := NewMeshCache(300, 10)

	a1 := &Asset{ContentHash: "hash-1", Bytes: make([]byte, 150)}
	a2 := &Asset{ContentHash: "hash-2", Bytes: make([]byte, 100)}
	a3 := &Asset{ContentHash: "hash-3", Bytes: make([]byte, 100)}

	cache.Put("hash-1", a1)
	cache.Put("hash-2", a2)
	// Total bytes: 250 <= 300

	if cache.Len() != 2 {
		t.Fatalf("expected 2 entries, got %d", cache.Len())
	}

	// Putting a3 adds 100 -> 350 > 300 -> evicts oldest (a1)
	cache.Put("hash-3", a3)

	if _, ok := cache.Get("hash-1"); ok {
		t.Fatal("expected hash-1 to be evicted by byte limit")
	}
	if cache.UsedBytes() > 300 {
		t.Fatalf("expected <= 300 bytes used, got %d", cache.UsedBytes())
	}
}

func TestMeshCache_Concurrent(t *testing.T) {
	cache := NewMeshCache(1<<20, 50)
	done := make(chan bool)

	for i := 0; i < 8; i++ {
		go func(workerID int) {
			for j := 0; j < 50; j++ {
				h := ContentHash(fmt.Sprintf("hash-%d-%d", workerID, j))
				cache.Put(h, &Asset{ContentHash: h, Bytes: []byte("sample")})
				cache.Get(h)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 8; i++ {
		<-done
	}

	if cache.Len() > 50 {
		t.Fatalf("expected <= 50 entries, got %d", cache.Len())
	}
}
