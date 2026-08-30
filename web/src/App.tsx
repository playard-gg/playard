import { LoginPage } from './components/login/LoginPage'
import { useAuthStore } from './store/authStore'

function App() {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated())
  const nickname = useAuthStore((state) => state.nickname)
  const logout = useAuthStore((state) => state.logout)

  if (!isAuthenticated) {
    return <LoginPage />
  }

  return (
    <main className="flex min-h-screen items-center justify-center bg-page">
      <div className="text-center">
        <p className="text-text-muted">Signed in as</p>
        <h1 className="text-3xl font-bold text-text-primary">{nickname}</h1>
        <button
          type="button"
          onClick={logout}
          className="mt-6 rounded-pill border border-accent-cyan px-6 py-2 font-bold text-accent-cyan transition-opacity hover:opacity-90"
        >
          Log out
        </button>
      </div>
    </main>
  )
}

export default App
