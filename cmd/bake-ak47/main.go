// bake-ak47 converts the pinned Lokeig FBX into bounded braille surface samples.
// The derived data is CC BY-NC 4.0; see animations/models/README.md.
package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/Nomadcxx/sysc-Go/internal/mesh"
)

type node struct {
	name     string
	props    []any
	children []node
}
type vec = mesh.Vec3
type triangle = mesh.Triangle
type point = mesh.Sample

func (n node) child(name string) node {
	for _, c := range n.children {
		if c.name == name {
			return c
		}
	}
	return node{}
}

// ponytail: FBX 7400 only, with bounded records and arrays. A different source
// format needs an offline converter update, never a runtime FBX dependency.
func parseNodes(data []byte, start, end, depth int) ([]node, error) {
	if depth > 32 {
		return nil, fmt.Errorf("FBX nesting exceeds 32")
	}
	var out []node
	for p := start; p < end; {
		if end-p < 13 {
			return nil, fmt.Errorf("truncated FBX node at %d", p)
		}
		stop := int(binary.LittleEndian.Uint32(data[p:]))
		count := int(binary.LittleEndian.Uint32(data[p+4:]))
		size := int(binary.LittleEndian.Uint32(data[p+8:]))
		namesize := int(data[p+12])
		p += 13
		if stop == 0 {
			if count != 0 || size != 0 || namesize != 0 {
				return nil, fmt.Errorf("invalid FBX terminator")
			}
			return out, nil
		}
		if stop > end || stop < p+namesize || count > 100000 || size > stop-p-namesize {
			return nil, fmt.Errorf("invalid FBX node bounds")
		}
		n := node{name: string(data[p : p+namesize])}
		p += namesize
		propsEnd := p + size
		r := bytes.NewReader(data[p:propsEnd])
		for j := 0; j < count; j++ {
			v, err := property(r)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", n.name, err)
			}
			n.props = append(n.props, v)
		}
		if r.Len() != 0 {
			return nil, fmt.Errorf("FBX property size mismatch")
		}
		var err error
		n.children, err = parseNodes(data, propsEnd, stop, depth+1)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
		p = stop
	}
	return out, nil
}
func property(r *bytes.Reader) (any, error) {
	typ, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	read := func(v any) error { return binary.Read(r, binary.LittleEndian, v) }
	switch typ {
	case 'Y':
		var v int16
		err = read(&v)
		return int64(v), err
	case 'C':
		v, e := r.ReadByte()
		return int64(v), e
	case 'I':
		var v int32
		err = read(&v)
		return int64(v), err
	case 'L':
		var v int64
		err = read(&v)
		return v, err
	case 'F':
		var v float32
		err = read(&v)
		return float64(v), err
	case 'D':
		var v float64
		err = read(&v)
		return v, err
	case 'S', 'R':
		var size uint32
		if err = read(&size); err != nil {
			return nil, err
		}
		if uint64(size) > uint64(r.Len()) {
			return nil, io.ErrUnexpectedEOF
		}
		b := make([]byte, size)
		_, err = io.ReadFull(r, b)
		return string(b), err
	case 'd', 'f', 'i', 'l', 'b':
		var h [3]uint32
		if err = read(&h); err != nil {
			return nil, err
		}
		width := map[byte]uint64{'d': 8, 'f': 4, 'i': 4, 'l': 8, 'b': 1}[typ]
		size := uint64(h[0]) * width
		if size > 32<<20 || uint64(h[2]) > uint64(r.Len()) || h[1] > 1 {
			return nil, fmt.Errorf("invalid FBX array bounds/encoding")
		}
		packed := make([]byte, h[2])
		if _, err = io.ReadFull(r, packed); err != nil {
			return nil, err
		}
		b := packed
		if h[1] == 1 {
			z, e := zlib.NewReader(bytes.NewReader(packed))
			if e != nil {
				return nil, e
			}
			b, err = io.ReadAll(io.LimitReader(z, int64(size)+1))
			closeErr := z.Close()
			if err != nil {
				return nil, err
			}
			if closeErr != nil {
				return nil, closeErr
			}
		}
		if uint64(len(b)) != size {
			return nil, fmt.Errorf("FBX array length mismatch")
		}
		switch typ {
		case 'd':
			v := make([]float64, h[0])
			err = binary.Read(bytes.NewReader(b), binary.LittleEndian, v)
			return v, err
		case 'i':
			v := make([]int32, h[0])
			err = binary.Read(bytes.NewReader(b), binary.LittleEndian, v)
			return v, err
		}
		return nil, nil
	}
	return nil, fmt.Errorf("unsupported FBX property %q", typ)
}
func readFBX(data []byte) ([]triangle, error) {
	if len(data) < 40 || len(data) > 8<<20 || string(data[:23]) != "Kaydara FBX Binary  \x00\x1a\x00" || binary.LittleEndian.Uint32(data[23:]) != 7400 {
		return nil, fmt.Errorf("expected bounded binary FBX 7400")
	}
	nodes, err := parseNodes(data, 27, len(data), 0)
	if err != nil {
		return nil, err
	}
	root := node{children: nodes}
	objects := root.child("Objects")
	models := map[int64]node{}
	for _, m := range objects.children {
		if m.name != "Model" {
			continue
		}
		if len(m.props) != 3 {
			return nil, fmt.Errorf("invalid model properties")
		}
		id, ok := m.props[0].(int64)
		if !ok {
			return nil, fmt.Errorf("invalid model ID")
		}
		models[id] = m
		for _, p := range m.child("Properties70").children {
			if len(p.props) < 7 {
				continue
			}
			name, _ := p.props[0].(string)
			switch name {
			case "GeometricRotation", "GeometricScaling", "PostRotation", "RotationOffset", "RotationPivot", "ScalingOffset", "ScalingPivot":
				want := 0.0
				if name == "GeometricScaling" {
					want = 1
				}
				for _, v := range p.props[4:7] {
					f, ok := v.(float64)
					if !ok || math.IsNaN(f) || math.IsInf(f, 0) || math.Abs(f-want) > 1e-8 {
						return nil, fmt.Errorf("unsupported %s transform %v on %v", name, p.props[4:], m.props)
					}
				}
			}
		}
	}
	parents := map[int64]int64{}
	for _, c := range root.child("Connections").children {
		if len(c.props) != 3 {
			continue
		}
		kind, _ := c.props[0].(string)
		if kind == "OO" {
			a, aok := c.props[1].(int64)
			b, bok := c.props[2].(int64)
			if !aok || !bok {
				return nil, fmt.Errorf("invalid connection")
			}
			if _, ok := models[a]; ok && b != 0 && models[b].name != "" {
				return nil, fmt.Errorf("nested model transforms unsupported")
			}
			if models[b].name != "" {
				parents[a] = b
			}
		}
	}
	var tris []triangle
	for _, g := range objects.children {
		if g.name != "Geometry" {
			continue
		}
		if len(g.props) != 3 {
			return nil, fmt.Errorf("invalid geometry properties")
		}
		id, ok := g.props[0].(int64)
		if !ok {
			return nil, fmt.Errorf("invalid geometry ID")
		}
		model := models[parents[id]]
		if model.name == "" {
			return nil, fmt.Errorf("geometry has no model")
		}
		values := map[string]vec{}
		values["Lcl Scaling"] = vec{1, 1, 1}
		for _, p := range model.child("Properties70").children {
			if len(p.props) == 5 {
				name, _ := p.props[0].(string)
				if name == "RotationOrder" {
					order, ok := p.props[4].(int64)
					if !ok || order != 0 {
						return nil, fmt.Errorf("non-XYZ rotation order unsupported")
					}
				}
			}
			if len(p.props) != 7 {
				continue
			}
			name, _ := p.props[0].(string)
			var v vec
			for k := range v {
				f, ok := p.props[k+4].(float64)
				if !ok || math.IsNaN(f) || math.IsInf(f, 0) {
					return nil, fmt.Errorf("invalid transform")
				}
				v[k] = f
			}
			values[name] = v
		}
		v, p := g.child("Vertices"), g.child("PolygonVertexIndex")
		if len(v.props) != 1 || len(p.props) != 1 {
			return nil, fmt.Errorf("missing mesh arrays")
		}
		verts, ok := v.props[0].([]float64)
		if !ok || len(verts)%3 != 0 {
			return nil, fmt.Errorf("invalid vertices")
		}
		indices, ok := p.props[0].([]int32)
		if !ok {
			return nil, fmt.Errorf("invalid polygon indices")
		}
		for _, f := range verts {
			if math.IsNaN(f) || math.IsInf(f, 0) || math.Abs(f) > 1e8 {
				return nil, fmt.Errorf("nonfinite/oversized vertex")
			}
		}
		var poly []vec
		for _, idx := range indices {
			end := idx < 0
			if end {
				idx = ^idx
			}
			if int64(idx)*3+2 >= int64(len(verts)) {
				return nil, fmt.Errorf("vertex index out of range")
			}
			q := vec{verts[3*idx], verts[3*idx+1], verts[3*idx+2]}
			for k := range q {
				q[k] = (q[k] + values["GeometricTranslation"][k]) * values["Lcl Scaling"][k]
			}
			q = rotate(q, values["Lcl Rotation"])
			q = rotate(q, values["PreRotation"])
			for k := range q {
				q[k] += values["Lcl Translation"][k]
				if math.IsNaN(q[k]) || math.IsInf(q[k], 0) || math.Abs(q[k]) > 1e8 {
					return nil, fmt.Errorf("invalid transformed vertex")
				}
			}
			q[1] = -q[1] // FBX Y-up to screen Y-down.
			poly = append(poly, q)
			if len(poly) > 64 {
				return nil, fmt.Errorf("polygon exceeds 64 corners")
			}
			if end {
				if len(poly) < 3 {
					return nil, fmt.Errorf("short polygon")
				}
				for j := 1; j+1 < len(poly); j++ {
					tris = append(tris, triangle{Vertices: [3]vec{poly[0], poly[j], poly[j+1]}})
				}
				poly = poly[:0]
			}
		}
		if len(poly) != 0 {
			return nil, fmt.Errorf("unterminated polygon")
		}
	}
	if len(tris) == 0 || len(tris) > 100000 {
		return nil, fmt.Errorf("empty/oversized mesh")
	}
	return tris, nil
}

