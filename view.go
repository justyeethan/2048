package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
)

// Terminal cells are roughly twice as tall as they are wide, so a tile that
// is 2N columns wide and N rows tall looks square. The view picks the
// largest size that fits the terminal.
type tileSize struct{ w, h int } // outer size, including the tile's border

var tileSizes = []tileSize{
	{14, 7}, // large
	{10, 5}, // medium
	{6, 3},  // compact
}

type box struct{ tl, tr, bl, br, h, v string }

var (
	lightBox   = box{"┌", "┐", "└", "┘", "─", "│"}
	thickBox   = box{"┏", "┓", "┗", "┛", "━", "┃"}
	roundedBox = box{"╭", "╮", "╰", "╯", "─", "│"}
)

var (
	colFrame = lipgloss.Color("#6c6c6c")
	colDim   = lipgloss.Color("#8a8a8a")
	colGold  = lipgloss.Color("#edc22e")
	colRed   = lipgloss.Color("#f65e3b")
	colEmpty = lipgloss.Color("#3f3f3f")

	tileColors = map[int]lipgloss.Color{
		2:    "#eee4da",
		4:    "#ede0c8",
		8:    "#f2b179",
		16:   "#f59563",
		32:   "#f67c5f",
		64:   "#f65e3b",
		128:  "#edcf72",
		256:  "#edcc61",
		512:  "#edc850",
		1024: "#edc53f",
		2048: "#edc22e",
	}
)

func tileColor(v int) lipgloss.Color {
	if v == 0 {
		return colEmpty
	}
	if c, ok := tileColors[v]; ok {
		return c
	}
	return "#b784f5" // beyond 2048
}

func paint(s string, c lipgloss.Color, bold bool) string {
	return lipgloss.NewStyle().Foreground(c).Bold(bold).Render(s)
}

func width(s string) int { return utf8.RuneCountInString(s) }

func center(s string, w int) string {
	pad := w - width(s)
	if pad <= 0 {
		return s
	}
	left := pad / 2
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", pad-left)
}

// label returns the text for a tile, shortened if it can't fit.
func label(v, w int) string {
	if v == 0 {
		return "·"
	}
	s := strconv.Itoa(v)
	if width(s) > w {
		s = strconv.Itoa(v/1024) + "k"
	}
	return s
}

// renderTile returns the tile as sz.h lines, each sz.w cells wide.
// Highlighted tiles (just merged or just spawned) get a heavy border.
func renderTile(v int, sz tileSize, highlight bool) []string {
	b := lightBox
	if highlight {
		b = thickBox
	}
	c := tileColor(v)
	iw, ih := sz.w-2, sz.h-2

	lines := make([]string, 0, sz.h)
	lines = append(lines, paint(b.tl+strings.Repeat(b.h, iw)+b.tr, c, highlight))
	side := paint(b.v, c, highlight)
	for i := 0; i < ih; i++ {
		content := strings.Repeat(" ", iw)
		if i == ih/2 {
			content = center(label(v, iw), iw)
		}
		lines = append(lines, side+paint(content, c, v != 0)+side)
	}
	lines = append(lines, paint(b.bl+strings.Repeat(b.h, iw)+b.br, c, highlight))
	return lines
}

// seg is a piece of text placed in a frame border: plain is used to measure,
// styled is what gets drawn.
type seg struct{ plain, styled string }

// fit returns the first candidate no wider than avail (or an empty seg).
func fit(avail int, cands ...seg) seg {
	for _, c := range cands {
		if width(c.plain) <= avail {
			return c
		}
	}
	return seg{}
}

