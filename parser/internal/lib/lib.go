package lib

import (
	"container/list"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"sync"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/parser/internal/parser2"
	"gradientcontrol.com/omega/core/program"
	gsllibs "gradientcontrol.com/omega/core/stdlib"
)

func init() {
	for _, libInfo := range gsllibs.All() {
		pkg, err := ParsePackage(libInfo.Sources)
		if err == nil && pkg != nil {
			program.RegisterStdPackage(libInfo.ImportPath, pkg)
		}
	}
}

// ParsePackage parses all of the sources in a package.
//
// Every assembly parses each package it imports, and parsing was the largest
// share of assembling a typical document: about a microsecond per byte, so
// 50–400 ms of library a compile. A package parsed before is cloned from the
// cache instead, which costs about a twenty-fifth of that. The caller owns
// the package it is given and may rewrite it in place.
func ParsePackage(sources []string) (*program.Package, error) {
	key := sourcesKey(sources)
	if pkg := parsed.get(key); pkg != nil {
		return pkg, nil
	}
	res, err := parser2.Parse(sources)
	if err != nil {
		return nil, err
	}
	if len(res.Errors) > 0 {
		if len(res.Errors) == 1 {
			return nil, &res.Errors[0]
		}
		errs := make([]error, len(res.Errors))
		for i := range res.Errors {
			errs[i] = &res.Errors[i]
		}
		return nil, errors.Join(errs...)
	}
	if res.Package != nil {
		parsed.put(key, proto.Clone(res.Package).(*program.Package), sourcesSize(sources))
	}
	return res.Package, nil
}

// parseCacheLimit bounds the source bytes whose parses are kept. The standard
// library is about 1 MB; the rest is room for documents and their packages.
const parseCacheLimit = 32 << 20

// TODO: Remove this hack and the cache entirely
// Instead, add an assembler option to replace the parser
// used by the assembler so that a caller can supply a caching parser.
var parsed = newParseCache(parseCacheLimit)

// parseCache keeps parsed packages by their sources, least recently used
// first out. It hands out clones and never its own copy.
type parseCache struct {
	mu      sync.Mutex
	entries map[[sha256.Size]byte]*list.Element
	order   *list.List
	size    int
	limit   int
}

type parseEntry struct {
	key  [sha256.Size]byte
	pkg  *program.Package
	size int
}

func newParseCache(limit int) *parseCache {
	return &parseCache{
		entries: make(map[[sha256.Size]byte]*list.Element),
		order:   list.New(),
		limit:   limit,
	}
}

func (c *parseCache) get(key [sha256.Size]byte) *program.Package {
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

func (c *parseCache) put(key [sha256.Size]byte, pkg *program.Package, size int) {
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
