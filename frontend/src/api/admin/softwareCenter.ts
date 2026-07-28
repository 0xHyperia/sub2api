import { apiClient } from '../client'
import type { GitHubRelease, ClientArchitecture, ClientPlatform, SoftwarePackageKind } from '@/utils/clientRelease'

export interface SoftwareDownloadAsset {
  name: string
  label: string
  platform: ClientPlatform
  architecture: ClientArchitecture
  format: string
  kind: SoftwarePackageKind
  url: string
  accelerated_url?: string
  size?: number
}

export interface SoftwareCenterItem {
  id: number
  source_type: 'github' | 'manual'
  source_url: string
  repository?: string
  repository_url?: string
  name: string
  description: string
  logo_url: string
  featured: boolean
  enabled: boolean
  sort_order: number
  supported_platforms: ClientPlatform[]
  asset_variants?: SoftwareDownloadAsset[]
  version?: string
  release_name?: string
  release_notes?: string
  published_at?: string
  release?: GitHubRelease
  release_fetched_at?: string
  last_error?: string
}

export interface SoftwareCenterPreview {
  repository: string
  repository_url: string
  name: string
  description: string
  logo_url: string
  supported_platforms: ClientPlatform[]
  release: GitHubRelease
}

export interface CreateSoftwareCenterItem {
  source_type: 'github' | 'manual'
  repository_url?: string
  source_url?: string
  name: string
  description: string
  logo_url: string
  featured: boolean
  enabled: boolean
  sort_order: number
  version?: string
  release_name?: string
  release_notes?: string
  published_at?: string | null
  asset_variants?: SoftwareDownloadAsset[]
}

export type UpdateSoftwareCenterItem = Partial<Omit<CreateSoftwareCenterItem, 'source_type' | 'repository_url'>>

export async function listSoftware() { const { data } = await apiClient.get<SoftwareCenterItem[]>('/admin/software-center'); return data }
export async function previewSoftware(repositoryUrl: string) { const { data } = await apiClient.post<SoftwareCenterPreview>('/admin/software-center/preview', { repository_url: repositoryUrl }); return data }
export async function createSoftware(input: CreateSoftwareCenterItem) { const { data } = await apiClient.post<SoftwareCenterItem>('/admin/software-center', input); return data }
export async function updateSoftware(id: number, input: UpdateSoftwareCenterItem) { const { data } = await apiClient.put<SoftwareCenterItem>(`/admin/software-center/${id}`, input); return data }
export async function deleteSoftware(id: number) { await apiClient.delete(`/admin/software-center/${id}`) }
export async function refreshSoftware(id: number) { const { data } = await apiClient.post<SoftwareCenterItem>(`/admin/software-center/${id}/refresh`); return data }
