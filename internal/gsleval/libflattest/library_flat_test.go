// Package libflattest checks the stand-ins the texture libraries declare: the
// colour or material a texture draws as when its part is too small on screen
// for it (name.simple). It is a package of its own so its dense sampling runs in
// its own test binary, under its own timeout, apart from gsleval's.
package libflattest

import (
	"context"
	"io/fs"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/internal/gsleval"
	"gradientcontrol.com/omega/core/internal/source"
)

// packageRoot is the GSL package tree.
const packageRoot = "../../../stdlib"

// tolerance is how far a declared colour constant may be from the texture's
// mean, in Oklab: about one just noticeable difference. materialTolerance is
// how far each of a material constant's twelve numbers may be from theirs.
const (
	tolerance         = 0.02
	materialTolerance = 0.02
)

// samples is how many points each texture's mean is taken over, in a cube
// half as wide as reach, far larger than any of the textures' features.
const (
	samples = 8000
	reach   = 400.0
)

// flatCase is a texture whose declared constants are checked: a declaration
// that binds c to its colour and, where it has one, m to its material, with
// parameters as the examples use them where they do.
type flatCase struct {
	decl     string
	material bool
	// faded checks the stand-in against the texture as it looks once its
	// detail has faded below a pixel (fadedPackages), not as it looks close
	// up: a finish whose fine relief becomes a streaked lobe from afar stands
	// in for that far look. Its stand-in may then vary over the part, as a
	// lay's direction does, so it is averaged over the same points.
	faded bool
}

