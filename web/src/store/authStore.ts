import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { apiRequest } from '../lib/api'
import { isTokenExpired } from '../lib/token'

interface LoginResponse {
  token: string
  player_id: string
  nickname: string
}

interface AuthState {
  token: string | null
  playerId: string | null
  nickname: string | null
  isAuthenticating: boolean
  error: string | null
  login: (nickname: string) => Promise<void>
  logout: () => void
  isAuthenticated: () => boolean
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      playerId: null,
      nickname: null,
      isAuthenticating: false,
      error: null,

      login: async (nickname: string) => {
        set({ isAuthenticating: true, error: null })
        try {
          const res = await apiRequest<LoginResponse>('/api/auth/login', {
            method: 'POST',
            body: { nickname },
          })
          set({
            token: res.token,
            playerId: res.player_id,
            nickname: res.nickname,
            isAuthenticating: false,
          })
        } catch (err: unknown) {
          const message = err instanceof Error ? err.message : 'Failed to log in'
          set({ isAuthenticating: false, error: message })
          throw err
        }
      },

      logout: () => set({ token: null, playerId: null, nickname: null, error: null }),

      isAuthenticated: () => {
        const { token } = get()
        return Boolean(token) && !isTokenExpired(token as string)
      },
    }),
    {
      name: 'playard-auth',
      partialize: (state) => ({
        token: state.token,
        playerId: state.playerId,
        nickname: state.nickname,
      }),
    },
  ),
)
