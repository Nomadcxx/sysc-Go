package animations

import (
	_ "embed"
	"fmt"
	"math"
	"sync"
	"time"
)

// Model-derived data and source licenses: see models/README.md.
//
//go:embed models/jesus-christ-statue.bin
var jesusChristStatueData []byte

//go:embed models/jason.bin
var jasonData []byte

//go:embed models/nymphe-der-flora.bin
var nympheDerFloraData []byte

//go:embed models/lowe.bin
var loweData []byte

//go:embed models/roman-gladiator.bin
var romanGladiatorData []byte

//go:embed models/grabfigur.bin
var grabfigurData []byte

const rotatingModelPeriod = 12 * time.Second

type rotatingModelSpec struct {
	data         []byte
	credit       string
	displayScale float64
	decode       sync.Once
	points       []meshPoint
}

func (s *rotatingModelSpec) surface() []meshPoint {
	s.decode.Do(func() { s.points = decodeMeshPoints(s.data) })
	return s.points
}

var rotatingModelCatalog = map[string]*rotatingModelSpec{
	"jesus-christ-statue": {data: jesusChristStatueData, credit: "Jesus Christ Statue / amcgi / CC BY 4.0", displayScale: .56},
	"jason":               {data: jasonData, credit: "Jason / noe-3d.at / CC BY-NC 4.0", displayScale: .56},
	"nymphe-der-flora":    {data: nympheDerFloraData, credit: "Nymphe der Flora / noe-3d.at / CC BY-NC 4.0", displayScale: .56},
	"lowe":                {data: loweData, credit: "Löwe / noe-3d.at / CC BY-NC 4.0", displayScale: .72},
	"roman-gladiator":     {data: romanGladiatorData, credit: "Sculpture of a Roman gladiator / Virtual Museums of Małopolska / CC BY 4.0", displayScale: .49},
	"grabfigur":           {data: grabfigurData, credit: "Grabfigur / noe-3d.at / CC BY-NC 4.0", displayScale: .58},
}

// RotatingModelAnimation keeps a statue upright and turns it around its vertical axis.
type RotatingModelAnimation struct {
	*meshSpin
	spec *rotatingModelSpec
}

// NewRotatingModelEffect creates one of the six attributed horizontal-turn effects.
func NewRotatingModelEffect(id string, width, height int, theme string) (*RotatingModelAnimation, error) {
	spec, ok := rotatingModelCatalog[id]
	if !ok {
		return nil, fmt.Errorf("unknown rotating model effect %q", id)
	}
	spin := newMeshSpin(width, height, theme, rotatingModelPeriod)
	spin.twoSidedLighting = false
	return &RotatingModelAnimation{meshSpin: spin, spec: spec}, nil
}

func (a *RotatingModelAnimation) Render() string {
	rotation := 2 * math.Pi * float64(a.elapsed) / float64(a.period)
	frame := renderMeshSpinAtPitch(a.meshSpin, a.spec.surface(), rotation, 0, 0, nil, 0, 0, a.spec.displayScale, 0)
	return meshCredit(frame, a.canvas.width, a.spec.credit)
}

var _ Animation = (*RotatingModelAnimation)(nil)
