//go:build cgo

package eval_test

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math"
	"sync"
	"testing"

	"gradientcontrol.com/omega/core/eval"
)

func oracleCube(x float32) string {
	vertices := [][3]float32{{-1, -1, -1}, {1, -1, -1}, {1, 1, -1}, {-1, 1, -1}, {-1, -1, 1}, {1, -1, 1}, {1, 1, 1}, {-1, 1, 1}}
	triangles := [][3]int{{0, 2, 1}, {0, 3, 2}, {4, 5, 6}, {4, 6, 7}, {0, 1, 5}, {0, 5, 4}, {3, 7, 6}, {3, 6, 2}, {0, 4, 7}, {0, 7, 3}, {1, 2, 6}, {1, 6, 5}}
	var packed []byte
	for _, triangle := range triangles {
		for _, index := range triangle {
			point := vertices[index]
			point[0] += x
			for _, value := range point {
				packed = binary.LittleEndian.AppendUint32(packed, math.Float32bits(value))
			}
		}
	}
	return base64.StdEncoding.EncodeToString(packed)
}

func TestOracleIsolationAcrossEvaluators(t *testing.T) {
	for _, kind := range []string{"bvhSdf", "vdbSdf"} {
		t.Run(kind, func(t *testing.T) {
			var evaluators [2]*eval.Evaluator
			for i := range evaluators {
				payload := oracleCube(float32(i * 10))
				evaluators[i] = eval.New(mockMgr, func(context.Context, map[string]string, []byte, bool) (*eval.EvalAssetsResponse, error) {
					assets := &eval.EvalAssetsResponse{}
					if kind == "bvhSdf" {
						assets.MeshOracles = []eval.MeshOraclePayload{{AssetID: 0, VerticesB64: payload}}
					} else {
						assets.VdbOracles = []eval.VdbOraclePayload{{AssetID: 0, VerticesB64: payload, VoxelSize: 0.1, HalfBandVoxels: 20}}
					}
					return assets, nil
				})
			}
			check := func(i int) {
				point := [3]float64{float64(i * 10), 0, 0}
				expr := fmt.Sprintf("%s(0)", kind)
				exprs := []string{expr, fmt.Sprintf("remap(%s, [%d,0,0])", expr, i*10)}
				results, err := evaluators[i].EvalGsl(t.Context(), nil, "return "+expr+";", "", exprs, [][3]float64{point})
				if err != nil {
					t.Error(err)
					return
				}
				for _, result := range results {
					if value := result.Samples[0][0].Value; math.Abs(float64(value)+1) > 0.05 {
						t.Errorf("evaluator %d: got %g, want -1", i, value)
					}
				}
			}
			for _, i := range []int{0, 1, 0, 1} {
				check(i)
			}
			var wg sync.WaitGroup
			for i := 0; i < 8; i++ {
				wg.Go(func() { check(i % 2) })
			}
			wg.Wait()
		})
	}
}
