// bake-model converts a pinned Sketchfab GLB into bounded braille samples.
// See animations/models/README.md for model attribution and bake commands.
// ponytail: This offline tool supports the pinned GLB 2.0 triangle subset; a
// changed source format updates this converter, never the runtime dependencies.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/Nomadcxx/sysc-Go/internal/mesh"
)

const (
	// The pinned St George source is 82.2 MiB; this remains a bounded offline read.
	maxGLBBytes  = 96 << 20
	maxNodes     = 10000
	maxMeshes    = 2048
	maxAccessors = 4096
	maxVertices  = 1000000
	maxIndices   = 3000000
	maxDepth     = 64
	glbJSON      = 0x4e4f534a
	glbBIN       = 0x004e4942
)

type gltfDocument struct {
	Asset struct {
		Version string `json:"version"`
	} `json:"asset"`
	Scene       *int             `json:"scene"`
	Scenes      []gltfScene      `json:"scenes"`
	Nodes       []gltfNode       `json:"nodes"`
	Meshes      []gltfMesh       `json:"meshes"`
	Accessors   []gltfAccessor   `json:"accessors"`
	BufferViews []gltfBufferView `json:"bufferViews"`
	Buffers     []gltfBuffer     `json:"buffers"`
}

type gltfScene struct {
	Nodes []int `json:"nodes"`
}

type gltfNode struct {
	Name        string    `json:"name"`
	Children    []int     `json:"children"`
	Mesh        *int      `json:"mesh"`
	Matrix      []float64 `json:"matrix"`
	Translation []float64 `json:"translation"`
	Rotation    []float64 `json:"rotation"`
	Scale       []float64 `json:"scale"`
}

type gltfMesh struct {
	Primitives []gltfPrimitive `json:"primitives"`
}

type gltfPrimitive struct {
	Attributes map[string]int `json:"attributes"`
	Indices    *int           `json:"indices"`
	Mode       *int           `json:"mode"`
}

type gltfAccessor struct {
	BufferView    *int            `json:"bufferView"`
	ByteOffset    int             `json:"byteOffset"`
	ComponentType int             `json:"componentType"`
	Count         int             `json:"count"`
	Type          string          `json:"type"`
	Normalized    bool            `json:"normalized"`
	Sparse        json.RawMessage `json:"sparse"`
}

type gltfBufferView struct {
	Buffer     int `json:"buffer"`
	ByteOffset int `json:"byteOffset"`
	ByteLength int `json:"byteLength"`
	ByteStride int `json:"byteStride"`
}

type gltfBuffer struct {
	ByteLength int    `json:"byteLength"`
	URI        string `json:"uri"`
}

type matrix4 [16]float64

func identity() matrix4 { return matrix4{0: 1, 5: 1, 10: 1, 15: 1} }

