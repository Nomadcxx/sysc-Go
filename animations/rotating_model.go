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

//go:embed models/laocoon-and-his-sons.bin
var laocoonData []byte

//go:embed models/napoleon-bust.bin
var napoleonBustData []byte

//go:embed models/brandenburg-gate.bin
var brandenburgGateData []byte

//go:embed models/louis-xiv.bin
var louisXIVData []byte

//go:embed models/engel.bin
var engelData []byte

//go:embed models/st-george.bin
var stGeorgeData []byte

//go:embed models/santa-muerte.bin
var santaMuerteData []byte

//go:embed models/aspasia.bin
var aspasiaData []byte

//go:embed models/st-nepomuk.bin
var stNepomukData []byte

//go:embed models/lion-statue.bin
var lionStatueData []byte

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
	"jesus-christ-statue":  {data: jesusChristStatueData, credit: "Jesus Christ Statue / amcgi / CC BY 4.0", displayScale: .56},
	"jason":                {data: jasonData, credit: "Jason / noe-3d.at / CC BY-NC 4.0", displayScale: .56},
	"nymphe-der-flora":     {data: nympheDerFloraData, credit: "Nymphe der Flora / noe-3d.at / CC BY-NC 4.0", displayScale: .56},
	"lowe":                 {data: loweData, credit: "Löwe / noe-3d.at / CC BY-NC 4.0", displayScale: .72},
	"roman-gladiator":      {data: romanGladiatorData, credit: "Sculpture of a Roman gladiator / Virtual Museums of Małopolska / CC BY 4.0", displayScale: .49},
	"grabfigur":            {data: grabfigurData, credit: "Grabfigur / noe-3d.at / CC BY-NC 4.0", displayScale: .58},
	"laocoon-and-his-sons": {data: laocoonData, credit: "Laocoon and His Sons / Rigsters / CC BY 4.0", displayScale: .62},
	"napoleon-bust":        {data: napoleonBustData, credit: "The bust of Napoleon Bonaparte / Virtual Museums of Małopolska / CC BY 4.0", displayScale: .62},
	"brandenburg-gate":     {data: brandenburgGateData, credit: "Brandenburg Gate / Resinnnn / CC BY 4.0", displayScale: .96},
	"louis-xiv":            {data: louisXIVData, credit: "Louis XIV / HoangHiepVu / CC BY 4.0", displayScale: .62},
	"engel":                {data: engelData, credit: "Engel / noe-3d.at / CC BY-NC 4.0", displayScale: .62},
	"st-george":            {data: stGeorgeData, credit: "Hl. Georg als Drachentöter / VIMUNE / CC BY-SA 4.0", displayScale: .62},
	"santa-muerte":         {data: santaMuerteData, credit: "Santa Muerte / Mateusz Woliński / CC BY-SA 4.0", displayScale: .60},
	"aspasia":              {data: aspasiaData, credit: "Aspasia / noe-3d.at / CC BY-NC 4.0", displayScale: .62},
	"st-nepomuk":           {data: stNepomukData, credit: "Hl. Nepomuk / noe-3d.at / CC BY-NC 4.0", displayScale: .60},
	"lion-statue":          {data: lionStatueData, credit: "Lion Statue / Thomas Flynn / CC0 1.0", displayScale: .62},
}

// RotatingModelAnimation keeps a statue upright and turns it around its vertical axis.
type RotatingModelAnimation struct {
	*meshSpin
	spec *rotatingModelSpec
}

// NewRotatingModelEffect creates an attributed horizontal-turn effect by catalog ID.
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
