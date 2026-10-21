// Package mesh owns the baked surface format and area sampler shared by model
// conversion tools and the embedded braille effects.
package mesh

import (
	"encoding/binary"
	"fmt"
	"math"
	"sort"
)

const (
	// ponytail: 24,000 samples remain the default; statue densities are raised
	// only after qualifying the frame budget at the largest supported grid.
	SampleCount    = 24000
	MaxSampleCount = 96000
	SampleBytes    = 12
	// MaxTriangles bounds geometry work in the offline model converters.
	MaxTriangles = 1000000
)

type Vec3 [3]float64
type Triangle [3]Vec3

type Sample struct {
	Position Vec3
	Normal   Vec3
}

type Bounds struct {
	Min Vec3
	Max Vec3
}

func sub(a, b Vec3) Vec3 { return Vec3{a[0] - b[0], a[1] - b[1], a[2] - b[2]} }

func cross(a, b Vec3) Vec3 {
	return Vec3{a[1]*b[2] - a[2]*b[1], a[2]*b[0] - a[0]*b[2], a[0]*b[1] - a[1]*b[0]}
}

func length(v Vec3) float64 { return math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2]) }

// SampleSurface area-weights triangles, normalizes the longest source axis to
// [-1,1], then places a deterministic low-discrepancy set of samples.
func SampleSurface(triangles []Triangle) ([]Sample, Bounds, error) {
	return SampleSurfaceN(triangles, SampleCount)
}

// SampleSurfaceN is SampleSurface with a bounded number of deterministic samples.
func SampleSurfaceN(triangles []Triangle, sampleCount int) ([]Sample, Bounds, error) {
	if len(triangles) == 0 || len(triangles) > MaxTriangles {
		return nil, Bounds{}, fmt.Errorf("empty or oversized mesh")
	}
	if sampleCount < SampleCount || sampleCount > MaxSampleCount {
		return nil, Bounds{}, fmt.Errorf("sample count must be between %d and %d", SampleCount, MaxSampleCount)
	}
	lo, hi := Vec3{math.Inf(1), math.Inf(1), math.Inf(1)}, Vec3{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
	for _, triangle := range triangles {
		for _, point := range triangle {
			for axis, value := range point {
				if math.IsNaN(value) || math.IsInf(value, 0) || math.Abs(value) > 1e8 {
					return nil, Bounds{}, fmt.Errorf("nonfinite or oversized vertex")
				}
				lo[axis] = math.Min(lo[axis], value)
				hi[axis] = math.Max(hi[axis], value)
			}
		}
	}
	bounds := Bounds{Min: lo, Max: hi}
	span := math.Max(hi[0]-lo[0], math.Max(hi[1]-lo[1], hi[2]-lo[2]))
	if span <= 0 || math.IsInf(span, 0) || math.IsNaN(span) {
		return nil, Bounds{}, fmt.Errorf("degenerate bounds")
	}
	normalized := append([]Triangle(nil), triangles...)
	for i := range normalized {
		for j := range normalized[i] {
			for axis := range normalized[i][j] {
				normalized[i][j][axis] = (normalized[i][j][axis] - (lo[axis]+hi[axis])/2) * 2 / span
			}
		}
	}

	areas := make([]float64, len(normalized))
	normals := make([]Vec3, len(normalized))
	total := 0.0
	for i, triangle := range normalized {
		normal := cross(sub(triangle[1], triangle[0]), sub(triangle[2], triangle[0]))
		area2 := length(normal)
		if area2 > 0 {
			for axis := range normal {
				normal[axis] /= area2
			}
		}
		normals[i] = normal
		total += area2 / 2
		areas[i] = total
	}
	if total <= 0 || math.IsInf(total, 0) || math.IsNaN(total) {
		return nil, Bounds{}, fmt.Errorf("zero or invalid surface area")
	}

	samples := make([]Sample, sampleCount)
	for i := range samples {
		index := sort.SearchFloat64s(areas, (float64(i)+.5)*total/float64(len(samples)))
		triangle := normalized[index]
		u := math.Sqrt(math.Mod((float64(i)+.5)*.7548776662466927, 1))
		v := math.Mod((float64(i)+.5)*.5698402909980532, 1)
		var position Vec3
		for axis := range position {
			position[axis] = triangle[0][axis]*(1-u) + triangle[1][axis]*u*(1-v) + triangle[2][axis]*u*v
		}
		samples[i] = Sample{Position: position, Normal: normals[index]}
	}
	return samples, bounds, nil
}

// Encode stores each sample as little-endian int16 position XYZ then normal
// XYZ, with values quantized from [-1,1].
func Encode(samples []Sample) ([]byte, error) {
	if len(samples) < SampleCount || len(samples) > MaxSampleCount {
		return nil, fmt.Errorf("sample count must be between %d and %d, got %d", SampleCount, MaxSampleCount, len(samples))
	}
	data := make([]byte, len(samples)*SampleBytes)
	for i, sample := range samples {
		for j, value := range [6]float64{
			sample.Position[0], sample.Position[1], sample.Position[2],
			sample.Normal[0], sample.Normal[1], sample.Normal[2],
		} {
			if math.IsNaN(value) || math.IsInf(value, 0) || math.Abs(value) > 1 {
				return nil, fmt.Errorf("sample %d component %d is outside [-1,1]", i, j)
			}
			binary.LittleEndian.PutUint16(data[i*SampleBytes+j*2:], uint16(int16(math.Round(value*32767))))
		}
	}
	return data, nil
}

func Decode(data []byte) ([]Sample, error) {
	if len(data)%SampleBytes != 0 || len(data)/SampleBytes < SampleCount || len(data)/SampleBytes > MaxSampleCount {
		return nil, fmt.Errorf("expected between %d and %d sample records, got %d bytes", SampleCount, MaxSampleCount, len(data))
	}
	samples := make([]Sample, len(data)/SampleBytes)
	for i := range samples {
		for axis := 0; axis < 3; axis++ {
			samples[i].Position[axis] = float64(int16(binary.LittleEndian.Uint16(data[i*SampleBytes+axis*2:]))) / 32767
			samples[i].Normal[axis] = float64(int16(binary.LittleEndian.Uint16(data[i*SampleBytes+6+axis*2:]))) / 32767
		}
	}
	return samples, nil
}
