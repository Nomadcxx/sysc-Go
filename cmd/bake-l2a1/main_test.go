package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"testing"

	"github.com/Nomadcxx/sysc-Go/internal/mesh"
)

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
	want := meshTriangle{{11, -22, 33}, {11, -24, 33}, {8, -22, 33}}
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
	want := meshTriangle{{5, -6, 7}, {6, -6, 7}, {5, -6, 8}}
	if len(triangles) != 1 || !sameTriangle(triangles[0], want) {
		t.Fatalf("matrix transformed triangles = %v, want %v", triangles, want)
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
	for i := range a {
		for axis := range a[i] {
			if math.Abs(a[i][axis]-b[i][axis]) > 1e-12 {
				return false
			}
		}
	}
	return true
}

func testGLB(t *testing.T, indices []uint16, meshNode map[string]any, rootNode map[string]any) []byte {
	t.Helper()
	var bin bytes.Buffer
	for _, value := range []float32{0, 0, 0, 1, 0, 0, 0, 1, 0} {
		if err := binary.Write(&bin, binary.LittleEndian, value); err != nil {
			t.Fatal(err)
		}
	}
	for _, index := range indices {
		if err := binary.Write(&bin, binary.LittleEndian, index); err != nil {
			t.Fatal(err)
		}
	}
	root := map[string]any{"children": []int{1}}
	for key, value := range rootNode {
		root[key] = value
	}
	doc := map[string]any{
		"asset":  map[string]any{"version": "2.0"},
		"scene":  0,
		"scenes": []any{map[string]any{"nodes": []int{0}}},
		"nodes":  []any{root, meshNode},
		"meshes": []any{map[string]any{"primitives": []any{map[string]any{
			"attributes": map[string]int{"POSITION": 0}, "indices": 1, "mode": 4,
		}}}},
		"accessors": []any{
			map[string]any{"bufferView": 0, "componentType": 5126, "count": 3, "type": "VEC3"},
			map[string]any{"bufferView": 1, "componentType": 5123, "count": len(indices), "type": "SCALAR"},
		},
		"bufferViews": []any{
			map[string]any{"buffer": 0, "byteOffset": 0, "byteLength": 36},
			map[string]any{"buffer": 0, "byteOffset": 36, "byteLength": 2 * len(indices)},
		},
		"buffers": []any{map[string]any{"byteLength": bin.Len()}},
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
