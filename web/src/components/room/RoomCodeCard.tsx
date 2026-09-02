import { useState } from 'react'
import { joinLink } from '../../lib/rooms'
import { Button } from '../ui/Button'

interface RoomCodeCardProps {
  code: string
}

/** Step 5 of the flow: the host's shareable code and link. */
export function RoomCodeCard({ code }: RoomCodeCardProps) {
  const [copied, setCopied] = useState<'code' | 'link' | null>(null)

  const copy = async (kind: 'code' | 'link') => {
    try {
      await navigator.clipboard.writeText(kind === 'code' ? code : joinLink(code))
      setCopied(kind)
      setTimeout(() => setCopied(null), 2000)
    } catch {
      // Clipboard access can be denied; the code is on screen to read out.
    }
  }

  return (
    <section
      aria-labelledby="room-code-heading"
      className="rounded-md border border-border bg-surface p-6"
    >
      <h2 id="room-code-heading" className="text-xs font-bold tracking-wide text-text-muted uppercase">
        Room code
      </h2>

      <p className="mt-3 font-mono text-4xl font-bold tracking-[0.3em] text-accent-yellow">{code}</p>

      <div className="mt-5 flex flex-wrap gap-3">
        <Button variant="secondary" className="px-5 py-2 text-sm" onClick={() => copy('code')}>
          {copied === 'code' ? 'Copied!' : 'Copy code'}
        </Button>
        <Button variant="secondary" className="px-5 py-2 text-sm" onClick={() => copy('link')}>
          {copied === 'link' ? 'Copied!' : 'Copy invite link'}
        </Button>
      </div>

      <p aria-live="polite" className="mt-4 text-xs text-text-muted">
        Send the link to a friend — it drops them straight into this room.
      </p>
    </section>
  )
}
