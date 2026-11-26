// Package theme defines foreground colors.
package theme

import (
	"os"
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
)

var names = [...]string{"kiwi", "contrast", "rainbow", "fall", "tide", "fire", "orchid"}

func Names() []string { return append([]string(nil), names[:]...) }
func Normalize(name string) string {
	name = strings.ToLower(name)
	if name == "ember" {
		return "fall"
	}
	for _, candidate := range names {
		if name == candidate {
			return name
		}
	}
	// Migrate the old default palette.
	return "kiwi"
}
func Label(name string) string {
	name = Normalize(name)
	return strings.ToUpper(name[:1]) + name[1:]
}
func Next(name string, delta int) string {
	name = Normalize(name)
	for i, candidate := range names {
		if candidate == name {
			return names[(i+delta%len(names)+len(names))%len(names)]
		}
	}
	return names[0]
}

type Palette struct {
	Primary, Muted, Bonus, Correct, Error, Secondary, LogoI lipgloss.Style
	Plain                                                   bool
	Flame                                                   bool
	Spectrum, LogoShades                                    []lipgloss.Style
}

// Accent applies a static rainbow to selected text without animation or backgrounds.
func (p Palette) Accent(text string) string {
	if len(p.Spectrum) == 0 {
		return p.Primary.Render(text)
	}
	var out strings.Builder
	i := 0
	for _, r := range text {
		if unicode.IsSpace(r) {
			out.WriteRune(r)
			continue
		}
		out.WriteString(p.Spectrum[i%len(p.Spectrum)].Render(string(r)))
		i++
	}
	return out.String()
}
func New(name string) Palette {
	name = Normalize(name)
	style := func(s string) lipgloss.Style { return lipgloss.NewStyle().Foreground(lipgloss.Color(s)) }
	_, noColor := os.LookupEnv("NO_COLOR")
	if noColor || os.Getenv("TERM") == "dumb" {
		primary := lipgloss.NewStyle().Bold(true)
		p := Palette{Primary: primary, Muted: lipgloss.NewStyle().Faint(true), Bonus: primary, Correct: lipgloss.NewStyle(), Error: primary, Secondary: lipgloss.NewStyle(), LogoI: primary, Plain: true}
		return p
	}
	p := Palette{Primary: style("#B7D58C"), Muted: style("#788479"), Bonus: style("#E1C38A"), Correct: style("#E1E8DA"), Error: style("#E58F80"), Secondary: style("#83BDB1"), LogoI: style("#B68A62")}
	p.LogoShades = titleShades["kiwi"]
	switch name {
	case "contrast":
		p.Primary = style("#FFFF00").Bold(true)
		p.Muted = style("#B8B8B8")
		p.Correct = style("#FFFFFF").Bold(true)
		p.Error = style("#FF657A")
		p.Bonus = style("#FFD75F")
		p.Secondary = style("#5FFFFF")
		p.LogoI = style("#D6A06B").Bold(true)
		p.LogoShades = titleShades["contrast"]
	case "rainbow":
		p.LogoShades = titleShades["rainbow"]
		p.Primary = style("#C4A7E7")
		p.Muted = style("#9393A8")
		p.Correct = style("#E8E5F0")
		p.Error = style("#FF7B93")
		p.Bonus = style("#F6D78B")
		p.Secondary = style("#89DCEB")
		for _, color := range []string{"#F28B9B", "#F2B880", "#E7D98B", "#A9D99B", "#89CDDD", "#B4A5E8"} {
			p.Spectrum = append(p.Spectrum, style(color))
		}
	case "fall":
		p.Primary = style("#F4B078")
		p.Muted = style("#A89385")
		p.Correct = style("#F7E5D5")
		p.Error = style("#F0778C")
		p.Bonus = style("#FFD18A")
		p.Secondary = style("#DF8F6A")
		p.LogoShades = titleShades["fall"]
	case "fire":
		p.Primary = style("#FFAC38")
		p.Muted = style("#A99082")
		p.Correct = style("#FFF0D0")
		p.Error = style("#FF6B81")
		p.Bonus = style("#FFE38A")
		p.Secondary = style("#F37B43")
		p.LogoI = style("#C4874F")
		p.LogoShades = titleShades["fire"]
		p.Flame = true
	case "tide":
		p.Primary = style("#7CC8D8")
		p.Muted = style("#79959E")
		p.Correct = style("#E0F3F1")
		p.Error = style("#F08B99")
		p.Bonus = style("#A5E2C3")
		p.Secondary = style("#80AEE8")
		p.LogoShades = titleShades["tide"]
	case "orchid":
		p.Primary = style("#C6A0E8")
		p.Muted = style("#9987AA")
		p.Correct = style("#F0E5FA")
		p.Error = style("#F18CAB")
		p.Bonus = style("#E9BAD3")
		p.Secondary = style("#9FADDF")
		p.LogoShades = titleShades["orchid"]
	}
	return p
}
