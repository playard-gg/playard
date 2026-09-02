import type { RoomPlayer } from '../../store/roomStore'

interface PlayerListProps {
  players: RoomPlayer[]
  maxPlayers: number
  minPlayers: number
  selfId: string | null
}

export function PlayerList({ players, maxPlayers, minPlayers, selfId }: PlayerListProps) {
  const missing = Math.max(0, minPlayers - players.length)

  return (
    <section aria-labelledby="players-heading" className="rounded-md border border-border bg-surface p-6">
      <div className="flex items-baseline justify-between">
        <h2 id="players-heading" className="text-xs font-bold tracking-wide text-text-muted uppercase">
          Players
        </h2>
        <p className="text-sm font-bold text-accent-yellow">
          {players.length}
          <span className="text-text-muted">/{maxPlayers}</span>
        </p>
      </div>

      <ul className="mt-4 flex flex-col gap-2">
        {players.map((player) => (
          <li
            key={player.id}
            className="flex items-center justify-between rounded-sm border border-border bg-card px-4 py-3"
          >
            <span className="flex min-w-0 items-center gap-2">
              <span
                className={`h-2 w-2 shrink-0 rounded-full ${player.connected ? 'bg-accent-cyan' : 'bg-text-muted'}`}
                aria-hidden="true"
              />
              <span className="truncate font-semibold text-text-primary">
                {player.nickname}
                {player.id === selfId && <span className="ml-1.5 text-text-muted">(you)</span>}
              </span>
              {player.is_host && (
                <span className="shrink-0 rounded-pill bg-accent-yellow-muted px-2 py-0.5 text-[0.6rem] font-extrabold tracking-wide text-accent-yellow uppercase">
                  Host
                </span>
              )}
            </span>

            <span className="ml-3 shrink-0 text-xs font-bold">
              {!player.connected ? (
                <span className="text-text-muted">Reconnecting…</span>
              ) : player.is_host ? (
                <span className="text-text-muted">Starts the game</span>
              ) : (
                <span className="text-accent-cyan">Ready</span>
              )}
            </span>
          </li>
        ))}

        {Array.from({ length: missing }, (_, i) => (
          <li
            key={`empty-${i}`}
            className="rounded-sm border border-dashed border-border px-4 py-3 text-sm text-text-muted"
          >
            Waiting for a player…
          </li>
        ))}
      </ul>
    </section>
  )
}
