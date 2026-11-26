package graph

import "math"

func raster(data []Point, cols, rows int, xmin, xmax, ymax float64) [][]rune {
	grid := make([][]rune, rows)
	for y := range grid {
		grid[y] = make([]rune, cols)
		for x := range grid[y] {
			grid[y][x] = '\u2800'
		}
	}
	pw, ph := cols*2, rows*4
	dot := func(x, y int) {
		if x < 0 || x >= pw || y < 0 || y >= ph {
			return
		}
		bits := [4][2]uint{{0, 3}, {1, 4}, {2, 5}, {6, 7}}
		grid[y/4][x/2] |= 1 << bits[y%4][x%2]
	}
	previousX, previousY := -1, -1
	for _, point := range data {
		x := int(math.Round((point.X - xmin) / (xmax - xmin) * float64(pw-1)))
		y := ph - 1 - int(math.Round(point.Y/ymax*float64(ph-1)))
		if previousX < 0 {
			dot(x, y)
		} else {
			dx, dy := x-previousX, y-previousY
			steps := max(abs(dx), abs(dy))
			for i := 0; i <= steps; i++ {
				if steps == 0 {
					dot(x, y)
				} else {
					dot(previousX+int(math.Round(float64(dx*i)/float64(steps))), previousY+int(math.Round(float64(dy*i)/float64(steps))))
				}
			}
		}
		previousX, previousY = x, y
	}
	return grid
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
