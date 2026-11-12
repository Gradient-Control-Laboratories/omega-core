package threemf

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"io"
	"math"
	"strconv"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// binarySTL builds a binary STL from triangles of three xyz corners each.
func binarySTL(tris [][3][3]float32) []byte {
	buf := make([]byte, 84+50*len(tris))
	binary.LittleEndian.PutUint32(buf[80:], uint32(len(tris)))
	for t, tri := range tris {
		base := 84 + t*50 + 12
		for c := 0; c < 3; c++ {
			for k := 0; k < 3; k++ {
				binary.LittleEndian.PutUint32(buf[base+c*12+k*4:], math.Float32bits(tri[c][k]))
			}
		}
	}
	return buf
}

func TestWeldSTLSharesCornersAndDropsCollapsed(t *testing.T) {
	negZero := float32(math.Copysign(0, -1))
	stl := binarySTL([][3][3]float32{
		{{0, 0, 0}, {1, 0, 0}, {0, 1, 0}},
		{{1, 0, 0}, {1, 1, 0}, {0, 1, 0}},
		{{negZero, 0, 0}, {0, 0, 1}, {1, 0, 0}}, // -0 welds to the first corner
		{{2, 2, 2}, {2, 2, 2}, {3, 3, 3}},       // collapses: dropped
	})
	m, err := WeldSTL(stl)
	require.NoError(t, err)
	assert.Equal(t, []float32{0, 0, 0, 1, 0, 0, 0, 1, 0, 1, 1, 0, 0, 0, 1, 2, 2, 2, 3, 3, 3}, m.Positions)
	assert.Equal(t, []uint32{0, 1, 2, 1, 3, 2, 0, 4, 1}, m.Indices)
	assert.False(t, math.Signbit(float64(m.Positions[0])), "-0 is written as 0")
}

func TestWeldSTLRejectsTruncated(t *testing.T) {
	_, err := WeldSTL(make([]byte, 40))
	require.Error(t, err)
	stl := binarySTL([][3][3]float32{{{0, 0, 0}, {1, 0, 0}, {0, 1, 0}}})
	_, err = WeldSTL(stl[:len(stl)-1])
	require.Error(t, err)
}

func TestCoordRoundTripsShortest(t *testing.T) {
	for _, v := range []float32{0, 1, -1.5, 0.1, 1.0000001, 12.345678, 1e-5, 1234567, 3.4028235e38} {
		s := coord(v)
		assert.False(t, strings.Contains(s, "e"), "no exponent form: %s", s)
		back, err := strconv.ParseFloat(s, 32)
		require.NoError(t, err)
		assert.Equal(t, v, float32(back), "%s round-trips", s)
	}
	assert.Equal(t, "0.1", coord(0.1))
	assert.Equal(t, "1", coord(1))
	assert.Equal(t, "0.00001", coord(1e-5))
}

func TestColorAttr(t *testing.T) {
	assert.Equal(t, "#A1B2C3", colorAttr("#a1b2c3"))
	assert.Equal(t, "#808080", colorAttr(""))
	assert.Equal(t, "#808080", colorAttr("#abc"))
	assert.Equal(t, "#808080", colorAttr("red"))
}

// TestWriteModelMatchesBrowserExporter pins the model document to the layout
// threemf.js writes for the same bodies.
func TestWriteModelMatchesBrowserExporter(t *testing.T) {
	bodies := []Body{
		{Name: `a<&>"'`, Color: "#ff0000", Mesh: Mesh{Positions: []float32{0, 0, 0, 1, 0, 0, 0, 1, 0}, Indices: []uint32{0, 1, 2}}},
		{Name: "b", Color: "nope", Mesh: Mesh{Positions: []float32{0.5, -2, 3, 1, 1, 1, 0, 0, 1}, Indices: []uint32{2, 1, 0}}},
	}
	var buf bytes.Buffer
	require.NoError(t, WriteModel(&buf, bodies))
	want := `<?xml version="1.0" encoding="UTF-8"?>
<model unit="millimeter" xml:lang="en-US" xmlns="http://schemas.microsoft.com/3dmanufacturing/core/2015/02" xmlns:m="http://schemas.microsoft.com/3dmanufacturing/material/2015/02">
 <metadata name="Application">Omega</metadata>
 <resources>
  <m:colorgroup id="1">
   <m:color color="#FF0000"/>
   <m:color color="#808080"/>
  </m:colorgroup>
  <object id="2" type="model" name="a&lt;&amp;&gt;&quot;&apos;" pid="1" pindex="0">
   <mesh>
    <vertices>
     <vertex x="0" y="0" z="0"/>
     <vertex x="1" y="0" z="0"/>
     <vertex x="0" y="1" z="0"/>
    </vertices>
    <triangles>
     <triangle v1="0" v2="1" v3="2"/>
    </triangles>
   </mesh>
  </object>
  <object id="3" type="model" name="b" pid="1" pindex="1">
   <mesh>
    <vertices>
     <vertex x="0.5" y="-2" z="3"/>
     <vertex x="1" y="1" z="1"/>
     <vertex x="0" y="0" z="1"/>
    </vertices>
    <triangles>
     <triangle v1="2" v2="1" v3="0"/>
    </triangles>
   </mesh>
  </object>
 </resources>
 <build>
  <item objectid="2"/>
  <item objectid="3"/>
 </build>
</model>
`
	assert.Equal(t, want, buf.String())
}

func TestWritePackage(t *testing.T) {
	stl := binarySTL([][3][3]float32{
		{{0, 0, 0}, {1, 0, 0}, {0, 1, 0}},
		{{0, 0, 0}, {0, 1, 0}, {0, 0, 1}},
	})
	m, err := WeldSTL(stl)
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, Write(&buf, []Body{{Name: "part", Color: "#00ff00", Mesh: m}}))

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)
	var names []string
	files := map[string]string{}
	for _, f := range zr.File {
		names = append(names, f.Name)
		rc, err := f.Open()
		require.NoError(t, err)
		b, err := io.ReadAll(rc)
		require.NoError(t, err)
		rc.Close()
		files[f.Name] = string(b)
	}
	assert.Equal(t, []string{"[Content_Types].xml", "_rels/.rels", "3D/3dmodel.model"}, names)
	assert.Equal(t, contentTypes, files["[Content_Types].xml"])
	assert.Equal(t, rels, files["_rels/.rels"])

	// The model is well-formed XML with the welded mesh in it.
	var model struct {
		Unit      string `xml:"unit,attr"`
		Resources struct {
			Objects []struct {
				Name     string     `xml:"name,attr"`
				Vertices []struct{} `xml:"mesh>vertices>vertex"`
				Tris     []struct{} `xml:"mesh>triangles>triangle"`
			} `xml:"object"`
		} `xml:"resources"`
	}
	require.NoError(t, xml.Unmarshal([]byte(files["3D/3dmodel.model"]), &model))
	assert.Equal(t, "millimeter", model.Unit)
	require.Len(t, model.Resources.Objects, 1)
	assert.Equal(t, "part", model.Resources.Objects[0].Name)
	assert.Len(t, model.Resources.Objects[0].Vertices, 4)
	assert.Len(t, model.Resources.Objects[0].Tris, 2)
	assert.True(t, strings.Contains(files["3D/3dmodel.model"], `<m:color color="#00FF00"/>`))
}

func TestWriteNoBodies(t *testing.T) {
	err := Write(io.Discard, nil)
	require.Error(t, err)
	assert.Equal(t, "nothing to write: no bodies", err.Error())
}
