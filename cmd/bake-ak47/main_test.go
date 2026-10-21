package main

import (
	"math"
	"os"
	"testing"
)

func TestAK47Bake(t *testing.T) {
	for _, b := range [][]byte{nil, []byte("not FBX"), append([]byte("Kaydara FBX Binary  \x00\x1a\x00"), 0xe8, 0x1c, 0, 0)} {
		if _, err := readFBX(b); err == nil {
			t.Fatal("accepted missing/truncated FBX")
		}
	}
	if path := os.Getenv("AK47_SOURCE"); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		tris, err := readFBX(data)
		if err != nil {
			t.Fatal(err)
		}
		if len(tris) != 17397 {
			t.Fatalf("only %d triangles", len(tris))
		}
		pts, err := sampleMesh(tris)
		if err != nil {
			t.Fatal(err)
		}
		if len(pts) != 24000 {
			t.Fatalf("got %d surface samples", len(pts))
		}
		for _, p := range pts {
			for _, v := range p.Position {
				if math.IsNaN(v) || math.IsInf(v, 0) || math.Abs(v) > 1 {
					t.Fatal("unbounded point")
				}
			}
			normalLength := math.Sqrt(p.Normal[0]*p.Normal[0] + p.Normal[1]*p.Normal[1] + p.Normal[2]*p.Normal[2])
			if math.Abs(normalLength-1) > 1e-9 {
				t.Fatal("nonunit normal")
			}
		}
		t.Logf("%d triangles, %d points", len(tris), len(pts))
	}
	// Every sample of a known plane must lie on it with a unit normal.
	pts, err := sampleMesh([]triangle{{vec{0, 0, 0}, vec{2, 0, 0}, vec{0, 2, 0}}})
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range pts {
		if p.Position[2] != 0 || p.Position[0]+p.Position[1] > 1e-12 || p.Position[0] < -1 || p.Position[1] < -1 || p.Normal != (vec{0, 0, 1}) {
			t.Fatal("sample left triangle surface")
		}
	}
	q := rotate(vec{0, 1, 0}, vec{-90, 0, 0})
	if math.Abs(q[2]+1) > 1e-9 || math.Abs(q[1]) > 1e-9 {
		t.Fatal("incorrect FBX pre-rotation")
	}
}