func parseGLB(data []byte, excludedNodeNames ...string) ([]mesh.Triangle, error) {
	if len(data) < 20 || len(data) > maxGLBBytes || string(data[:4]) != "glTF" || binary.LittleEndian.Uint32(data[4:8]) != 2 || uint64(binary.LittleEndian.Uint32(data[8:12])) != uint64(len(data)) {
		return nil, fmt.Errorf("expected a bounded GLB version 2 file")
	}
	var jsonChunk, binChunk []byte
	for pos, chunk := 12, 0; pos < len(data); chunk++ {
		if len(data)-pos < 8 {
			return nil, fmt.Errorf("truncated GLB chunk header")
		}
		chunkSize := uint64(binary.LittleEndian.Uint32(data[pos : pos+4]))
		typ := binary.LittleEndian.Uint32(data[pos+4 : pos+8])
		pos += 8
		if chunkSize > uint64(len(data)-pos) || chunkSize%4 != 0 {
			return nil, fmt.Errorf("invalid GLB chunk bounds")
		}
		size := int(chunkSize)
		payload := data[pos : pos+size]
		switch typ {
		case glbJSON:
			if chunk != 0 || jsonChunk != nil {
				return nil, fmt.Errorf("GLB JSON must be the first and only JSON chunk")
			}
			jsonChunk = payload
		case glbBIN:
			if chunk != 1 || binChunk != nil {
				return nil, fmt.Errorf("GLB binary data must be the second and only BIN chunk")
			}
			binChunk = payload
		default:
			return nil, fmt.Errorf("unsupported GLB chunk type %#x", typ)
		}
		pos += size
	}
	if jsonChunk == nil || binChunk == nil {
		return nil, fmt.Errorf("GLB must contain JSON and binary chunks")
	}
	var doc gltfDocument
	if err := json.Unmarshal(jsonChunk, &doc); err != nil {
		return nil, fmt.Errorf("invalid GLB JSON: %w", err)
	}
	if doc.Asset.Version != "2.0" || len(doc.Nodes) == 0 || len(doc.Nodes) > maxNodes || len(doc.Meshes) == 0 || len(doc.Meshes) > maxMeshes || len(doc.Accessors) == 0 || len(doc.Accessors) > maxAccessors {
		return nil, fmt.Errorf("unsupported or oversized glTF document")
	}
	if len(doc.Buffers) != 1 || doc.Buffers[0].URI != "" || doc.Buffers[0].ByteLength < 0 || doc.Buffers[0].ByteLength > len(binChunk) || len(binChunk)-doc.Buffers[0].ByteLength > 3 {
		return nil, fmt.Errorf("expected one bounded GLB-embedded buffer")
	}
	if len(doc.BufferViews) == 0 || len(doc.BufferViews) > maxAccessors {
		return nil, fmt.Errorf("missing or oversized glTF buffer views")
	}
	for _, view := range doc.BufferViews {
		if view.Buffer != 0 || view.ByteOffset < 0 || view.ByteLength <= 0 || view.ByteOffset > doc.Buffers[0].ByteLength-view.ByteLength || view.ByteStride < 0 || view.ByteStride > 252 {
			return nil, fmt.Errorf("invalid glTF buffer view")
		}
	}
	if len(doc.Scenes) == 0 || len(doc.Scenes) > maxNodes {
		return nil, fmt.Errorf("glTF has no scene")
	}
	sceneIndex := 0
	if doc.Scene != nil {
		sceneIndex = *doc.Scene
	}
	if sceneIndex < 0 || sceneIndex >= len(doc.Scenes) || len(doc.Scenes[sceneIndex].Nodes) == 0 {
		return nil, fmt.Errorf("invalid or empty selected scene")
	}

	var triangles []mesh.Triangle
	excludedNodes := make(map[string]struct{}, len(excludedNodeNames))
	for _, name := range excludedNodeNames {
		if name != "" {
			excludedNodes[name] = struct{}{}
		}
	}
	state := make([]uint8, len(doc.Nodes))
	var visit func(int, matrix4, int, bool) error
	visit = func(index int, parent matrix4, depth int, skipMesh bool) error {
		if index < 0 || index >= len(doc.Nodes) || depth > maxDepth {
			return fmt.Errorf("invalid node index or scene depth")
		}
		if state[index] != 0 {
			return fmt.Errorf("scene graph contains a cycle or multiply referenced node")
		}
		state[index] = 1
		node := doc.Nodes[index]
		_, excluded := excludedNodes[node.Name]
		skipMesh = skipMesh || excluded
		local, err := nodeTransform(node)
		if err != nil {
			return fmt.Errorf("node %d: %w", index, err)
		}
		world := multiply(parent, local)
		if !boundedMatrix(world) {
			return fmt.Errorf("node %d has an invalid world transform", index)
		}
		if node.Mesh != nil && !skipMesh {
			if *node.Mesh < 0 || *node.Mesh >= len(doc.Meshes) {
				return fmt.Errorf("node %d references an invalid mesh", index)
			}
			if err := appendMesh(&doc, binChunk, *node.Mesh, world, &triangles); err != nil {
				return fmt.Errorf("node %d: %w", index, err)
			}
		}
		for _, child := range node.Children {
			if err := visit(child, world, depth+1, skipMesh); err != nil {
				return err
			}
		}
		state[index] = 2
		return nil
	}
	for _, root := range doc.Scenes[sceneIndex].Nodes {
		if err := visit(root, identity(), 0, false); err != nil {
			return nil, err
		}
	}
	if len(triangles) == 0 {
		return nil, fmt.Errorf("selected scene contains no triangle mesh")
	}
	return triangles, nil
}

