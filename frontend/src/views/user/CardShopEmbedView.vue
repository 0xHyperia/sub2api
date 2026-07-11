<template>
  <div class="card-shop-embed">
    <div v-if="!shopUrl" class="flex min-h-[420px] items-center justify-center px-6 text-center">
      <div>
        <Icon name="link" size="xl" class="mx-auto mb-3 text-gray-300 dark:text-dark-600" />
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('purchase.notConfiguredDesc') }}</p>
      </div>
    </div>

    <template v-else>
      <a
        :href="shopUrl"
        target="_blank"
        rel="noopener noreferrer"
        class="btn btn-secondary btn-sm card-shop-open"
      >
        <Icon name="externalLink" size="sm" />
        {{ t('purchase.openInNewTab') }}
      </a>
      <iframe
        :src="shopUrl"
        :title="t('payment.tabIframe')"
        class="card-shop-frame"
        allow="payment; clipboard-write"
        allowfullscreen
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'

const { t } = useI18n()
const appStore = useAppStore()
const shopUrl = computed(() => appStore.cachedPublicSettings?.purchase_subscription_url?.trim() || '')
</script>

<style scoped>
.card-shop-embed {
  position: relative;
  min-height: min(820px, calc(100vh - 190px));
  overflow: hidden;
  border: 1px solid rgb(229 231 235);
  border-radius: 8px;
  background: white;
}

.card-shop-open {
  position: absolute;
  z-index: 10;
  top: 14px;
  right: 14px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  box-shadow: 0 2px 8px rgb(15 23 42 / 10%);
}

.card-shop-frame {
  display: block;
  width: 100%;
  height: min(820px, calc(100vh - 190px));
  min-height: 640px;
  border: 0;
  background: white;
}

:global(.dark) .card-shop-embed {
  border-color: rgb(55 65 81);
  background: rgb(17 24 39);
}

@media (max-width: 767px) {
  .card-shop-embed,
  .card-shop-frame {
    min-height: calc(100vh - 170px);
    height: calc(100vh - 170px);
  }
}
</style>
