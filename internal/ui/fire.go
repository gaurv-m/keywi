package ui

import "math"

// Heat rises through swaying tongues, with gold cores and red tips.
func fireShade(x, y, height, frame, count int) int {
	depth := float64(y) / float64(max(1, height-1))
	t := float64(frame)
	sway := 1.3 * math.Sin(t*0.22+float64(y)*0.55)
	flow := fireNoise((float64(x)+sway)*0.19, float64(y)*0.38+t*0.24)
	sparks := fireNoise(float64(x)*0.53, float64(y)*0.81+t*0.52)
	heat := 0.22 + depth*0.60 + (flow-0.5)*0.58 + (sparks-0.5)*0.16
	if height == 1 {
		heat += 0.3
	}
	return int(math.Round(min(1, max(0, heat)) * float64(max(0, count-1))))
}

func fireNoise(x, y float64) float64 {
	x0, y0 := int(math.Floor(x)), int(math.Floor(y))
	u, v := soften(x-float64(x0)), soften(y-float64(y0))
	sample := func(x, y int) float64 {
		return float64(titleHash(x, y, 17)&65535) / 65535
	}
	top := mix(sample(x0, y0), sample(x0+1, y0), u)
	bottom := mix(sample(x0, y0+1), sample(x0+1, y0+1), u)
	return mix(top, bottom, v)
}
