<template>
  <div class="space-y-4">
    <section
      data-testid="profile-overview-hero"
      class="card min-w-0 overflow-hidden"
    >
      <div class="flex min-w-0 flex-col gap-4 p-4 sm:flex-row sm:items-center sm:p-5">
        <div
          class="flex h-16 w-16 shrink-0 items-center justify-center overflow-hidden rounded-panel border border-outline-strong bg-brand text-xl font-semibold text-brand-foreground"
        >
          <img
            v-if="avatarUrl"
            :src="avatarUrl"
            :alt="displayName"
            class="h-full w-full object-cover"
          >
          <span v-else>{{ avatarInitial }}</span>
        </div>

        <div class="min-w-0 flex-1">
          <div class="flex min-w-0 flex-wrap items-center gap-2">
            <h2 class="min-w-0 break-words text-lg font-semibold text-foreground sm:text-xl">
              {{ displayName }}
            </h2>
            <span :class="['badge', user?.role === 'admin' ? 'badge-primary' : 'badge-gray']">
              {{ user?.role === 'admin' ? t('profile.administrator') : t('profile.user') }}
            </span>
            <span :class="['badge', user?.status === 'active' ? 'badge-success' : 'badge-danger']">
              {{ user?.status === 'active' ? t('common.active') : t('common.disabled') }}
            </span>
          </div>
          <p v-if="primaryEmailDisplay" class="mt-1 break-all text-sm text-foreground-muted">
            {{ primaryEmailDisplay }}
          </p>
        </div>
      </div>

      <dl class="profile-overview-metrics">
        <div data-testid="profile-overview-metric-balance">
          <dt>{{ t('profile.accountBalance') }}</dt>
          <dd>{{ formatCurrency(user?.balance || 0) }}</dd>
        </div>
        <div data-testid="profile-overview-metric-concurrency">
          <dt>{{ t('profile.concurrencyLimit') }}</dt>
          <dd>{{ user?.concurrency || 0 }}</dd>
        </div>
        <div data-testid="profile-overview-metric-member-since">
          <dt>{{ t('profile.memberSince') }}</dt>
          <dd>{{ memberSinceLabel }}</dd>
        </div>
      </dl>
    </section>

    <div data-testid="profile-main-column" class="space-y-4">
      <section data-testid="profile-basics-panel" class="card min-w-0 overflow-hidden">
        <header class="card-header">
          <h3 class="text-base font-semibold text-foreground">{{ t('profile.basicsTitle') }}</h3>
          <p class="mt-1 text-sm text-foreground-muted">{{ t('profile.basicsDescription') }}</p>
        </header>

        <div class="profile-basics-grid">
          <div
            class="min-w-0 p-4 sm:p-5"
          >
            <ProfileAvatarCard :user="user" embedded />
          </div>

          <div class="min-w-0 border-t border-outline p-4 sm:p-5 md:border-l md:border-t-0">
            <ProfileEditForm :initial-username="user?.username || ''" embedded />
          </div>
        </div>
      </section>

      <section data-testid="profile-auth-bindings-panel" class="card min-w-0 overflow-hidden p-4 sm:p-5">
        <ProfileIdentityBindingsSection
          :user="user"
          :linuxdo-enabled="linuxdoEnabled"
          :dingtalk-enabled="dingtalkEnabled"
          :oidc-enabled="oidcEnabled"
          :oidc-provider-name="oidcProviderName"
          :wechat-enabled="wechatEnabled"
          :wechat-open-enabled="wechatOpenEnabled"
          :wechat-mp-enabled="wechatMpEnabled"
          embedded
          compact
        />
      </section>
    </div>

    <div data-testid="profile-side-column">
      <section v-if="sourceHints.length" class="card min-w-0 overflow-hidden">
        <header class="card-header">
          <h3 class="text-base font-semibold text-foreground">{{ t('profile.linkedProfileSources') }}</h3>
          <p class="mt-1 text-sm text-foreground-muted">{{ t('profile.linkedProfileSourcesDescription') }}</p>
        </header>

        <ul class="divide-y divide-outline">
          <li v-for="hint in sourceHints" :key="hint.key" class="flex min-w-0 items-start gap-3 px-4 py-3 text-sm text-foreground-muted sm:px-5">
            <Icon name="link" size="sm" class="mt-0.5 shrink-0 text-foreground-subtle" aria-hidden="true" />
            <span class="min-w-0 break-words">{{ hint.text }}</span>
          </li>
        </ul>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import ProfileAvatarCard from '@/components/user/profile/ProfileAvatarCard.vue'
import ProfileEditForm from '@/components/user/profile/ProfileEditForm.vue'
import ProfileIdentityBindingsSection from '@/components/user/profile/ProfileIdentityBindingsSection.vue'
import type { User, UserAuthBindingStatus, UserAuthProvider, UserProfileSourceContext } from '@/types'

const props = withDefaults(defineProps<{
  user: User | null
  linuxdoEnabled?: boolean
  dingtalkEnabled?: boolean
  oidcEnabled?: boolean
  oidcProviderName?: string
  wechatEnabled?: boolean
  wechatOpenEnabled?: boolean
  wechatMpEnabled?: boolean
}>(), {
  linuxdoEnabled: false,
  dingtalkEnabled: false,
  oidcEnabled: false,
  oidcProviderName: 'OIDC',
  wechatEnabled: false,
  wechatOpenEnabled: undefined,
  wechatMpEnabled: undefined,
})

const { t } = useI18n()

