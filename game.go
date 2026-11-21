package main

import "math/rand"

// Size is the width and height of the board.
const Size = 4

// WinTile is the tile value that wins the game.
const WinTile = 2048

type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)

type Board [Size][Size]int

type Pos struct{ R, C int }

// Game holds the pure game state; it knows nothing about the terminal.
type Game struct {
	Board  Board
	Score  int
	Spawn  *Pos             // tile spawned by the last move (for highlighting)
	Merged [Size][Size]bool // tiles produced by a merge on the last move
	rng    *rand.Rand
}

func NewGame(rng *rand.Rand) *Game {
	g := &Game{rng: rng}
	g.spawnTile()
	g.spawnTile()
	g.Spawn = nil
	return g
}

// Clone returns a copy of the game state (used for undo).
func (g *Game) Clone() *Game {
	c := *g
	return &c
}

// coord maps (line i, position j along the line) to a board cell, where
// position 0 is the edge tiles slide towards.
func coord(d Direction, i, j int) (int, int) {
	switch d {
	case Left:
		return i, j
	case Right:
		return i, Size - 1 - j
	case Up:
		return j, i
	default: // Down
		return Size - 1 - j, i
	}
}

// slideLine slides and merges one line towards index 0.
// It returns the new line, points gained, which cells were merges, and
// whether anything changed.
func slideLine(line [Size]int) ([Size]int, int, [Size]bool, bool) {
	var out [Size]int
	var merged [Size]bool
	score, pos, last := 0, 0, 0
	for _, v := range line {
		if v == 0 {
			continue
		}
		if v == last {
			out[pos-1] = v * 2
			merged[pos-1] = true
			score += v * 2
			last = 0 // a merged tile can't merge again this move
		} else {
			out[pos] = v
			pos++
			last = v
		}
	}
	return out, score, merged, out != line
}

// Move applies a move. It returns false (and changes nothing) if the move
// doesn't shift any tile.
func (g *Game) Move(d Direction) bool {
	var next Board
	var merged [Size][Size]bool
	gained, moved := 0, false
	for i := 0; i < Size; i++ {
		var line [Size]int
		for j := 0; j < Size; j++ {
			r, c := coord(d, i, j)
			line[j] = g.Board[r][c]
		}
		out, s, m, changed := slideLine(line)
		moved = moved || changed
		gained += s
		for j := 0; j < Size; j++ {
			r, c := coord(d, i, j)
			next[r][c] = out[j]
			merged[r][c] = m[j]
		}
	}
	if !moved {
		return false
	}
	g.Board = next
	g.Merged = merged
	g.Score += gained
	g.spawnTile()
	return true
}

func (g *Game) spawnTile() {
	var empty []Pos
	for r := 0; r < Size; r++ {
		for c := 0; c < Size; c++ {
			if g.Board[r][c] == 0 {
				empty = append(empty, Pos{r, c})
			}
		}
	}
	if len(empty) == 0 {
		g.Spawn = nil
		return
	}
	p := empty[g.rng.Intn(len(empty))]
	v := 2
	if g.rng.Intn(10) == 0 {
		v = 4
	}
	g.Board[p.R][p.C] = v
	g.Spawn = &p
}

// CanMove reports whether any move is still possible.
func (g *Game) CanMove() bool {
	for r := 0; r < Size; r++ {
		for c := 0; c < Size; c++ {
			v := g.Board[r][c]
			if v == 0 {
				return true
			}
			if c+1 < Size && g.Board[r][c+1] == v {
				return true
			}
			if r+1 < Size && g.Board[r+1][c] == v {
				return true
			}
		}
	}
	return false
}

// MaxTile returns the largest tile on the board.
func (g *Game) MaxTile() int {
	m := 0
	for _, row := range g.Board {
		for _, v := range row {
			if v > m {
				m = v
			}
		}
	}
	return m
}
