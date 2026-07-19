<template>
  <div class="space-y-3 lg:hidden" role="region" :aria-label="columns.name">
    <template v-if="loading">
      <div
        v-for="index in 3"
        :key="index"
        class="animate-pulse rounded-panel border border-outline bg-surface p-4"
      >
        <div class="h-5 w-36 rounded bg-surface-subtle"></div>
        <div class="mt-2 h-3 w-full rounded bg-surface-subtle"></div>
        <div class="mt-4 h-11 rounded-panel bg-surface-subtle"></div>
      </div>
    </template>

    <div
      v-else-if="rows.length === 0"
      class="rounded-panel border border-outline bg-surface px-4 py-12 text-center"
    >
      <Icon name="inbox" size="xl" class="mx-auto mb-3 h-12 w-12 text-foreground-subtle" />
      <p class="text-sm text-foreground-subtle">{{ emptyLabel }}</p>
    </div>

    <article
      v-for="(channel, chIdx) in rows"
      v-else
      :key="`mobile-${channel.name}-${chIdx}`"
      class="overflow-hidden rounded-panel border border-outline bg-surface"
    >
      <header class="border-b border-outline px-4 py-3">
        <h2 class="text-sm font-semibold text-foreground">{{ channel.name }}</h2>
        <p v-if="channel.description" class="mt-1 text-xs leading-5 text-foreground-subtle">
          {{ channel.description }}
        </p>
      </header>

      <div class="divide-y divide-outline">
        <details
          v-for="section in channel.platforms"
          :key="`mobile-${channel.name}-${section.platform}`"
          class="group/platform bg-surface"
        >
          <summary
            class="flex min-h-14 cursor-pointer list-none items-center gap-3 px-4 py-2.5 transition-colors hover:bg-surface-subtle focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-focus"
          >
            <span
              :class="[
                'inline-flex shrink-0 items-center gap-1 rounded-control border px-2 py-0.5 text-[11px] font-medium uppercase',
                platformBadgeClass(section.platform),
              ]"
            >
              <PlatformIcon :platform="section.platform as GroupPlatform" size="xs" />
              {{ section.platform }}
            </span>
            <span class="min-w-0 flex-1 text-right text-[11px] text-foreground-subtle">
              {{ columns.groups }} {{ section.groups.length }} · {{ columns.supportedModels }} {{ section.supported_models.length }}
            </span>
            <Icon
              name="chevronDown"
              size="xs"
              class="shrink-0 text-foreground-subtle transition-transform group-open/platform:rotate-180"
            />
          </summary>

          <div class="space-y-4 border-t border-outline bg-surface-subtle/45 px-4 py-3">
            <section>
              <h3 class="mb-2 text-[11px] font-semibold uppercase text-foreground-muted">
                {{ columns.groups }}
              </h3>
              <div v-if="section.groups.length" class="space-y-2">
                <div
                  v-for="group in section.groups"
                  :key="`mobile-group-${group.id}`"
                  class="flex min-w-0 items-center gap-2"
                >
                  <Icon
                    :name="group.is_exclusive ? 'shield' : 'globe'"
                    size="xs"
                    class="h-3.5 w-3.5 shrink-0 text-foreground-subtle"
                    :title="group.is_exclusive ? t('availableChannels.exclusiveTooltip') : t('availableChannels.publicTooltip')"
                  />
                  <div class="min-w-0 flex-1">
                    <GroupBadge
                      :name="group.name"
                      :platform="group.platform as GroupPlatform"
                      :subscription-type="(group.subscription_type || 'standard') as SubscriptionType"
                      :rate-multiplier="group.rate_multiplier"
                      :user-rate-multiplier="userGroupRates[group.id] ?? null"
                      always-show-rate
                    />
                  </div>
                  <span
                    v-if="hasPeakRate(group)"
                    class="inline-flex shrink-0 items-center gap-1 rounded-control bg-warning-subtle px-1.5 py-0.5 text-[10px] font-medium text-warning-foreground"
                    :title="peakRateTitle(group)"
                  >
                    <Icon name="clock" size="xs" class="h-3 w-3" />
                    {{ peakRateLabel(group) }}
                  </span>
                </div>
              </div>
              <span v-else class="text-xs text-foreground-subtle">-</span>
            </section>

            <section>
              <h3 class="mb-2 text-[11px] font-semibold uppercase text-foreground-muted">
                {{ columns.supportedModels }}
              </h3>
              <div v-if="section.supported_models.length" class="flex flex-wrap gap-1.5">
                <SupportedModelChip
                  v-for="model in section.supported_models"
                  :key="`mobile-model-${section.platform}-${model.name}`"
                  :model="model"
                  :pricing-key-prefix="pricingKeyPrefix"
                  :no-pricing-label="noPricingLabel"
                  :show-platform="false"
                  :platform-hint="section.platform"
                />
              </div>
              <span v-else class="text-xs text-foreground-subtle">{{ noModelsLabel }}</span>
            </section>
          </div>
        </details>
      </div>
    </article>
  </div>

  <div class="table-wrapper hidden max-w-full overflow-x-auto lg:block" role="region" :aria-label="columns.name" tabindex="0">
    <table class="min-w-[960px] w-full border-collapse text-sm">
      <thead>
        <tr class="border-b border-outline text-xs font-medium uppercase tracking-wide text-foreground-subtle bg-surface/50">
          <th class="w-[180px] px-4 py-3 text-center">{{ columns.name }}</th>
          <th class="w-[200px] px-4 py-3 text-left">{{ columns.description }}</th>
          <th class="w-[140px] px-4 py-3 text-left">{{ columns.platform }}</th>
          <th class="px-4 py-3 text-left">{{ columns.groups }}</th>
          <th class="px-4 py-3 text-left">{{ columns.supportedModels }}</th>
        </tr>
      </thead>
      <tbody v-if="loading">
        <tr>
          <td colspan="5" class="py-10 text-center">
            <Icon name="refresh" size="lg" class="inline-block animate-spin text-foreground-subtle" />
          </td>
        </tr>
      </tbody>
      <tbody v-else-if="rows.length === 0">
        <tr>
          <td colspan="5" class="py-12 text-center">
            <Icon name="inbox" size="xl" class="mx-auto mb-3 h-12 w-12 text-foreground-subtle" />
            <p class="text-sm text-foreground-subtle">{{ emptyLabel }}</p>
          </td>
        </tr>
      </tbody>
      <!-- 每个渠道一个 tbody：首行 td rowspan 渠道名，后续行只渲染其余三列。
           tbody 之间强分隔线表达"渠道边界"，tbody 内部用淡分隔线区分平台。 -->
      <tbody
        v-else
        v-for="(channel, chIdx) in rows"
        :key="`${channel.name}-${chIdx}`"
        class="border-b-2 last:border-b-0 border-outline-strong"
      >
        <tr
          v-for="(section, secIdx) in channel.platforms"
          :key="`${channel.name}-${section.platform}`"
          class="transition-colors hover:bg-surface-subtle"
          :class="{ 'border-t border-outline/50': secIdx > 0 }"
        >
          <!-- 渠道名：只在第一行渲染并用 rowspan 纵向合并 -->
          <td
            v-if="secIdx === 0"
            :rowspan="channel.platforms.length"
            class="px-4 py-3 text-center align-middle font-medium text-foreground"
          >
            {{ channel.name }}
          </td>

          <!-- 描述：独立一列，同样用 rowspan 纵向合并 -->
          <td
            v-if="secIdx === 0"
            :rowspan="channel.platforms.length"
            class="px-4 py-3 align-middle text-xs text-foreground-subtle"
          >
            <template v-if="channel.description">{{ channel.description }}</template>
            <span v-else class="text-foreground-subtle">-</span>
          </td>

          <!-- 平台徽章 -->
          <td class="align-top px-4 py-3">
            <span
              :class="[
                'inline-flex items-center gap-1 rounded-control border px-2 py-0.5 text-[11px] font-medium uppercase',
                platformBadgeClass(section.platform),
              ]"
            >
              <PlatformIcon :platform="section.platform as GroupPlatform" size="xs" />
              {{ section.platform }}
            </span>
          </td>

          <!-- 分组：专属分组在前（紫色 shield 行），公开分组在后（灰色 globe 行）。 -->
          <td class="align-top px-4 py-3">
            <div class="flex flex-col gap-1.5">
              <div
                v-if="exclusiveGroups(section).length > 0"
                class="flex flex-wrap items-center gap-1.5"
              >
                <span
                  class="inline-flex items-center gap-0.5 text-[10px] font-medium uppercase text-foreground-muted"
                  :title="t('availableChannels.exclusiveTooltip')"
                >
                  <Icon name="shield" size="xs" class="h-3 w-3" />
                  {{ t('availableChannels.exclusive') }}
                </span>
                <div
                  v-for="g in exclusiveGroups(section)"
                  :key="`ex-${g.id}`"
                  class="inline-flex flex-wrap items-center gap-1"
                >
                  <GroupBadge
                    :name="g.name"
                    :platform="g.platform as GroupPlatform"
                    :subscription-type="(g.subscription_type || 'standard') as SubscriptionType"
                    :rate-multiplier="g.rate_multiplier"
                    :user-rate-multiplier="userGroupRates[g.id] ?? null"
                    always-show-rate
                  />
                  <span
                    v-if="hasPeakRate(g)"
                    class="inline-flex items-center gap-1 rounded-control bg-warning-subtle px-1.5 py-0.5 text-[10px] font-medium text-warning-foreground"
                    :title="peakRateTitle(g)"
                  >
                    <Icon name="clock" size="xs" class="h-3 w-3" />
                    {{ peakRateLabel(g) }}
                  </span>
                </div>
              </div>
              <div
                v-if="publicGroups(section).length > 0"
                class="flex flex-wrap items-center gap-1.5"
              >
                <span
                  class="inline-flex items-center gap-0.5 text-[10px] font-medium uppercase text-foreground-subtle"
                  :title="t('availableChannels.publicTooltip')"
                >
                  <Icon name="globe" size="xs" class="h-3 w-3" />
                  {{ t('availableChannels.public') }}
                </span>
                <div
                  v-for="g in publicGroups(section)"
                  :key="`pub-${g.id}`"
                  class="inline-flex flex-wrap items-center gap-1"
                >
                  <GroupBadge
                    :name="g.name"
                    :platform="g.platform as GroupPlatform"
                    :subscription-type="(g.subscription_type || 'standard') as SubscriptionType"
                    :rate-multiplier="g.rate_multiplier"
                    :user-rate-multiplier="userGroupRates[g.id] ?? null"
                    always-show-rate
                  />
                  <span
                    v-if="hasPeakRate(g)"
                    class="inline-flex items-center gap-1 rounded-control bg-warning-subtle px-1.5 py-0.5 text-[10px] font-medium text-warning-foreground"
                    :title="peakRateTitle(g)"
                  >
                    <Icon name="clock" size="xs" class="h-3 w-3" />
                    {{ peakRateLabel(g) }}
                  </span>
                </div>
              </div>
              <span v-if="section.groups.length === 0" class="text-xs text-foreground-subtle">-</span>
            </div>
          </td>

          <!-- 支持模型 -->
          <td class="align-top px-4 py-3">
            <div class="flex flex-wrap gap-1">
              <SupportedModelChip
                v-for="m in section.supported_models"
                :key="`${section.platform}-${m.name}`"
                :model="m"
                :pricing-key-prefix="pricingKeyPrefix"
                :no-pricing-label="noPricingLabel"
                :show-platform="false"
                :platform-hint="section.platform"
              />
              <span v-if="section.supported_models.length === 0" class="text-xs text-foreground-subtle">
                {{ noModelsLabel }}
              </span>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import SupportedModelChip from './SupportedModelChip.vue'
