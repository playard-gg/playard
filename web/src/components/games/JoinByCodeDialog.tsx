import { useState } from 'react'
import type { FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { Dialog } from '../ui/Dialog'
import { Button } from '../ui/Button'
import { joinRoom } from '../../lib/rooms'
import { useAuthStore } from '../../store/authStore'

const CODE_LENGTH = 6

interface JoinByCodeDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

/** Step 3 of the flow: a friend read you the code, or you typed it off a link. */
export function JoinByCodeDialog({ open, onOpenChange }: JoinByCodeDialogProps) {
  const token = useAuthStore((state) => state.token)
  const navigate = useNavigate()
  const [code, setCode] = useState('')
  const [pending, setPending] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (!token || code.length !== CODE_LENGTH) return

    setPending(true)
    setError(null)
    try {
      const room = await joinRoom(code, token)
      navigate(`/room/${room.code}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not join that room')
      setPending(false)
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={onOpenChange}
      title="Join with a code"
      description="Six characters, from whoever invited you."
    >
      <form onSubmit={handleSubmit}>
        <label htmlFor="room-code" className="sr-only">
          Room code
        </label>
        <input
          id="room-code"
          name="room-code"
          autoComplete="off"
          autoCapitalize="characters"
          spellCheck={false}
          maxLength={CODE_LENGTH}
          value={code}
          onChange={(event) => setCode(event.target.value.toUpperCase().replace(/[^A-Z0-9]/g, ''))}
          placeholder="ABC234"
          className="w-full rounded-md border border-border bg-surface px-4 py-4 text-center font-mono text-2xl font-bold tracking-[0.4em] text-text-primary uppercase placeholder:tracking-[0.4em] placeholder:text-text-muted focus:border-accent-cyan focus:outline-none"
        />

        {error && (
          <p role="alert" className="mt-3 text-center text-sm text-accent-pink">
            {error}
          </p>
        )}

        <Button type="submit" className="mt-5 w-full" disabled={pending || code.length !== CODE_LENGTH}>
          {pending ? 'Joining…' : 'Join room →'}
        </Button>
      </form>
    </Dialog>
  )
}
