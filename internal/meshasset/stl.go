package meshasset

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ContentHash is a hex-encoded SHA-256 of raw asset bytes, used as the cache
// key that lets re-imports of identical content skip the bake step.
type ContentHash string

// HashBytes returns the ContentHash for the given file bytes.
func HashBytes(b []byte) ContentHash {
	sum := sha256.Sum256(b)
	return ContentHash(hex.EncodeToString(sum[:]))
}

// ParseSTL auto-detects binary vs ASCII STL and dispatches. Both formats
// return the same TriMesh + content-hash pair. Non-watertight input is fine —
// the downstream bake uses generalized winding number which handles triangle
// soup, but the STL itself must be well-formed as a file.
func ParseSTL(data []byte) (*TriMesh, ContentHash, error) {
	if looksLikeBinarySTL(data) {
		return ParseBinarySTL(data)
	}
	return ParseASCIISTL(data)
}

// MarshalBinarySTL writes mesh as binary STL bytes.
func MarshalBinarySTL(mesh *TriMesh) ([]byte, error) {
	if mesh == nil || len(mesh.Triangles) == 0 {
		return nil, errors.New("mesh has no triangles")
	}
	var buf bytes.Buffer
	buf.Grow(84 + len(mesh.Triangles)*50)
	buf.Write(make([]byte, 80))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(len(mesh.Triangles)))
	for _, t := range mesh.Triangles {
		for i := 0; i < 3; i++ {
			_ = binary.Write(&buf, binary.LittleEndian, float32(0))
		}
		for _, v := range [...][3]float64{t.V0, t.V1, t.V2} {
			for i := 0; i < 3; i++ {
				_ = binary.Write(&buf, binary.LittleEndian, float32(v[i]))
			}
		}
		_ = binary.Write(&buf, binary.LittleEndian, uint16(0))
	}
	return buf.Bytes(), nil
}

// looksLikeBinarySTL reports whether data matches a binary STL's exact
// expected size. The file-size check is the canonical way to disambiguate
// because ASCII STL files can also start with the literal word "solid" (and
// some legit binary STLs do too, so a prefix sniff is unreliable).
func looksLikeBinarySTL(data []byte) bool {
	if len(data) < 84 {
		return false
	}
	tryCount := binary.LittleEndian.Uint32(data[80:84])
	expected := 84 + int(tryCount)*50
	return len(data) == expected
}

// ParseBinarySTL parses a binary STL file into a TriMesh plus its content
// hash.
//
// Binary STL layout:
//
//	  80 bytes : arbitrary header (ignored)
//	   4 bytes : uint32 little-endian triangle count
//	per tri:
//	  12 bytes : normal   (float32 x 3) - ignored; we recompute if needed
//	  36 bytes : vertices (float32 x 9)
//	   2 bytes : attribute byte count    (ignored)
func ParseBinarySTL(data []byte) (*TriMesh, ContentHash, error) {
	if len(data) < 84 {
		return nil, "", errors.New("binary STL too short (need at least 84 bytes)")
	}
	tryCount := binary.LittleEndian.Uint32(data[80:84])
	expected := 84 + int(tryCount)*50
	if len(data) < expected {
		return nil, "", fmt.Errorf("binary STL truncated: header says %d triangles (needs %d bytes), got %d",
			tryCount, expected, len(data))
	}

	mesh := &TriMesh{
		Triangles: make([]Triangle, 0, tryCount),
		Bbox:      emptyBbox(),
	}

	off := 84
	for i := uint32(0); i < tryCount; i++ {
		off += 12 // skip normal
		var tri Triangle
		readVec := func() [3]float64 {
			x := float64(math.Float32frombits(binary.LittleEndian.Uint32(data[off+0 : off+4])))
			y := float64(math.Float32frombits(binary.LittleEndian.Uint32(data[off+4 : off+8])))
			z := float64(math.Float32frombits(binary.LittleEndian.Uint32(data[off+8 : off+12])))
			off += 12
			return [3]float64{x, y, z}
		}
		tri.V0 = readVec()
		tri.V1 = readVec()
		tri.V2 = readVec()
		off += 2 // attribute byte count
		mesh.Triangles = append(mesh.Triangles, tri)
		accumulateBbox(&mesh.Bbox, tri)
	}

	if len(mesh.Triangles) == 0 {
		return nil, "", errors.New("STL contains no triangles")
	}

	return mesh, HashBytes(data), nil
}

