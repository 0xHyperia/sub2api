<template>
  <BaseDialog
    :show="show"
    :title="title"
    width="wide"
    @close="$emit('close')"
  >
    <div v-if="loading" class="py-8 text-center text-sm text-foreground-subtle">
      {{ t('common.loading') }}
    </div>
    <div v-else-if="!detail" class="py-8 text-center text-sm text-foreground-subtle">
      {{ t('channelStatus.detailLoadError') }}
    </div>
    <div v-else role="region" :aria-label="title">
      <div class="space-y-2 sm:hidden">
        <article
          v-for="model in detail.models"
          :key="`mobile-${model.model}`"
          class="rounded-panel border border-outline bg-surface p-3"
        >
          <header class="flex min-w-0 items-start justify-between gap-3">
            <h3 class="min-w-0 break-words text-sm font-semibold text-foreground">{{ model.model }}</h3>
            <span
              class="inline-flex shrink-0 items-center rounded-full px-2 py-0.5 text-[11px]"
              :class="statusBadgeClass(model.latest_status)"
            >
              {{ statusLabel(model.latest_status) }}
            </span>
          </header>
          <div class="mt-3 flex items-end justify-between gap-3 border-b border-outline pb-3">
            <span class="text-[10px] text-foreground-subtle">{{ t('channelStatus.detailColumns.latestLatency') }}</span>
            <strong class="font-mono text-base tabular-nums text-foreground">{{ formatLatency(model.latest_latency_ms) }}</strong>
          </div>
          <dl class="mt-3 grid grid-cols-2 gap-x-3 gap-y-3">
            <div>
              <dt class="text-[10px] text-foreground-subtle">{{ t('channelStatus.detailColumns.availability7d') }}</dt>
              <dd class="mt-0.5 font-mono text-xs tabular-nums text-foreground">{{ formatPercent(model.availability_7d) }}</dd>
            </div>
            <div>
              <dt class="text-[10px] text-foreground-subtle">{{ t('channelStatus.detailColumns.availability15d') }}</dt>
              <dd class="mt-0.5 font-mono text-xs tabular-nums text-foreground">{{ formatPercent(model.availability_15d) }}</dd>
            </div>
            <div>
              <dt class="text-[10px] text-foreground-subtle">{{ t('channelStatus.detailColumns.availability30d') }}</dt>
              <dd class="mt-0.5 font-mono text-xs tabular-nums text-foreground">{{ formatPercent(model.availability_30d) }}</dd>
            </div>
            <div>
              <dt class="text-[10px] text-foreground-subtle">{{ t('channelStatus.detailColumns.avgLatency7d') }}</dt>
              <dd class="mt-0.5 font-mono text-xs tabular-nums text-foreground">{{ formatLatency(model.avg_latency_7d_ms) }}</dd>
            </div>
          </dl>
        </article>
      </div>

      <div class="hidden overflow-x-auto sm:block" tabindex="0">
      <table class="min-w-[760px] w-full text-left text-sm">
        <thead class="border-b border-outline">
          <tr class="text-xs uppercase tracking-wider text-foreground-subtle">
            <th class="py-2 pr-3">{{ t('channelStatus.detailColumns.model') }}</th>
            <th class="py-2 pr-3">{{ t('channelStatus.detailColumns.latestStatus') }}</th>
            <th class="py-2 pr-3">{{ t('channelStatus.detailColumns.latestLatency') }}</th>
            <th class="py-2 pr-3">{{ t('channelStatus.detailColumns.availability7d') }}</th>
            <th class="py-2 pr-3">{{ t('channelStatus.detailColumns.availability15d') }}</th>
            <th class="py-2 pr-3">{{ t('channelStatus.detailColumns.availability30d') }}</th>
            <th class="py-2 pr-3">{{ t('channelStatus.detailColumns.avgLatency7d') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="m in detail.models"
            :key="m.model"
            class="border-b border-outline"
          >
            <td class="py-2 pr-3 font-medium text-foreground">{{ m.model }}</td>
            <td class="py-2 pr-3">
              <span
                class="inline-flex items-center rounded-full px-2 py-0.5 text-[11px]"
                :class="statusBadgeClass(m.latest_status)"
              >
                {{ statusLabel(m.latest_status) }}
              </span>
            </td>
            <td class="py-2 pr-3 tabular-nums text-foreground-muted">{{ formatLatency(m.latest_latency_ms) }}</td>
            <td class="py-2 pr-3 tabular-nums text-foreground-muted">{{ formatPercent(m.availability_7d) }}</td>
            <td class="py-2 pr-3 tabular-nums text-foreground-muted">{{ formatPercent(m.availability_15d) }}</td>
            <td class="py-2 pr-3 tabular-nums text-foreground-muted">{{ formatPercent(m.availability_30d) }}</td>
            <td class="py-2 pr-3 tabular-nums text-foreground-muted">{{ formatLatency(m.avg_latency_7d_ms) }}</td>
          </tr>
        </tbody>
      </table>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end">
        <button type="button" @click="$emit('close')" class="btn btn-secondary">
          {{ t('channelStatus.closeDetail') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import {
  status as fetchChannelMonitorDetail,
  type UserMonitorDetail,
} from '@/api/channelMonitor'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { useChannelMonitorFormat } from '@/composables/useChannelMonitorFormat'

const props = defineProps<{
  show: boolean
  monitorId: number | null
  title: string
}>()

defineEmits<{
  (e: 'close'): void
}>()

const { t } = useI18n()
const appStore = useAppStore()
const { statusLabel, statusBadgeClass, formatLatency, formatPercent } = useChannelMonitorFormat()

const detail = ref<UserMonitorDetail | null>(null)
const loading = ref(false)

async function load(id: number) {
  detail.value = null
  loading.value = true
  try {
    detail.value = await fetchChannelMonitorDetail(id)
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('channelStatus.detailLoadError')))
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.show, props.monitorId] as const,
  ([show, id]) => {
    if (!show) {
      detail.value = null
      return
    }
    if (id != null) void load(id)
  },
  { immediate: true },
)
</script>
