import { describe, expect, it } from 'vitest'
import { normalizeModelVendor, resolveModelVendor, resolveModelVendors } from '../modelVendor'

describe('modelVendor', () => {
  it.each([
    ['gpt-5.4', 'openai'],
    ['claude-opus-4-6', 'claude'],
    ['gemini-3-pro', 'gemini'],
    ['deepseek-chat', 'deepseek'],
    ['qwen3-max', 'qwen'],
    ['kimi-k2.5', 'moonshot'],
    ['grok-4', 'xai'],
  ])('resolves %s to %s', (model, vendor) => {
    expect(resolveModelVendor(model)).toBe(vendor)
  })

  it('normalizes provider aliases to model vendor icons', () => {
    expect(normalizeModelVendor('anthropic')).toBe('claude')
    expect(normalizeModelVendor('vertex_ai-language-models')).toBe('gemini')
    expect(normalizeModelVendor('dashscope')).toBe('qwen')
  })

  it('deduplicates vendors for group icon selection', () => {
    expect(resolveModelVendors(['gpt-5.4', 'gpt-4.1', 'claude-sonnet-4-6']))
      .toEqual(['claude', 'openai'])
    expect(resolveModelVendors(['unknown-model'])).toEqual([])
    expect(resolveModelVendors(['gpt-5.4', 'unknown-model'])).toEqual([])
  })
})