func appendMesh(doc *gltfDocument, bin []byte, meshIndex int, world matrix4, out *[]mesh.Triangle) error {
	primitives := doc.Meshes[meshIndex].Primitives
	if len(primitives) == 0 || len(primitives) > maxAccessors {
		return fmt.Errorf("mesh has no primitives or too many primitives")
	}
	// The Y conversion reflects the source winding; a reflected node transform
	// cancels that parity change.
	reverseWinding := determinant3x3(world) >= 0
	for _, primitive := range primitives {
		mode := 4
		if primitive.Mode != nil {
			mode = *primitive.Mode
		}
		if mode != 4 {
			return fmt.Errorf("unsupported primitive mode %d", mode)
		}
		positionIndex, ok := primitive.Attributes["POSITION"]
		if !ok {
			return fmt.Errorf("triangle primitive has no POSITION accessor")
		}
		positions, err := readPositions(doc, bin, positionIndex)
		if err != nil {
			return err
		}
		var sourceNormals []mesh.Vec3
		if normalIndex, ok := primitive.Attributes["NORMAL"]; ok {
			sourceNormals, err = readNormals(doc, bin, normalIndex, len(positions))
			if err != nil {
				return err
			}
		}
		var indices []uint32
		if primitive.Indices != nil {
			indices, err = readIndices(doc, bin, *primitive.Indices)
			if err != nil {
				return err
			}
		} else {
			if len(positions)%3 != 0 {
				return fmt.Errorf("non-indexed vertex count is not divisible by three")
			}
			indices = make([]uint32, len(positions))
			for i := range indices {
				indices[i] = uint32(i)
			}
		}
		if len(indices)%3 != 0 || len(indices) > maxIndices {
			return fmt.Errorf("invalid triangle index count")
		}
		for i := 0; i < len(indices); i += 3 {
			var triangle mesh.Triangle
			normalsUsable := len(sourceNormals) != 0
			for corner := 0; corner < 3; corner++ {
				sourceCorner := corner
				if reverseWinding && corner > 0 {
					sourceCorner = 3 - corner
				}
				index := indices[i+sourceCorner]
				if uint64(index) >= uint64(len(positions)) {
					return fmt.Errorf("vertex index %d is out of range", index)
				}
				point := transformPoint(world, positions[index])
				point[1] = -point[1] // glTF Y-up to the screen's downward-positive axis.
				if !boundedVec(point) {
					return fmt.Errorf("transformed position is nonfinite or oversized")
				}
				triangle.Vertices[corner] = point
				if normalsUsable {
					normal, ok := transformNormal(world, sourceNormals[index])
					if !ok {
						normalsUsable = false
						continue
					}
					normal[1] = -normal[1]
					triangle.Normals[corner] = normal
				}
			}
			if !normalsUsable {
				triangle.Normals = [3]mesh.Vec3{}
			}
			*out = append(*out, triangle)
			if len(*out) > mesh.MaxTriangles {
				return fmt.Errorf("scene exceeds %d triangles", mesh.MaxTriangles)
			}
		}
	}
	return nil
}

func accessorData(doc *gltfDocument, bin []byte, index, itemSize, componentSize, maxCount int) (gltfAccessor, []byte, int, gltfBufferView, error) {
	if index < 0 || index >= len(doc.Accessors) {
		return gltfAccessor{}, nil, 0, gltfBufferView{}, fmt.Errorf("accessor index out of range")
	}
	a := doc.Accessors[index]
	if a.BufferView == nil || a.Count <= 0 || a.Count > maxCount || a.ByteOffset < 0 || (len(a.Sparse) != 0 && !bytes.Equal(bytes.TrimSpace(a.Sparse), []byte("null"))) {
		return gltfAccessor{}, nil, 0, gltfBufferView{}, fmt.Errorf("invalid, sparse, or oversized accessor")
	}
	if *a.BufferView < 0 || *a.BufferView >= len(doc.BufferViews) {
		return gltfAccessor{}, nil, 0, gltfBufferView{}, fmt.Errorf("accessor buffer view index out of range")
	}
	view := doc.BufferViews[*a.BufferView]
	stride := view.ByteStride
	if stride == 0 {
		stride = itemSize
	} else if stride < itemSize || stride%4 != 0 {
		return gltfAccessor{}, nil, 0, gltfBufferView{}, fmt.Errorf("invalid accessor byte stride")
	}
	if a.ByteOffset%componentSize != 0 || view.ByteOffset%componentSize != 0 {
		return gltfAccessor{}, nil, 0, gltfBufferView{}, fmt.Errorf("misaligned accessor data")
	}
	span := (a.Count-1)*stride + itemSize
	if span < itemSize || a.ByteOffset > view.ByteLength-span || view.ByteOffset > len(bin)-view.ByteLength {
		return gltfAccessor{}, nil, 0, gltfBufferView{}, fmt.Errorf("accessor data exceeds buffer view")
	}
	start := view.ByteOffset + a.ByteOffset
	return a, bin[start : start+span], stride, view, nil
}

