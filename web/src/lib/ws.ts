const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'

export type ConnectionStatus = 'connecting' | 'open' | 'reconnecting' | 'closed'

export interface OutboundMessage {
  type: string
  payload?: unknown
}

export interface InboundMessage {
  type: string
  data?: unknown
}

interface SocketOptions {
  code: string
  token: string
  onMessage: (message: InboundMessage) => void
  onStatus: (status: ConnectionStatus) => void
}

const INITIAL_RETRY_MS = 500
const MAX_RETRY_MS = 8000

/**
 * RoomSocket keeps one connection to a room alive, reconnecting with backoff.
 * The server holds the player's seat during the gap, so a reconnect resumes
 * the same room rather than rejoining a new one.
 */
export class RoomSocket {
  private socket: WebSocket | null = null
  private retryMs = INITIAL_RETRY_MS
  private retryTimer: ReturnType<typeof setTimeout> | null = null
  private closedByUs = false
  private readonly options: SocketOptions

  constructor(options: SocketOptions) {
    this.options = options
  }

  connect(): void {
    this.closedByUs = false
    this.options.onStatus(this.retryMs === INITIAL_RETRY_MS ? 'connecting' : 'reconnecting')

    const url = new URL('/api/ws', API_BASE_URL)
    url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
    url.searchParams.set('code', this.options.code)
    url.searchParams.set('token', this.options.token)

    const socket = new WebSocket(url)
    this.socket = socket

    socket.onopen = () => {
      this.retryMs = INITIAL_RETRY_MS
      this.options.onStatus('open')
    }

    socket.onmessage = (event) => {
      try {
        this.options.onMessage(JSON.parse(event.data as string) as InboundMessage)
      } catch {
        // A malformed frame is the server's problem, not something the UI can
        // act on — drop it rather than tearing down a working connection.
      }
    }

    socket.onclose = () => {
      this.socket = null
      if (this.closedByUs) {
        this.options.onStatus('closed')
        return
      }
      this.scheduleReconnect()
    }

    socket.onerror = () => socket.close()
  }

  send(message: OutboundMessage): void {
    if (this.socket?.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify(message))
    }
  }

  close(): void {
    this.closedByUs = true
    if (this.retryTimer) clearTimeout(this.retryTimer)
    this.retryTimer = null
    this.socket?.close()
    this.socket = null
    this.options.onStatus('closed')
  }

  private scheduleReconnect(): void {
    this.options.onStatus('reconnecting')
    this.retryTimer = setTimeout(() => this.connect(), this.retryMs)
    this.retryMs = Math.min(this.retryMs * 2, MAX_RETRY_MS)
  }
}
