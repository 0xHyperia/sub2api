<template>
  <div>
    <div class="mb-2 flex items-center justify-between gap-2">
      <label class="input-label mb-0">{{ t('admin.accounts.opencodeGo.protocolRules.title') }}</label>
      <button
        type="button"
        class="text-xs text-brand hover:text-brand "
        @click="restoreDefaults"
      >
        {{ t('admin.accounts.opencodeGo.protocolRules.restoreDefaults') }}
      </button>
    </div>
    <p class="input-hint mb-2">{{ t('admin.accounts.opencodeGo.protocolRules.hint') }}</p>
    <div v-if="rows.length > 0" class="mb-2 space-y-2">
      <div
        v-for="(row, index) in rows"
        :key="getRowKey(row)"
        class="flex items-center gap-2"
      >
        <input
          v-model="row.pattern"
          type="text"
          class="input flex-1 font-mono text-sm"
          :placeholder="t('admin.accounts.opencodeGo.protocolRules.patternPlaceholder')"
          :data-testid="`opencode-go-protocol-pattern-${index}`"
        />
        <select
          v-model="row.protocol"
          class="input w-44 shrink-0"
          :data-testid="`opencode-go-protocol-select-${index}`"
        >
          <option value="chat_completions">
            {{ t('admin.accounts.cnProviders.apiProtocol.chatCompletions') }}
          </option>
          <option value="responses">
            {{ t('admin.accounts.cnProviders.apiProtocol.responses') }}
          </option>
          <option value="anthropic">
            {{ t('admin.accounts.cnProviders.apiProtocol.anthropic') }}
          </option>
        </select>
        <button
          type="button"
          class="rounded-lg p-2 text-danger-foreground transition-colors hover:bg-danger-subtle hover:text-danger-foreground "
          :aria-label="t('admin.accounts.opencodeGo.protocolRules.remove')"
          @click="removeRow(index)"
        >
          <Icon name="trash" size="sm" />
        </button>
      </div>
    </div>
    <div
      class="mb-2 flex items-center gap-2 rounded-lg border border-dashed border-outline bg-surface-subtle px-3 py-2 text-xs text-foreground-subtle "
      data-testid="opencode-go-protocol-fallback"
    >
      <span class="flex-1 font-mono">*</span>
      <span>{{ t('admin.accounts.opencodeGo.protocolRules.fallback') }}</span>
    </div>
    <button
      type="button"
      class="w-full rounded-lg border-2 border-dashed border-outline px-4 py-2 text-foreground-muted transition-colors hover:border-outline hover:text-foreground-muted "
      data-testid="opencode-go-protocol-add-rule"
      @click="addRow"
    >
      {{ t('admin.accounts.opencodeGo.protocolRules.add') }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { createStableObjectKeyResolver } from '@/utils/stableObjectKey'
import {
  cloneOpenCodeGoProtocolRules,
  defaultOpenCodeProtocolRules,
  type OpenCodeAccountMode,
  type OpenCodeGoProtocolRule
} from '@/components/account/credentialsBuilder'

const props = withDefaults(defineProps<{
  rows: OpenCodeGoProtocolRule[]
  plan?: OpenCodeAccountMode
}>(), {
  plan: 'go'
})

const emit = defineEmits<{
  (e: 'update:rows', rows: OpenCodeGoProtocolRule[]): void
}>()

const { t } = useI18n()
const getRowKey = createStableObjectKeyResolver<OpenCodeGoProtocolRule>('opencode-go-protocol-rule')

const addRow = () => {
  emit('update:rows', [...props.rows, { pattern: '', protocol: 'chat_completions' }])
}

const removeRow = (index: number) => {
  emit('update:rows', props.rows.filter((_, i) => i !== index))
}

const restoreDefaults = () => {
  emit('update:rows', cloneOpenCodeGoProtocolRules(defaultOpenCodeProtocolRules(props.plan)))
}
</script>
