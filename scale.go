package main

func DownScale2(fs []float32, ncols int, nrows int) ([]float32, int, int) {
	tcols := ncols / 2
	trows := nrows / 2
	ts := make([]float32, tcols*trows)

	var rbase, tbase int
	for row := 0; row < trows; row++ {
		for col := 0; col < tcols; col++ {
			rcol := col * 2
			ts[tbase+col] = fs[rbase+rcol] + fs[rbase+rcol+1]
		}
		rbase += ncols
		for col := 0; col < tcols; col++ {
			rcol := col * 2
			ts[tbase+col] = (fs[rbase+rcol] + fs[rbase+rcol+1] + ts[tbase+col]) / 4
		}
		rbase += ncols
		tbase += tcols
	}

	return ts, tcols, trows
}

func DownScaleN(fs []float32, ncols int, nrows int, scaleDownFactor int) ([]float32, int, int) {
	tcols := ncols / scaleDownFactor
	trows := nrows / scaleDownFactor
	ts := make([]float32, tcols*trows)

	var rbase int
outer:
	for row := 0; row < nrows; row++ {
		if row/scaleDownFactor >= trows {
			continue outer
		}
	inner:
		for col := 0; col < ncols; col++ {
			if col/scaleDownFactor >= tcols {
				break inner
			}
			ts[(tcols*(row/scaleDownFactor))+(col/scaleDownFactor)] += fs[rbase+col] / float32(scaleDownFactor)
		}
		rbase += ncols
	}

	return ts, tcols, trows
}
