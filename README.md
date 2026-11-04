# Barrel Jumper

Created from [gameskeleton](https://github.com/comprofix/gameskeleton) v0.5.1.

## Requirements

- Go 1.27 or later
- **Linux only:** cgo plus the X11/GL/ALSA development packages:
  `libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev pkg-config`

## Build and run

```sh
go run ./cmd/barreljumper                 # run the game
go run -tags debug ./cmd/barreljumper     # with the debug overlay
go generate ./assets               # rebuild sprite sheets
go vet ./...
go test ./...
go build -o bin/ ./cmd/barreljumper
```

## Personal documents

The private project instructions, design and plan live in project-notes.
The local checkout uses ignored links for AGENTS.md, CLAUDE.md, DESIGN.md and PLAN.md.

## License

MIT. See [LICENSE](LICENSE).
