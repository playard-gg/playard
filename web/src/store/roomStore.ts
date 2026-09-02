import { create } from 'zustand'
import { RoomSocket, type ConnectionStatus, type InboundMessage } from '../lib/ws'
import type { RoomStatus, Visibility } from '../lib/rooms'

export interface RoomPlayer {
  id: string
  nickname: string
  connected: boolean
  is_host: boolean
}

export interface RoomView {
  code: string
  game_id: string
  game_name: string
  visibility: Visibility
  status: RoomStatus
  host_id: string
  min_players: number
  max_players: number
  players: RoomPlayer[]
  can_start: boolean
  game_view?: unknown
}

interface RoomState {
  socket: RoomSocket | null
  connection: ConnectionStatus
  room: RoomView | null
  error: string | null
  connect: (code: string, token: string) => void
  disconnect: () => void
  start: () => void
  leave: () => void
  clearError: () => void
}

/**
 * Mirrors whatever the server broadcasts. The client never computes room or
 * game state itself — it renders `room` and sends intents.
 */
export const useRoomStore = create<RoomState>()((set, get) => ({
  socket: null,
  connection: 'closed',
  room: null,
  error: null,

  connect: (code, token) => {
    get().socket?.close()

    const socket = new RoomSocket({
      code,
      token,
      onStatus: (connection) => set({ connection }),
      onMessage: (message: InboundMessage) => {
        switch (message.type) {
          case 'room_state':
          case 'game_started':
            set({ room: message.data as RoomView, error: null })
            break
          case 'error':
            set({ error: (message.data as { message: string }).message })
            break
        }
      },
    })

    set({ socket, room: null, error: null })
    socket.connect()
  },

  disconnect: () => {
    get().socket?.close()
    set({ socket: null, room: null, connection: 'closed', error: null })
  },

  start: () => get().socket?.send({ type: 'start' }),
  leave: () => {
    get().socket?.send({ type: 'leave' })
    get().disconnect()
  },
  clearError: () => set({ error: null }),
}))
