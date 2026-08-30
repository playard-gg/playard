interface TokenClaims {
  player_id: string
  nickname: string
  issued_at: number
  expires_at: number
}

/** Decodes the (unverified) claims from a session token to check expiry client-side. */
export function decodeTokenClaims(token: string): TokenClaims | null {
  const [encodedPayload] = token.split('.')
  if (!encodedPayload) return null

  try {
    const base64 = encodedPayload.replace(/-/g, '+').replace(/_/g, '/')
    const json = atob(base64)
    return JSON.parse(json) as TokenClaims
  } catch {
    return null
  }
}

export function isTokenExpired(token: string): boolean {
  const claims = decodeTokenClaims(token)
  if (!claims) return true
  return Date.now() >= claims.expires_at * 1000
}
