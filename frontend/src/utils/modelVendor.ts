export type ModelVendorIconKey =
  | 'openai'
  | 'claude'
  | 'gemini'
  | 'zhipu'
  | 'qwen'
  | 'deepseek'
  | 'mistral'
  | 'meta'
  | 'cohere'
  | 'yi'
  | 'xai'
  | 'moonshot'
  | 'doubao'
  | 'minimax'
  | 'wenxin'
  | 'spark'
  | 'hunyuan'
  | 'cloudflare'
  | 'midjourney'
  | 'perplexity'
  | 'jina'
  | 'openrouter'
  | 'suno'
  | 'ollama'
  | 'ai360'
  | 'dify'
  | 'coze'

const vendorAliases: Record<string, ModelVendorIconKey> = {
  openai: 'openai',
  azure: 'openai',
  anthropic: 'claude',
  claude: 'claude',
  google: 'gemini',
  gemini: 'gemini',
  vertex: 'gemini',
  vertex_ai: 'gemini',
  'vertex_ai-language-models': 'gemini',
  zhipu: 'zhipu',
  glm: 'zhipu',
  alibaba: 'qwen',
  dashscope: 'qwen',
  qwen: 'qwen',
  deepseek: 'deepseek',
  mistral: 'mistral',
  meta: 'meta',
  cohere: 'cohere',
  yi: 'yi',
  xai: 'xai',
  grok: 'xai',
  moonshot: 'moonshot',
  kimi: 'moonshot',
  bytedance: 'doubao',
  volcengine: 'doubao',
  doubao: 'doubao',
  minimax: 'minimax',
  baidu: 'wenxin',
  wenxin: 'wenxin',
  iflytek: 'spark',
  spark: 'spark',
  tencent: 'hunyuan',
  hunyuan: 'hunyuan',
  cloudflare: 'cloudflare',
  midjourney: 'midjourney',
  perplexity: 'perplexity',
  jina: 'jina',
  openrouter: 'openrouter',
  suno: 'suno',
  ollama: 'ollama',
  ai360: 'ai360',
  '360': 'ai360',
  dify: 'dify',
  coze: 'coze',
}

export function normalizeModelVendor(vendor?: string | null): ModelVendorIconKey | null {
  const normalized = vendor?.trim().toLowerCase()
  return normalized ? vendorAliases[normalized] ?? null : null
}

export function resolveModelVendor(model?: string | null): ModelVendorIconKey | null {
  const value = model?.trim().toLowerCase() ?? ''
  if (!value) return null

  if (/^(gpt|chatgpt|codex|o[1-9](?:[-_.]|$)|dall-e|whisper|tts-1|text-embedding-3|text-moderation)/.test(value)
    || /(babbage|davinci|curie|ada)/.test(value)) return 'openai'
  if (value.includes('claude')) return 'claude'
  if (/(gemini|gemma|learnlm|imagen-|veo-)/.test(value)) return 'gemini'
  if (/(chatglm|cogview|cogvideo|(^|[/_.-])glm)/.test(value)) return 'zhipu'
  if (/(qwen|qwq)/.test(value)) return 'qwen'
  if (value.includes('deepseek')) return 'deepseek'
  if (/(mistral|mixtral|codestral|pixtral|voxtral|magistral)/.test(value)) return 'mistral'
  if (value.includes('llama')) return 'meta'
  if (/(command|c4ai-|(^|[/_.-])embed-)/.test(value)) return 'cohere'
  if (/^yi(?:[- _]|$)/.test(value)) return 'yi'
  if (value.includes('grok')) return 'xai'
  if (/(moonshot|kimi)/.test(value)) return 'moonshot'
  if (value.includes('doubao')) return 'doubao'
  if (/(abab|minimax)/.test(value)) return 'minimax'
  if (/(ernie|wenxin)/.test(value)) return 'wenxin'
  if (value.includes('spark')) return 'spark'
  if (value.includes('hunyuan')) return 'hunyuan'
  if (value.includes('@cf/')) return 'cloudflare'
  if (/(mj_|midjourney)/.test(value)) return 'midjourney'
  if (/(perplexity|pplx)/.test(value)) return 'perplexity'
  if (value.includes('jina')) return 'jina'
  if (value.includes('openrouter')) return 'openrouter'
  if (value.includes('suno')) return 'suno'
  if (value.includes('ollama')) return 'ollama'
  if (value.includes('360')) return 'ai360'
  if (value.includes('dify')) return 'dify'
  if (value.includes('coze')) return 'coze'
  return null
}

export function resolveModelVendors(models?: readonly string[] | null): ModelVendorIconKey[] {
  const vendors = new Set<ModelVendorIconKey>()
  for (const model of models ?? []) {
    const vendor = resolveModelVendor(model)
    if (!vendor) return []
    vendors.add(vendor)
  }
  return [...vendors].sort()
}
