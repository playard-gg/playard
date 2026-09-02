import type { GameMetadata } from '../../lib/rooms'

// Accents rotate across the catalog purely as a chip background (a muted
// variant), which DESIGN_SYSTEM.md permits — the accent carries no action
// meaning here.
const chipAccents = [
  'bg-accent-yellow-muted text-accent-yellow',
  'bg-accent-cyan-muted text-accent-cyan',
  'bg-accent-pink-muted text-accent-pink',
]

interface GameCardProps {
  game: GameMetadata
  index: number
  onSelect: (game: GameMetadata) => void
}

export function GameCard({ game, index, onSelect }: GameCardProps) {
  const players =
    game.min_players === game.max_players
      ? `${game.min_players} players`
      : `${game.min_players}–${game.max_players} players`

  return (
    <button
      type="button"
      onClick={() => onSelect(game)}
      className="group rounded-md border border-border bg-surface p-6 text-left transition-[transform,border-color] duration-200 hover:-translate-y-1 hover:border-accent-yellow focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent-cyan"
    >
      <span
        className={`flex h-12 w-12 items-center justify-center rounded-sm text-2xl ${chipAccents[index % chipAccents.length]}`}
        aria-hidden="true"
      >
        {game.emoji}
      </span>

      <h3 className="mt-5 text-lg font-bold text-text-primary">{game.name}</h3>
      <p className="mt-1 text-sm text-text-muted">{game.tagline}</p>

      <dl className="mt-5 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs font-semibold text-text-muted">
        <div>
          <dt className="sr-only">Players</dt>
          <dd>{players}</dd>
        </div>
        <span aria-hidden="true">·</span>
        <div>
          <dt className="sr-only">Pacing</dt>
          <dd>{game.pacing === 'realtime' ? 'Real time' : 'Turn based'}</dd>
        </div>
      </dl>

      <span className="mt-5 inline-block text-sm font-extrabold text-accent-yellow opacity-0 transition-opacity group-hover:opacity-100 group-focus-visible:opacity-100">
        Play →
      </span>
    </button>
  )
}
