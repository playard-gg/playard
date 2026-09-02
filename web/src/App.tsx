import { Navigate, Route, Routes } from 'react-router-dom'
import { GamesPage } from './components/games/GamesPage'
import { RoomPage } from './components/room/RoomPage'
import { JoinRedirect } from './routes/JoinRedirect'
import { RequireAuth } from './routes/RequireAuth'

function App() {
  return (
    <Routes>
      <Route
        path="/"
        element={
          <RequireAuth>
            <GamesPage />
          </RequireAuth>
        }
      />
      <Route
        path="/join/:code"
        element={
          <RequireAuth>
            <JoinRedirect />
          </RequireAuth>
        }
      />
      <Route
        path="/room/:code"
        element={
          <RequireAuth>
            <RoomPage />
          </RequireAuth>
        }
      />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

export default App
