package animations

import "math"

func (a *ArtworkAnimation) drawWasteland(t float64) {
	a.view(960, 540, true)
	left := (960 - float64(a.dotW)/a.sx) / 2
	right := 960 - left
	phase := t * 2 * math.Pi / 16
	// Flat colour planes establish depth before any animated detail is drawn.
	a.rect(left, 0, right-left, 330, artWarmShade)
	a.rect(left, 65, right-left, 225, artWarm)
	a.rect(left, 206, right-left, 84, artViolet)
	a.ellipse(232, 145, 48, 48, artSun)
	// Cloud bands drift independently; their irregular edges remain readable.
	for k := 0; k < 15; k++ {
		x := float64((k*193)%1130) - 110 + 34*math.Sin(phase+float64(k)*.71)
		y := 15 + float64((k*29)%130)
		w := 92 + float64((k*17)%95)
		a.poly(artWarmShade, artPoint{x, y + 21}, artPoint{x + 14, y + 5}, artPoint{x + 37, y + 9}, artPoint{x + 48, y - 7}, artPoint{x + 76, y - 11}, artPoint{x + 94, y + 3}, artPoint{x + w, y}, artPoint{x + w + 21, y + 19}, artPoint{x + w - 18, y + 26}, artPoint{x + 41, y + 23})
		a.poly(artSkinShade, artPoint{x + 31, y + 19}, artPoint{x + 53, y + 11}, artPoint{x + 73, y + 18}, artPoint{x + w - 17, y + 17}, artPoint{x + w - 28, y + 24}, artPoint{x + 48, y + 23})
		if k%3 == 0 {
			a.line(artPoint{x + 48, y + 25}, artPoint{x + w - 35, y + 25}, .5, artSkin)
		}
	}
	a.rect(180, 139, 119, 4, artWarm)
	a.rect(202, 155, 103, 3, artWarm)
	// Ruined skyline, deliberately small and hazy behind the near ridges.
	for k := 0; k < 9; k++ {
		x := 101 + float64(k)*16
		h := 18 + float64((k*37)%92)
		a.poly(artWarmShade, artPoint{x, 281}, artPoint{x - 3, 266 - h}, artPoint{x + 3, 259 - h}, artPoint{x + 7, 269 - h}, artPoint{x + 9, 264 - h}, artPoint{x + 12, 277 - h}, artPoint{x + 18, 280})
		for floor := 0; floor < 4; floor++ {
			a.rect(x+3, 279-h+float64(floor)*12, 3, 2, artSkinShade)
		}
	}
	a.poly(artWarmShade, artPoint{left, 299}, artPoint{119, 281}, artPoint{226, 283}, artPoint{369, 310}, artPoint{430, 266}, artPoint{476, 298}, artPoint{506, 279}, artPoint{561, 307}, artPoint{667, 282}, artPoint{800, 315}, artPoint{right, 284}, artPoint{right, 540}, artPoint{left, 540})
	a.poly(artBloodShade, artPoint{left, 355}, artPoint{124, 340}, artPoint{256, 357}, artPoint{398, 342}, artPoint{438, 321}, artPoint{500, 343}, artPoint{548, 320}, artPoint{667, 352}, artPoint{778, 326}, artPoint{right, 347}, artPoint{right, 540}, artPoint{left, 540})
	// Broken planes and slanted beams interrupt the far ridges.
	a.poly(artBloodShade, artPoint{413, 310}, artPoint{417, 249}, artPoint{454, 298}, artPoint{457, 279}, artPoint{478, 309})
	a.poly(artWarmShade, artPoint{419, 260}, artPoint{429, 269}, artPoint{434, 296}, artPoint{426, 290})
	a.line(artPoint{576, 320}, artPoint{599, 297}, 2, artBloodShade)
	a.line(artPoint{599, 297}, artPoint{623, 315}, 2, artBloodShade)
	// The vehicle slows and turns around; its trail follows its direction.
	vx := 495 - 137*math.Cos(phase)
	speed := math.Sin(phase)
	direction := math.Copysign(1, speed)
	a.rect(vx, 316, 10, 3, artInk)
	a.rect(vx+3, 312, 5, 4, artInk)
	a.ellipse(vx+2, 320, 1.5, 1.5, artShade)
	a.ellipse(vx+8, 320, 1.5, 1.5, artShade)
	for k := 0; k < int(7*math.Abs(speed)); k++ {
		x := vx + 5 - direction*(10+float64(k)*6)
		a.line(artPoint{x, 317 + float64(k%2)*2}, artPoint{x + direction*4, 318}, .5, artWarm)
	}
	a.line(artPoint{550, 280}, artPoint{554, 334}, 3, artInk)
	a.line(artPoint{552, 296}, artPoint{563, 308}, 2, artInk)
	// The reflective channel carries travelling horizontal ripples.
	a.poly(artWarmShade, artPoint{140, 392}, artPoint{625, 389}, artPoint{664, 483}, artPoint{225, 497})
	for k := 0; k < 34; k++ {
		y := 397 + float64(k)*3
		x := 173 + float64((k*61)%375) + 14*math.Sin(phase*2+float64(k))
		a.line(artPoint{x, y}, artPoint{x + 18 + float64(k%9)*5, y}, .7, artWarm)
	}
	a.poly(artShade, artPoint{217, 398}, artPoint{262, 397}, artPoint{286, 510}, artPoint{216, 510})
	// Ripples cut the figure's reflection; they do not erase the real silhouette.
	for k := 0; k < 15; k++ {
		y := 406 + float64(k)*7
		shift := 6 * math.Sin(phase*3+float64(k)*.7)
		a.rect(216+shift, y, 76, 1.6, artWarmShade)
	}
	a.poly(artSkinShade, artPoint{401, 361}, artPoint{442, 344}, artPoint{479, 353}, artPoint{489, 378}, artPoint{415, 384})
	a.line(artPoint{402, 361}, artPoint{442, 344}, 2, artWarm)
	a.poly(artInk, artPoint{418, 366}, artPoint{450, 356}, artPoint{476, 363}, artPoint{484, 378}, artPoint{418, 383})
	// Foreground banks and angular wreckage frame the figure.
	a.poly(artInk, artPoint{left, 153}, artPoint{0, 153}, artPoint{35, 175}, artPoint{48, 203}, artPoint{88, 214}, artPoint{64, 234}, artPoint{106, 256}, artPoint{76, 279}, artPoint{106, 321}, artPoint{140, 363}, artPoint{112, 408}, artPoint{171, 449}, artPoint{221, 459}, artPoint{240, 508}, artPoint{341, 520}, artPoint{388, 540}, artPoint{left, 540})
	a.poly(artInk, artPoint{805, 0}, artPoint{866, 0}, artPoint{917, 77}, artPoint{945, 65}, artPoint{960, 540}, artPoint{674, 540}, artPoint{721, 488}, artPoint{693, 464}, artPoint{660, 441}, artPoint{633, 422}, artPoint{590, 418}, artPoint{670, 380}, artPoint{733, 365}, artPoint{741, 304}, artPoint{699, 256}, artPoint{748, 277}, artPoint{751, 211}, artPoint{803, 193}, artPoint{809, 129}, artPoint{777, 117}, artPoint{817, 78})
	a.rect(960, 0, math.Max(0, right-960), 540, artInk)
	a.poly(artMid, artPoint{815, 5}, artPoint{856, -8}, artPoint{893, 48}, artPoint{858, 66})
	a.poly(artInk, artPoint{821, 11}, artPoint{849, 4}, artPoint{877, 44}, artPoint{856, 54})
	a.line(artPoint{815, 5}, artPoint{856, -8}, 1, artSkinShade)
	a.poly(artMid, artPoint{790, 156}, artPoint{831, 143}, artPoint{855, 215}, artPoint{818, 231})
	a.poly(artInk, artPoint{798, 164}, artPoint{825, 158}, artPoint{842, 211}, artPoint{822, 220})
	a.line(artPoint{791, 155}, artPoint{832, 143}, 1, artWarmShade)
	a.poly(artShade, artPoint{859, 90}, artPoint{906, 73}, artPoint{955, 90}, artPoint{944, 270}, artPoint{881, 254})
	for k := 0; k < 7; k++ {
		y := 102 + float64(k)*21
		a.line(artPoint{868, y}, artPoint{939, y + 7}, 2, artBloodShade)
		a.line(artPoint{873, y + 5}, artPoint{934, y + 11}, 1, artMid)
	}
	a.poly(artBloodShade, artPoint{765, 319}, artPoint{824, 279}, artPoint{868, 302}, artPoint{854, 376}, artPoint{780, 386}, artPoint{748, 349})
	a.poly(artInk, artPoint{770, 325}, artPoint{824, 288}, artPoint{860, 306}, artPoint{847, 369}, artPoint{783, 378}, artPoint{758, 349})
	a.curve([4]artPoint{{759, 347}, {767, 306}, {811, 291}, {845, 322}}, 3, artMid)
	a.curve([4]artPoint{{788, 374}, {812, 341}, {824, 344}, {848, 366}}, 3, artMid)
	a.poly(artMid, artPoint{861, 305}, artPoint{956, 280}, artPoint{968, 334}, artPoint{886, 356})
	a.poly(artInk, artPoint{870, 313}, artPoint{950, 293}, artPoint{956, 326}, artPoint{890, 345})
	for k := 0; k < 9; k++ {
		x := 889 + float64(k)*7
		a.line(artPoint{x, 307 - float64(k)*1.7}, artPoint{x + 6, 337 - float64(k)*1.7}, 1, artBloodShade)
	}
	for k := 0; k < 5; k++ {
		x := 844 + float64(k)*19
		y := 448 - float64(k)*11
		a.curve([4]artPoint{{x - 69, y + 21}, {x - 7, y - 54}, {x + 51, y - 23}, {x + 52, y + 34}}, 4, artMid)
		a.curve([4]artPoint{{x - 68, y + 24}, {x - 6, y - 48}, {x + 45, y - 18}, {x + 47, y + 33}}, 1.2, artBloodShade)
	}
	// Rivets, torn braces and scratches keep the foreground readable up close.
	for k := 0; k < 18; k++ {
		x := 805 + float64((k*37)%139)
		y := 121 + float64((k*59)%340)
		a.ellipse(x, y, 1.5, 1.5, artWarmShade)
	}
	a.line(artPoint{790, 214}, artPoint{821, 306}, 3, artBloodShade)
	a.line(artPoint{855, 215}, artPoint{816, 290}, 3, artBloodShade)
	a.poly(artMid, artPoint{744, 235}, artPoint{761, 227}, artPoint{783, 268}, artPoint{766, 277})
	a.line(artPoint{744, 235}, artPoint{761, 227}, 1, artWarmShade)
	a.curve([4]artPoint{{904, 435}, {875, 421}, {852, 485}, {803, 484}}, 3, artMid)
	a.curve([4]artPoint{{719, 414}, {774, 369}, {767, 278}, {810, 256}}, 4, artBloodShade)
	a.curve([4]artPoint{{732, 422}, {787, 381}, {784, 290}, {825, 269}}, 1, artWarmShade)
	a.line(artPoint{808, 138}, artPoint{853, 50}, 5, artBloodShade)
	a.line(artPoint{47, 207}, artPoint{38, 131}, 3, artInk)
	a.curve([4]artPoint{{38, 140}, {41, 113}, {61, 100}, {65, 105}}, 4, artInk)
	a.curve([4]artPoint{{102, 356}, {98, 320}, {131, 326}, {143, 302}}, 2, artInk)
	// Continue broken beams and curved pipework through the ultrawide bank.
	if right > 980 {
		a.poly(artShade, artPoint{932, 67}, artPoint{right + 30, 112}, artPoint{right + 20, 154}, artPoint{939, 108})
		a.line(artPoint{932, 67}, artPoint{right + 30, 112}, 2, artMid)
		a.poly(artMid, artPoint{950, 230}, artPoint{right + 20, 163}, artPoint{right + 20, 181}, artPoint{958, 246})
		a.line(artPoint{962, 235}, artPoint{right + 20, 173}, 4, artInk)
		a.poly(artBloodShade, artPoint{966, 272}, artPoint{right + 10, 300}, artPoint{right + 10, 370}, artPoint{978, 338})
		for k := 0; k < 5; k++ {
			y := 284 + float64(k)*11
			a.line(artPoint{985, y}, artPoint{right + 10, y + 28}, 2, artInk)
		}
		for k := 0; k < 3; k++ {
			y := 392 + float64(k)*22
			a.curve([4]artPoint{{942, y + 60}, {1006, y - 32}, {right - 26, y - 25}, {right + 10, y + 72}}, 4, artMid)
			a.curve([4]artPoint{{948, y + 60}, {1010, y - 25}, {right - 31, y - 18}, {right + 7, y + 74}}, 1, artWarmShade)
		}
		for k := 0; k < 12; k++ {
			x := 981 + float64(k%4)*(right-981)/4
			y := 133 + float64(k/4)*122
			a.ellipse(x, y, 1.5, 1.5, artWarmShade)
		}
	}
	// The lone figure breathes and shifts the hanging fabric in the wind.
	wind := 2 * math.Sin(phase*4)
	a.poly(artInk, artPoint{224, 316}, artPoint{236, 302}, artPoint{250, 303}, artPoint{258, 334}, artPoint{254 + wind, 369}, artPoint{224, 370}, artPoint{221, 340})
	a.ellipse(240, 298, 8, 10, artShade)
	a.rect(240, 294, 8, 3, artCyan)
	a.line(artPoint{241, 304}, artPoint{249, 312}, 2, artBlue)
	a.line(artPoint{231, 366}, artPoint{228 + wind, 397}, 3, artInk)
	a.line(artPoint{249, 366}, artPoint{251 - wind, 398}, 3, artInk)
	a.line(artPoint{221, 334}, artPoint{216 + wind, 361}, 3, artInk)
	a.line(artPoint{240, 320}, artPoint{236 + wind, 350}, 1, artShade)
	// Windblown particles cross different depth planes at different speeds.
	for k := 0; k < 40; k++ {
		depth := float64(k%3 + 1)
		age := artPhase(t+float64(k)*.41, 16/depth)
		x := 110 + age*700
		y := 220 + float64((k*37)%220) - 18*math.Sin(age*math.Pi)
		c := uint8(artWarmShade)
		if k%3 == 0 {
			c = artInk
		}
		a.line(artPoint{x, y}, artPoint{x + depth*2, y - 1}, .5, c)
	}
	// Birds travel in the far sky with wings, rather than pixel twinkles.
	for k := 0; k < 3; k++ {
		age := artPhase(t+float64(k)*.6, 16)
		x := -45 + age*1050
		y := 214 + float64(k)*10 + 7*math.Sin(phase)
		flap := 2.5 * math.Sin(t*math.Pi*2+float64(k))
		a.line(artPoint{x - 3, y + flap}, artPoint{x, y}, .6, artBloodShade)
		a.line(artPoint{x, y}, artPoint{x + 3, y + flap}, .6, artBloodShade)
	}
}