func rotate(p, degrees vec) vec {
	for axis, angle := range degrees {
		s, c := math.Sincos(angle * math.Pi / 180)
		a, b := (axis+1)%3, (axis+2)%3
		p[a], p[b] = c*p[a]-s*p[b], s*p[a]+c*p[b]
	}
	return p
}
func sampleMesh(tris []triangle) ([]point, error) {
	samples, bounds, err := mesh.SampleSurface(tris)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "source bounds %v .. %v\n", bounds.Min, bounds.Max)
	return samples, nil
}
func run() error {
	if len(os.Args) != 3 {
		return fmt.Errorf("usage: bake-ak47 source/AKM.fbx animations/models/ak47.bin")
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, (8<<20)+1))
	if err != nil {
		return err
	}
	if len(data) > 8<<20 {
		return fmt.Errorf("source exceeds 8 MiB")
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != "152f608c545ebc691a711c72fdb3ea7637304c4b83d472c4217fd9f4571639b4" {
		return fmt.Errorf("source digest differs from credited AKM.fbx")
	}
	tris, err := readFBX(data)
	if err != nil {
		return err
	}
	points, err := sampleMesh(tris)
	if err != nil {
		return err
	}
	out, err := mesh.Encode(points)
	if err != nil {
		return err
	}
	if err = os.WriteFile(os.Args[2], out, 0644); err != nil {
		return err
	}
	fmt.Printf("baked %d triangles into %d samples (%d bytes)\n", len(tris), len(points), len(out))
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
