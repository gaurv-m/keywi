package sound

import (
	"context"
	_ "embed"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

//go:embed assets/shatter.wav
var shatter []byte

type Player struct {
	command string
	quiet   bool
	closed  bool
	mu      sync.Mutex
	ctx     context.Context
	cancel  context.CancelFunc
}

func New() *Player {
	ctx, cancel := context.WithCancel(context.Background())
	p := &Player{ctx: ctx, cancel: cancel}
	commands := []string{"paplay", "aplay"}
	if runtime.GOOS == "darwin" {
		commands = []string{"afplay"}
	}
	for _, name := range commands {
		if path, err := exec.LookPath(name); err == nil {
			p.command, p.quiet = path, name == "aplay"
			break
		}
	}
	return p
}

func (p *Player) Available() bool { return p != nil && p.command != "" }
func (p *Player) Close() {
	p.cancel()
	p.mu.Lock()
	p.closed = true
	p.mu.Unlock()
}

// Play drops overlapping effects and never writes to the terminal.
func (p *Player) Play() {
	if !p.Available() || !p.mu.TryLock() {
		return
	}
	defer p.mu.Unlock()
	if p.closed {
		return
	}
	ctx, cancel := context.WithTimeout(p.ctx, 3*time.Second)
	defer cancel()
	if ctx.Err() != nil {
		return
	}
	f, err := os.CreateTemp("", "keywi-shatter-*.wav")
	if err != nil {
		return
	}
	defer os.Remove(f.Name())
	_, err = f.Write(shatter)
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		return
	}
	args := []string{f.Name()}
	if p.quiet {
		args = append([]string{"-q"}, args...)
	}
	_ = exec.CommandContext(ctx, p.command, args...).Run()
}
