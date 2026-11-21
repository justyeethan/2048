package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type state int

const (
	playing state = iota
	won
	lost
)

type model struct {
	game        *Game
	prev        *Game // snapshot for a single undo
	best        int
	state       state
	keepPlaying bool // player chose to continue past 2048
	width       int
	height      int
	rng         *rand.Rand
}

func newModel() model {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return model{game: NewGame(rng), best: loadBest(), rng: rng}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			saveBest(m.best)
			return m, tea.Quit
		case "r":
			m.game = NewGame(m.rng)
			m.prev = nil
			m.state = playing
			m.keepPlaying = false
			return m, nil
		case "u":
			if m.prev != nil {
				m.game = m.prev
				m.prev = nil
				m.state = playing
			}
			return m, nil
		case "c":
			if m.state == won {
				m.state = playing
				m.keepPlaying = true
			}
			return m, nil
		}

		if m.state != playing {
			return m, nil
		}

		dir, ok := keyDirection(msg.String())
		if !ok {
			return m, nil
		}
		snapshot := m.game.Clone()
		if !m.game.Move(dir) {
			return m, nil
		}
		m.prev = snapshot
		if m.game.Score > m.best {
			m.best = m.game.Score
			saveBest(m.best)
		}
		switch {
		case !m.keepPlaying && m.game.MaxTile() >= WinTile:
			m.state = won
		case !m.game.CanMove():
			m.state = lost
		}
	}
	return m, nil
}

func keyDirection(k string) (Direction, bool) {
	switch k {
	case "up", "k", "w":
		return Up, true
	case "down", "j", "s":
		return Down, true
	case "left", "h", "a":
		return Left, true
	case "right", "l", "d":
		return Right, true
	}
	return 0, false
}

// ---------- best score persistence ----------

func bestPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "term2048", "best")
}

func loadBest() int {
	p := bestPath()
	if p == "" {
		return 0
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return 0
	}
	n, _ := strconv.Atoi(strings.TrimSpace(string(b)))
	return n
}

func saveBest(n int) {
	p := bestPath()
	if p == "" {
		return
	}
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	_ = os.WriteFile(p, []byte(strconv.Itoa(n)), 0o644)
}

func main() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
