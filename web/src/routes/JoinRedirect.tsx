import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { joinRoom } from '../lib/rooms'
import { useAuthStore } from '../store/authStore'

/**
 * Resolves a shared /join/:code link. The code is checked before the socket
 * opens so a dead or full room gives a real message instead of a silent
 * connection failure.
 */
export function JoinRedirect() {
  const { code = '' } = useParams()
  const token = useAuthStore((state) => state.token)
  const navigate = useNavigate()
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!token || !code) return

    let active = true
    joinRoom(code, token)
      .then((room) => active && navigate(`/room/${room.code}`, { replace: true }))
      .catch((err: unknown) => {
        if (active) setError(err instanceof Error ? err.message : 'Could not join that room')
      })

    return () => {
      active = false
    }
  }, [code, token, navigate])

  return (
    <main className="flex min-h-screen items-center justify-center bg-page px-6">
      <div className="max-w-sm text-center">
        {error ? (
          <>
            <h1 className="text-2xl font-bold text-text-primary">That link didn&apos;t work</h1>
            <p role="alert" className="mt-2 text-accent-pink">
              {error}
            </p>
            <Link
              to="/"
              className="mt-6 inline-block rounded-pill border border-accent-cyan px-6 py-3 font-extrabold text-accent-cyan"
            >
              Browse games
            </Link>
          </>
        ) : (
          <p className="text-text-muted">Finding room {code}…</p>
        )}
      </div>
    </main>
  )
}
