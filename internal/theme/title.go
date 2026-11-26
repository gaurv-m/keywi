package theme

import (
	"fmt"
	"math"

	"github.com/charmbracelet/lipgloss"
)

// Build the small shade tables once, rather than mixing RGB colours per glyph.
var titleShades = map[string][]lipgloss.Style{
	"kiwi":     titleRamp([]uint32{0x7FA45E, 0xA4C779, 0xC8DEA5, 0xEBF2D5}, false),
	"contrast": titleRamp([]uint32{0xD0A700, 0xFFDB00, 0xFFFF66, 0xFFFFEE}, true),
	"rainbow":  titleRamp([]uint32{0xF28B9B, 0xF2B880, 0xE7D98B, 0xA9D99B, 0x89CDDD, 0xB4A5E8, 0xF28B9B}, false),
	"fall":     titleRamp([]uint32{0xC47750, 0xEC955C, 0xF4BE79, 0xFFE1AB}, false),
	"fire":     titleRamp([]uint32{0x9C2820, 0xD83D22, 0xFA6A21, 0xFFAA2B, 0xFFD65A, 0xFFF1BB}, false),
	"tide":     titleRamp([]uint32{0x709ED8, 0x70BACE, 0x83D3CA, 0xBCEBD5}, false),
	"orchid":   titleRamp([]uint32{0xA787CD, 0xC3A2E3, 0xDCAEDC, 0xF2CCD9}, false),
}

func titleRamp(stops []uint32, bold bool) []lipgloss.Style {
	const count = 24
	shades := make([]lipgloss.Style, count)
	for i := range shades {
		position := float64(i) * float64(len(stops)-1) / (count - 1)
		left := int(position)
		right := min(left+1, len(stops)-1)
		weight := position - float64(left)
		channel := func(shift uint) int {
			a := float64((stops[left] >> shift) & 255)
			b := float64((stops[right] >> shift) & 255)
			return int(math.Round(a + (b-a)*weight))
		}
		color := fmt.Sprintf("#%02X%02X%02X", channel(16), channel(8), channel(0))
		shades[i] = lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(bold)
	}
	return shades
}
