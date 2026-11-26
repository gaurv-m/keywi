// Keywi is an entirely offline terminal typing game.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gaurv-m/keywi/internal/app"
	"github.com/gaurv-m/keywi/internal/sound"
	"github.com/gaurv-m/keywi/internal/storage"
	"github.com/gaurv-m/keywi/internal/theme"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"
)

var version = "dev"
var commit = "none"
var buildID = "source"

func run() error {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Printf("keywi %s (%s) · build %s\n", version, commit, buildID)
		return nil
	}
	if !isatty.IsTerminal(os.Stdin.Fd()) || !isatty.IsTerminal(os.Stdout.Fd()) {
		return fmt.Errorf("keywi requires an interactive terminal (use --version for build information)")
	}
	theme.ConfigureRenderer()
	s, warnings, unlock, err := storage.OpenLocked()
	if err != nil {
		return err
	}
	defer unlock()
	audio := sound.New()
	defer audio.Close()
	// Restore the terminal on exit.
	model := app.New(s, warnings)
	model.Audio = audio
	p := tea.NewProgram(model, tea.WithAltScreen())
	final, err := p.Run()
	for _, w := range warnings {
		fmt.Fprintln(os.Stderr, w)
	}
	if err != nil {
		return err
	}
	if model, ok := final.(app.Model); ok {
		return model.PersistenceError()
	}
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "keywi:", err)
		os.Exit(1)
	}
}
