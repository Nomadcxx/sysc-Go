package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/Nomadcxx/sysc-Go/internal/mesh"
)

func TestVerifySourceDigest(t *testing.T) {
	data := []byte("pinned model")
	digest := fmt.Sprintf("%x", sha256.Sum256(data))
	if err := verifySourceDigest(data, digest); err != nil {
		t.Fatalf("verifySourceDigest() = %v", err)
	}
	if err := verifySourceDigest(data, strings.ToUpper(digest)); err != nil {
		t.Fatalf("verifySourceDigest() rejected uppercase hex: %v", err)
	}
	for _, invalid := range []string{"", digest[:len(digest)-1], strings.Repeat("z", len(digest)), strings.Repeat("0", len(digest))} {
		if err := verifySourceDigest(data, invalid); err == nil {
			t.Fatalf("verifySourceDigest() accepted %q", invalid)
		}
	}
}

func TestParseGLBAppliesParentAndNodeTransforms(t *testing.T) {
	triangles, err := parseGLB(testGLB(t, []uint16{0, 1, 2}, map[string]any{
		"mesh":        0,
		"scale":       []float64{2, 3, 4},
		"rotation":    []float64{0, 0, 0.7071067811865476, 0.7071067811865476},
		"translation": []float64{1, 2, 3},
	}, map[string]any{
		"children":    []int{1},
		"translation": []float64{10, 20, 30},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if len(triangles) != 1 {
		t.Fatalf("got %d triangles", len(triangles))
	}
	want := meshTriangle{Vertices: [3]mesh.Vec3{{11, -22, 33}, {8, -22, 33}, {11, -24, 33}}}
	if !sameTriangle(triangles[0], want) {
		t.Fatalf("transformed triangle = %v, want %v", triangles[0], want)
	}
}

func TestParseGLBAppliesColumnMajorMatrix(t *testing.T) {
	matrix := []float64{1, 0, 0, 0, 0, 0, 1, 0, 0, -1, 0, 0, 5, 6, 7, 1}
	triangles, err := parseGLB(testGLB(t, []uint16{0, 1, 2}, map[string]any{
		"mesh":   0,
		"matrix": matrix,
	}, nil))
	if err != nil {
		t.Fatal(err)
	}
	want := meshTriangle{Vertices: [3]mesh.Vec3{{5, -6, 7}, {5, -6, 8}, {6, -6, 7}}}
	if len(triangles) != 1 || !sameTriangle(triangles[0], want) {
		t.Fatalf("matrix transformed triangles = %v, want %v", triangles, want)
	}
}

func TestParseGLBPreservesOutwardNormalsAfterYConversion(t *testing.T) {
	for _, test := range []struct {
		name      string
		transform map[string]any
	}{
		{name: "identity", transform: map[string]any{}},
		{name: "reflected node", transform: map[string]any{"scale": []float64{-1, 1, 1}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			normals := [][3]float32{{0, 0, 1}, {0, 0, 1}, {0, 0, 1}}
			triangles, err := parseGLB(testGLB(t, []uint16{0, 1, 2}, map[string]any{"mesh": 0}, test.transform, normals...))
			if err != nil {
				t.Fatal(err)
			}
			triangle := triangles[0].Vertices
			ab := mesh.Vec3{triangle[1][0] - triangle[0][0], triangle[1][1] - triangle[0][1], triangle[1][2] - triangle[0][2]}
			ac := mesh.Vec3{triangle[2][0] - triangle[0][0], triangle[2][1] - triangle[0][1], triangle[2][2] - triangle[0][2]}
			normalZ := ab[0]*ac[1] - ab[1]*ac[0]
			if normalZ <= 0 {
				t.Fatalf("converted outward +Z normal = %g, want positive", normalZ)
			}
			if triangles[0].Normals[0][2] <= 0 {
				t.Fatalf("transformed authored normal = %v, want to agree with converted winding", triangles[0].Normals[0])
			}
		})
	}
}

func TestParseGLBInterpolatesAuthoredNormals(t *testing.T) {
	normals := [][3]float32{{0, 0, 1}, {.6, 0, .8}, {0, .6, .8}}
	triangles, err := parseGLB(testGLB(t, []uint16{0, 1, 2}, map[string]any{"mesh": 0}, nil, normals...))
	if err != nil {
		t.Fatal(err)
	}
	samples, _, err := mesh.SampleSurfaceN(triangles, mesh.SampleCount)
	if err != nil {
		t.Fatal(err)
	}
	u := math.Sqrt(.5 * .7548776662466927)
	v := .5 * .5698402909980532
	weights := [3]float64{1 - u, u * (1 - v), u * v}
	// The baker reverses vertex order when the Y conversion reflects winding.
	want := mesh.Vec3{.6 * weights[2], -.6 * weights[1], weights[0] + .8*weights[1] + .8*weights[2]}
	length := math.Sqrt(want[0]*want[0] + want[1]*want[1] + want[2]*want[2])
	for axis := range want {
		want[axis] /= length
	}
	for axis := range want {
		if math.Abs(samples[0].Normal[axis]-want[axis]) > 1e-7 {
			t.Fatalf("sample normal = %v, want interpolated source normal %v", samples[0].Normal, want)
		}
	}
}

func TestParseGLBTransformsAuthoredNormals(t *testing.T) {
	for _, test := range []struct {
		name      string
		transform map[string]any
		normal    [3]float32
		want      mesh.Vec3
	}{
		{
			name:      "non-uniform scale",
			transform: map[string]any{"mesh": 0, "scale": []float64{2, 1, 1}},
			normal:    [3]float32{.70710677, .70710677, 0},
			want:      mesh.Vec3{1 / math.Sqrt(5), -2 / math.Sqrt(5), 0},
		},
		{
			name: "rotation and non-uniform scale",
			transform: map[string]any{
				"mesh": 0, "scale": []float64{2, 1, 4},
				"rotation": []float64{math.Sqrt(.5), 0, 0, math.Sqrt(.5)},
			},
			normal: [3]float32{0, 0, 1},
			want:   mesh.Vec3{0, 1, 0},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			normals := [][3]float32{test.normal, test.normal, test.normal}
			triangles, err := parseGLB(testGLB(t, []uint16{0, 1, 2}, test.transform, nil, normals...))
			if err != nil {
				t.Fatal(err)
			}
			samples, _, err := mesh.SampleSurfaceN(triangles, mesh.SampleCount)
			if err != nil {
				t.Fatal(err)
			}
			for axis := range test.want {
				if math.Abs(samples[0].Normal[axis]-test.want[axis]) > 1e-6 {
					t.Fatalf("sample normal = %v, want inverse-transpose transformed normal %v", samples[0].Normal, test.want)
				}
			}
		})
	}
}

func TestParseGLBReadsNormalizedShortNormals(t *testing.T) {
	normals := [][3]int16{{32767, 0, 0}, {32767, 0, 0}, {32767, 0, 0}}
	triangles, err := parseGLB(testGLBWithShortNormals(t, []uint16{0, 1, 2}, map[string]any{"mesh": 0}, nil, normals...))
	if err != nil {
		t.Fatal(err)
	}
	samples, _, err := mesh.SampleSurfaceN(triangles, mesh.SampleCount)
	if err != nil {
		t.Fatal(err)
	}
	if want := (mesh.Vec3{1, 0, 0}); samples[0].Normal != want {
		t.Fatalf("sample normal = %v, want normalized signed SHORT normal %v", samples[0].Normal, want)
	}
}

func TestParseGLBReadsNormalizedByteNormals(t *testing.T) {
	normals := [][3]int8{{-128, 0, 0}, {-128, 0, 0}, {-128, 0, 0}}
	triangles, err := parseGLB(testGLBWithByteNormals(t, []uint16{0, 1, 2}, map[string]any{"mesh": 0}, nil, normals...))
	if err != nil {
		t.Fatal(err)
	}
	samples, _, err := mesh.SampleSurfaceN(triangles, mesh.SampleCount)
	if err != nil {
		t.Fatal(err)
	}
	if want := (mesh.Vec3{-1, 0, 0}); samples[0].Normal != want {
		t.Fatalf("sample normal = %v, want normalized signed BYTE normal %v", samples[0].Normal, want)
	}
}

func TestParseGLBUsesFaceNormalsWhenAuthoredNormalsAreMissing(t *testing.T) {
	triangles, err := parseGLB(testGLB(t, []uint16{0, 1, 2}, map[string]any{"mesh": 0}, nil))
	if err != nil {
		t.Fatal(err)
	}
	samples, _, err := mesh.SampleSurfaceN(triangles, mesh.SampleCount)
	if err != nil {
		t.Fatal(err)
	}
	if want := (mesh.Vec3{0, 0, 1}); samples[0].Normal != want {
		t.Fatalf("sample normal = %v, want geometric face normal %v", samples[0].Normal, want)
	}
}

func TestParseGLBFallsBackWhenNodeTransformCannotTransformNormals(t *testing.T) {
	normals := [][3]float32{{0, 0, 1}, {0, 0, 1}, {0, 0, 1}}
	triangles, err := parseGLB(testGLB(t, []uint16{0, 1, 2}, map[string]any{
		"mesh": 0, "scale": []float64{1, 1, 0},
	}, nil, normals...))
	if err != nil {
		t.Fatal(err)
	}
	samples, _, err := mesh.SampleSurfaceN(triangles, mesh.SampleCount)
	if err != nil {
		t.Fatal(err)
	}
	if want := (mesh.Vec3{0, 0, 1}); samples[0].Normal != want {
		t.Fatalf("sample normal = %v, want geometric fallback %v", samples[0].Normal, want)
	}
}

func TestParseGLBRejectsZeroAuthoredNormals(t *testing.T) {
	normals := [][3]float32{{0, 0, 1}, {0, 0, 0}, {0, 0, 1}}
	if _, err := parseGLB(testGLB(t, []uint16{0, 1, 2}, map[string]any{"mesh": 0}, nil, normals...)); err == nil {
		t.Fatal("accepted a zero authored normal")
	}
}

func TestParseGLBExcludesNamedSubtreesOnlyWhenRequested(t *testing.T) {
	for _, name := range []string{"308 casing_30", "308 cartridge_31"} {
		t.Run(name, func(t *testing.T) {
			data := testGLB(t, []uint16{0, 1, 2}, map[string]any{"mesh": 0}, map[string]any{"name": name})
			if triangles, err := parseGLB(data); err != nil || len(triangles) != 1 {
				t.Fatalf("parseGLB() = %d triangles, %v; want the unfiltered subtree", len(triangles), err)
			}
			if _, err := parseGLB(data, name); err == nil || !strings.Contains(err.Error(), "no triangle mesh") {
				t.Fatalf("parseGLB() error = %v, want selected scene with no triangles after subtree omission", err)
			}
		})
	}
}

func TestParseGLBRejectsBadChunkBoundsAndIndices(t *testing.T) {
	data := testGLB(t, []uint16{0, 1, 2}, map[string]any{"mesh": 0}, nil)
	binary.LittleEndian.PutUint32(data[12:16], ^uint32(0))
	if _, err := parseGLB(data); err == nil {
		t.Fatal("accepted chunk length outside the file")
	}
	if _, err := parseGLB(testGLB(t, []uint16{0, 1, 3}, map[string]any{"mesh": 0}, nil)); err == nil {
		t.Fatal("accepted an index outside the position accessor")
	}
}

type meshTriangle = mesh.Triangle

func sameTriangle(a, b mesh.Triangle) bool {
	for i := range a.Vertices {
		for axis := range a.Vertices[i] {
			if math.Abs(a.Vertices[i][axis]-b.Vertices[i][axis]) > 1e-12 {
				return false
			}
		}
	}
	return true
}

func testGLB(t *testing.T, indices []uint16, meshNode map[string]any, rootNode map[string]any, normalValues ...[3]float32) []byte {
	t.Helper()
	var normalData bytes.Buffer
	for _, normal := range normalValues {
		if err := binary.Write(&normalData, binary.LittleEndian, normal); err != nil {
			t.Fatal(err)
		}
	}
	return testGLBData(t, indices, meshNode, rootNode, normalData.Bytes(), len(normalValues), 5126, false, 4)
}

func testGLBWithShortNormals(t *testing.T, indices []uint16, meshNode map[string]any, rootNode map[string]any, normalValues ...[3]int16) []byte {
	t.Helper()
	var normalData bytes.Buffer
	for _, normal := range normalValues {
		if err := binary.Write(&normalData, binary.LittleEndian, normal); err != nil {
			t.Fatal(err)
		}
	}
	return testGLBData(t, indices, meshNode, rootNode, normalData.Bytes(), len(normalValues), 5122, true, 2)
}

func testGLBWithByteNormals(t *testing.T, indices []uint16, meshNode map[string]any, rootNode map[string]any, normalValues ...[3]int8) []byte {
	t.Helper()
	var normalData bytes.Buffer
	for _, normal := range normalValues {
		if err := binary.Write(&normalData, binary.LittleEndian, normal); err != nil {
			t.Fatal(err)
		}
	}
	return testGLBData(t, indices, meshNode, rootNode, normalData.Bytes(), len(normalValues), 5120, true, 1)
}

func testGLBData(t *testing.T, indices []uint16, meshNode map[string]any, rootNode map[string]any, normalData []byte, normalCount, normalComponentType int, normalNormalized bool, normalComponentSize int) []byte {
	t.Helper()
	var bin bytes.Buffer
	for _, value := range []float32{0, 0, 0, 1, 0, 0, 0, 1, 0} {
		if err := binary.Write(&bin, binary.LittleEndian, value); err != nil {
			t.Fatal(err)
		}
	}
	bin.Write(normalData)
	bin.Write(make([]byte, (4-bin.Len()%4)%4))
	indexOffset := bin.Len()
	for _, index := range indices {
		if err := binary.Write(&bin, binary.LittleEndian, index); err != nil {
			t.Fatal(err)
		}
	}
	root := map[string]any{"children": []int{1}}
	for key, value := range rootNode {
		root[key] = value
	}
	attributes := map[string]int{"POSITION": 0}
	accessors := []any{
		map[string]any{"bufferView": 0, "componentType": 5126, "count": 3, "type": "VEC3"},
		map[string]any{"bufferView": 1, "componentType": 5123, "count": len(indices), "type": "SCALAR"},
	}
	bufferViews := []any{
		map[string]any{"buffer": 0, "byteOffset": 0, "byteLength": 36},
		map[string]any{"buffer": 0, "byteOffset": indexOffset, "byteLength": 2 * len(indices)},
	}
	if normalCount > 0 {
		attributes["NORMAL"] = 2
		accessors[1].(map[string]any)["bufferView"] = 2
		accessors = append(accessors, map[string]any{"bufferView": 1, "componentType": normalComponentType, "count": normalCount, "normalized": normalNormalized, "type": "VEC3"})
		bufferViews = []any{
			map[string]any{"buffer": 0, "byteOffset": 0, "byteLength": 36},
			map[string]any{"buffer": 0, "byteOffset": 36, "byteLength": 3 * normalComponentSize * normalCount},
			map[string]any{"buffer": 0, "byteOffset": indexOffset, "byteLength": 2 * len(indices)},
		}
	}
	doc := map[string]any{
		"asset":  map[string]any{"version": "2.0"},
		"scene":  0,
		"scenes": []any{map[string]any{"nodes": []int{0}}},
		"nodes":  []any{root, meshNode},
		"meshes": []any{map[string]any{"primitives": []any{map[string]any{
			"attributes": attributes, "indices": 1, "mode": 4,
		}}}},
		"accessors":   accessors,
		"bufferViews": bufferViews,
		"buffers":     []any{map[string]any{"byteLength": bin.Len()}},
	}
	if meshNode["matrix"] != nil {
		delete(meshNode, "scale")
		delete(meshNode, "rotation")
		delete(meshNode, "translation")
	}
	jsonData, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	for len(jsonData)%4 != 0 {
		jsonData = append(jsonData, ' ')
	}
	binData := bin.Bytes()
	for len(binData)%4 != 0 {
		binData = append(binData, 0)
	}
	total := 12 + 8 + len(jsonData) + 8 + len(binData)
	data := make([]byte, total)
	copy(data, "glTF")
	binary.LittleEndian.PutUint32(data[4:8], 2)
	binary.LittleEndian.PutUint32(data[8:12], uint32(total))
	pos := 12
	binary.LittleEndian.PutUint32(data[pos:pos+4], uint32(len(jsonData)))
	binary.LittleEndian.PutUint32(data[pos+4:pos+8], 0x4e4f534a)
	copy(data[pos+8:], jsonData)
	pos += 8 + len(jsonData)
	binary.LittleEndian.PutUint32(data[pos:pos+4], uint32(len(binData)))
	binary.LittleEndian.PutUint32(data[pos+4:pos+8], 0x004e4942)
	copy(data[pos+8:], binData)
	return data
}
