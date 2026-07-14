<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.testAccountConnection')"
    width="normal"
    @close="handleClose"
  >
    <div class="space-y-4">
      <!-- Account Info Card -->
      <div
        v-if="account"
        class="flex min-w-0 items-center justify-between gap-3 rounded-panel border border-outline bg-surface-subtle p-3"
      >
        <div class="flex min-w-0 items-center gap-3">
          <div
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-control border border-outline bg-surface text-foreground-muted"
          >
            <Icon name="play" size="md" :stroke-width="2" />
          </div>
          <div class="min-w-0">
            <div class="truncate font-semibold text-foreground" :title="account.name">{{ account.name }}</div>
            <div class="flex flex-wrap items-center gap-1.5 text-xs text-foreground-subtle">
              <span
                class="rounded bg-outline px-1.5 py-0.5 text-[10px] font-medium uppercase bg-outline-strong"
              >
                {{ account.type }}
              </span>
              <span>{{ t('admin.accounts.account') }}</span>
            </div>
          </div>
        </div>
        <span
          :class="[
            'shrink-0 rounded-control border px-2.5 py-1 text-xs font-semibold',
            account.status === 'active'
              ? 'border-success/30 bg-success-subtle text-success-foreground'
              : 'border-outline bg-surface text-foreground-muted'
          ]"
        >
          {{ account.status }}
        </span>
      </div>

      <div class="space-y-1.5">
        <label class="text-sm font-medium text-foreground-muted">
          {{ t('admin.accounts.selectTestModel') }}
        </label>
        <Select
          v-model="selectedModelId"
          :options="availableModels"
          :disabled="loadingModels || status === 'connecting'"
          value-key="id"
          label-key="display_name"
          :placeholder="loadingModels ? t('common.loading') + '...' : t('admin.accounts.selectTestModel')"
        />
        <div
          v-if="modelsLoadFailed"
          class="mt-2 flex flex-col gap-2 rounded-panel border border-danger/30 bg-danger-subtle p-3 text-danger-foreground sm:flex-row sm:items-center sm:justify-between"
          role="alert"
          data-testid="account-test-models-error"
        >
          <span class="text-xs font-medium">{{ t('admin.accounts.testModelsLoadFailed') }}</span>
          <button type="button" class="btn btn-secondary btn-sm" :disabled="loadingModels" @click="loadAvailableModels">
            <Icon name="refresh" size="sm" :class="{ 'animate-spin': loadingModels }" />
            {{ t('common.retry') }}
          </button>
        </div>
      </div>

      <div v-if="isOpenAIAccount" class="space-y-1.5">
        <label class="text-sm font-medium text-foreground-muted">
          {{ t('admin.accounts.openai.testMode') }}
        </label>
        <Select
          v-model="testMode"
          :options="openAITestModeOptions"
          :disabled="status === 'connecting'"
        />
      </div>

      <div v-if="supportsImageTest" class="space-y-1.5">
        <TextArea
          v-model="testPrompt"
          :label="t('admin.accounts.imagePromptLabel')"
          :placeholder="t('admin.accounts.imagePromptPlaceholder')"
          :hint="t('admin.accounts.imageTestHint')"
          :disabled="status === 'connecting'"
          rows="3"
        />
      </div>

      <!-- Terminal Output -->
      <div class="group relative">
        <div
          ref="terminalRef"
          class="max-h-[240px] min-h-[120px] overflow-y-auto rounded-panel border border-outline-strong bg-foreground p-4 font-mono text-sm text-foreground-muted"
        >
          <!-- Status Line -->
          <div v-if="status === 'idle'" class="flex items-center gap-2 text-foreground-subtle">
            <Icon name="play" size="sm" :stroke-width="2" />
            <span>{{ t('admin.accounts.readyToTest') }}</span>
          </div>
          <div v-else-if="status === 'connecting'" class="flex items-center gap-2 text-warning-foreground">
            <Icon name="refresh" size="sm" class="animate-spin" :stroke-width="2" />
            <span>{{ t('admin.accounts.connectingToApi') }}</span>
          </div>

          <!-- Output Lines -->
          <div v-for="(line, index) in outputLines" :key="index" :class="line.class">
            {{ line.text }}
          </div>

          <!-- Streaming Content -->
          <div v-if="streamingContent" class="text-success-foreground">
            {{ streamingContent }}<span class="animate-pulse">_</span>
          </div>

          <!-- Result Status -->
          <div
            v-if="status === 'success'"
            class="mt-3 flex items-center gap-2 border-t border-outline-strong pt-3 text-success-foreground"
          >
            <Icon name="check" size="sm" :stroke-width="2" />
            <span>{{ t('admin.accounts.testCompleted') }}</span>
          </div>
          <div
            v-else-if="status === 'error'"
            class="mt-3 flex items-center gap-2 border-t border-outline-strong pt-3 text-danger-foreground"
          >
            <Icon name="x" size="sm" :stroke-width="2" />
            <span>{{ errorMessage }}</span>
          </div>
        </div>

        <!-- Copy Button -->
        <button
          v-if="outputLines.length > 0"
          type="button"
          @click="copyOutput"
          class="absolute right-2 top-2 flex h-8 w-8 items-center justify-center rounded-control bg-foreground text-foreground-subtle opacity-100 transition-colors hover:bg-surface/20 hover:text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white sm:opacity-0 sm:group-hover:opacity-100 sm:focus:opacity-100"
          :title="t('admin.accounts.copyOutput')"
          :aria-label="t('admin.accounts.copyOutput')"
        >
          <Icon name="copy" size="sm" :stroke-width="2" />
        </button>
      </div>

      <div v-if="generatedImages.length > 0" class="space-y-2">
        <div class="text-xs font-medium text-foreground-muted">
          {{ t('admin.accounts.imagePreview') }}
        </div>
        <div class="flex flex-wrap justify-center gap-3">
          <button
            v-for="(image, index) in generatedImages"
            :key="`${image.url}-${index}`"
            type="button"
            class="group/img relative min-w-0 overflow-hidden rounded-panel border border-outline bg-surface text-left transition-colors hover:border-outline-strong focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus"
            :aria-label="`${t('admin.accounts.imagePreview')} ${index + 1}`"
            @click="previewImageUrl = image.url"
          >
            <img :src="image.url" :alt="`test-image-${index + 1}`" class="max-h-[360px] w-full object-contain" />
            <div class="absolute inset-0 flex items-center justify-center bg-black/0 transition-colors group-hover/img:bg-black/20">
              <Icon name="eye" size="lg" class="text-white opacity-0 drop-shadow-floating transition-opacity group-hover/img:opacity-100" :stroke-width="2" />
            </div>
            <div class="border-t px-3 py-1.5 text-xs border-outline-strong text-foreground-muted">
              {{ image.mimeType || 'image/*' }}
            </div>
          </button>
        </div>
      </div>

      <!-- Image Lightbox -->
      <Teleport to="body">
        <Transition name="fade">
          <div
            v-if="previewImageUrl"
            class="fixed inset-0 z-[100] flex items-center justify-center bg-black/80 p-4"
            role="dialog"
            aria-modal="true"
            :aria-label="t('admin.accounts.imagePreview')"
            @click.self="previewImageUrl = ''"
            @keydown.esc="previewImageUrl = ''"
          >
            <button
              type="button"
              class="absolute right-4 top-4 flex h-10 w-10 items-center justify-center rounded-control bg-black/60 text-white transition-colors hover:bg-black/80 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white"
              :aria-label="t('common.close')"
              @click="previewImageUrl = ''"
            >
              <Icon name="x" size="lg" :stroke-width="2" />
            </button>
            <img
              :src="previewImageUrl"
              alt="preview"
              class="max-h-[90vh] max-w-[90vw] rounded-panel object-contain"
            />
          </div>
        </Transition>
      </Teleport>

      <!-- Test Info -->
      <div class="flex flex-wrap items-center justify-between gap-2 px-1 text-xs text-foreground-subtle">
        <div class="flex items-center gap-3">
          <span class="flex items-center gap-1">
            <Icon name="grid" size="sm" :stroke-width="2" />
            {{ t('admin.accounts.testModel') }}
          </span>
        </div>
        <span class="flex items-center gap-1">
          <Icon name="chat" size="sm" :stroke-width="2" />
          {{
            supportsImageTest
              ? t('admin.accounts.imageTestMode')
              : t('admin.accounts.testPrompt')
          }}
        </span>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button
          type="button"
          @click="handleClose"
          class="btn btn-secondary"
        >
          {{ t('common.close') }}
        </button>
        <button
          type="button"
          @click="startTest"
          :disabled="status === 'connecting' || !selectedModelId"
          :class="[
            'btn btn-primary',
            status === 'connecting' || !selectedModelId
              ? 'cursor-not-allowed'
              : status === 'success'
                ? 'bg-success text-white hover:bg-success/90'
                : status === 'error'
                  ? 'bg-warning text-warning-foreground hover:bg-warning/90'
                  : ''
          ]"
        >
          <Icon
            v-if="status === 'connecting'"
            name="refresh"
            size="sm"
            class="animate-spin"
            :stroke-width="2"
          />
          <Icon v-else-if="status === 'idle'" name="play" size="sm" :stroke-width="2" />
          <Icon v-else name="refresh" size="sm" :stroke-width="2" />
          <span>
            {{
              status === 'connecting'
                ? t('admin.accounts.testing')
                : status === 'idle'
                  ? t('admin.accounts.startTest')
                  : t('admin.accounts.retry')
            }}
          </span>
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch, nextTick, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import TextArea from '@/components/common/TextArea.vue'
import { Icon } from '@/components/icons'
import { useClipboard } from '@/composables/useClipboard'
import { buildApiUrl } from '@/api/client'
import { ADMIN_UI_REQUEST_HEADER } from '@/api/adminUIRequest'
import { adminAPI } from '@/api/admin'
import type { Account, ClaudeModel } from '@/types'

