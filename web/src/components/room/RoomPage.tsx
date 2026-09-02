import { useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { RoomCodeCard } from './RoomCodeCard'
import { PlayerList } from './PlayerList'
import { Button } from '../ui/Button'
import { useAuthStore } from '../../store/authStore'
import { useRoomStore } from '../../store/roomStore'

const connectionLabels: Record<string, string> = {
  connecting: 'Connecting…',
  reconnecting: 'Reconnecting…',
  closed: 'Disconnected',
}

/** Being in the room is being ready, so the wait is only ever for more players. */
function waitingLabel(missing: number): string {
  return `Waiting for ${missing} more player${missing === 1 ? '' : 's'}`
}

/** Steps 5–8: the lobby everyone converges on, and the Start button. */
export function RoomPage() {
  const { code = '' } = useParams()
  const navigate = useNavigate()

  const token = useAuthStore((state) => state.token)
  const playerId = useAuthStore((state) => state.playerId)

  const room = useRoomStore((state) => state.room)
  const connection = useRoomStore((state) => state.connection)
  const error = useRoomStore((state) => state.error)
  const connect = useRoomStore((state) => state.connect)
  const disconnect = useRoomStore((state) => state.disconnect)
  const start = useRoomStore((state) => state.start)
  const leave = useRoomStore((state) => state.leave)

  useEffect(() => {
    if (!token || !code) return
    connect(code, token)
    return () => disconnect()
  }, [code, token, connect, disconnect])

  const self = room?.players.find((player) => player.id === playerId) ?? null
  const isHost = self?.is_host ?? false

  const handleLeave = () => {
    leave()
    navigate('/')
  }

  return (
    <div className="min-h-screen bg-page">
      <header className="mx-auto flex max-w-3xl items-center justify-between px-6 py-7">
        <div>
          <p className="text-xs font-bold tracking-wide text-text-muted uppercase">Lobby</p>
          <h1 className="text-2xl font-bold text-text-primary">{room?.game_name ?? 'Loading…'}</h1>
        </div>
        <Button variant="ghost" className="px-5 py-2 text-sm" onClick={handleLeave}>
          Leave
        </Button>
      </header>

      <main className="mx-auto max-w-3xl px-6 pb-20">
        {connection !== 'open' && (
          <p role="status" className="mb-5 rounded-sm border border-border bg-surface px-4 py-3 text-sm text-text-muted">
            {connectionLabels[connection] ?? 'Connecting…'} Your seat is held for a moment while you
            reconnect.
          </p>
        )}

        {error && (
          <p role="alert" className="mb-5 rounded-sm border border-accent-pink bg-accent-pink-muted px-4 py-3 text-sm text-accent-pink">
            {error}
          </p>
        )}

        {room?.status === 'in_game' ? (
          <section className="rounded-lg border border-border bg-card p-12 text-center">
            <p className="text-xs font-bold tracking-wide text-accent-pink uppercase">Game started</p>
            <h2 className="mt-3 text-3xl font-bold text-text-primary">{room.game_name}</h2>
            <p className="mt-3 text-text-muted">
              The board lands here next — this build stops at the starting whistle.
            </p>
          </section>
        ) : (
          <div className="flex flex-col gap-5">
            <RoomCodeCard code={room?.code ?? code} />

            {room && (
              <PlayerList
                players={room.players}
                minPlayers={room.min_players}
                maxPlayers={room.max_players}
                selfId={playerId}
              />
            )}

            {room &&
              (isHost ? (
                <Button className="w-full" disabled={!room.can_start} onClick={start}>
                  {room.can_start ? 'Start game →' : waitingLabel(room.min_players - room.players.length)}
                </Button>
              ) : (
                <p
                  role="status"
                  className="rounded-sm border border-border bg-surface px-4 py-3 text-center text-sm text-text-muted"
                >
                  {room.can_start
                    ? 'You’re in — waiting for the host to start.'
                    : `You’re in — ${waitingLabel(room.min_players - room.players.length).toLowerCase()}.`}
                </p>
              ))}
          </div>
        )}
      </main>
    </div>
  )
}