var flatCases = []flatCase{
	{decl: "c = playful.checkerboard(10, color.black, color.white);"},
	{decl: "c = playful.checkerboard(3, [0.8, 0.2, 0.1], [0.1, 0.3, 0.9]);"},
	{decl: "c = playful.polkaDot(10, 0.3, [0, 0, 1], color.white, [0.8, 0.1, 0.1]);"},
	{decl: "c = playful.cellSkin(5, 0.5, color.teal, color.white);"},
	{decl: "c = playful.cellSkin(10, 1, [0.2, 0.1, 0.05], [0.9, 0.8, 0.6]);"},
	{decl: "m, c = playful.glitter(material.polishedMetal, color.black, 1.0, color.white, 0.25);", material: true},
	{decl: "m, c = playful.glitter(material.matte, [0.1, 0.2, 0.5], 2.0, [1.0, 0.9, 0.6], 0.15);", material: true},
	{decl: "c = stone.granite(4, color.gray, color.black, color.white);"},
	{decl: "c = stone.granite(4, [0.55, 0.53, 0.52], [0.10, 0.10, 0.11], [0.85, 0.83, 0.80]);"},
	{decl: "c = stone.granite(7, [0.20, 0.19, 0.19], [0.04, 0.04, 0.05], [0.66, 0.67, 0.70]);"},
	{decl: "c = stone.marble(25, 0.4, 0.016, 0.4, color.white, color.gray);"},
	{decl: "c = stone.calacatta([1, 0.2, 0.1]);"},
	{decl: "c = stone.statuario([1, 0.2, 0.1]);"},
	{decl: "c = stone.carrara([1, 0.2, 0.1]);"},
	{decl: "c = stone.flagstone(20, [0.60, 0.55, 0.50], [0.20, 0.20, 0.20]);"},
	{decl: "m = stone.graniteFinish(4, 0.3); c = color.white;", material: true},
	{decl: "m = stone.graniteFacetFinish(4, 0.3, 0.05, 0.4); c = color.white;", material: true},
	{decl: "c = playful.confetti(8, color.cyclicRainbow);"},
	{decl: "c = playful.rainbowSheen(10, color.cyclicRainbow);"},
	{decl: "m, c = playful.glitterTwinkle(material.polishedMetal, color.black, 1.0, color.white, 0.25, 8);", material: true},
	{decl: "m, c = natural.water(8, 3.0, 0.15);", material: true},
	{decl: "m, c = natural.lava(15, [0.12, 0.09, 0.08], [1.0, 0.4, 0.05], 4.0, 5.0);", material: true},
	{decl: "m, c = natural.cloudMaterial(0.5, color.white, 10, 3, 0.5);", material: true},
	{decl: "m, c = weathering.rustedSteel(0.5);", material: true},
	{decl: "m, c = weathering.patinaBronze(0.3);", material: true},
	{decl: "m, c, h, s = animated.ocean(10, 10, [1, 0.3, 0], 0.6);", material: true},
	{decl: "m, c, h, s = animated.lavaLake(8, 20, (x) { return [0.1, 0, 0]; }, 20);", material: true},
	{decl: "m, c, h, s = animated.rain(10, 2);", material: true},
	{decl: "m, c, h, s = animated.river(2, (x) { return [2.5, 0, 0]; }, 0.4);", material: true},
	{decl: "m, c, h, s = animated.mudPot(6, 2, 0.3);", material: true},
	{decl: "m, c, h, s = animated.iceFloes(8, 120, (x) { return [0.01, 0, 0]; }, 0.5, 0.3);", material: true},
	{decl: "m, c, h, s = animated.sandRipples(7.65, 9, [1, 0.2, 0]);", material: true},
	{decl: "steelMat, steelColor = metal.steel(); m, c, h, s = machined.turned(steelMat, steelColor, [0, 0, 0], [1, 0, 0], 0.2, 0.8);", material: true, faded: true},
	{decl: "steelMat, steelColor = metal.steel(); m, c, h, s = machined.faced(steelMat, steelColor, [0, 0, 0], [1, 0, 0], 0.2, 0.8);", material: true, faded: true},
	{decl: "aluminumMat, aluminumColor = metal.aluminum(); m, c, h, s = machined.faceMilled(aluminumMat, aluminumColor, [0, 0, 0], [1, 0, 0], [0, 0, 1], 25, 0.3, 0.8);", material: true, faded: true},
	{decl: "steelMat, steelColor = metal.steel(); m, c, h, s = machined.ground(steelMat, steelColor, [0, 1, 0], 0.3);", material: true},
	{decl: "stainlessMat, stainlessColor = metal.stainless(); m, c, h, s = machined.brushed(stainlessMat, stainlessColor, [1, 0, 0], 150);", material: true},
	{decl: "aluminumMat, aluminumColor = metal.aluminum(); m, c, h, s = machined.blasted(aluminumMat, aluminumColor, 300, 1);", material: true},
	{decl: "aluminumMat, aluminumColor = metal.aluminum(); m, c, h, s = coating.anodized(aluminumMat, aluminumColor, [0.2, 0.35, 0.8]);", material: true},
	{decl: "aluminumMat, aluminumColor = metal.aluminum(); m, c, h, s = coating.hardAnodized(aluminumMat, aluminumColor, 50);", material: true},
	{decl: "steelMat, steelColor = metal.steel(); m, c, h, s = coating.blackOxide(steelMat, 1);", material: true},
	{decl: "steelMat, steelColor = metal.steel(); m, c, h, s = coating.phosphate(steelMat, 1);", material: true},
	{decl: "m, c, h, s = coating.powderCoat([0.7, 0.1, 0.08], 80, 0.6);", material: true},
	{decl: "m, c, h, s = coating.wrinkleCoat([0.05, 0.05, 0.05], 0.6);", material: true},
	{decl: "m, c, h, s = coating.sandCoat([0.1, 0.12, 0.1], 0.2);", material: true},
	{decl: "m, c, h, s = coating.hammertone([0.2, 0.35, 0.3], 4);", material: true},
	{decl: "steelMat, steelColor = metal.steel(); m, c, h, s = coating.chippedPaint(steelMat, steelColor, [0.1, 0.3, 0.15], [0.6, 0.35, 0.3], 60, 0, 1);", material: true},
	{decl: "m, c, h, s = coating.galvanized(4, 0);", material: true},
	{decl: "m, c, h, s = coating.galvanized(4, 0.6);", material: true},
	{decl: "m, c, h, s = printed.fdm([0.85, 0.35, 0.1], [0, 0, 1], [0, 0, 0], 0.2, 40);", material: true, faded: true},
	{decl: "m, c, h, s = printed.resin([0.6, 0.62, 0.64], [0, 0, 1], [0, 0, 0], 0.05, 0);", material: true, faded: true},
	{decl: "m, c, h, s = printed.resin([0.9, 0.92, 0.95], [0, 0, 1], [0, 0, 0], 0.05, 0.9);", material: true, faded: true},
	{decl: "m, c, h, s = printed.powderBed([0.85, 0.84, 0.8], 0.06);", material: true},
	{decl: "titaniumMat, titaniumColor = metal.titanium(); m, c, h, s = printed.laserMelted(titaniumMat, titaniumColor, 10);", material: true, faded: true},
	{decl: "steelMat, steelColor = metal.steel(); m, c, h, s = machined.knurled(steelMat, steelColor, [0, 0, 0], [0, 0, 1], 16, 1.0, 1, 1);", material: true},
	{decl: "m = wood.speciesMaterial(wood.sugarMapleNumber, [0, 0, -500], [0, 0, 1], wood.defaultLook); c = wood.speciesColor(wood.sugarMapleNumber, [0, 0, -500], [0, 0, 1], wood.defaultLook);", material: true},
	{decl: "m = wood.speciesMaterial(wood.whiteOakNumber, [0, 0, -500], [0, 0, 1], wood.defaultLook); c = wood.speciesColor(wood.whiteOakNumber, [0, 0, -500], [0, 0, 1], wood.defaultLook);", material: true},
	{decl: "m = wood.speciesMaterial(wood.walnutNumber, [0, 0, -500], [0, 0, 1], wood.defaultLook); c = wood.speciesColor(wood.walnutNumber, [0, 0, -500], [0, 0, 1], wood.defaultLook);", material: true},
}

