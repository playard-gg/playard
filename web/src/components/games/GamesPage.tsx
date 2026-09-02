import { useEffect, useState } from 'react'
import { GameCard } from './GameCard'
import { GameActionDialog } from './GameActionDialog'
import { JoinByCodeDialog } from './JoinByCodeDialog'
import { Button } from '../ui/Button'
import { fetchGames, type GameMetadata } from '../../lib/rooms'
import { useAuthStore } from '../../store/authStore'

/** Step 2 of the flow: the catalog, and the two ways into a room from here. */
export function GamesPage() {
  const nickname = useAuthStore((state) => state.nickname)
  const logout = useAuthStore((state) => state.logout)

  const [games, setGames] = useState<GameMetadata[] | null>(null)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [selected, setSelected] = useState<GameMetadata | null>(null)
  const [joinOpen, setJoinOpen] = useState(false)

  useEffect(() => {
    let active = true
    fetchGames()
      .then((res) => active && setGames(res.games))
      .catch((err: unknown) => {
        if (active) setLoadError(err instanceof Error ? err.message : 'Could not load games')
      })
    return () => {
      active = false
    }
  }, [])

  return (
    <div className="min-h-screen bg-page">
      <header className="mx-auto flex max-w-5xl items-center justify-between px-6 py-7">
        <p className="text-sm text-text-muted">
          Playing as <span className="font-bold text-text-primary">{nickname}</span>
        </p>
        <div className="flex items-center gap-3">
          <Button variant="secondary" className="px-5 py-2 text-sm" onClick={() => setJoinOpen(true)}>
            Join with code
          </Button>
          <Button variant="ghost" className="px-5 py-2 text-sm" onClick={logout}>
            Log out
          </Button>
        </div>
      </header>

      <main className="mx-auto max-w-5xl px-6 pb-20">
        <span className="inline-flex items-center gap-1.5 rounded-pill bg-accent-pink px-3 py-1.5 text-[0.68rem] font-extrabold tracking-wide text-page uppercase">
          ★ Live now
        </span>
        <h1 className="mt-4 text-4xl font-bold tracking-tight text-text-primary sm:text-5xl">
          Pick a <span className="text-accent-yellow">game</span>
        </h1>
        <p className="mt-2 max-w-md text-text-muted">
          Start a room and share the link, or jump in with whoever&apos;s already playing.
        </p>

        {loadError && (
          <p role="alert" className="mt-10 text-accent-pink">
            {loadError}
          </p>
        )}

        {games === null && !loadError && (
          <div className="mt-10 grid gap-5 sm:grid-cols-2 lg:grid-cols-3" aria-hidden="true">
            {[0, 1, 2].map((i) => (
              <div key={i} className="h-56 animate-pulse rounded-md border border-border bg-surface" />
            ))}
          </div>
        )}

        {games && (
          <section aria-label="Available games" className="mt-10 grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
            {games.map((game, index) => (
              <GameCard key={game.id} game={game} index={index} onSelect={setSelected} />
            ))}
          </section>
        )}
      </main>

      <GameActionDialog game={selected} onClose={() => setSelected(null)} />
      <JoinByCodeDialog open={joinOpen} onOpenChange={setJoinOpen} />
    </div>
  )
}