import type { UserAvailableChannel, UserAvailableGroup, UserChannelPlatformSection } from '@/api/channels'
import type { GroupPlatform, SubscriptionType } from '@/types'
import { platformBadgeClass } from '@/utils/platformColors'
import { useAppStore } from '@/stores/app'
import { hasPeakRate as groupHasPeakRate, formatPeakRateWindow, serverTimezoneLabel } from '@/utils/peak-rate'

const props = defineProps<{
  columns: {
    name: string
    description: string
    platform: string
    groups: string
    supportedModels: string
  }
  rows: UserAvailableChannel[]
  loading: boolean
  pricingKeyPrefix: string
  noPricingLabel: string
  noModelsLabel: string
  emptyLabel: string
  /** 用户专属倍率（group_id → multiplier）；无专属时由 GroupBadge 仅显示默认倍率。 */
  userGroupRates: Record<number, number>
}>()

// Suppress unused warning — props is accessed via template automatically but
// the explicit reference here keeps the linter from flagging userGroupRates.
void props.userGroupRates

const { t } = useI18n()

function exclusiveGroups(section: UserChannelPlatformSection): UserAvailableGroup[] {
  return section.groups.filter((g) => g.is_exclusive)
}

function publicGroups(section: UserChannelPlatformSection): UserAvailableGroup[] {
  return section.groups.filter((g) => !g.is_exclusive)
}

const appStore = useAppStore()

function hasPeakRate(group: UserAvailableGroup): boolean {
  return groupHasPeakRate(group)
}

function peakRateLabel(group: UserAvailableGroup): string {
  return formatPeakRateWindow(group, serverTimezoneLabel(appStore.cachedPublicSettings?.server_utc_offset))
}

function peakRateTitle(group: UserAvailableGroup): string {
  return t('common.peakRateTooltip', { window: peakRateLabel(group) }) + t('common.peakRateImageNote')
}
</script>
