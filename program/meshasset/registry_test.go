package meshasset

import "testing"

func sampleMesh() *TriMesh {
	tri := Triangle{
		V0: [3]float64{0, 0, 0},
		V1: [3]float64{1, 0, 0},
		V2: [3]float64{0, 1, 0},
	}
	return &TriMesh{
		Triangles: []Triangle{tri},
		Bbox:      Bbox{Min: [3]float64{-0.5, -0.5, -0.5}, Max: [3]float64{0.5, 0.5, 0.5}},
	}
}

func TestRegistryPutGetRemove(t *testing.T) {
	r := NewRegistry()
	mesh := sampleMesh()
	a, hit := r.Put(0, "cube.stl", mesh, "h1", []byte("stl"))
	if hit {
		t.Fatalf("first put should be a miss")
	}
	if got, ok := r.Get(0); !ok || got != a {
		t.Fatalf("Get(0) returned %v/%v, want %v/true", got, ok, a)
	}

	r.Remove(0)
	if _, ok := r.Get(0); ok {
		t.Fatalf("expected index 0 to be removed")
	}
}

func TestRegistryContentHashReuse(t *testing.T) {
	r := NewRegistry()
	mesh := sampleMesh()

	a0, hit := r.Put(0, "cube.stl", mesh, "sharedhash", []byte("stl"))
	if hit {
		t.Fatalf("first put should be a miss")
	}
	a1, hit := r.Put(1, "cube_copy.stl", mesh, "sharedhash", []byte("stl"))
	if !hit {
		t.Fatalf("second put with same hash should be a hit")
	}
	if a0 != a1 {
		t.Fatalf("same-hash indices should share the same Asset pointer")
	}

	// Removing one index leaves the other intact.
	r.Remove(0)
	if got, ok := r.Get(1); !ok || got != a0 {
		t.Fatalf("index 1 should still resolve to the shared asset after removing 0")
	}
}

func TestRegistryRequireMissing(t *testing.T) {
	r := NewRegistry()
	if _, err := r.Require(42); err == nil {
		t.Fatalf("expected error for missing index")
	}
}

func TestRegistryFindByFilenameCaseInsensitive(t *testing.T) {
	r := NewRegistry()
	mesh := sampleMesh()
	r.PutStep(-1, "part.stl", "part.hodge", "part.step", mesh, "h2", []byte("stl"), []byte("hodge"), []byte("step"))

	testCases := []string{"part.STEP", "PART.sTeP", "PART.STEP", "PART.HODGE", "PART.STL", "part.stp", "PART.sTp", "PART.STP"}
	for _, tc := range testCases {
		if got, ok := r.FindByFilename(tc); !ok || got == nil {
			t.Fatalf("FindByFilename(%q) failed to locate asset registered with \"part.step\"", tc)
		}
	}
}

func TestRegistryRetainsTransientAssets(t *testing.T) {
	r := NewRegistry()
	mesh := sampleMesh()
	r.PutStep(-1, "a.stl", "a.hodge", "a.step", mesh, "a", nil, nil, nil)
	r.PutStep(-1, "b.stl", "b.hodge", "b.step", mesh, "b", nil, nil, nil)

	for _, name := range []string{"a.step", "b.step"} {
		if _, ok := r.FindByFilename(name); !ok {
			t.Fatalf("missing transient asset %q", name)
		}
	}
}

func TestRegistryFindBySourceDisambiguatesFilename(t *testing.T) {
	r := NewRegistry()
	mesh := sampleMesh()
	a, _ := r.PutStep(-1, "part.stl", "part.hodge", "part.step", mesh, "a", nil, nil, nil)
	b, _ := r.PutStep(-1, "part.stl", "part.hodge", "part.step", mesh, "b", nil, nil, nil)
	r.BindSource("parts/A", a, "part.stl", "part.hodge", "part.step")
	r.BindSource("parts/B", b, "part.stl", "part.hodge", "part.step")

	if got, ok := r.FindBySource("parts/A", "part.stp"); !ok || got != a {
		t.Fatalf("parts/A resolved %v/%v, want first asset", got, ok)
	}
	if got, ok := r.FindBySource("parts/B", "part.step"); !ok || got != b {
		t.Fatalf("parts/B resolved %v/%v, want second asset", got, ok)
	}
}

func TestRegistryClaimUploadReplacesStoredSource(t *testing.T) {
	r := NewRegistry()
	mesh := sampleMesh()
	stored, _ := r.PutStep(-1, "part.stl", "part.hodge", "part.step", mesh, "old", nil, nil, nil)
	r.BindSource("parts/A", stored, "part.step")
	uploaded, _ := r.PutStep(0, "part.stl", "part.hodge", "part.step", mesh, "new", nil, nil, nil)
	r.BindUpload(uploaded, "part.stl", "part.hodge", "part.step")

	if got, ok := r.ClaimUpload("parts/A", "part.stp"); !ok || got != uploaded {
		t.Fatalf("claimed %v/%v, want replacement upload", got, ok)
	}
	if got, ok := r.FindBySource("parts/A", "part.step"); !ok || got != uploaded {
		t.Fatalf("stored source is %v/%v, want replacement upload", got, ok)
	}
	if _, ok := r.ClaimUpload("parts/B", "part.step"); ok {
		t.Fatalf("claimed upload more than once")
	}
}

func TestRegistryRemoveAsset(t *testing.T) {
	r := NewRegistry()
	placeholder := r.PutError(-1, "bracket.hodge", "missing file")
	r.BindSource("parts/A", placeholder, "bracket.hodge", "bracket.step")
	r.BindUpload(placeholder, "bracket.hodge")

	if got, ok := r.FindByFilename("bracket.hodge"); !ok || got != placeholder {
		t.Fatalf("placeholder not found by filename before remove")
	}
	if got, ok := r.FindBySource("parts/A", "bracket.step"); !ok || got != placeholder {
		t.Fatalf("placeholder not found by source before remove")
	}

	r.RemoveAsset(placeholder)

	if _, ok := r.FindByFilename("bracket.hodge"); ok {
		t.Fatalf("placeholder still found by filename after RemoveAsset")
	}
	if _, ok := r.FindBySource("parts/A", "bracket.step"); ok {
		t.Fatalf("placeholder still found by source after RemoveAsset")
	}

	mesh := sampleMesh()
	replacement, _ := r.PutStep(-1, "bracket.stl", "bracket.hodge", "bracket.step", mesh, "hash1", nil, nil, nil)
	r.BindSource("parts/A", replacement, "bracket.hodge", "bracket.step")

	if got, ok := r.FindBySource("parts/A", "bracket.step"); !ok || got != replacement {
		t.Fatalf("replacement asset not found by source")
	}
}
