package main

import (
	"math/rand"
	"testing"
)

func TestSlideLine(t *testing.T) {
	cases := []struct {
		in, want [Size]int
		score    int
		changed  bool
	}{
		{[Size]int{0, 0, 0, 0}, [Size]int{0, 0, 0, 0}, 0, false},
		{[Size]int{2, 0, 0, 0}, [Size]int{2, 0, 0, 0}, 0, false},
		{[Size]int{0, 0, 0, 2}, [Size]int{2, 0, 0, 0}, 0, true},
		{[Size]int{2, 2, 0, 0}, [Size]int{4, 0, 0, 0}, 4, true},
		{[Size]int{2, 0, 2, 0}, [Size]int{4, 0, 0, 0}, 4, true},
		{[Size]int{2, 2, 2, 2}, [Size]int{4, 4, 0, 0}, 8, true},
		{[Size]int{2, 2, 2, 0}, [Size]int{4, 2, 0, 0}, 4, true},
		{[Size]int{4, 4, 8, 0}, [Size]int{8, 8, 0, 0}, 8, true}, // no chain merge
		{[Size]int{2, 4, 8, 16}, [Size]int{2, 4, 8, 16}, 0, false},
		{[Size]int{8, 0, 8, 8}, [Size]int{16, 8, 0, 0}, 16, true},
	}
	for _, c := range cases {
		got, score, _, changed := slideLine(c.in)
		if got != c.want || score != c.score || changed != c.changed {
			t.Errorf("slideLine(%v) = %v,%d,%v; want %v,%d,%v",
				c.in, got, score, changed, c.want, c.score, c.changed)
		}
	}
}

func TestMoveDirections(t *testing.T) {
	start := Board{
		{2, 0, 0, 2},
		{0, 4, 0, 4},
		{0, 0, 0, 0},
		{2, 0, 0, 0},
	}
	want := map[Direction]Board{
		Left:  {{4, 0, 0, 0}, {8, 0, 0, 0}, {0, 0, 0, 0}, {2, 0, 0, 0}},
		Right: {{0, 0, 0, 4}, {0, 0, 0, 8}, {0, 0, 0, 0}, {0, 0, 0, 2}},
		Up:    {{4, 4, 0, 2}, {0, 0, 0, 4}, {0, 0, 0, 0}, {0, 0, 0, 0}},
		Down:  {{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 2}, {4, 4, 0, 4}},
	}
	for d, w := range want {
		g := &Game{Board: start, rng: rand.New(rand.NewSource(1))}
		if !g.Move(d) {
			t.Fatalf("dir %d: expected move", d)
		}
		// Remove the spawned tile before comparing.
		if g.Spawn == nil {
			t.Fatalf("dir %d: expected a spawned tile", d)
		}
		if w[g.Spawn.R][g.Spawn.C] != 0 {
			t.Fatalf("dir %d: spawned onto an occupied cell", d)
		}
		g.Board[g.Spawn.R][g.Spawn.C] = 0
		if g.Board != w {
			t.Errorf("dir %d: got %v want %v", d, g.Board, w)
		}
	}
}

func TestNoOpMoveDoesNotSpawn(t *testing.T) {
	b := Board{{2, 4, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}}
	g := &Game{Board: b, rng: rand.New(rand.NewSource(1))}
	if g.Move(Left) {
		t.Fatal("move left should be a no-op")
	}
	if g.Board != b {
		t.Fatal("board changed on no-op move")
	}
}

func TestCanMove(t *testing.T) {
	stuck := Board{{2, 4, 2, 4}, {4, 2, 4, 2}, {2, 4, 2, 4}, {4, 2, 4, 2}}
	if (&Game{Board: stuck}).CanMove() {
		t.Error("expected no moves")
	}
	vert := stuck
	vert[3][3] = 4 // matches the 4 above it
	if !(&Game{Board: vert}).CanMove() {
		t.Error("expected a vertical merge to be possible")
	}
}

func TestNewGameHasTwoTiles(t *testing.T) {
	g := NewGame(rand.New(rand.NewSource(42)))
	n := 0
	for _, row := range g.Board {
		for _, v := range row {
			if v != 0 {
				n++
				if v != 2 && v != 4 {
					t.Errorf("unexpected start tile %d", v)
				}
			}
		}
	}
	if n != 2 {
		t.Errorf("got %d tiles, want 2", n)
	}
}

func TestRandomPlayIsConsistent(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	g := NewGame(rng)
	for i := 0; i < 5000 && g.CanMove(); i++ {
		before := g.Score
		g.Move(Direction(rng.Intn(4)))
		if g.Score < before {
			t.Fatal("score decreased")
		}
	}
	if g.MaxTile() < 64 {
		t.Errorf("random play ended suspiciously early, max tile %d", g.MaxTile())
	}
}
