package mesh

import (
	"math"
	"testing"
)

func TestSurfaceSampleEncodingRoundTrip(t *testing.T) {
	triangles := []Triangle{{{0, 0, 0}, {2, 0, 0}, {0, 2, 0}}}
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

func TestSurfaceSamplerRejectsInvalidMeshes(t *testing.T) {
	for _, triangles := range [][]Triangle{
		nil,
		{{{0, 0, 0}, {1, 1, 1}, {2, 2, 2}}},
		{{{0, 0, 0}, {math.NaN(), 1, 0}, {0, 1, 0}}},
	} {
		if _, _, err := SampleSurface(triangles); err == nil {
			t.Fatal("accepted empty, degenerate, or non-finite mesh")
		}
	}
	if _, err := Decode([]byte{1}); err == nil {
		t.Fatal("accepted truncated sample data")
	}
}
