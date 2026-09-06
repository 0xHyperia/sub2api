import type { GroupPlatform, ReasoningEffortMapping, ReasoningEffortMatchType } from '@/types'

const effortValues = ['minimal', 'low', 'medium', 'high', 'xhigh', 'max'] as const
const sourceValues = ['none', ...effortValues] as const
export const reasoningEffortOverLimitDowngrade = 'downgrade'
export const reasoningEffortOverLimitDeny = 'deny'

export interface ReasoningEffortMappingPair { id: string; from: string; to: string }
export interface ReasoningEffortMappingRow { id: string; match_type: ReasoningEffortMatchType | ''; model: string; pairs: ReasoningEffortMappingPair[] }
export type ReasoningEffortMappingErrorCode = 'fromRequired' | 'toRequired' | 'duplicateFrom' | 'duplicateScope' | 'unsupportedFrom' | 'unsupportedTo' | 'unsupportedMatchType'
export type ReasoningEffortMappingErrors = Record<string, Partial<Record<'from' | 'to' | 'match_type' | 'duplicateScope', ReasoningEffortMappingErrorCode>>>
let nextID = 0
const id = (prefix: string) => `${prefix}-${++nextID}`

export const supportsReasoningEffortPolicyPlatform = (platform: GroupPlatform) => platform === 'openai' || platform === 'composite'
export const reasoningEffortOptionsForPlatform = (platform: GroupPlatform) => (supportsReasoningEffortPolicyPlatform(platform) ? effortValues : []).map(value => ({ value, label: value }))
export const reasoningEffortSourceOptionsForPlatform = (platform: GroupPlatform) => (supportsReasoningEffortPolicyPlatform(platform) ? sourceValues : []).map(value => ({ value, label: value }))
export const normalizeReasoningEffortForPlatform = (platform: GroupPlatform, value?: string | null) => {
  const normalized = value?.trim().toLowerCase() ?? ''
  return reasoningEffortOptionsForPlatform(platform).some(option => option.value === normalized) ? normalized : ''
}
export const normalizeReasoningEffortSourceForPlatform = (platform: GroupPlatform, value?: string | null) => value?.trim().toLowerCase() === 'none' && supportsReasoningEffortPolicyPlatform(platform) ? 'none' : normalizeReasoningEffortForPlatform(platform, value)
export const normalizeReasoningEffortMatchType = (value?: string | null): ReasoningEffortMatchType | '' => ['exact', 'prefix', 'suffix'].includes(value?.trim().toLowerCase() ?? '') ? value!.trim().toLowerCase() as ReasoningEffortMatchType : ''

export function createReasoningEffortMappingPair(pair: Partial<Pick<ReasoningEffortMapping, 'from' | 'to'>> = {}): ReasoningEffortMappingPair { return { id: id('reasoning-effort-pair'), from: pair.from ?? '', to: pair.to ?? '' } }
export function createReasoningEffortMappingRow(mapping: Partial<ReasoningEffortMapping> = {}): ReasoningEffortMappingRow {
  const matchType = normalizeReasoningEffortMatchType(mapping.match_type)
  return { id: id('reasoning-effort-group'), match_type: matchType || '', model: mapping.model?.trim() ?? '', pairs: [createReasoningEffortMappingPair(mapping)] }
}
export function reasoningEffortMappingsToRows(mappings?: ReasoningEffortMapping[] | null, platform: GroupPlatform = 'openai'): ReasoningEffortMappingRow[] {
  const rows: ReasoningEffortMappingRow[] = []
  for (const mapping of mappings ?? []) {
    const from = normalizeReasoningEffortSourceForPlatform(platform, mapping.from); const to = normalizeReasoningEffortForPlatform(platform, mapping.to)
    if (!from || !to) continue
    const match_type = normalizeReasoningEffortMatchType(mapping.match_type); const model = mapping.model?.trim() ?? ''
    const existing = rows.find(row => row.match_type === match_type && row.model.toLowerCase() === model.toLowerCase())
    if (existing) existing.pairs.push(createReasoningEffortMappingPair({ from, to }))
    else {
      const row = createReasoningEffortMappingRow({ ...mapping, from, to, model })
      row.match_type = match_type
      rows.push(row)
    }
  }
  return rows
}
export const reasoningEffortMappingsToAPI = (rows: ReasoningEffortMappingRow[]): ReasoningEffortMapping[] => rows.flatMap(row => row.pairs.map(pair => ({ from: pair.from.trim(), to: pair.to.trim(), ...(row.match_type ? { match_type: row.match_type } : {}), ...(row.model.trim() ? { model: row.model.trim() } : {}) })))
export function validateReasoningEffortMappings(rows: ReasoningEffortMappingRow[], platform: GroupPlatform = 'openai'): ReasoningEffortMappingErrors {
  const errors: ReasoningEffortMappingErrors = {}; const scopes = new Map<string, string>(); const duplicateSources = new Map<string, string[]>()
  rows.forEach(row => {
    if (row.model.trim() && !row.match_type) errors[row.id] = { ...errors[row.id], match_type: 'unsupportedMatchType' }
    const scope = `${row.match_type}\0${row.model.trim().toLowerCase()}`
    if (row.model.trim()) { if (scopes.has(scope)) errors[row.id] = { ...errors[row.id], duplicateScope: 'duplicateScope' }; else scopes.set(scope, row.id) }
    row.pairs.forEach(pair => {
      const from = pair.from.trim(); const to = pair.to.trim()
      const target = row.pairs.length === 1 ? row.id : pair.id
      if (!from) errors[target] = { ...errors[target], from: 'fromRequired' }; else if (!normalizeReasoningEffortSourceForPlatform(platform, from)) errors[target] = { ...errors[target], from: 'unsupportedFrom' }; else { const key = `${row.model.trim() ? row.id : ''}\0${from.toLowerCase()}`; duplicateSources.set(key, [...(duplicateSources.get(key) ?? []), target]) }
      if (!to) errors[target] = { ...errors[target], to: 'toRequired' }; else if (!normalizeReasoningEffortForPlatform(platform, to)) errors[target] = { ...errors[target], to: 'unsupportedTo' }
    })
  })
  duplicateSources.forEach(ids => { if (ids.length > 1) ids.forEach(target => { errors[target] = { ...errors[target], from: 'duplicateFrom' } }) })
  return errors
}
