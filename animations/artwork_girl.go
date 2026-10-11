package animations

import "math"

func (a *ArtworkAnimation) drawGirl(t float64) {
	a.view(675, 336, true)
	// The camera keeps the portrait full-height without stretching the face.
	sway := 2.4 * math.Sin(t*2*math.Pi/12)
	breathe := 1.3 * math.Sin(t*2*math.Pi/4)
	a.ox += sway * a.sx
	a.oy += breathe * a.sy
	// Hair and neck run past the screen edges, retaining the reference close-up.
	a.poly(artInk, artPoint{212, -20}, artPoint{455, -20}, artPoint{482, 110}, artPoint{461, 235}, artPoint{415, 360}, artPoint{260, 360}, artPoint{209, 218}, artPoint{194, 98})
	a.poly(artSkinShade, artPoint{224, 134}, artPoint{445, 130}, artPoint{448, 219}, artPoint{405, 320}, artPoint{377, 350}, artPoint{285, 350}, artPoint{245, 285}, artPoint{221, 213})
	a.poly(artSkin, artPoint{245, 132}, artPoint{432, 132}, artPoint{426, 216}, artPoint{386, 307}, artPoint{361, 333}, artPoint{306, 333}, artPoint{272, 293}, artPoint{249, 234})
	// Broad faceted cheeks and the nasal ridge retain their edges in block cells.
	a.poly(artLight, artPoint{243, 138}, artPoint{284, 132}, artPoint{302, 163}, artPoint{280, 195}, artPoint{272, 239}, artPoint{252, 219})
	a.poly(artSkinShade, artPoint{283, 188}, artPoint{306, 165}, artPoint{317, 208}, artPoint{303, 244}, artPoint{279, 236})
	a.poly(artLight, artPoint{366, 149}, artPoint{428, 137}, artPoint{418, 192}, artPoint{388, 207}, artPoint{362, 187})
	a.poly(artSkinShade, artPoint{407, 204}, artPoint{431, 185}, artPoint{418, 257}, artPoint{384, 309}, artPoint{376, 287})
	a.poly(artLight, artPoint{307, 291}, artPoint{367, 289}, artPoint{363, 319}, artPoint{328, 328}, artPoint{308, 315})
	a.poly(artSkinShade, artPoint{224, 144}, artPoint{237, 143}, artPoint{236, 195}, artPoint{223, 213}, artPoint{214, 182})
	a.line(artPoint{229, 152}, artPoint{226, 188}, 2, artLight)
	// Small soft light planes drift across the cheeks while the head moves.
	light := math.Sin(t * 2 * math.Pi / 6)
	a.poly(artLight, artPoint{392 + light*6, 170}, artPoint{410, 154}, artPoint{411, 181}, artPoint{393 + light*6, 191})
	a.poly(artSkin, artPoint{276 + light*3, 203}, artPoint{291, 199}, artPoint{289, 220}, artPoint{276, 231})
	// Nose, nostrils and the injury remain distinct from the visor.
	a.poly(artLight, artPoint{328, 141}, artPoint{344, 141}, artPoint{358, 184}, artPoint{348, 196}, artPoint{322, 195}, artPoint{314, 184})
	a.poly(artSkinShade, artPoint{343, 153}, artPoint{359, 181}, artPoint{351, 195}, artPoint{339, 189})
	a.poly(artInk, artPoint{315, 182}, artPoint{325, 182}, artPoint{329, 189}, artPoint{320, 192})
	a.poly(artInk, artPoint{345, 185}, artPoint{357, 181}, artPoint{355, 192}, artPoint{346, 195})
	a.line(artPoint{336, 170}, artPoint{333, 205}, 3, artBloodShade)
	a.poly(artBlood, artPoint{331, 184}, artPoint{335, 185}, artPoint{333, 220}, artPoint{328, 227}, artPoint{328, 214})
	// Blood accumulates, forms a narrow thread and detaches into a falling drop.
	age := artPhase(t, 3)
	thread := 8 + 24*math.Min(age/.58, 1)
	a.line(artPoint{332, 209}, artPoint{331 + math.Sin(t*2*math.Pi/3)*.8, 209 + thread}, 1.6, artBlood)
	if age < .65 {
		a.ellipse(331, 209+thread, 2.2, 2.5+age*2, artBlood)
	} else {
		fall := (age - .65) / .35
		a.ellipse(331+fall*2, 241+120*fall*fall, 1.7, 3, artBlood)
	}
	a.line(artPoint{349, 193}, artPoint{346, 220 + 9*math.Sin(t*2*math.Pi/3)}, 1.2, artBloodShade)
	// Lips part with the breathing cycle instead of holding a still face.
	mouth := 1.5 + .8*math.Sin(t*2*math.Pi/4+.8)
	a.poly(artSkinShade, artPoint{305, 238}, artPoint{321, 227}, artPoint{336, 232}, artPoint{350, 228}, artPoint{370, 242}, artPoint{354, 264}, artPoint{319, 263})
	a.poly(artInk, artPoint{304, 245}, artPoint{324, 237}, artPoint{344, 238}, artPoint{370, 245}, artPoint{353, 249 + mouth}, artPoint{325, 252 + mouth})
	a.poly(artViolet, artPoint{308, 244}, artPoint{326, 233}, artPoint{335, 238}, artPoint{349, 235}, artPoint{364, 243}, artPoint{344, 244}, artPoint{326, 242})
	a.poly(artViolet, artPoint{312, 251 + mouth}, artPoint{333, 253 + mouth}, artPoint{361, 249 + mouth}, artPoint{349, 260}, artPoint{327, 262})
	a.line(artPoint{324, 258}, artPoint{345, 258}, 1.7, artLight)
	a.line(artPoint{320, 275}, artPoint{351, 276}, 1, artSkinShade)
	// The visor's footprint stays attached to the face. Patterns scan through it.
	a.poly(artInk, artPoint{209, 34}, artPoint{451, 34}, artPoint{460, 143}, artPoint{373, 148}, artPoint{345, 139}, artPoint{324, 139}, artPoint{290, 150}, artPoint{210, 147})
	a.poly(artShade, artPoint{217, 46}, artPoint{443, 46}, artPoint{449, 127}, artPoint{375, 131}, artPoint{345, 122}, artPoint{322, 123}, artPoint{282, 136}, artPoint{217, 133})
	for k := 0; k < 35; k++ {
		band := k % 9
		y := 52 + float64(band)*8.7
		speed := 221 * float64(k%4+1) / 12
		x := 218 + math.Mod(float64(k*47)+t*speed, 221)
		w := 6 + float64((k*13)%31)
		w = math.Min(w, 444-x)
		c := uint8(artBlue)
		if k%4 == 0 {
			c = artCyan
		}
		if k%7 == 0 {
			c = artViolet
		}
		if k%11 == 0 {
			c = artWarm
		}
		a.rect(x, y, w, 3+float64(k%3)*2, c)
		if k%3 == 0 {
			a.rect(x+2, y+6, math.Min(w, 9), 1, artLight)
		}
	}
	scan := 52 + 81*artPhase(t, 2)
	a.rect(218, scan, 226, 1.1, artCyan)
	// Tiny display fragments change location and length; brief glitches shift rows.
	for k := 0; k < 7; k++ {
		p := artPhase(t+float64(k)*.23, 1.5)
		x := 231 + float64((k*31)%169)
		a.rect(x, 64+float64(k)*9, 4+18*p, 1, artLight)
	}
	a.line(artPoint{217, 144}, artPoint{281, 148}, 2, artBlue)
	a.line(artPoint{380, 145}, artPoint{454, 141}, 2, artViolet)
	// Side fragments occupy the wide screen without widening the subject.
	for k := 0; k < 12; k++ {
		age := artPhase(t+float64(k)*.6, 4)
		x := 190 - float64(k%3)*22 - 16*age
		if k%2 == 0 {
			x = 472 + float64(k%3)*22 + 16*age
		}
		y := 29 + float64(k*19%104)
		if age < .6 {
			a.rect(x, y, 3+float64(k%4)*3, 1+float64(k%3), artBlue)
		}
	}
}