func readPositions(doc *gltfDocument, bin []byte, index int) ([]mesh.Vec3, error) {
	a, data, stride, _, err := accessorData(doc, bin, index, 12, 4, maxVertices)
	if err != nil {
		return nil, err
	}
	if a.ComponentType != 5126 || a.Type != "VEC3" || a.Normalized {
		return nil, fmt.Errorf("POSITION must be an unnormalized FLOAT VEC3 accessor")
	}
	positions := make([]mesh.Vec3, a.Count)
	for i := range positions {
		for axis := 0; axis < 3; axis++ {
			value := float64(math.Float32frombits(binary.LittleEndian.Uint32(data[i*stride+axis*4:])))
			if math.IsNaN(value) || math.IsInf(value, 0) || math.Abs(value) > 1e8 {
				return nil, fmt.Errorf("POSITION contains nonfinite or oversized data")
			}
			positions[i][axis] = value
		}
	}
	return positions, nil
}

func readNormals(doc *gltfDocument, bin []byte, index, expectedCount int) ([]mesh.Vec3, error) {
	if index < 0 || index >= len(doc.Accessors) {
		return nil, fmt.Errorf("normal accessor index out of range")
	}
	accessor := doc.Accessors[index]
	componentSize, itemSize := 0, 0
	switch accessor.ComponentType {
	case 5126:
		componentSize, itemSize = 4, 12
	case 5120:
		componentSize, itemSize = 1, 3
	case 5122:
		componentSize, itemSize = 2, 6
	default:
		return nil, fmt.Errorf("unsupported NORMAL component type %d", accessor.ComponentType)
	}
	a, data, stride, _, err := accessorData(doc, bin, index, itemSize, componentSize, maxVertices)
	if err != nil {
		return nil, err
	}
	if a.Type != "VEC3" || a.Count != expectedCount || (a.ComponentType == 5126 && a.Normalized) || (a.ComponentType != 5126 && !a.Normalized) {
		return nil, fmt.Errorf("NORMAL must be a matching FLOAT or normalized signed BYTE/SHORT VEC3 accessor")
	}
	normals := make([]mesh.Vec3, a.Count)
	for i := range normals {
		for axis := 0; axis < 3; axis++ {
			component := data[i*stride+axis*componentSize:]
			switch a.ComponentType {
			case 5126:
				normals[i][axis] = float64(math.Float32frombits(binary.LittleEndian.Uint32(component)))
			case 5120:
				value := float64(int8(component[0])) / 127
				normals[i][axis] = math.Max(-1, value)
			case 5122:
				value := float64(int16(binary.LittleEndian.Uint16(component))) / 32767
				normals[i][axis] = math.Max(-1, value)
			}
		}
		if !boundedVec(normals[i]) {
			return nil, fmt.Errorf("NORMAL contains nonfinite or oversized data")
		}
		var ok bool
		normals[i], ok = normalize(normals[i])
		if !ok {
			return nil, fmt.Errorf("NORMAL contains a zero direction")
		}
	}
	return normals, nil
}

func readIndices(doc *gltfDocument, bin []byte, index int) ([]uint32, error) {
	if index < 0 || index >= len(doc.Accessors) {
		return nil, fmt.Errorf("index accessor out of range")
	}
	accessor := doc.Accessors[index]
	itemSize := 0
	switch accessor.ComponentType {
	case 5121:
		itemSize = 1
	case 5123:
		itemSize = 2
	case 5125:
		itemSize = 4
	default:
		return nil, fmt.Errorf("unsupported index component type %d", accessor.ComponentType)
	}
	a, data, stride, view, err := accessorData(doc, bin, index, itemSize, itemSize, maxIndices)
	if err != nil {
		return nil, err
	}
	if a.Type != "SCALAR" || a.Normalized || view.ByteStride != 0 {
		return nil, fmt.Errorf("indices must be an unnormalized, tightly packed SCALAR accessor")
	}
	indices := make([]uint32, a.Count)
	for i := range indices {
		switch itemSize {
		case 1:
			indices[i] = uint32(data[i*stride])
		case 2:
			indices[i] = uint32(binary.LittleEndian.Uint16(data[i*stride:]))
		case 4:
			indices[i] = binary.LittleEndian.Uint32(data[i*stride:])
		}
	}
	return indices, nil
}