// TestTextureStandInMatchesMean checks each texture's declared colour
// stand-in against its mean colour: sampled densely, clamped as the viewer
// draws it, and averaged in Oklab, where the textures fade and mix. A
// material's stand-in is checked against its material's mean, number by
// number. The stand-ins are read by evaluating the same call against a copy
// of the libraries whose simpleValue returns its stand-in, so nothing here
// restates a formula.
func TestTextureStandInMatchesMean(t *testing.T) {
	if testing.Short() {
		t.Skip("dense sampling of every texture")
	}
	textured := gsleval.New(source.NewLocalManager(packageRoot), nil)
	faded := gsleval.New(source.NewLocalManager(fadedPackages(t)), nil)
	flat := gsleval.New(source.NewLocalManager(constantPackages(t)), nil)
	ctx := context.Background()

	r := rand.New(rand.NewSource(7))
	pts := make([][3]float64, samples)
	for i := range pts {
		pts[i] = [3]float64{(r.Float64() - 0.5) * reach, (r.Float64() - 0.5) * reach, (r.Float64() - 0.5) * reach}
	}

	for _, fc := range flatCases {
		t.Run(fc.decl, func(t *testing.T) {
			doc := "import \"color\";\nimport \"material\";\nimport \"material/collection/playful\";\nimport \"material/bulk/stone\";\n" +
				fc.decl + "\nreturn sphere(1);\n"
			tex, standPts := textured, pts[:1]
			if fc.faded {
				tex, standPts = faded, pts
			}
			if fc.material {
				checkMaterial(t, tex, flat, doc, pts, standPts)
			}
			groups, err := tex.EvalGsl(ctx, nil, doc, "", []string{"c"}, pts)
			if err != nil {
				t.Fatalf("texture: %v", err)
			}
			var mean [3]float64
			for i := range pts {
				lab := oklab([3]float64{
					float64(groups[0].Samples[0][i].Value),
					float64(groups[0].Samples[1][i].Value),
					float64(groups[0].Samples[2][i].Value),
				})
				for k := range mean {
					mean[k] += lab[k] / samples
				}
			}
			cg, err := flat.EvalGsl(ctx, nil, doc, "", []string{"c"}, standPts)
			if err != nil {
				t.Fatalf("constant: %v", err)
			}
			var constant [3]float64
			for i := range standPts {
				lab := oklab([3]float64{
					float64(cg[0].Samples[0][i].Value),
					float64(cg[0].Samples[1][i].Value),
					float64(cg[0].Samples[2][i].Value),
				})
				for k := range constant {
					constant[k] += lab[k] / float64(len(standPts))
				}
			}
			d := math.Sqrt(sq(mean[0]-constant[0]) + sq(mean[1]-constant[1]) + sq(mean[2]-constant[2]))
			t.Logf("mean L %.4f a %+.4f b %+.4f, constant L %.4f a %+.4f b %+.4f, distance %.4f",
				mean[0], mean[1], mean[2], constant[0], constant[1], constant[2], d)
			if d > tolerance {
				t.Errorf("declared constant is %.4f from the texture's mean (at most %.2f)", d, tolerance)
			}
		})
	}
}

