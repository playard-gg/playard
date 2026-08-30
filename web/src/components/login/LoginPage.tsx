import { useState } from 'react'
import type { CSSProperties, FormEvent } from 'react'
import { useAuthStore } from '../../store/authStore'
import './login.css'

type Accent = 'yellow' | 'cyan' | 'pink'

interface Floater {
  emoji: string
  accent: Accent
  size: number
  animation: 'animate-drift' | 'animate-drift-slow'
  duration: string
  style: CSSProperties
}

const floaters: Floater[] = [
  { emoji: '🎮', accent: 'yellow', size: 46, animation: 'animate-drift', duration: '6s', style: { top: '12%', left: '8%' } },
  { emoji: '⚡', accent: 'cyan', size: 38, animation: 'animate-drift-slow', duration: '8s', style: { top: '22%', right: '10%' } },
  { emoji: '🏆', accent: 'pink', size: 42, animation: 'animate-drift', duration: '7.5s', style: { bottom: '16%', left: '10%' } },
  { emoji: '🎲', accent: 'yellow', size: 34, animation: 'animate-drift-slow', duration: '6.5s', style: { bottom: '22%', right: '8%' } },
  { emoji: '★', accent: 'cyan', size: 30, animation: 'animate-drift', duration: '9s', style: { top: '8%', left: '45%' } },
]

interface Spark {
  accent: Accent
  delay: string
  style: CSSProperties
}

const sparks: Spark[] = [
  { accent: 'yellow', delay: '0s', style: { top: '15%', left: '25%' } },
  { accent: 'cyan', delay: '0.6s', style: { top: '35%', left: '80%' } },
  { accent: 'pink', delay: '1.2s', style: { top: '65%', left: '15%' } },
  { accent: 'yellow', delay: '0.3s', style: { top: '75%', left: '70%' } },
  { accent: 'cyan', delay: '0.9s', style: { top: '50%', left: '50%' } },
  { accent: 'pink', delay: '1.5s', style: { top: '10%', left: '70%' } },
]

const floaterAccentClasses: Record<Accent, string> = {
  yellow: 'border-accent-yellow/40 bg-accent-yellow/15 text-accent-yellow',
  cyan: 'border-accent-cyan/40 bg-accent-cyan/15 text-accent-cyan',
  pink: 'border-accent-pink/40 bg-accent-pink/15 text-accent-pink',
}

const sparkAccentClasses: Record<Accent, string> = {
  yellow: 'bg-accent-yellow',
  cyan: 'bg-accent-cyan',
  pink: 'bg-accent-pink',
}

export function LoginPage() {
  const [nickname, setNickname] = useState('')
  const login = useAuthStore((state) => state.login)
  const isAuthenticating = useAuthStore((state) => state.isAuthenticating)
  const error = useAuthStore((state) => state.error)

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const trimmed = nickname.trim()
    if (trimmed.length < 2) return
    await login(trimmed).catch(() => {
      /* error is surfaced via the store's error state */
    })
  }

  return (
    <main className="login-page relative flex min-h-screen items-center justify-center overflow-hidden bg-page">
      <div className="bg-glow fixed inset-0 z-0" aria-hidden="true" />
      <div className="bg-grid fixed inset-0 z-0" aria-hidden="true" />

      <div className="pointer-events-none fixed inset-0 z-0" aria-hidden="true">
        {floaters.map((floater, i) => (
          <div
            key={i}
            className={`absolute flex items-center justify-center rounded-sm border text-lg ${floaterAccentClasses[floater.accent]} ${floater.animation}`}
            style={{ width: floater.size, height: floater.size, animationDuration: floater.duration, ...floater.style }}
          >
            {floater.emoji}
          </div>
        ))}
        {sparks.map((spark, i) => (
          <div
            key={i}
            className={`absolute h-[5px] w-[5px] rounded-full animate-twinkle ${sparkAccentClasses[spark.accent]}`}
            style={{ animationDelay: spark.delay, ...spark.style }}
          />
        ))}
      </div>

      <section
        aria-labelledby="login-heading"
        className="relative z-10 mx-6 w-full max-w-sm rounded-lg border border-border bg-card/85 p-9 text-center shadow-2xl backdrop-blur-lg"
      >
        <span className="mb-4 inline-flex items-center gap-1.5 rounded-pill bg-accent-pink px-3 py-1.5 text-[0.68rem] font-extrabold tracking-wide text-page uppercase">
          ★ Live now
        </span>

        <h1 id="login-heading" className="text-2xl font-bold tracking-tight text-text-primary">
          Pick your <span className="text-accent-yellow">nickname</span>
        </h1>
        <p className="mt-1.5 mb-7 text-sm text-text-muted">No email, no password. Just jump in.</p>

        <form onSubmit={handleSubmit} className="text-left">
          <label htmlFor="nickname" className="mb-1.5 block text-xs font-bold tracking-wide text-text-muted uppercase">
            Nickname
          </label>
          <input
            id="nickname"
            name="nickname"
            type="text"
            autoComplete="off"
            minLength={2}
            maxLength={20}
            required
            value={nickname}
            onChange={(event) => setNickname(event.target.value)}
            placeholder="e.g. Kurama99"
            className="w-full rounded-md border border-border bg-surface px-3.5 py-3 font-semibold text-text-primary placeholder:font-normal placeholder:text-text-muted focus:border-accent-cyan focus:outline-none"
          />

          {error && (
            <p role="alert" className="mt-3 text-sm text-accent-pink">
              {error}
            </p>
          )}

          <button
            type="submit"
            disabled={isAuthenticating || nickname.trim().length < 2}
            className="mt-5 w-full rounded-pill bg-linear-to-br from-accent-yellow to-[#f0a93c] px-6 py-3.5 font-extrabold text-page shadow-lg shadow-accent-yellow/25 transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {isAuthenticating ? 'Joining…' : 'Play Now →'}
          </button>
        </form>

        <p className="mt-4 text-xs text-text-muted">
          <span className="mr-1.5 inline-block h-1.5 w-1.5 rounded-full bg-accent-cyan shadow-[0_0_8px_var(--color-accent-cyan)]" />
          You&apos;ll join the lobby instantly
        </p>
      </section>
    </main>
  )
}