func nodeTransform(node gltfNode) (matrix4, error) {
	if len(node.Matrix) != 0 {
		if len(node.Matrix) != 16 || len(node.Translation)+len(node.Rotation)+len(node.Scale) != 0 {
			return matrix4{}, fmt.Errorf("matrix must contain 16 values and cannot be combined with TRS")
		}
		var out matrix4
		for i, value := range node.Matrix {
			if !finiteBounded(value) {
				return matrix4{}, fmt.Errorf("matrix contains invalid values")
			}
			out[i] = value
		}
		if math.Abs(out[3])+math.Abs(out[7])+math.Abs(out[11])+math.Abs(out[15]-1) > 1e-8 {
			return matrix4{}, fmt.Errorf("projective node matrices are unsupported")
		}
		return out, nil
	}
	translation, err := triple(node.Translation, mesh.Vec3{0, 0, 0})
	if err != nil {
		return matrix4{}, err
	}
	scale, err := triple(node.Scale, mesh.Vec3{1, 1, 1})
	if err != nil {
		return matrix4{}, err
	}
	rotation := [4]float64{0, 0, 0, 1}
	if len(node.Rotation) != 0 {
		if len(node.Rotation) != 4 {
			return matrix4{}, fmt.Errorf("rotation must contain four values")
		}
		norm := 0.0
		for i, value := range node.Rotation {
			if !finiteBounded(value) {
				return matrix4{}, fmt.Errorf("rotation contains invalid values")
			}
			rotation[i] = value
			norm += value * value
		}
		if norm <= 1e-20 {
			return matrix4{}, fmt.Errorf("rotation quaternion has zero length")
		}
		norm = math.Sqrt(norm)
		for i := range rotation {
			rotation[i] /= norm
		}
	}
	x, y, z, w := rotation[0], rotation[1], rotation[2], rotation[3]
	rot := identity()
	rot[0], rot[4], rot[8] = 1-2*y*y-2*z*z, 2*x*y-2*z*w, 2*x*z+2*y*w
	rot[1], rot[5], rot[9] = 2*x*y+2*z*w, 1-2*x*x-2*z*z, 2*y*z-2*x*w
	rot[2], rot[6], rot[10] = 2*x*z-2*y*w, 2*y*z+2*x*w, 1-2*x*x-2*y*y
	for row := 0; row < 3; row++ {
		rot[row] *= scale[0]
		rot[4+row] *= scale[1]
		rot[8+row] *= scale[2]
	}
	rot[12], rot[13], rot[14] = translation[0], translation[1], translation[2]
	return rot, nil
}

func triple(values []float64, fallback mesh.Vec3) (mesh.Vec3, error) {
	if len(values) == 0 {
		return fallback, nil
	}
	if len(values) != 3 {
		return mesh.Vec3{}, fmt.Errorf("translation or scale must contain three values")
	}
	var out mesh.Vec3
	for i, value := range values {
		if !finiteBounded(value) {
			return mesh.Vec3{}, fmt.Errorf("translation or scale contains invalid values")
		}
		out[i] = value
	}
	return out, nil
}

func multiply(a, b matrix4) matrix4 {
	var out matrix4
	for column := 0; column < 4; column++ {
		for row := 0; row < 4; row++ {
			for k := 0; k < 4; k++ {
				out[column*4+row] += a[k*4+row] * b[column*4+k]
			}
		}
	}
	return out
}

func transformPoint(m matrix4, point mesh.Vec3) mesh.Vec3 {
	return mesh.Vec3{
		m[0]*point[0] + m[4]*point[1] + m[8]*point[2] + m[12],
		m[1]*point[0] + m[5]*point[1] + m[9]*point[2] + m[13],
		m[2]*point[0] + m[6]*point[1] + m[10]*point[2] + m[14],
	}
}