function normalizeBindingStatus(binding: boolean | UserAuthBindingStatus | undefined): boolean | null {
  if (typeof binding === 'boolean') {
    return binding
  }
  if (!binding) {
    return null
  }
  if (typeof binding.bound === 'boolean') {
    return binding.bound
  }
  return Boolean(binding.provider_subject || binding.issuer || binding.provider_key)
}

function isEmailBound(user: User | null | undefined): boolean {
  if (typeof user?.email_bound === 'boolean') {
    return user.email_bound
  }

  const nested = user?.auth_bindings?.email ?? user?.identity_bindings?.email
  const normalized = normalizeBindingStatus(nested)
  return normalized ?? false
}

const avatarUrl = computed(() => props.user?.avatar_url?.trim() || '')
const displayName = computed(() => props.user?.username?.trim() || props.user?.email?.trim() || t('profile.user'))
const primaryEmailDisplay = computed(() => {
  const email = props.user?.email?.trim() || ''
  if (!email) {
    return ''
  }
  if (email.endsWith('.invalid') && !isEmailBound(props.user)) {
    return ''
  }
  return email
})
const avatarInitial = computed(() => displayName.value.charAt(0).toUpperCase() || 'U')
const memberSinceLabel = computed(() => {
  const raw = props.user?.created_at?.trim()
  if (!raw) {
    return '-'
  }

  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) {
    return '-'
  }

  return new Intl.DateTimeFormat(undefined, {
    year: 'numeric',
    month: 'short',
  }).format(date)
})

const providerLabels = computed<Record<UserAuthProvider, string>>(() => ({
  email: t('profile.authBindings.providers.email'),
  linuxdo: t('profile.authBindings.providers.linuxdo'),
  dingtalk: t('profile.authBindings.providers.dingtalk'),
  oidc: t('profile.authBindings.providers.oidc', { providerName: props.oidcProviderName }),
  wechat: t('profile.authBindings.providers.wechat'),
  github: 'GitHub',
  google: 'Google'
}))

function formatCurrency(value: number): string {
  return `$${value.toFixed(2)}`
}

function normalizeProvider(value: string): UserAuthProvider | null {
  const normalized = value.trim().toLowerCase()
  if (
    normalized === 'email' ||
    normalized === 'linuxdo' ||
    normalized === 'dingtalk' ||
    normalized === 'wechat' ||
    normalized === 'github' ||
    normalized === 'google'
  ) {
    return normalized
  }
  if (normalized === 'oidc' || normalized.startsWith('oidc:') || normalized.startsWith('oidc/')) {
    return 'oidc'
  }
  return null
}

function readObjectString(source: Record<string, unknown>, ...keys: string[]): string {
  for (const key of keys) {
    const value = source[key]
    if (typeof value === 'string' && value.trim()) {
      return value.trim()
    }
  }
  return ''
}

function resolveThirdPartySource(
  rawSource: string | UserProfileSourceContext | null | undefined
): { provider: UserAuthProvider; label: string } | null {
  if (!rawSource) {
    return null
  }

  if (typeof rawSource === 'string') {
    const provider = normalizeProvider(rawSource)
    if (!provider || provider === 'email') {
      return null
    }
    return {
      provider,
      label: providerLabels.value[provider]
    }
  }

  const sourceRecord = rawSource as Record<string, unknown>
  const provider = normalizeProvider(
    readObjectString(sourceRecord, 'provider', 'source', 'provider_type', 'auth_provider')
  )
  if (!provider || provider === 'email') {
    return null
  }

  const explicitLabel = readObjectString(
    sourceRecord,
    'provider_label',
    'label',
    'provider_name',
    'providerName'
  )

  return {
    provider,
    label: explicitLabel || providerLabels.value[provider]
  }
}

const sourceHints = computed(() => {
  const currentUser = props.user
  if (!currentUser) {
    return []
  }

  const hints: Array<{ key: string; text: string }> = []
  const avatarSource = resolveThirdPartySource(
    currentUser.profile_sources?.avatar ?? currentUser.avatar_source
  )
  const usernameSource = resolveThirdPartySource(
    currentUser.profile_sources?.username ??
      currentUser.profile_sources?.display_name ??
      currentUser.profile_sources?.nickname ??
      currentUser.display_name_source ??
      currentUser.username_source ??
      currentUser.nickname_source
  )

  if (avatarSource) {
    hints.push({
      key: 'avatar',
      text: t('profile.authBindings.source.avatar', { providerName: avatarSource.label })
    })
  }

  if (usernameSource) {
    hints.push({
      key: 'username',
      text: t('profile.authBindings.source.username', { providerName: usernameSource.label })
    })
  }

  return hints
})
</script>

<style scoped>
.profile-overview-metrics {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1px;
  border-top: 1px solid var(--ui-border);
  background: var(--ui-border);
}

.profile-overview-metrics > div {
  min-width: 0;
  padding: 0.75rem 1rem;
  background: var(--ui-surface-subtle);
}

.profile-overview-metrics dt {
  overflow-wrap: anywhere;
  color: var(--ui-text-subtle);
  font-size: 0.75rem;
  line-height: 1.125rem;
}

.profile-overview-metrics dd {
  margin-top: 0.125rem;
  overflow-wrap: anywhere;
  color: var(--ui-text);
  font-size: 1rem;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  line-height: 1.5rem;
}

.profile-basics-grid {
  display: grid;
  min-width: 0;
}

@media (max-width: 479px) {
  .profile-overview-metrics {
    grid-template-columns: 1fr;
  }
}

@media (min-width: 768px) {
  .profile-basics-grid {
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  }
}
</style>
