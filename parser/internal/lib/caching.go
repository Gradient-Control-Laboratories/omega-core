package lib

import (
	"container/list"
	"crypto/sha256"
	"encoding/binary"
	"sync"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/program"
)

type stdCachedPkg struct {
	key  [sha256.Size]byte
	pkg  *program.Package
	size int
}

var stdCachedPkgs []stdCachedPkg

func registerStdCachedPkg(sources []string, pkg *program.Package) {
	stdCachedPkgs = append(stdCachedPkgs, stdCachedPkg{
		key:  sourcesKey(sources),
		pkg:  proto.Clone(pkg).(*program.Package),
		size: sourcesSize(sources),
	})
}

// defaultCacheLimit bounds the source bytes whose parses are kept by default (32 MiB).
// The standard library is about 1 MB; the rest is room for documents and their packages.
const defaultCacheLimit = 32 << 20

// CachingParser wraps ParsePackage with an in-memory LRU cache keyed by source
// content hash. It hands out clones of cached packages and never its own copy,
// so callers may rewrite the returned *program.Package in place.
type CachingParser struct {
	// Read-only immutable fields set at construction
	limit int

	// Lock for accessing the fields below
	mu      sync.Mutex
	entries map[[sha256.Size]byte]*list.Element
	order   *list.List
	size    int
}

type parseEntry struct {
	key  [sha256.Size]byte
	pkg  *program.Package
	size int
}

// NewCachingParser creates a new CachingParser pre-populated with standard library packages.
func NewCachingParser() *CachingParser {
	c := &CachingParser{
		limit:   defaultCacheLimit,
		entries: make(map[[sha256.Size]byte]*list.Element),
		order:   list.New(),
	}
	for _, sp := range stdCachedPkgs {
		c.put(sp.key, sp.pkg, sp.size)
	}
	return c
}

// ParsePackage parses all of the sources in a package, returning a cloned
// package from the cache when available and caching error-free parses.
func (c *CachingParser) ParsePackage(sources []string) *Result {
	key := sourcesKey(sources)
	if pkg := c.get(key); pkg != nil {
		return &Result{Package: pkg}
	}
	res := ParsePackage(sources)
	if len(res.Errors) > 0 {
		return res
	}
	c.put(key, proto.Clone(res.Package).(*program.Package), sourcesSize(sources))
	return res
}

func (c *CachingParser) get(key [sha256.Size]byte) *program.Package {
	c.mu.Lock()
	el, ok := c.entries[key]
	if ok {
		c.order.MoveToFront(el)
	}
	c.mu.Unlock()
	if !ok {
		return nil
	}
	// The cached package is never written after it is stored, so cloning it
	// outside the lock is safe.
	return proto.Clone(el.Value.(*parseEntry).pkg).(*program.Package)
}

func (c *CachingParser) put(key [sha256.Size]byte, pkg *program.Package, size int) {
	if size > c.limit {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.entries[key]; ok {
		c.order.MoveToFront(el)
		return
	}
	c.entries[key] = c.order.PushFront(&parseEntry{key: key, pkg: pkg, size: size})
	c.size += size
	for c.size > c.limit {
		oldest := c.order.Back()
		e := oldest.Value.(*parseEntry)
		c.order.Remove(oldest)
		delete(c.entries, e.key)
		c.size -= e.size
	}
}

// sourcesKey identifies a package by its sources, each prefixed with its
// length so that no two different lists of files share a key.
func sourcesKey(sources []string) [sha256.Size]byte {
	h := sha256.New()
	var n [8]byte
	for _, src := range sources {
		binary.LittleEndian.PutUint64(n[:], uint64(len(src)))
		h.Write(n[:])
		h.Write([]byte(src))
	}
	var key [sha256.Size]byte
	copy(key[:], h.Sum(nil))
	return key
}

func sourcesSize(sources []string) int {
	n := 0
	for _, src := range sources {
		n += len(src)
	}
	return n
}
