package runner

import (
	"testing"
)

func TestParseGitLsTreeSizesAndCompareGoldenCategory(t *testing.T) {
	sample := []byte(
		"100644 blob 1d2867fe7929ff20bc11974aa95a72533b1db3cf   10372\tagent/fidget/internal/convert/testdata/bike.glsl\n" +
			"100644 blob 68fced12a73afb34e25840f7ece6a37750a28e7d   10211\tagent/fidget/internal/convert/testdata/biplane.glsl\n",
	)
	mainSizes := ParseGitLsTreeSizes(sample)
	if len(mainSizes) != 2 {
		t.Fatalf("len(mainSizes) = %d, want 2", len(mainSizes))
	}

	diskSizes := map[string]int64{
		"agent/fidget/internal/convert/testdata/bike.glsl":    10000, // modified (-372 B)
		"agent/fidget/internal/convert/testdata/biplane.glsl": 10211, // unchanged
		"agent/fidget/internal/convert/testdata/new.glsl":     500,   // added
	}

	cat := CompareGoldenCategory(diskSizes, mainSizes, true)
	if cat.FileCount != 3 {
		t.Errorf("FileCount = %d, want 3", cat.FileCount)
	}
	if cat.ActualBytes != 20711 {
		t.Errorf("ActualBytes = %d, want 20711", cat.ActualBytes)
	}
	if cat.MainBytes != 20583 {
		t.Errorf("MainBytes = %d, want 20583", cat.MainBytes)
	}
	if len(cat.Diffs) != 2 {
		t.Fatalf("len(Diffs) = %d, want 2: %+v", len(cat.Diffs), cat.Diffs)
	}
	if cat.Diffs[0].Status != "modified" || cat.Diffs[1].Status != "added" {
		t.Errorf("unexpected diff statuses: %+v", cat.Diffs)
	}
}