// borderLine draws  ╭─ left ───────── right ─╮  with inner width `inner`.
func borderLine(lc, h, rc string, inner int, fc lipgloss.Color, left seg, right []seg) string {
	used := 2 // one h at each end
	l := ""
	if left.plain != "" {
		l = " " + left.styled + " "
		used += width(left.plain) + 2
	}
	r := ""
	if rs := fit(inner-used-3, right...); rs.plain != "" {
		r = " " + rs.styled + " "
		used += width(rs.plain) + 2
	}
	fill := inner - used
	if fill < 0 {
		fill = 0
	}
	return paint(lc+h, fc, false) + l + paint(strings.Repeat(h, fill), fc, false) + r + paint(h+rc, fc, false)
}

func (m model) pickSize() (tileSize, bool) {
	if m.width == 0 || m.height == 0 {
		return tileSizes[1], true
	}
	for _, s := range tileSizes {
		w := Size*s.w + 4 // frame (2) + 1 space padding each side
		h := Size*s.h + 2 // frame top and bottom
		if w <= m.width && h <= m.height {
			return s, true
		}
	}
	return tileSizes[len(tileSizes)-1], false
}

func (m model) View() string {
	sz, ok := m.pickSize()
	if !ok {
		need := fmt.Sprintf("Terminal too small: need %d×%d, have %d×%d",
			Size*sz.w+4, Size*sz.h+2, m.width, m.height)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			paint(need, colRed, true))
	}

	g := m.game
	inner := Size*sz.w + 2

	fc := colFrame
	switch m.state {
	case won:
		fc = colGold
	case lost:
		fc = colRed
	}

	score, best := strconv.Itoa(g.Score), strconv.Itoa(m.best)
	title := seg{"2048", paint("2048", colGold, true)}
	scores := []seg{
		{"score " + score + " · best " + best,
			paint("score ", colDim, false) + paint(score, "#ffffff", true) +
				paint(" · best ", colDim, false) + paint(best, colGold, true)},
		{score + " · " + best,
			paint(score, "#ffffff", true) + paint(" · ", colDim, false) + paint(best, colGold, true)},
		{score, paint(score, "#ffffff", true)},
	}

	var footer []seg
	switch m.state {
	case won:
		footer = []seg{
			{"★ 2048! c keep going · r new game", paint("★ 2048!", colGold, true) + paint(" c keep going · r new game", colDim, false)},
			{"★ 2048! c / r", paint("★ 2048!", colGold, true) + paint(" c / r", colDim, false)},
			{"★ win", paint("★ win", colGold, true)},
		}
	case lost:
		footer = []seg{
			{"game over · u undo · r new game", paint("game over", colRed, true) + paint(" · u undo · r new game", colDim, false)},
			{"game over · u · r", paint("game over", colRed, true) + paint(" · u · r", colDim, false)},
			{"game over", paint("game over", colRed, true)},
		}
	default:
		footer = []seg{
			{"←↑↓→ move · u undo · r restart · q quit", paint("←↑↓→ move · u undo · r restart · q quit", colDim, false)},
			{"←↑↓→ · u undo · r new · q quit", paint("←↑↓→ · u undo · r new · q quit", colDim, false)},
			{"u undo · q quit", paint("u undo · q quit", colDim, false)},
		}
	}

	b := roundedBox
	side := paint(b.v, fc, false)
	out := []string{borderLine(b.tl, b.h, b.tr, inner, fc, title, scores)}
	for r := 0; r < Size; r++ {
		tiles := make([][]string, Size)
		for c := 0; c < Size; c++ {
			hl := g.Merged[r][c] || (g.Spawn != nil && g.Spawn.R == r && g.Spawn.C == c)
			tiles[c] = renderTile(g.Board[r][c], sz, hl)
		}
		for i := 0; i < sz.h; i++ {
			var line strings.Builder
			line.WriteString(side + " ")
			for c := 0; c < Size; c++ {
				line.WriteString(tiles[c][i])
			}
			line.WriteString(" " + side)
			out = append(out, line.String())
		}
	}
	out = append(out, borderLine(b.bl, b.h, b.br, inner, fc, fit(inner-4, footer...), nil))

	view := strings.Join(out, "\n")
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, view)
	}
	return view
}
