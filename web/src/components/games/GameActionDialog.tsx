import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Dialog } from '../ui/Dialog'
import { Button } from '../ui/Button'
import { createRoom, quickmatch, type GameMetadata } from '../../lib/rooms'
import { useAuthStore } from '../../store/authStore'

interface GameActionDialogProps {
  game: GameMetadata | null
  onClose: () => void
}

/** Step 4 of the flow: create a private room, or get matched with strangers. */
export function GameActionDialog({ game, onClose }: GameActionDialogProps) {
  const token = useAuthStore((state) => state.token)
  const navigate = useNavigate()
  const [pending, setPending] = useState<'create' | 'quickmatch' | null>(null)
  const [error, setError] = useState<string | null>(null)

  const go = async (action: 'create' | 'quickmatch') => {
    if (!game || !token) return
    setPending(action)
    setError(null)
    try {
      const room =
        action === 'create'
          ? await createRoom(game.id, 'private', token)
          : await quickmatch(game.id, token)
      navigate(`/room/${room.code}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong')
      setPending(null)
    }
  }

  return (
    <Dialog
      open={game !== null}
      onOpenChange={(open) => !open && onClose()}
      title={game ? `Play ${game.name}` : ''}
      description="Bring your own crowd, or get dropped in with whoever's around."
    >
      <div className="flex flex-col gap-3">
        <Button variant="primary" disabled={pending !== null} onClick={() => go('create')}>
          {pending === 'create' ? 'Creating room…' : 'Create a room'}
        </Button>
        <p className="-mt-1 text-center text-xs text-text-muted">
          You get a code and a link to share with friends.
        </p>

        <Button
          variant="secondary"
          className="mt-3"
          disabled={pending !== null}
          onClick={() => go('quickmatch')}
        >
          {pending === 'quickmatch' ? 'Finding a room…' : 'Play with strangers'}
        </Button>
        <p className="-mt-1 text-center text-xs text-text-muted">
          Drops you into an open room, or opens a fresh one.
        </p>

        {error && (
          <p role="alert" className="mt-2 text-center text-sm text-accent-pink">
            {error}
          </p>
        )}
      </div>
    </Dialog>
  )
}
