package skillcmd

import (
	"regexp"
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/assembler"
	"github.com/Gradient-Control-Laboratories/omega-core/backend/wgsl"
	"github.com/Gradient-Control-Laboratories/omega-core/program/preamble"
)

// gslBlockRE matches a ```gsl fenced block, possibly indented under a list
// item. A ```gsl-fragment block (a counter-example, or a snippet that leans on
// names it does not define) does not match: the word must end at the newline.
var gslBlockRE = regexp.MustCompile("(?m)^([ \t]*)```gsl\n((?s:.*?))^[ \t]*```")

// TestDocExamplesBuild: every ```gsl block in the GSL author skill is a whole
// program that assembles and generates WGSL, so an example cannot rot into
// one that teaches a wrong call. Fragments are tagged ```gsl-fragment.
func TestDocExamplesBuild(t *testing.T) {
	blocks := gslBlockRE.FindAllStringSubmatch(preamble.GSLSkillIndex(), -1)
	if len(blocks) == 0 {
		t.Fatal("no ```gsl blocks found in the skill")
	}
	for _, block := range blocks {
		src := dedent(block[2], block[1])
		res, err := assembler.Assemble([]string{src})
		if err != nil {
			t.Errorf("example does not assemble: %v\n%s", err, src)
			continue
		}
		if _, err := wgsl.Render(src, res.Program); err != nil {
			t.Errorf("example does not generate WGSL: %v\n%s", err, src)
		}
	}
}

// dedent removes the fence's own indentation from every line of the block.
func dedent(src, indent string) string {
	lines := strings.Split(src, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimPrefix(line, indent)
	}
	return strings.Join(lines, "\n")
}
