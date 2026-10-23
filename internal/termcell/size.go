// Package termcell applies terminal cell geometry to effects that support it.
package termcell

func applySize(effect any, cols, rows, pixelWidth, pixelHeight int) {
	sized, ok := effect.(interface{ SetCellAspect(int, int) })
	if !ok || cols <= 0 || rows <= 0 || pixelWidth <= 0 || pixelHeight <= 0 {
		return
	}
	w, h := (pixelWidth+cols/2)/cols, (pixelHeight+rows/2)/rows
	if w > 0 && h > 0 && w <= 1024 && h <= 1024 {
		sized.SetCellAspect(w, h)
	}
}
