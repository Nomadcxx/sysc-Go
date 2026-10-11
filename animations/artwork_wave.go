package animations

import "math"

func (a *ArtworkAnimation) drawWave(t float64) {
	a.view(1000, 620, true)
	left := (1000 - float64(a.dotW)/a.sx) / 2
	right := 1000 - left
	phase := t * 2 * math.Pi / 8
	rise := 28 * math.Sin(phase)
	curl := 48 * math.Sin(phase+.7)
	breaking := 55 * math.Max(0, math.Sin(phase-.8))
	// The bay stays open around Fuji; the foreground swells move through it.
	a.poly(artShade, artPoint{left, 410}, artPoint{right, 350}, artPoint{right, 620}, artPoint{left, 620})
	for band := 0; band < 12; band++ {
		y := 365 + float64(band)*23
		a.curve([4]artPoint{{left, y + 20*math.Sin(phase+float64(band))}, {330, y - 35}, {650, y + 22*math.Sin(phase-float64(band)*.4)}, {right, y - 15}}, .8, artBlue)
	}
	a.poly(artSkinShade, artPoint{582, 372}, artPoint{666, 278}, artPoint{752, 372})
	a.poly(artLight, artPoint{636, 312}, artPoint{666, 278}, artPoint{688, 304}, artPoint{671, 303}, artPoint{663, 313}, artPoint{653, 306})
	a.line(artPoint{595, 373}, artPoint{749, 373}, 1, artMid)
	// Extend the offshore flank; the lip and Fuji retain their physical scale.
	edge := math.Min(-45, left-45)
	outer := [4]artPoint{{edge, 590}, {edge, 150}, {102, -16 + rise}, {300, 76 + rise}}
	lip := [4]artPoint{{300, 76 + rise}, {420, 53 + rise}, {529 + curl, 142 + rise + breaking*.6}, {484 + curl, 218 + rise + breaking}}
	inside := [4]artPoint{{484 + curl, 218 + rise + breaking}, {429 + curl, 128 + rise + breaking}, {342, 353}, {677, 620}}
	var shape [126]artPoint
	n := 0
	for _, curve := range [][4]artPoint{outer, lip, inside} {
		for i := 0; i <= 40; i++ {
			shape[n] = artBezier(curve, float64(i)/40)
			n++
		}
	}
	shape[n] = artPoint{edge, 620}
	n++
	a.poly(artBlue, shape[:n]...)
	// Ink grooves bend with the body and travel toward the breaking edge.
	for k := 0; k < 14; k++ {
		u := float64(k) / 14
		drift := 9 * math.Sin(phase-u*7)
		start := artPoint{edge + 20 + u*(375-edge), 635}
		end := artPoint{93 + u*325, 105 + rise + u*39}
		p := [4]artPoint{start, {8 + u*155, 365 + drift}, {5 + u*180, 170 + rise + drift}, end}
		a.curve(p, 1.3, artShade)
		if k%3 == 0 {
			a.curve([4]artPoint{{start.x + 10, start.y}, {p[1].x + 8, p[1].y}, {p[2].x + 8, p[2].y}, {end.x + 8, end.y}}, .8, artCyan)
		}
	}
	a.curve(outer, 18, artLight)
	a.curve(lip, 24, artWhite)
	// Foam veins fold back into the underside of the breaking lip.
	for k := 0; k < 10; k++ {
		u := .35 + float64(k)*.055
		root := artBezier(lip, u)
		a.curve([4]artPoint{root, {root.x - 12, root.y + 12}, {root.x - 35, root.y + 30 + breaking*.3}, {root.x - 26, root.y + 48}}, 2.3, artLight)
	}
	// Branching foam fingers articulate along the crest, with different phases.
	for k := 0; k < 15; k++ {
		u := .04 + float64(k)*.063
		root := artBezier(lip, u)
		theta := -1.45 + u*2.3 + .22*math.Sin(phase+float64(k)*.8)
		length := 23 + 11*math.Sin(phase+float64(k)*.7)
		tip := artPoint{root.x + math.Cos(theta)*length, root.y + math.Sin(theta)*length}
		a.curve([4]artPoint{root, {root.x + math.Cos(theta)*length*.5, root.y - 14}, {tip.x - 10, tip.y - 8}, tip}, 5, artWhite)
		a.line(tip, artPoint{tip.x + 10*math.Sin(phase+u*9), tip.y + 9}, 2, artLight)
		a.line(artPoint{(root.x + tip.x) / 2, (root.y + tip.y) / 2}, artPoint{tip.x - 12, tip.y - 5}, 2.5, artWhite)
	}
	// Spray follows ballistic arcs and returns to the crest at staggered ages.
	for k := 0; k < 70; k++ {
		age := artPhase(t+float64(k)*.157, 2)
		u := float64((k*37)%100) / 100
		root := artBezier(lip, u)
		x := root.x + (35+float64(k%11)*8)*age
		y := root.y - (90+float64(k%7)*14)*age + 190*age*age
		if y < root.y+25 {
			a.ellipse(x, y, 1.0+float64(k%3)*.35, 1.7, artWhite)
		}
	}
	// Foreground rollers spread into the bay with travelling white caps.
	for k := 0; k < 3; k++ {
		y := 514 + float64(k)*49
		p := [4]artPoint{{330, y + 50}, {545, y - 80 + 18*math.Sin(phase+float64(k))}, {780, y + 40}, {math.Max(1060, right), y - 30}}
		a.curve(p, 7, artLight)
		a.curve([4]artPoint{{p[0].x, p[0].y + 14}, {p[1].x, p[1].y + 14}, {p[2].x, p[2].y + 14}, {p[3].x, p[3].y + 14}}, 2, artShade)
	}
	// Long fishing boats sit beneath the crest, rising on the passing rollers.
	for k := 0; k < 3; k++ {
		x := 451 + float64(k)*164
		y := 432 + float64(k%2)*62 + 8*math.Sin(phase+float64(k)*1.4)
		a.poly(artWarmShade, artPoint{x - 100, y}, artPoint{x + 98, y - 12}, artPoint{x + 74, y + 14}, artPoint{x - 70, y + 25})
		a.curve([4]artPoint{{x - 101, y - 2}, {x - 35, y + 8}, {x + 47, y - 3}, {x + 99, y - 13}}, 2, artLight)
		a.line(artPoint{x - 71, y + 23}, artPoint{x + 75, y + 12}, 2, artInk)
		for rower := 0; rower < 7; rower++ {
			xx := x - 61 + float64(rower)*19
			yy := y + 2 - float64(rower)*1.2
			a.ellipse(xx, yy, 3, 3, artLight)
			a.line(artPoint{xx, yy + 4}, artPoint{xx + 6, yy + 9}, 2, artInk)
			a.line(artPoint{xx + 4, yy + 7}, artPoint{xx + 20, yy + 22}, .7, artWarm)
		}
	}
	// Small streaks give water flow a direction between the larger forms.
	for k := 0; k < 40; k++ {
		x := 570 + float64((k*47)%490) - 45*math.Sin(phase+float64(k)*.4)
		y := 391 + float64((k*23)%190)
		a.line(artPoint{x, y}, artPoint{x + 8 + float64(k%8), y - 2}, .5, artCyan)
	}
}
