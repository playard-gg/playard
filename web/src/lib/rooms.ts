import { apiRequest } from './api'

export type Pacing = 'turn_based' | 'realtime'
export type TeamMode = 'none' | 'ffa' | 'fixed_teams'
export type OutcomeType = 'none' | 'single_winner' | 'ranked' | 'team_win'
export type RoomStatus = 'lobby' | 'in_game' | 'finished'
export type Visibility = 'public' | 'private'

export interface GameMetadata {
  id: string
  name: string
  tagline: string
  emoji: string
  min_players: number
  max_players: number
  pacing: Pacing
  team_mode: TeamMode
  outcome_type: OutcomeType
}

export interface RoomRef {
  code: string
  game_id: string
}

export interface RoomSummary {
  code: string
  game_id: string
  game_name: string
  status: RoomStatus
  player_count: number
  max_players: number
}

export function fetchGames(): Promise<{ games: GameMetadata[] }> {
  return apiRequest('/api/games')
}

export function createRoom(gameId: string, visibility: Visibility, token: string): Promise<RoomRef> {
  return apiRequest('/api/rooms', { method: 'POST', body: { game_id: gameId, visibility }, token })
}

export function quickmatch(gameId: string, token: string): Promise<RoomRef> {
  return apiRequest('/api/rooms/quickmatch', { method: 'POST', body: { game_id: gameId }, token })
}

/** Pre-flight check that a code leads somewhere joinable, before opening a socket. */
export function joinRoom(code: string, token: string): Promise<RoomRef> {
  return apiRequest('/api/rooms/join', { method: 'POST', body: { code }, token })
}

export function fetchRoomSummary(code: string, token: string): Promise<RoomSummary> {
  return apiRequest(`/api/rooms/${encodeURIComponent(code)}`, { token })
}

/** The shareable link a host sends to friends. */
export function joinLink(code: string): string {
  return `${window.location.origin}/join/${code}`
}
