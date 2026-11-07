# Barrel Jumper

A construction worker climbs a building site, jumping barrels and fireballs,
to rescue the captive held at the top by a hulking barrel thrower.

Barrel Jumper is an independent rebuild of the Classic mode of CHAMP Kong
(DOS, 1996). Its rules and timing come from studying how the original plays.
All of its code, artwork, sound and music are new. It is not affiliated with
or endorsed by the authors of CHAMP Kong.

## Download

Builds for Windows and Linux (64-bit) are on the
[Releases](https://github.com/comprofix/barreljumper/releases) page, with
`SHA256SUMS.txt` to check them against.

- **Windows:** unzip `barreljumper-vX.Y.Z-windows-amd64.zip` and run `barreljumper.exe`.
- **Linux:** extract `barreljumper-vX.Y.Z-linux-amd64.tar.gz` and run `./barreljumper`.
  It needs OpenGL, X11 and ALSA, which desktop systems already have.

## How to play

Each level is a row of screens. Clear one to move on to the next.

- **Barrels:** climb to the top while the thrower rolls barrels down the
  girders. Barrels that reach the oil drum can set loose fireballs.
- **Rivets:** walk or jump over all eight rivets to remove them and bring
  the structure down.
- **Elevators:** ride the lifts up and keep clear of the bouncing springs.
- **Conveyor:** cross the moving belts, dodge the pies and time the
  retracting ladders.

Jump up to a hammer to grab it, and it smashes what is in your way for a
while. You can't climb or jump while you hold it. Touching a hazard, or falling too far, costs a
life.

### Scoring

| Event | Points |
|---|---|
| Jump one hazard | 100 |
| Jump two hazards at once | 300 |
| Jump three or more at once | 800 |
| Smash a barrel or fireball with the hammer | 300, 500 or 800 |
| Remove a rivet | 100 |
| Collect a lost item | 300 to 800, by level |
| Clear a screen | the bonus left |

The bonus starts at 5000 on level 1, 1000 more on each later level up to 8000,
and counts down as you play. If it reaches zero you lose a life. You earn an
extra life at 7,000 points, then at 50,000 and every 50,000 after that.

### Options

Before a game, choose:

- **Players:** one or two, taking turns.
- **Starting level:** 1 to 8.
- **Skill:** for each player.
  - Easy: five lives, and a held hammer guards you.
  - Arcade: three lives, as the original plays.
  - Hard: three lives and faster throws.

## Controls

| Action | Keyboard | Gamepad |
|---|---|---|
| Move and climb | Arrow keys or WASD | D-pad or left stick |
| Jump; confirm or skip in menus and scenes | Space or Enter | A |
| Pause | Esc or P | Start |
| Back (menus) | Esc or Backspace | B |
| Restart game | Ctrl+R | |
| Main menu | Ctrl+M | Select |
| Sound on or off | Ctrl+S | |
| Change scale mode | F9 | |
| Quit | F10 | |
| Fullscreen | F11 or Alt+Enter | |

The right mouse button also pauses and goes back.

Settings and high scores are saved in `settings.json` and `scores.json` next
to the game.

## Build and run

Requirements:

- Go 1.27 or later.
- **Linux only:** cgo and the X11, GL and ALSA development packages:
  `libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev pkg-config`

```sh
go run ./cmd/barreljumper                 # run the game
go run -tags debug ./cmd/barreljumper     # with the debug overlay
go generate ./assets                      # rebuild the sprite sheets
go vet ./...
go test ./...
go build -o bin/ ./cmd/barreljumper
```

Command-line options, for testing:

| Option | Meaning |
|---|---|
| `-seed n` | Use a fixed random seed, so runs repeat. |
| `-level n` | Start new games on level `n`. These games keep no high scores. |
| `-screen n` | Start new games on screen `n` of the level. These games keep no high scores. |

## Credits

- Built with [Ebitengine](https://ebitengine.org) on
  [Game Skeleton](https://github.com/comprofix/gameskeleton) v0.5.1.
- Colours from the Resurrect 64 palette by Kerrie Lake.
- Original game: CHAMP Kong, by John Champeau.

## License

MIT. See [LICENSE](LICENSE).
