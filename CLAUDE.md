# Playard

Free, open-source multiplayer game hub. Players pick a nickname, join or
create a room via a shareable code, and play simple games together in
real time. No accounts, no signup.

License: MIT.

## Tech Stack

- **Backend**: Go — HTTP API + WebSocket server (`gorilla/websocket`)
- **Frontend**: TypeScript + React + Vite
- **Styling**: Tailwind CSS (see `DESIGN_SYSTEM.md` — dark arcade theme,
  follow it exactly, don't introduce new colors ad hoc)
- **Client state**: Zustand
- **UI primitives**: Radix UI (unstyled) for dialogs/modals/dropdowns,
  styled with our Tailwind tokens — don't pull in a full component kit
  (no MUI/Chakra/shadcn-as-dependency)

## Repo Structure (monorepo)

```
playard/
├── server/          # Go backend
│   ├── cmd/         # entrypoint(s)
│   ├── internal/
│   │   ├── room/    # room/lobby lifecycle, player sessions
│   │   ├── ws/       # websocket hub, connection handling
│   │   └── games/    # one subpackage per game (see Game Architecture)
│   └── go.mod
├── web/             # React frontend
│   ├── src/
│   │   ├── components/   # organized by feature, not by type
│   │   ├── games/        # one folder per game's client UI
│   │   ├── hooks/
│   │   ├── store/        # Zustand stores
│   │   └── lib/
│   └── package.json
├── DESIGN_SYSTEM.md
└── project_plan/
```

## Game Architecture — Plug-and-Play

Adding a new game must never require touching the core room/matchmaking
engine. Each game is self-contained on both sides:

**Server side** (`server/internal/games/<game>/`): implements a common
`Game` interface. The room/hub only knows about this interface, never
about specific game rules.

```
Game interface {
    Metadata() GameMetadata   // id, name, minPlayers, maxPlayers,
                               // pacing (turn_based | realtime),
                               // teamMode (ffa | fixed_teams | none),
                               // outcomeType (none | single_winner |
                               //              ranked | team_win)
    Init(players []Player, config Config) State
    ApplyAction(state State, playerID string, action Action) (State, error)
    CheckEnd(state State) *Result
    ViewFor(state State, playerID string) ClientState  // per-player view;
        // most games return the same view for everyone, but this exists
        // so asymmetric-role games (imposter/traitor/secret-word style)
        // can hide state from specific players without a special case
        // in the room/hub layer
}
```

The `Metadata()` call lets the room/lobby treat games generically —
validating player counts, showing "waiting for players," and rendering
game-agnostic lobby chrome — without any game-specific logic leaking
into the shared room code.

**Client side** (`web/src/games/<game>/`): owns its own board/UI
components and reads/writes through the shared room WebSocket
connection + a per-game slice of client state. The lobby/room shell
(player list, room code, chat if any) is shared chrome around whatever
game is active.

When adding a game, check whether the interface actually fits before
extending it — don't bend the shared interface into a special case for
one game.

## Multiplayer Model

- WebSocket connection per player, established after joining a room
- Room = lobby (room code, player list, ready state) + active game
  instance + game-agnostic result state (winner, points, "play again?")
- Anonymous nickname only, scoped to the room/session — nothing
  persisted server-side beyond the room's lifetime
- Keep server the source of truth for game state; client renders and
  sends intents, never computes authoritative outcomes itself
- **Reconnects**: a dropped connection should not instantly remove a
  player from an in-progress game. Hold their seat for a short grace
  period (e.g. 30–60s) keyed to their session token, and let them
  rejoin the same room/game state on reconnect. Decide this explicitly
  now rather than defaulting to "disconnect = gone."

## Abuse & Rate Limiting (no-accounts tradeoff)

No signup means no auth layer between a request and the API, so:

- Room codes must be long/random enough to resist brute-force guessing
  into someone else's room
- Rate-limit room creation and join attempts per IP
- Keep this minimal for MVP — the point is to avoid an easy accidental
  DoS, not to build a full trust & safety system

## Coding Style

- Simplicity over abstraction — this is a small OSS project. Don't
  build a generic "game engine" beyond what's needed to keep games
  pluggable. Three similar lines beat a premature abstraction.
- Go: idiomatic Go, small focused packages, explicit error handling,
  no unnecessary interfaces beyond the `Game` plug-and-play boundary
- React: function components, hooks, colocate a game's logic with its
  own folder rather than spreading it across shared directories
- Follow `DESIGN_SYSTEM.md` for all UI — dark surfaces only, the three
  defined accent colors with their fixed meanings (yellow = primary
  action, cyan = secondary, pink = live/urgent), pill/rounded corners

## Testing

- Go: table-driven unit tests for each game's rules/state machine
  (win conditions, invalid moves, edge cases) — this is where bugs
  actually hurt in a multiplayer game, prioritize it over UI tests
- Frontend: unit test game logic/hooks that don't need a live socket;
  don't chase heavy E2E coverage for a project this size unless a
  specific flow keeps breaking

## Commands

_(fill in once the projects are scaffolded, e.g. `go run ./server/cmd/...`,
`npm run dev` in `web/`, lint/format commands)_