func transformNormal(m matrix4, normal mesh.Vec3) (mesh.Vec3, bool) {
	a00, a01, a02 := m[0], m[4], m[8]
	a10, a11, a12 := m[1], m[5], m[9]
	a20, a21, a22 := m[2], m[6], m[10]
	c00 := a11*a22 - a12*a21
	c01 := a12*a20 - a10*a22
	c02 := a10*a21 - a11*a20
	c10 := a02*a21 - a01*a22
	c11 := a00*a22 - a02*a20
	c12 := a01*a20 - a00*a21
	c20 := a01*a12 - a02*a11
	c21 := a02*a10 - a00*a12
	c22 := a00*a11 - a01*a10
	determinant := a00*c00 + a01*c01 + a02*c02
	if determinant == 0 || math.IsNaN(determinant) || math.IsInf(determinant, 0) {
		return mesh.Vec3{}, false
	}
	// The cofactor matrix points along the inverse-transpose normal; determinant
	// sign preserves orientation for reflected node transforms.
	direction := mesh.Vec3{
		c00*normal[0] + c01*normal[1] + c02*normal[2],
		c10*normal[0] + c11*normal[1] + c12*normal[2],
		c20*normal[0] + c21*normal[1] + c22*normal[2],
	}
	if determinant < 0 {
		for axis := range direction {
			direction[axis] = -direction[axis]
		}
	}
	return normalize(direction)
}

func normalize(vector mesh.Vec3) (mesh.Vec3, bool) {
	length := math.Hypot(math.Hypot(vector[0], vector[1]), vector[2])
	if length == 0 || math.IsNaN(length) || math.IsInf(length, 0) {
		return mesh.Vec3{}, false
	}
	for axis := range vector {
		vector[axis] /= length
	}
	return vector, true
}

func determinant3x3(m matrix4) float64 {
	a, b, c := m[0], m[4], m[8]
	d, e, f := m[1], m[5], m[9]
	g, h, i := m[2], m[6], m[10]
	return a*(e*i-f*h) - b*(d*i-f*g) + c*(d*h-e*g)
}

func finiteBounded(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && math.Abs(value) <= 1e8
}

func boundedVec(v mesh.Vec3) bool {
	return finiteBounded(v[0]) && finiteBounded(v[1]) && finiteBounded(v[2])
}

func boundedMatrix(m matrix4) bool {
	for _, value := range m {
		if !finiteBounded(value) {
			return false
		}
	}
	return true
}

func verifySourceDigest(data []byte, expected string) error {
	want, err := hex.DecodeString(expected)
	if err != nil || len(want) != sha256.Size {
		return fmt.Errorf("expected source SHA-256 must be 64 hexadecimal characters")
	}
	got := sha256.Sum256(data)
	if !bytes.Equal(want, got[:]) {
		return fmt.Errorf("source SHA-256 mismatch: got %x", got)
	}
	return nil
}

func run() error {
	flags := flag.NewFlagSet("bake-model", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), "usage: bake-model -sha256 DIGEST [-samples COUNT] [-exclude-node NAME] source.glb output.bin")
		flags.PrintDefaults()
	}
	expectedDigest := flags.String("sha256", "", "expected source SHA-256 (required)")
	sampleCount := flags.Int("samples", mesh.SampleCount, "number of samples (24,000–96,000)")
	var excludedNodes []string
	flags.Func("exclude-node", "omit a named node and all its children (repeatable)", func(name string) error {
		if name == "" {
			return fmt.Errorf("excluded node name cannot be empty")
		}
		excludedNodes = append(excludedNodes, name)
		return nil
	})
	if err := flags.Parse(os.Args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 2 {
		return fmt.Errorf("usage: bake-model -sha256 DIGEST [-samples COUNT] [-exclude-node NAME] source.glb output.bin")
	}
	if *expectedDigest == "" {
		return fmt.Errorf("-sha256 is required")
	}
	if *sampleCount < mesh.SampleCount || *sampleCount > mesh.MaxSampleCount {
		return fmt.Errorf("-samples must be between %d and %d", mesh.SampleCount, mesh.MaxSampleCount)
	}
	f, err := os.Open(flags.Arg(0))
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxGLBBytes+1))
	if err != nil {
		return err
	}
	if len(data) > maxGLBBytes {
		return fmt.Errorf("source exceeds %d MiB", maxGLBBytes>>20)
	}
	if err := verifySourceDigest(data, *expectedDigest); err != nil {
		return err
	}
	triangles, err := parseGLB(data, excludedNodes...)
	if err != nil {
		return err
	}
	samples, bounds, err := mesh.SampleSurfaceN(triangles, *sampleCount)
	if err != nil {
		return err
	}
	encoded, err := mesh.Encode(samples)
	if err != nil {
		return err
	}
	if err = os.WriteFile(flags.Arg(1), encoded, 0644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "source bounds %v .. %v\n", bounds.Min, bounds.Max)
	fmt.Printf("baked %d triangles into %d samples (%d bytes)\n", len(triangles), len(samples), len(encoded))
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
