package graph

// Smooth averages interval rates by elapsed time, without changing the source.
func Smooth(points []Point, window float64) []Point {
	var data []Point
	previous := 0.0
	for _, p := range points {
		if !finite(p.X) || !finite(p.Y) || p.X <= previous {
			continue
		}
		p.Y = max(0, p.Y)
		data = append(data, p)
		previous = p.X
	}
	if window <= 0 || !finite(window) {
		return data
	}
	output := make([]Point, 0, len(data))
	area, previous, left := 0.0, 0.0, 0
	for i, p := range data {
		area += p.Y * (p.X - previous)
		previous = p.X
		start := max(0, p.X-window)
		for left < i && data[left].X <= start {
			beginning := 0.0
			if left > 0 {
				beginning = data[left-1].X
			}
			area -= data[left].Y * (data[left].X - beginning)
			left++
		}
		beginning := 0.0
		if left > 0 {
			beginning = data[left-1].X
		}
		partial := data[left].Y * (start - beginning)
		output = append(output, Point{p.X, max(0, (area-partial)/(p.X-start))})
	}
	return output
}
