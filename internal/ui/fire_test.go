package ui

import (
	"testing"

	"github.com/gaurv-m/keywi/assets"
	"github.com/gaurv-m/keywi/internal/theme"
)

func TestFireHasHotCoresAndActiveFlicker(t *testing.T) {
	top, bottom, changes := 0, 0, 0
	shades := map[int]bool{}
	for frame := 0; frame < 60; frame++ {
		for x := 0; x < 42; x++ {
			top += fireShade(x, 0, 9, frame, 24)
			bottom += fireShade(x, 8, 9, frame, 24)
			for y := 0; y < 9; y++ {
				shade := fireShade(x, y, 9, frame, 24)
				if shade < 0 || shade >= 24 {
					t.Fatal("Fire escaped its palette")
				}
				shades[shade] = true
				if shade != fireShade(x, y, 9, frame+1, 24) {
					changes++
				}
			}
		}
	}
	if bottom-top < 42*60*8 {
		t.Fatal("Fire lost its hot base and cooler tips")
	}
	if len(shades) < 16 || changes < 42*9*6 {
		t.Fatal("Fire stopped flickering across its shades")
	}
}

func TestFireStaticStillUsesItsHeatPalette(t *testing.T) {
	logoColors(t)
	p := theme.New("fire")
	art := assets.Logo()
	static := Logo(art, p)
	if Plain(static) != art || static == p.Primary.Render(art) {
		t.Fatal("static Fire lost its palette or changed the artwork")
	}
	if static != LogoFrame(art, p, 0) || static == LogoFrame(art, p, 5) {
		t.Fatal("static and moving Fire are inconsistent")
	}
}
