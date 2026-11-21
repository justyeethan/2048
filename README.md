# 2048 (terminal)

2048 for the terminal, written in Go with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss).

## Run

```sh
go mod tidy   # only needed the first time if go.sum is missing anything
go run .
```

Or build a binary: `go build -o 2048 . && ./2048`

## Controls

| Key | Action |
| --- | --- |
| ← ↑ ↓ → / `w a s d` / `h j k l` | Move tiles |
| `u` | Undo last move |
| `r` | New game |
| `c` | Keep playing after reaching 2048 |
| `q` / `Esc` / `Ctrl+C` | Quit |

Your best score is saved in your user config folder (`~/Library/Application Support/term2048/best` on macOS).

## Tests

```sh
go test ./...
```

The game rules live in `game.go` with no terminal dependencies. `main.go` holds the Bubble Tea model and key handling, and `view.go` draws the board with box-drawing characters. Tiles are twice as wide as they are tall so they look square, and the board picks the largest of three tile sizes that fits your terminal.
