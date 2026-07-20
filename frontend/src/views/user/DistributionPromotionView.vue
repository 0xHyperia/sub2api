<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-[1440px] space-y-5">
      <header class="flex min-w-0 items-center justify-between gap-3">
        <div class="min-w-0"><h1 class="page-title">推广中心</h1><p class="page-description">分享专属链接并发展客户</p></div>
        <button class="btn btn-secondary btn-icon shrink-0" :disabled="loading" title="刷新" aria-label="刷新" @click="load">
          <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
        </button>
      </header>
      <DistributionNav />

      <div v-if="loading && !overview" class="card flex min-h-48 items-center justify-center"><LoadingSpinner /></div>
      <template v-else-if="overview">
        <section v-if="!promotionEnabled" class="rounded-panel border border-warning/40 bg-warning-subtle p-4 text-sm text-warning-foreground">
          当前代理状态为“{{ statusText }}”，推广链接暂时不能绑定新客户。
        </section>

        <section class="grid min-w-0 gap-4 lg:grid-cols-[280px_minmax(0,1fr)]">
          <div class="card min-w-0 flex flex-col items-center p-5 text-center">
            <h2 class="text-base font-semibold">推广二维码</h2>
            <p class="mt-1 text-sm text-foreground-subtle">客户扫码后进入注册页面</p>
            <div class="mt-4 flex size-52 items-center justify-center rounded-panel border border-outline bg-white p-3">
              <canvas ref="qrCanvas" class="size-full" aria-label="推广链接二维码"></canvas>
            </div>
            <button class="btn btn-secondary mt-4 w-full" :disabled="!promotionEnabled" @click="downloadQr">
              <Icon name="download" size="sm" />下载二维码
            </button>
          </div>

          <div class="card min-w-0 p-4 sm:p-5">
            <div class="flex items-start justify-between gap-3">
              <div><h2 class="text-base font-semibold">专属推广信息</h2><p class="mt-1 text-sm text-foreground-subtle">返佣比例 {{ rate }}</p></div>
              <span class="badge" :class="promotionEnabled ? 'badge-success' : 'badge-warning'">{{ promotionEnabled ? '可推广' : statusText }}</span>
            </div>

            <div class="mt-5 space-y-4">
              <div>
                <p class="input-label">推广码</p>
                <div class="flex min-w-0 items-center gap-2 rounded-control border border-outline bg-surface-subtle p-2 pl-3">
                  <code class="min-w-0 flex-1 truncate font-mono text-sm">{{ overview.agent.promotion_code }}</code>
                  <button class="btn btn-secondary btn-icon shrink-0" title="复制推广码" aria-label="复制推广码" :disabled="!promotionEnabled" @click="copyText(overview.agent.promotion_code, '推广码已复制')"><Icon name="copy" size="sm" /></button>
                </div>
              </div>
              <div>
                <p class="input-label">推广链接</p>
                <div class="flex min-w-0 items-center gap-2 rounded-control border border-outline bg-surface-subtle p-2 pl-3">
                  <code class="min-w-0 flex-1 truncate text-sm" :title="promotionLink">{{ promotionLink }}</code>
                  <button class="btn btn-primary btn-icon shrink-0" title="复制推广链接" aria-label="复制推广链接" :disabled="!promotionEnabled" @click="copyText(promotionLink, '推广链接已复制')"><Icon name="copy" size="sm" /></button>
                </div>
              </div>
            </div>

            <dl class="mt-5 grid grid-cols-2 gap-4 border-t border-outline pt-4 sm:grid-cols-3">
              <div><dt class="text-xs text-foreground-subtle">直属客户</dt><dd class="mt-1 font-semibold">{{ overview.customer_count }}</dd></div>
              <div><dt class="text-xs text-foreground-subtle">付费客户</dt><dd class="mt-1 font-semibold">{{ overview.paying_customer_count }}</dd></div>
              <div><dt class="text-xs text-foreground-subtle">累计实付</dt><dd class="mt-1 font-semibold">¥{{ Number(overview.customer_paid_cny).toFixed(2) }}</dd></div>
            </dl>
          </div>
        </section>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import QRCode from 'qrcode'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import DistributionNav from '@/components/distribution/DistributionNav.vue'
import { getDistributionOverview, type DistributionOverview } from '@/api/distribution'
import { useAppStore } from '@/stores/app'

const app = useAppStore()
const overview = ref<DistributionOverview | null>(null)
const loading = ref(false)
const qrCanvas = ref<HTMLCanvasElement | null>(null)
const promotionLink = computed(() => overview.value ? `${window.location.origin}/register?agent=${encodeURIComponent(overview.value.agent.promotion_code)}` : '')
const promotionEnabled = computed(() => overview.value?.agent.status === 'active')
const statusText = computed(() => overview.value?.agent.status === 'suspended' ? '已暂停' : '已撤销')
const rate = computed(() => `${((overview.value?.agent.effective_rate_bps || 0) / 100).toFixed(2)}%`)

async function renderQr() {
  await nextTick()
  if (!qrCanvas.value || !promotionLink.value) return
  await QRCode.toCanvas(qrCanvas.value, promotionLink.value, { width: 184, margin: 1, errorCorrectionLevel: 'M' })
}

async function load() {
  loading.value = true
  try {
    overview.value = await getDistributionOverview()
    await renderQr()
  } catch {
    app.showError('加载推广信息失败')
  } finally {
    loading.value = false
  }
}

async function copyText(value: string, message: string) {
  await navigator.clipboard.writeText(value)
  app.showSuccess(message)
}

function downloadQr() {
  if (!qrCanvas.value) return
  const anchor = document.createElement('a')
  anchor.download = `agent-${overview.value?.agent.promotion_code || 'promotion'}.png`
  anchor.href = qrCanvas.value.toDataURL('image/png')
  anchor.click()
}

onMounted(load)
</script>
