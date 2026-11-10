package stdpkg_test

import (
	"testing"
	"testing/fstest"

	"gradientcontrol.com/omega/core/internal/stdpkg"
	_ "gradientcontrol.com/omega/core/stdlib"
)

func TestSourcesHash_CollisionResistance(t *testing.T) {
	h1 := stdpkg.SourcesHash([]string{"ab", "c"})
	h2 := stdpkg.SourcesHash([]string{"a", "bc"})
	if h1 == h2 {
		t.Fatalf("SourcesHash([\"ab\", \"c\"]) collided with SourcesHash([\"a\", \"bc\"]): %x", h1)
	}

	h1Again := stdpkg.SourcesHash([]string{"ab", "c"})
	if h1 != h1Again {
		t.Fatalf("SourcesHash is non-deterministic: %x != %x", h1, h1Again)
	}
}

func TestLoadSources_PackageYamlOrderAndValidation(t *testing.T) {
	fsys := fstest.MapFS{
		"builtin/package.yaml": &fstest.MapFile{Data: []byte("sources:\n  - base.gsl\n  - 2d.gsl\n")},
		"builtin/2d.gsl":       &fstest.MapFile{Data: []byte("// 2d")},
		"builtin/base.gsl":     &fstest.MapFile{Data: []byte("// base")},
		"builtin/package.pb":   &fstest.MapFile{Data: []byte("ignored-non-gsl")},
	}

	got, err := stdpkg.LoadSources(fsys, "builtin")
	if err != nil {
		t.Fatalf("LoadSources(\"builtin\") unexpected error: %v", err)
	}
	if len(got) != 2 || got[0] != "// base" || got[1] != "// 2d" {
		t.Fatalf("LoadSources(\"builtin\") = %q, want [\"// base\", \"// 2d\"]", got)
	}

	// Unlisted .gsl file alongside package.yaml must be rejected.
	fsysUnlisted := fstest.MapFS{
		"builtin/package.yaml": &fstest.MapFile{Data: []byte("sources:\n  - base.gsl\n")},
		"builtin/base.gsl":     &fstest.MapFile{Data: []byte("// base")},
		"builtin/extra.gsl":    &fstest.MapFile{Data: []byte("// extra")},
	}
	if _, err := stdpkg.LoadSources(fsysUnlisted, "builtin"); err == nil {
		t.Fatal("expected error for unlisted .gsl file alongside package.yaml, got nil")
	}

	// Missing .gsl file referenced by package.yaml must be rejected.
	fsysMissing := fstest.MapFS{
		"builtin/package.yaml": &fstest.MapFile{Data: []byte("sources:\n  - base.gsl\n  - missing.gsl\n")},
		"builtin/base.gsl":     &fstest.MapFile{Data: []byte("// base")},
	}
	if _, err := stdpkg.LoadSources(fsysMissing, "builtin"); err == nil {
		t.Fatal("expected error for missing file in package.yaml, got nil")
	}
}

