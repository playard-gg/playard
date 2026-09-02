import type { ReactNode } from 'react'
import { LoginPage } from '../components/login/LoginPage'
import { useAuthStore } from '../store/authStore'

/**
 * Renders the login screen in place of the route's content when the player has
 * no valid session, without changing the URL. A shared /join/:code link
 * therefore resolves to the room right after login rather than dropping the
 * player on the games list.
 */
export function RequireAuth({ children }: { children: ReactNode }) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated())
  return isAuthenticated ? <>{children}</> : <LoginPage />
}
