package preamble

import (
	"regexp"
	"strings"
	"testing"

	gsllibs "gradientcontrol.com/omega/core/stdlib"
)

// TestEveryPackageHasASummary: the index is only useful if each line says
// what the package holds.
func TestEveryPackageHasASummary(t *testing.T) {
	for _, name := range PackageNames() {
		if PackageSummary(name) == "" {
			t.Errorf("%s has no summary: start its first source with a `///` block and a blank line", name)
		}
	}
}

// TestPackageDocsListEveryDocumentedMember checks the extractor against an
// independent reading of the sources: every column-0 declaration with a `///`
// line directly above it must appear in its package's documentation, and so
// in the full reference. A doc form the extractor misreads fails here.
func TestPackageDocsListEveryDocumentedMember(t *testing.T) {
	decl := regexp.MustCompile(`^(?:@[A-Za-z_]\w*(?:\([^)]*\))?\s+)?([A-Za-z_]\w*)\s*[(=]`)
	full := GSLAuthorSkill()
	for _, lib := range gsllibs.All() {
		doc, err := PackageDescriptiveText(lib.ImportPath)
		if err != nil {
			t.Fatal(err)
		}
		for _, src := range lib.Sources {
			lines := strings.Split(src, "\n")
			for i := 1; i < len(lines); i++ {
				m := decl.FindStringSubmatch(lines[i])
				if m == nil || !isPublicDocLine(strings.TrimSpace(lines[i-1])) {
					continue
				}
				if !strings.Contains(doc, "`"+m[1]) {
					t.Errorf("%s: documented member %s is missing from its docs", lib.ImportPath, m[1])
				}
				if !strings.Contains(full, "`"+m[1]) {
					t.Errorf("%s: documented member %s is missing from the full reference", lib.ImportPath, m[1])
				}
			}
		}
	}
}

// TestSummaryIsNotAMemberDoc: the blank line under a summary detaches it, so
// adding one cannot change the full reference's text for any member.
func TestSummaryIsNotAMemberDoc(t *testing.T) {
	for _, name := range PackageNames() {
		summary := PackageSummary(name)
		doc, _ := PackageDescriptiveText(name)
		if strings.Count(doc, summary) != 1 {
			t.Errorf("%s: the summary appears %d times in its docs, want once (as the summary)", name, strings.Count(doc, summary))
		}
	}
}