func TestLoadSources_SingleDirnameGslConvention(t *testing.T) {
	fsys := fstest.MapFS{
		"lattice/beam/beam.gsl":   &fstest.MapFile{Data: []byte("// beam")},
		"lattice/beam/package.pb": &fstest.MapFile{Data: []byte("pb")},
	}

	got, err := stdpkg.LoadSources(fsys, "lattice/beam")
	if err != nil {
		t.Fatalf("LoadSources(\"lattice/beam\") unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != "// beam" {
		t.Fatalf("LoadSources(\"lattice/beam\") = %q, want [\"// beam\"]", got)
	}

	// Wrongly named single .gsl file without package.yaml must be rejected.
	fsysWrongName := fstest.MapFS{
		"lattice/beam/other.gsl": &fstest.MapFile{Data: []byte("// other")},
	}
	if _, err := stdpkg.LoadSources(fsysWrongName, "lattice/beam"); err == nil {
		t.Fatal("expected error when <dirname>.gsl is missing, got nil")
	}

	// Extra .gsl file without package.yaml must be rejected.
	fsysExtra := fstest.MapFS{
		"color/color.gsl":  &fstest.MapFile{Data: []byte("// color")},
		"color/helper.gsl": &fstest.MapFile{Data: []byte("// helper")},
	}
	if _, err := stdpkg.LoadSources(fsysExtra, "color"); err == nil {
		t.Fatal("expected error when extra .gsl file exists without package.yaml, got nil")
	}
}

func TestRegistry_ProtoAndHashValidation(t *testing.T) {
	colorSrc := "red = [1, 0, 0];"
	colorHash := stdpkg.SourcesHash([]string{colorSrc})
	validPayload := []byte("marshaled-color-proto")
	validPB := append(colorHash[:], validPayload...)

	staleSrc := "vec = [1, 2, 3];"
	otherHash := stdpkg.SourcesHash([]string{"old = 1;"})
	stalePB := append(otherHash[:], []byte("stale-proto")...)

	fsys := fstest.MapFS{
		"color/color.gsl":   &fstest.MapFile{Data: []byte(colorSrc)},
		"color/package.pb":  &fstest.MapFile{Data: validPB},
		"vector/vector.gsl": &fstest.MapFile{Data: []byte(staleSrc)},
		"vector/package.pb": &fstest.MapFile{Data: stalePB},
		"map/map.gsl":       &fstest.MapFile{Data: []byte("m = 1;")},
		"ease/ease.gsl":     &fstest.MapFile{Data: []byte("e = 1;")},
		"ease/package.pb":   &fstest.MapFile{Data: []byte("short")},
	}

	reg := stdpkg.NewRegistry(fsys, []stdpkg.Entry{
		{ImportPath: "color"},
		{ImportPath: "vector"},
		{ImportPath: "map"},
		{ImportPath: "ease"},
	})

	// 1. Valid package.pb matches SourcesHash and returns sliced payload + hash lookup.
	payload, ok := reg.Proto("color")
	if !ok || string(payload) != string(validPayload) {
		t.Fatalf("Proto(\"color\") = (%q, %v), want (%q, true)", payload, ok, validPayload)
	}
	raw, ok := reg.RawProto("color")
	if !ok || len(raw) != len(validPB) {
		t.Fatalf("RawProto(\"color\") len = %d, ok = %v, want %d, true", len(raw), ok, len(validPB))
	}
	if imp, ok := reg.ImportPathByHash(colorHash); !ok || imp != "color" {
		t.Fatalf("ImportPathByHash(colorHash) = (%q, %v), want (\"color\", true)", imp, ok)
	}

	// 2. Stale package.pb must be rejected by Proto and ImportPathByHash, while RawProto still returns raw bytes.
	if _, ok := reg.Proto("vector"); ok {
		t.Fatal("Proto(\"vector\") succeeded on stale package.pb, want false")
	}
	if _, ok := reg.ImportPathByHash(stdpkg.SourcesHash([]string{staleSrc})); ok {
		t.Fatal("ImportPathByHash succeeded for stale vector package.pb, want false")
	}
	if _, ok := reg.ImportPathByHash(otherHash); ok {
		t.Fatal("ImportPathByHash matched stale header hash, want false")
	}
	if rawVec, ok := reg.RawProto("vector"); !ok || len(rawVec) != len(stalePB) {
		t.Fatalf("RawProto(\"vector\") = (%d bytes, %v), want (%d bytes, true)", len(rawVec), ok, len(stalePB))
	}

	// 3. Missing package.pb (bootstrap case) still serves Sources cleanly.
	srcs, ok := reg.Sources("map")
	if !ok || len(srcs) != 1 || srcs[0] != "m = 1;" {
		t.Fatalf("Sources(\"map\") = (%q, %v), want ([\"m = 1;\"], true)", srcs, ok)
	}
	if _, ok := reg.Proto("map"); ok {
		t.Fatal("Proto(\"map\") returned true when package.pb is missing")
	}
	if _, ok := reg.RawProto("map"); ok {
		t.Fatal("RawProto(\"map\") returned true when package.pb is missing")
	}

	// 4. Truncated (< 32 bytes) package.pb is rejected by Proto.
	if _, ok := reg.Proto("ease"); ok {
		t.Fatal("Proto(\"ease\") returned true on truncated <32-byte package.pb")
	}
}

func TestDefaultRegisteredStdlib(t *testing.T) {
	all := stdpkg.All()
	if len(all) == 0 {
		t.Fatal("stdpkg.All() returned 0 packages")
	}
	srcs, ok := stdpkg.Sources("builtin")
	if !ok || len(srcs) == 0 {
		t.Fatalf("stdpkg.Sources(\"builtin\") = (%d files, %v), want (>0, true)", len(srcs), ok)
	}
	if _, ok := stdpkg.Sources("nonexistent"); ok {
		t.Fatal("stdpkg.Sources(\"nonexistent\") returned true")
	}
	_, _ = stdpkg.Proto("builtin")
	_, _ = stdpkg.RawProto("builtin")
	_, _ = stdpkg.ImportPathByHash([stdpkg.HashSize]byte{})
}
