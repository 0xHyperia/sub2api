import { apiClient } from './client'

export interface AuthorizationRequestParams {
  response_type: string
  client_id: string
  redirect_uri: string
  scope: string
  code_challenge: string
  code_challenge_method: string
  state?: string
  device_name?: string
  platform?: string
  installation_id: string
}

export interface AuthorizationRequestResult {
  request_id: string
  expires_in: number
  authorization_uri: string
}

export interface AppAuthorizationContext {
  request_id: string
  client_id: string
  client_name?: string
  device_name?: string
  platform?: string
  scopes: string[]
  expires_at?: string
}

export type AuthorizationDecision = 'allow' | 'deny'

export interface AuthorizationDecisionResult {
  redirect_uri: string
}

export interface AppAuthorizationGrant {
  id: string | number
  client_id: string
  client_name?: string
  platform?: string
  scopes: string[]
  session_count: number
  first_authorized_at: string
  last_authorized_at: string
  last_used_at?: string | null
}

export interface AppAuthorizationSession {
  id: string | number
  device_name: string
  platform?: string
  created_at: string
  last_used_at?: string | null
}

type RawAuthorizationContext = Omit<AppAuthorizationContext, 'scopes'> & {
  scopes?: string[] | string
  scope?: string[] | string
}

type RawAuthorizationGrant = Omit<AppAuthorizationGrant, 'scopes'> & {
  scopes?: string[] | string
  scope?: string[] | string
}

function normalizeScopes(scopes?: string[] | string): string[] {
  if (Array.isArray(scopes)) return scopes
  return typeof scopes === 'string' ? scopes.split(/\s+/).filter(Boolean) : []
}

export async function createAuthorizationRequest(
  params: AuthorizationRequestParams
): Promise<AuthorizationRequestResult> {
  const { data } = await apiClient.post<AuthorizationRequestResult>(
    '/app-auth/authorize/requests',
    params
  )
  return data
}

export async function getAuthorizationContext(
  requestId: string
): Promise<AppAuthorizationContext> {
  const { data } = await apiClient.get<RawAuthorizationContext>('/app-auth/authorize/context', {
    params: { request_id: requestId }
  })
  return { ...data, scopes: normalizeScopes(data.scopes ?? data.scope) }
}

export async function submitAuthorizationDecision(
  requestId: string,
  decision: AuthorizationDecision
): Promise<AuthorizationDecisionResult> {
  const { data } = await apiClient.post<AuthorizationDecisionResult>(
    '/app-auth/authorize/decision',
    { request_id: requestId, decision }
  )
  return data
}

export async function listAuthorizationGrants(): Promise<AppAuthorizationGrant[]> {
  const { data } = await apiClient.get<
    RawAuthorizationGrant[] | { items?: RawAuthorizationGrant[]; grants?: RawAuthorizationGrant[] }
  >('/app-auth/grants')
  const grants = Array.isArray(data) ? data : (data.items ?? data.grants ?? [])
  return grants.map((grant) => ({
    ...grant,
    scopes: normalizeScopes(grant.scopes ?? grant.scope)
  }))
}

export async function revokeAuthorizationGrant(id: string | number): Promise<void> {
  await apiClient.delete(`/app-auth/grants/${encodeURIComponent(String(id))}`)
}

export async function listAuthorizationSessions(
  grantId: string | number
): Promise<AppAuthorizationSession[]> {
  const { data } = await apiClient.get<AppAuthorizationSession[]>(
    `/app-auth/grants/${encodeURIComponent(String(grantId))}/sessions`
  )
  return Array.isArray(data) ? data : []
}

export async function renameAuthorizationSession(
  sessionId: string | number,
  deviceName: string
): Promise<void> {
  await apiClient.patch(`/app-auth/sessions/${encodeURIComponent(String(sessionId))}`, {
    device_name: deviceName
  })
}

export async function revokeAuthorizationSession(sessionId: string | number): Promise<void> {
  await apiClient.delete(`/app-auth/sessions/${encodeURIComponent(String(sessionId))}`)
}

export async function revokeOtherAuthorizationSessions(
  grantId: string | number,
  keepSessionId: string | number
): Promise<number> {
  const { data } = await apiClient.post<{ revoked_count?: number }>(
    `/app-auth/grants/${encodeURIComponent(String(grantId))}/sessions/${encodeURIComponent(String(keepSessionId))}/revoke-others`
  )
  return data.revoked_count ?? 0
}

export const appAuthAPI = {
  createAuthorizationRequest,
  getAuthorizationContext,
  submitAuthorizationDecision,
  listAuthorizationGrants,
  revokeAuthorizationGrant,
  listAuthorizationSessions,
  renameAuthorizationSession,
  revokeAuthorizationSession,
  revokeOtherAuthorizationSessions
}