// checkMaterial checks m's declared stand-in against m's mean, each of
// its numbers on its own: a record's rows are blended linearly. The stand-in
// is averaged over standPts: one point for a constant, every point for a
// faded case's.
func checkMaterial(t *testing.T, textured, flat *gsleval.Evaluator, doc string, pts, standPts [][3]float64) {
	t.Helper()
	ctx := context.Background()
	groups, err := textured.EvalGsl(ctx, nil, doc, "", []string{"m"}, pts)
	if err != nil {
		t.Fatalf("material: %v", err)
	}
	cg, err := flat.EvalGsl(ctx, nil, doc, "", []string{"m"}, standPts)
	if err != nil {
		t.Fatalf("material constant: %v", err)
	}
	worst, at := 0.0, 0
	for k := range groups[0].Samples {
		mean := 0.0
		for i := range pts {
			mean += float64(groups[0].Samples[k][i].Value) / float64(len(pts))
		}
		stand := 0.0
		for i := range standPts {
			stand += float64(cg[0].Samples[k][i].Value) / float64(len(standPts))
		}
		if d := math.Abs(mean - stand); d > worst {
			worst, at = d, k
		}
	}
	t.Logf("material constant within %.4f of the mean (number %d the furthest)", worst, at)
	if worst > materialTolerance {
		t.Errorf("material constant is %.4f from the mean in number %d (at most %.2f)", worst, at, materialTolerance)
	}
}

// fadedPixel is the pixel footprint, in mm, fadedPackages reads: far wider
// than any feature the libraries fade, so every fadeDetail is 0.
const fadedPixel = "1000.0"

// fadedPackages copies the GSL package tree with every pixelSize() replaced by
// fadedPixel, so a texture evaluates as it looks once all its fine detail has
// faded below a pixel.
func fadedPackages(t *testing.T) string {
	t.Helper()
	src, err := filepath.Abs(packageRoot)
	if err != nil {
		t.Fatal(err)
	}
	dst := t.TempDir()
	replaced := 0
	err = filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".gsl") {
			replaced += strings.Count(string(body), "pixelSize()")
			body = []byte(strings.ReplaceAll(string(body), "pixelSize()", fadedPixel))
		}
		return os.WriteFile(target, body, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	if replaced == 0 {
		t.Fatal("no pixelSize() in the libraries to fade")
	}
	return dst
}

func sq(x float64) float64 { return x * x }

// oklab converts an sRGB colour, clamped to [0, 1] as the viewer draws it, to
// Oklab.
func oklab(c [3]float64) [3]float64 {
	var l [3]float64
	for i, v := range c {
		v = math.Min(1, math.Max(0, v))
		if v <= 0.04045 {
			l[i] = v / 12.92
		} else {
			l[i] = math.Pow((v+0.055)/1.055, 2.4)
		}
	}
	lm := 0.4122214708*l[0] + 0.5363325363*l[1] + 0.0514459929*l[2]
	mm := 0.2119034982*l[0] + 0.6806995451*l[1] + 0.1073969566*l[2]
	sm := 0.0883024619*l[0] + 0.2817188376*l[1] + 0.6299787005*l[2]
	lm, mm, sm = math.Cbrt(lm), math.Cbrt(mm), math.Cbrt(sm)
	return [3]float64{
		0.2104542553*lm + 0.7936177850*mm - 0.0040720468*sm,
		1.9779984951*lm - 2.4285922050*mm + 0.4505937099*sm,
		0.0259040371*lm + 0.7827717662*mm - 0.8086757660*sm,
	}
}

// simpleValueBody is builtin's simpleValue as written, and standBody the same
// marker returning its stand-in instead.
const (
	simpleValueBody = "simpleValue(value, stand) {\n  return value;\n}"
	standBody       = "simpleValue(value, stand) {\n  return stand;\n}"
)

// constantPackages copies the package tree with builtin's simpleValue
// returning its stand-in, so every texture that declares one evaluates to it.
func constantPackages(t *testing.T) string {
	t.Helper()
	src, err := filepath.Abs(packageRoot)
	if err != nil {
		t.Fatal(err)
	}
	dst := t.TempDir()
	swapped := false
	err = filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if rel == filepath.Join("builtin", "base.gsl") {
			s := string(body)
			if !strings.Contains(s, simpleValueBody) {
				t.Fatalf("builtin/base.gsl: simpleValue is not written as the test expects")
			}
			body = []byte(strings.Replace(s, simpleValueBody, standBody, 1))
			swapped = true
		}
		return os.WriteFile(target, body, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	if !swapped {
		t.Fatal("builtin/base.gsl not found")
	}
	return dst
}
