# Playard

Free, open-source multiplayer game hub. Pick a nickname, create or join a
room with a shareable code, and play simple games together in real time.
No accounts, no signup.

Early development — the room and lobby flow works end to end; the games
themselves are still being built.

## Tech Stack

- **Backend** — Go, HTTP API + WebSocket, no database (rooms live in
  memory for the lifetime of the room)
- **Frontend** — TypeScript, React, Vite
- **Styling** — Tailwind CSS, dark arcade theme (see
  [DESIGN_SYSTEM.md](DESIGN_SYSTEM.md))
- **Client state** — Zustand
- **UI primitives** — Radix UI, styled with our own tokens

## Getting Started

Requirements: Go 1.24+ and Node 20+.

```bash
git clone https://github.com/kaviraj-j/playard.git
cd playard
(cd web && npm install)
make dev
```

That runs the Go API and the Vite dev server together.

### Commands

| Command | What it does |
| --- | --- |
| `make dev` | Run server and web together |
| `make dev-server` | Go server only |
| `make dev-web` | Vite dev server only |
| `make build` | Build the server binary and the frontend bundle |
| `make run` | Build, then serve both |
| `make clean` | Remove build artifacts |

Tests: `cd server && go test ./...`

## Contributing

Games are plug-and-play: each one is self-contained on both sides — a Go
package implementing the shared `Game` interface on the server, and its
own UI folder on the client — so adding a game never touches the room or
matchmaking engine.

See [CLAUDE.md](CLAUDE.md) for architecture and conventions.
