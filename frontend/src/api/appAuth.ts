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

export interface AppAuthorizationDevice {
  id: string | number
  client_id: string
  client_name?: string
  device_name?: string
  platform?: string
  scopes: string[]
  created_at: string
  last_used_at?: string | null
}

type RawAuthorizationContext = Omit<AppAuthorizationContext, 'scopes'> & {
  scopes?: string[] | string
  scope?: string[] | string
}

type RawAuthorizationDevice = Omit<AppAuthorizationDevice, 'scopes'> & {
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

export async function listAuthorizationDevices(): Promise<AppAuthorizationDevice[]> {
  const { data } = await apiClient.get<
    RawAuthorizationDevice[] | { items?: RawAuthorizationDevice[]; devices?: RawAuthorizationDevice[] }
  >('/app-auth/devices')
  const devices = Array.isArray(data) ? data : (data.items ?? data.devices ?? [])
  return devices.map((device) => ({
    ...device,
    scopes: normalizeScopes(device.scopes ?? device.scope)
  }))
}

export async function revokeAuthorizationDevice(id: string | number): Promise<void> {
  await apiClient.delete(`/app-auth/devices/${encodeURIComponent(String(id))}`)
}

export const appAuthAPI = {
  createAuthorizationRequest,
  getAuthorizationContext,
  submitAuthorizationDecision,
  listAuthorizationDevices,
  revokeAuthorizationDevice
}