const { t } = useI18n()
const { copyToClipboard } = useClipboard()

interface OutputLine {
  text: string
  class: string
}

interface PreviewImage {
  url: string
  mimeType?: string
}

const props = defineProps<{
  show: boolean
  account: Account | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const terminalRef = ref<HTMLElement | null>(null)
const status = ref<'idle' | 'connecting' | 'success' | 'error'>('idle')
const outputLines = ref<OutputLine[]>([])
const streamingContent = ref('')
const errorMessage = ref('')
const availableModels = ref<ClaudeModel[]>([])
const selectedModelId = ref('')
const testPrompt = ref('')
const loadingModels = ref(false)
const modelsLoadFailed = ref(false)
let modelsLoadSequence = 0
let abortController: AbortController | null = null
const generatedImages = ref<PreviewImage[]>([])
const previewImageUrl = ref('')
const testMode = ref<'default' | 'compact'>('default')
const isOpenAIAccount = computed(() => props.account?.platform === 'openai')
const openAITestModeOptions = computed(() => [
  { value: 'default', label: t('admin.accounts.openai.testModeDefault') },
  { value: 'compact', label: t('admin.accounts.openai.testModeCompact') }
])
const prioritizedGeminiModels = ['gemini-3.1-flash-image', 'gemini-2.5-flash-image', 'gemini-3.5-flash', 'gemini-2.5-flash', 'gemini-2.5-pro', 'gemini-3-flash-preview', 'gemini-3-pro-preview', 'gemini-2.0-flash']
const supportsGeminiImageTest = computed(() => {
  const modelID = selectedModelId.value.toLowerCase()
  if (!modelID.startsWith('gemini-') || !modelID.includes('-image')) return false

  return props.account?.platform === 'gemini' || (props.account?.platform === 'antigravity' && props.account?.type === 'apikey')
})

const supportsOpenAIImageTest = computed(() => {
  const modelID = selectedModelId.value.toLowerCase()
  if (!modelID.startsWith('gpt-image-')) return false
  return props.account?.platform === 'openai'
})

const supportsImageTest = computed(() => supportsGeminiImageTest.value || supportsOpenAIImageTest.value)

const sortTestModels = (models: ClaudeModel[]) => {
  const priorityMap = new Map(prioritizedGeminiModels.map((id, index) => [id, index]))

  return [...models].sort((a, b) => {
    const aPriority = priorityMap.get(a.id) ?? Number.MAX_SAFE_INTEGER
    const bPriority = priorityMap.get(b.id) ?? Number.MAX_SAFE_INTEGER
    if (aPriority !== bPriority) return aPriority - bPriority
    return 0
  })
}

// Load available models when modal opens
watch(
  () => props.show,
  async (newVal) => {
    if (newVal && props.account) {
      testPrompt.value = ''
      testMode.value = 'default'
      resetState()
      await loadAvailableModels()
    } else {
      modelsLoadSequence += 1
      abortStream()
    }
  }
)

watch(selectedModelId, () => {
  if (supportsImageTest.value && !testPrompt.value.trim()) {
    testPrompt.value = t('admin.accounts.imagePromptDefault')
  }
})

const loadAvailableModels = async () => {
  if (!props.account) return

  const account = props.account
  const sequence = ++modelsLoadSequence
  loadingModels.value = true
  modelsLoadFailed.value = false
  availableModels.value = []
  selectedModelId.value = '' // Reset selection before loading
  try {
    const models = await adminAPI.accounts.getAvailableModels(account.id)
    if (sequence !== modelsLoadSequence) return
    availableModels.value = account.platform === 'gemini' || account.platform === 'antigravity'
      ? sortTestModels(models)
      : models
    // Default selection by platform
    if (availableModels.value.length > 0) {
      if (account.platform === 'gemini') {
        selectedModelId.value = availableModels.value[0].id
      } else {
        // Try to select Sonnet as default, otherwise use first model
        const sonnetModel = availableModels.value.find((m) => m.id.includes('sonnet'))
        selectedModelId.value = sonnetModel?.id || availableModels.value[0].id
      }
    }
  } catch (error) {
    if (sequence !== modelsLoadSequence) return
    console.error('Failed to load available models:', error)
    modelsLoadFailed.value = true
  } finally {
    if (sequence === modelsLoadSequence) loadingModels.value = false
  }
}

const resetState = () => {
  status.value = 'idle'
  outputLines.value = []
  streamingContent.value = ''
  errorMessage.value = ''
  generatedImages.value = []
  previewImageUrl.value = ''
}

const handleClose = () => {
  abortStream()
  emit('close')
}

const abortStream = () => {
  if (abortController) {
    abortController.abort()
    abortController = null
  }
}

onUnmounted(() => {
  modelsLoadSequence += 1
  abortStream()
})

const addLine = (text: string, className: string = 'text-foreground-subtle') => {
  outputLines.value.push({ text, class: className })
  scrollToBottom()
}

const scrollToBottom = async () => {
  await nextTick()
  if (terminalRef.value) {
    terminalRef.value.scrollTop = terminalRef.value.scrollHeight
  }
}

const startTest = async () => {
  if (!props.account || !selectedModelId.value) return

  resetState()
  status.value = 'connecting'
  addLine(t('admin.accounts.startingTestForAccount', { name: props.account.name }), 'text-info-foreground')
  addLine(t('admin.accounts.testAccountTypeLabel', { type: props.account.type }), 'text-foreground-subtle')
  addLine('', 'text-foreground-subtle')

  abortStream()

  abortController = new AbortController()

  try {
    const requestBody: {
      model_id: string
      prompt: string
      mode?: 'default' | 'compact'
    } = {
      model_id: selectedModelId.value,
      prompt: supportsImageTest.value ? testPrompt.value.trim() : ''
    }
    if (isOpenAIAccount.value) {
      requestBody.mode = testMode.value
    }

    // Use the configured API base; EventSource does not support POST.
    const url = buildApiUrl(`/admin/accounts/${props.account.id}/test`)

    // Use fetch with streaming for SSE since EventSource doesn't support POST
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${localStorage.getItem('auth_token')}`,
        'Content-Type': 'application/json',
        [ADMIN_UI_REQUEST_HEADER]: '1'
      },
      body: JSON.stringify(requestBody),
      signal: abortController.signal
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const reader = response.body?.getReader()
    if (!reader) {
      throw new Error('No response body')
    }

    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (line.startsWith('data: ')) {
          const jsonStr = line.slice(6).trim()
          if (jsonStr) {
            try {
              const event = JSON.parse(jsonStr)
              handleEvent(event)
            } catch (e) {
              console.error('Failed to parse SSE event:', e)
            }
          }
        }
      }
    }
  } catch (error: unknown) {
    if (error instanceof DOMException && error.name === 'AbortError') {
      status.value = 'idle'
      return
    }
    status.value = 'error'
    const msg = error instanceof Error ? error.message : 'Unknown error'
    errorMessage.value = msg
    addLine(`Error: ${msg}`, 'text-danger-foreground')
  }
}

const handleEvent = (event: {
  type: string
  text?: string
  model?: string
  success?: boolean
  error?: string
  image_url?: string
  mime_type?: string
}) => {
  switch (event.type) {
    case 'test_start':
      addLine(t('admin.accounts.connectedToApi'), 'text-success-foreground')
      if (event.model) {
        addLine(t('admin.accounts.usingModel', { model: event.model }), 'text-info-foreground')
      }
      addLine(
        supportsImageTest.value
            ? t('admin.accounts.sendingImageRequest')
            : t('admin.accounts.sendingTestMessage'),
        'text-foreground-subtle'
      )
      addLine('', 'text-foreground-subtle')
      addLine(t('admin.accounts.response'), 'text-warning-foreground')
      break

    case 'content':
      if (event.text) {
        streamingContent.value += event.text
        scrollToBottom()
      }
      break

    case 'image':
      if (event.image_url) {
        generatedImages.value.push({
          url: event.image_url,
          mimeType: event.mime_type
        })
        addLine(t('admin.accounts.imageReceived', { count: generatedImages.value.length }), 'text-brand')
      }
      break

    case 'status':
      if (event.text) {
        addLine(event.text, 'text-info-foreground')
      }
      break

    case 'test_complete':
      // Move streaming content to output lines
      if (streamingContent.value) {
        addLine(streamingContent.value, 'text-success-foreground')
        streamingContent.value = ''
      }
      if (event.success) {
        status.value = 'success'
      } else {
        status.value = 'error'
        errorMessage.value = event.error || 'Test failed'
      }
      break

    case 'error':
      status.value = 'error'
      errorMessage.value = event.error || 'Unknown error'
      if (streamingContent.value) {
        addLine(streamingContent.value, 'text-success-foreground')
        streamingContent.value = ''
      }
      break
  }
}

const copyOutput = () => {
  const text = outputLines.value.map((l) => l.text).join('\n')
  copyToClipboard(text, t('admin.accounts.outputCopied'))
}
</script>

<style>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
