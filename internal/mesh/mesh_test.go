package mesh

import (
	"math"
	"testing"
)

func TestSurfaceSampleEncodingRoundTrip(t *testing.T) {
	triangles := []Triangle{{Vertices: [3]Vec3{{0, 0, 0}, {2, 0, 0}, {0, 2, 0}}}}
	samples, bounds, err := SampleSurface(triangles)
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) != SampleCount || bounds.Min != (Vec3{0, 0, 0}) || bounds.Max != (Vec3{2, 2, 0}) {
		t.Fatalf("unexpected samples/bounds: %d, %+v", len(samples), bounds)
	}
	for _, sample := range samples {
		if sample.Position[2] != 0 || sample.Position[0]+sample.Position[1] > 1e-12 || sample.Normal != (Vec3{0, 0, 1}) {
			t.Fatal("sample left the normalized triangle surface")
		}
	}

	data, err := Encode(samples)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != SampleCount*SampleBytes {
		t.Fatalf("encoded %d bytes", len(data))
	}
	decoded, err := Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	for i := range samples {
		for axis := range samples[i].Position {
			if math.Abs(samples[i].Position[axis]-decoded[i].Position[axis]) > 1.0/32767 {
				t.Fatalf("position sample %d axis %d did not round-trip", i, axis)
			}
			if math.Abs(samples[i].Normal[axis]-decoded[i].Normal[axis]) > 1.0/32767 {
				t.Fatalf("normal sample %d axis %d did not round-trip", i, axis)
			}
		}
	}
}

func TestSurfaceSamplingInterpolatesAndNormalizesVertexNormals(t *testing.T) {
	triangle := Triangle{
		Vertices: [3]Vec3{{0, 0, 0}, {2, 0, 0}, {0, 2, 0}},
		Normals:  [3][3]float32{{0, 0, 1}, {0, 1, 1}, {1, 0, 1}},
	}
	samples, _, err := SampleSurfaceN([]Triangle{triangle}, SampleCount)
	if err != nil {
		t.Fatal(err)
	}
	first, last := samples[0].Normal, samples[len(samples)-1].Normal
	if first == last {
		t.Fatal("sampled vertex normals were not interpolated across the triangle")
	}
	for i, sample := range samples {
		if math.Abs(length(sample.Normal)-1) > 1e-12 {
			t.Fatalf("sample %d normal length = %g, want 1", i, length(sample.Normal))
		}
	}
}

func TestSurfaceSamplingAndCodecSupportHigherSampleCounts(t *testing.T) {
	triangles := []Triangle{{Vertices: [3]Vec3{{0, 0, 0}, {2, 0, 0}, {0, 2, 0}}}}
	samples, _, err := SampleSurfaceN(triangles, SampleCount*2)
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) != SampleCount*2 {
		t.Fatalf("got %d samples, want %d", len(samples), SampleCount*2)
	}
	data, err := Encode(samples)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded) != len(samples) {
		t.Fatalf("decoded %d samples, want %d", len(decoded), len(samples))
	}
}

func TestSurfaceSamplerRejectsInvalidMeshes(t *testing.T) {
	for _, triangles := range [][]Triangle{
		nil,
		{{Vertices: [3]Vec3{{0, 0, 0}, {1, 1, 1}, {2, 2, 2}}}},
		{{Vertices: [3]Vec3{{0, 0, 0}, {math.NaN(), 1, 0}, {0, 1, 0}}}},
	} {
		if _, _, err := SampleSurface(triangles); err == nil {
			t.Fatal("accepted empty, degenerate, or non-finite mesh")
		}
	}
	if _, err := Decode([]byte{1}); err == nil {
		t.Fatal("accepted truncated sample data")
	}
	for _, count := range []int{SampleCount - 1, MaxSampleCount + 1} {
		if _, _, err := SampleSurfaceN([]Triangle{{Vertices: [3]Vec3{{0, 0, 0}, {1, 0, 0}, {0, 1, 0}}}}, count); err == nil {
			t.Fatalf("accepted invalid sample count %d", count)
		}
	}
}
