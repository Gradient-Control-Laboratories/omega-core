package meshasset

import (
	"container/list"
	"sync"
)

// MeshCache is a thread-safe, bounded, content-addressed LRU cache for *Asset objects.
type MeshCache struct {
	// Read-only immutable fields set at construction
	maxBytes   int64
	maxEntries int

	// Lock for accessing the fields below
	mu        sync.Mutex
	usedBytes int64
	entries   map[ContentHash]*list.Element
	lruList   *list.List
}

type cacheEntry struct {
	hash      ContentHash
	asset     *Asset
	sizeBytes int64
}

// NewMeshCache constructs a bounded LRU cache with the specified limits.
func NewMeshCache(maxBytes int64, maxEntries int) *MeshCache {
	if maxBytes <= 0 {
		maxBytes = 1 << 30 // 1 GiB default
	}
	if maxEntries <= 0 {
		maxEntries = 128 // 128 default
	}
	return &MeshCache{
		maxBytes:   maxBytes,
		maxEntries: maxEntries,
		entries:    make(map[ContentHash]*list.Element),
		lruList:    list.New(),
	}
}

// Get looks up an asset by its content hash, updating LRU recency on hit.
func (c *MeshCache) Get(hash ContentHash) (*Asset, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	el, ok := c.entries[hash]
	if !ok {
		return nil, false
	}
	c.lruList.MoveToFront(el)
	return el.Value.(*cacheEntry).asset, true
}

// Put inserts or updates an asset in the cache, evicting older entries as needed.
func (c *MeshCache) Put(hash ContentHash, asset *Asset) {
	if asset == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	size := int64(len(asset.Bytes) + len(asset.StepBytes) + len(asset.StepSourceBytes))
	if asset.Mesh != nil {
		size += int64(len(asset.Mesh.Triangles) * 72)
	}
	if size == 0 {
		size = 1024
	}

	if el, ok := c.entries[hash]; ok {
		entry := el.Value.(*cacheEntry)
		c.usedBytes -= entry.sizeBytes
		entry.asset = asset
		entry.sizeBytes = size
		c.usedBytes += size
		c.lruList.MoveToFront(el)
		c.evictLocked()
		return
	}

	entry := &cacheEntry{
		hash:      hash,
		asset:     asset,
		sizeBytes: size,
	}
	el := c.lruList.PushFront(entry)
	c.entries[hash] = el
	c.usedBytes += size
	c.evictLocked()
}

func (c *MeshCache) evictLocked() {
	for (len(c.entries) > c.maxEntries || c.usedBytes > c.maxBytes) && c.lruList.Len() > 0 {
		oldest := c.lruList.Back()
		if oldest == nil {
			break
		}
		entry := oldest.Value.(*cacheEntry)
		c.lruList.Remove(oldest)
		delete(c.entries, entry.hash)
		c.usedBytes -= entry.sizeBytes
	}
}

// Len returns the current number of cached entries.
func (c *MeshCache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.entries)
}

// UsedBytes returns the current memory usage of cached entries.
func (c *MeshCache) UsedBytes() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.usedBytes
}