// ParseASCIISTL parses an ASCII STL file. The format is:
//
//	solid <optional name>
//	  facet normal <nx> <ny> <nz>
//	    outer loop
//	      vertex <x> <y> <z>
//	      vertex <x> <y> <z>
//	      vertex <x> <y> <z>
//	    endloop
//	  endfacet
//	  ...
//	endsolid <optional name>
//
// We only care about the vertex lines; normals are recomputed at bake time
// if needed. Whitespace and line endings are flexible; comments aren't part
// of the ASCII STL spec so we don't handle them.
func ParseASCIISTL(data []byte) (*TriMesh, ContentHash, error) {
	mesh := &TriMesh{
		Bbox: emptyBbox(),
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	// Allow very large single-line facets if anyone ever writes one line per
	// triangle. Default bufio.Scanner line limit is 64KB; go up to 1MB.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var (
		current      Triangle
		vertCount    int
		lineNum      int
		inFacet      bool
		inLoop       bool
		facetStarted bool // at least one facet seen
	)

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Case-insensitive keyword dispatch.
		lower := strings.ToLower(line)
		switch {
		case strings.HasPrefix(lower, "solid"), strings.HasPrefix(lower, "endsolid"):
			// ignore — we don't use the solid name
		case strings.HasPrefix(lower, "facet"):
			if inFacet {
				return nil, "", fmt.Errorf("ASCII STL line %d: nested facet", lineNum)
			}
			inFacet = true
			facetStarted = true
			vertCount = 0
			current = Triangle{}
		case strings.HasPrefix(lower, "outer loop"):
			if !inFacet {
				return nil, "", fmt.Errorf("ASCII STL line %d: 'outer loop' without 'facet'", lineNum)
			}
			inLoop = true
		case strings.HasPrefix(lower, "vertex"):
			if !inLoop {
				return nil, "", fmt.Errorf("ASCII STL line %d: 'vertex' outside 'outer loop'", lineNum)
			}
			v, err := parseASCIIVertex(line)
			if err != nil {
				return nil, "", fmt.Errorf("ASCII STL line %d: %w", lineNum, err)
			}
			switch vertCount {
			case 0:
				current.V0 = v
			case 1:
				current.V1 = v
			case 2:
				current.V2 = v
			default:
				return nil, "", fmt.Errorf("ASCII STL line %d: facet has more than 3 vertices", lineNum)
			}
			vertCount++
		case strings.HasPrefix(lower, "endloop"):
			if !inLoop {
				return nil, "", fmt.Errorf("ASCII STL line %d: 'endloop' without 'outer loop'", lineNum)
			}
			if vertCount != 3 {
				return nil, "", fmt.Errorf("ASCII STL line %d: facet has %d vertices, need exactly 3", lineNum, vertCount)
			}
			inLoop = false
		case strings.HasPrefix(lower, "endfacet"):
			if !inFacet {
				return nil, "", fmt.Errorf("ASCII STL line %d: 'endfacet' without 'facet'", lineNum)
			}
			mesh.Triangles = append(mesh.Triangles, current)
			accumulateBbox(&mesh.Bbox, current)
			inFacet = false
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, "", fmt.Errorf("ASCII STL scan failed: %w", err)
	}
	if !facetStarted {
		return nil, "", errors.New("ASCII STL contains no facets")
	}
	if len(mesh.Triangles) == 0 {
		return nil, "", errors.New("STL contains no triangles")
	}

	return mesh, HashBytes(data), nil
}

func parseASCIIVertex(line string) ([3]float64, error) {
	// Format: `vertex <x> <y> <z>`. Accept any whitespace between tokens.
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return [3]float64{}, fmt.Errorf("malformed vertex line %q", line)
	}
	var v [3]float64
	for i := 0; i < 3; i++ {
		f, err := strconv.ParseFloat(fields[i+1], 64)
		if err != nil {
			return [3]float64{}, fmt.Errorf("bad vertex coordinate %q: %w", fields[i+1], err)
		}
		v[i] = f
	}
	return v, nil
}

func emptyBbox() Bbox {
	return Bbox{
		Min: [3]float64{math.Inf(+1), math.Inf(+1), math.Inf(+1)},
		Max: [3]float64{math.Inf(-1), math.Inf(-1), math.Inf(-1)},
	}
}

func accumulateBbox(b *Bbox, tri Triangle) {
	for _, v := range [...][3]float64{tri.V0, tri.V1, tri.V2} {
		for a := 0; a < 3; a++ {
			if v[a] < b.Min[a] {
				b.Min[a] = v[a]
			}
			if v[a] > b.Max[a] {
				b.Max[a] = v[a]
			}
		}
	}
}
