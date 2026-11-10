package parser

import (
	"container/list"
	"crypto/sha256"
	"sync"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/internal/stdpkg"
	"gradientcontrol.com/omega/core/parser/internal/parser2"
	"gradientcontrol.com/omega/core/program"
)

// defaultCacheLimit bounds the non-stdlib source bytes whose parses are kept by default (4 MiB).
const defaultCacheLimit = 4 << 20

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

// NewCachingParser creates a new CachingParser backed by an LRU cache for user
// packages and the shared pre-parsed standard library cache in package program.
func NewCachingParser() *CachingParser {
	return &CachingParser{
		limit:   defaultCacheLimit,
		entries: make(map[[sha256.Size]byte]*list.Element),
		order:   list.New(),
	}
}

// ParsePackage parses all of the sources in a package, returning a cloned
// package from the shared standard library cache or LRU cache when available
// and caching error-free parses.
func (c *CachingParser) ParsePackage(sources []string) (*Result, error) {
	key := stdpkg.SourcesHash(sources)
	// Resolve unmodified standard library packages through the shared lazy
	// cache in package program before touching c.mu. This avoids lock
	// contention, keeps the ~1.44 MB standard library out of the LRU eviction
	// budget, and falls through automatically to live parsing if a developer
	// edited a .gsl file in stdlib/ before regenerating package.pb.
	// Invariant: ImportPathByHash only succeeds when importPath has a
	// hash-validated package.pb payload, so GetStdPackage is non-nil.
	if stdpkg.StdLibRegistry != nil {
		if importPath, ok := stdpkg.StdLibRegistry.ImportPathByHash(key); ok {
			return &Result{Package: program.GetStdPackage(importPath)}, nil
		}
	}
	if pkg := c.get(key); pkg != nil {
		return &Result{Package: pkg}, nil
	}
	res, err := parser2.Parse(sources)
	if err == nil {
		c.put(key, proto.Clone(res.Package).(*program.Package), sourcesSize(sources))
	}
	return res, err
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

func sourcesSize(sources []string) int {
	n := 0
	for _, src := range sources {
		n += len(src)
	}
	return n
}
