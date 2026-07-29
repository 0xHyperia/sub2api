<template>
  <section
    class="card min-w-0 overflow-hidden"
    data-testid="authorized-apps-card"
    aria-labelledby="profile-authorized-apps-title"
  >
    <header class="card-header">
      <h2 id="profile-authorized-apps-title" class="text-base font-semibold text-foreground">
        {{ t('profile.authorizedApps.title') }}
      </h2>
      <p class="mt-1 text-sm text-foreground-muted">
        {{ t('profile.authorizedApps.description') }}
      </p>
    </header>

    <div class="p-4 sm:p-5">
      <div v-if="loading" class="flex min-h-24 items-center justify-center" role="status" :aria-label="t('common.loading')">
        <span class="h-7 w-7 animate-spin rounded-full border-2 border-outline-strong border-t-foreground" aria-hidden="true" />
      </div>

      <div
        v-else-if="loadError"
        data-testid="authorized-apps-load-error"
        class="flex min-w-0 flex-col gap-4 rounded-panel border border-danger/20 bg-danger-subtle p-4 sm:flex-row sm:items-center sm:justify-between"
        role="alert"
      >
        <div class="flex min-w-0 items-start gap-3">
          <span class="flex h-10 w-10 shrink-0 items-center justify-center rounded-control border border-danger/20 bg-surface/60 text-danger-foreground" aria-hidden="true">
            <Icon name="exclamationTriangle" size="md" />
          </span>
          <p class="min-w-0 break-words text-sm font-medium text-danger-foreground">{{ loadError }}</p>
        </div>
        <button type="button" class="btn btn-secondary shrink-0" @click="loadGrants">
          <Icon name="refresh" size="sm" aria-hidden="true" />
          {{ t('profile.authorizedApps.retry') }}
        </button>
      </div>

      <div v-else-if="grants.length === 0" class="flex min-h-28 flex-col items-center justify-center gap-3 text-center text-sm text-foreground-muted">
        <span class="flex h-10 w-10 items-center justify-center rounded-control border border-outline bg-surface-subtle text-foreground-subtle" aria-hidden="true">
          <Icon name="key" size="md" />
        </span>
        <p>{{ t('profile.authorizedApps.empty') }}</p>
      </div>

      <ul v-else class="divide-y divide-outline overflow-hidden rounded-panel border border-outline">
        <li v-for="grant in grants" :key="grant.id" class="min-w-0">
          <div class="grid min-w-0 grid-cols-[2.5rem_minmax(0,1fr)_auto] items-start gap-3 p-3 sm:p-4">
            <span class="flex h-10 w-10 shrink-0 items-center justify-center rounded-control border border-outline bg-surface-subtle text-foreground-muted" aria-hidden="true">
              <Icon name="key" size="md" />
            </span>
            <div class="min-w-0">
              <div class="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1">
                <p class="min-w-0 break-words text-sm font-medium text-foreground">{{ grant.client_name || grant.client_id }}</p>
                <span class="min-w-0 break-all text-xs text-foreground-subtle">
                  {{ t('profile.authorizedApps.activeSessions', { count: grant.session_count }) }}
                </span>
              </div>
              <p class="mt-1 break-words text-sm text-foreground-muted">
                {{ grant.platform || t('profile.authorizedApps.unknownPlatform') }}
                <span aria-hidden="true"> · </span>
                {{ t('profile.authorizedApps.authorizedAt', { date: formatDateTime(grant.last_authorized_at) }) }}
              </p>
              <p v-if="grant.last_used_at" class="mt-1 break-words text-xs text-foreground-subtle">
                {{ t('profile.authorizedApps.lastUsedAt', { date: formatDateTime(grant.last_used_at) }) }}
              </p>
            </div>
            <div class="flex shrink-0 items-center gap-1">
              <button
                type="button"
                class="btn btn-ghost btn-icon"
                :aria-label="expandedGrantIds.has(String(grant.id)) ? t('profile.authorizedApps.hideDevices') : t('profile.authorizedApps.manageDevices')"
                :title="expandedGrantIds.has(String(grant.id)) ? t('profile.authorizedApps.hideDevices') : t('profile.authorizedApps.manageDevices')"
                :aria-expanded="expandedGrantIds.has(String(grant.id))"
                @click="toggleGrant(grant)"
              >
                <Icon :name="expandedGrantIds.has(String(grant.id)) ? 'chevronUp' : 'chevronDown'" size="md" aria-hidden="true" />
              </button>
              <button
                type="button"
                class="btn btn-ghost btn-icon text-danger-foreground"
                :aria-label="t('profile.authorizedApps.revoke')"
                :title="t('profile.authorizedApps.revoke')"
                @click="grantToRevoke = grant"
              >
                <Icon name="trash" size="md" aria-hidden="true" />
              </button>
            </div>
          </div>

          <div v-if="expandedGrantIds.has(String(grant.id))" class="border-t border-outline bg-surface-subtle/50 px-3 py-3 sm:px-4">
            <div v-if="sessionLoadingIds.has(String(grant.id))" class="flex min-h-20 items-center justify-center" role="status" :aria-label="t('common.loading')">
              <span class="h-5 w-5 animate-spin rounded-full border-2 border-outline-strong border-t-foreground" aria-hidden="true" />
            </div>
            <div v-else-if="sessionErrors[String(grant.id)]" class="flex items-center justify-between gap-3 text-sm text-danger-foreground" role="alert">
              <span>{{ sessionErrors[String(grant.id)] }}</span>
              <button type="button" class="btn btn-secondary shrink-0" @click="loadSessions(grant.id)">{{ t('profile.authorizedApps.retry') }}</button>
            </div>
            <p v-else-if="(sessionsByGrant[String(grant.id)] || []).length === 0" class="py-3 text-center text-sm text-foreground-muted">
              {{ t('profile.authorizedApps.noActiveSessions') }}
            </p>
            <ul v-else class="space-y-2" data-testid="authorization-sessions">
              <li v-for="session in sessionsByGrant[String(grant.id)]" :key="session.id" class="flex min-w-0 flex-col gap-3 rounded-control border border-outline bg-surface p-3 sm:flex-row sm:items-center">
                <span class="flex h-9 w-9 shrink-0 items-center justify-center rounded-control border border-outline bg-surface-subtle text-foreground-muted" aria-hidden="true">
                  <Icon :name="session.platform === 'web' ? 'globe' : 'terminal'" size="sm" />
                </span>
                <div class="min-w-0 flex-1">
                  <form v-if="editingSessionId === String(session.id)" class="flex min-w-0 gap-2" @submit.prevent="saveSessionName(session)">
                    <input v-model="editingDeviceName" class="input min-w-0 flex-1" maxlength="200" :placeholder="t('profile.authorizedApps.deviceNamePlaceholder')" autofocus />
                    <button type="submit" class="btn btn-primary shrink-0" :disabled="!editingDeviceName.trim()">{{ t('profile.authorizedApps.save') }}</button>
                    <button type="button" class="btn btn-secondary shrink-0" @click="cancelRename">{{ t('common.cancel') }}</button>
                  </form>
                  <template v-else>
                    <p class="break-words text-sm font-medium text-foreground">{{ session.device_name || t('profile.authorizedApps.unnamedDevice') }}</p>
                    <p class="mt-1 break-words text-xs text-foreground-subtle">
                      {{ session.platform || t('profile.authorizedApps.unknownPlatform') }}
                      <span aria-hidden="true"> · </span>
                      {{ session.last_used_at
                        ? t('profile.authorizedApps.lastUsedAt', { date: formatDateTime(session.last_used_at) })
                        : t('profile.authorizedApps.signedInAt', { date: formatDateTime(session.created_at) }) }}
                    </p>
                  </template>
                </div>
                <div v-if="editingSessionId !== String(session.id)" class="flex shrink-0 items-center gap-1 self-end sm:self-auto">
                  <button type="button" class="btn btn-ghost btn-icon" :aria-label="t('profile.authorizedApps.rename')" :title="t('profile.authorizedApps.rename')" @click="beginRename(session)">
                    <Icon name="edit" size="sm" aria-hidden="true" />
                  </button>
                  <button
                    v-if="(sessionsByGrant[String(grant.id)] || []).length > 1"
                    type="button"
                    class="btn btn-ghost px-2 text-xs"
                    @click="othersToRevoke = { grant, session }"
                  >
                    {{ t('profile.authorizedApps.signOutOthers') }}
                  </button>
                  <button type="button" class="btn btn-ghost btn-icon text-danger-foreground" :aria-label="t('profile.authorizedApps.signOut')" :title="t('profile.authorizedApps.signOut')" @click="sessionToRevoke = { grant, session }">
                    <Icon name="login" size="sm" aria-hidden="true" />
                  </button>
                </div>
              </li>
            </ul>
          </div>
        </li>
      </ul>
    </div>

    <ConfirmDialog
      :show="grantToRevoke !== null"
      :title="t('profile.authorizedApps.revokeTitle')"
      :message="t('profile.authorizedApps.revokeMessage', { app: grantToRevoke?.client_name || grantToRevoke?.client_id || '' })"
      :confirm-text="t('profile.authorizedApps.revoke')"
      danger
      @confirm="revokeGrant"
      @cancel="grantToRevoke = null"
    />
    <ConfirmDialog
      :show="sessionToRevoke !== null"
      :title="t('profile.authorizedApps.signOutTitle')"
      :message="t('profile.authorizedApps.signOutMessage', { device: sessionDisplayName(sessionToRevoke?.session) })"
      :confirm-text="t('profile.authorizedApps.signOut')"
      danger
      @confirm="revokeSession"
      @cancel="sessionToRevoke = null"
    />
    <ConfirmDialog
      :show="othersToRevoke !== null"
      :title="t('profile.authorizedApps.signOutOthersTitle')"
      :message="t('profile.authorizedApps.signOutOthersMessage', { device: sessionDisplayName(othersToRevoke?.session) })"
      :confirm-text="t('profile.authorizedApps.signOutOthers')"
      danger
      @confirm="revokeOtherSessions"
      @cancel="othersToRevoke = null"
    />
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  listAuthorizationGrants,
  listAuthorizationSessions,
  renameAuthorizationSession,
  revokeAuthorizationGrant,
  revokeAuthorizationSession,
  revokeOtherAuthorizationSessions,
  type AppAuthorizationGrant,
  type AppAuthorizationSession
} from '@/api/appAuth'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import { Icon } from '@/components/icons'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { formatDateTime } from '@/utils/format'

type SessionSelection = { grant: AppAuthorizationGrant; session: AppAuthorizationSession }

const { t } = useI18n()
const appStore = useAppStore()
const loading = ref(true)
const loadError = ref('')
const grants = ref<AppAuthorizationGrant[]>([])
const expandedGrantIds = ref(new Set<string>())
const sessionLoadingIds = ref(new Set<string>())
const sessionsByGrant = ref<Record<string, AppAuthorizationSession[]>>({})
const sessionErrors = ref<Record<string, string>>({})
const grantToRevoke = ref<AppAuthorizationGrant | null>(null)
const sessionToRevoke = ref<SessionSelection | null>(null)
const othersToRevoke = ref<SessionSelection | null>(null)
const editingSessionId = ref('')
const editingDeviceName = ref('')

function setMembership(target: typeof expandedGrantIds, id: string, enabled: boolean): void {
  const next = new Set(target.value)
  enabled ? next.add(id) : next.delete(id)
  target.value = next
}

async function loadGrants(): Promise<void> {
  loading.value = true
  loadError.value = ''
  try {
    grants.value = await listAuthorizationGrants()
  } catch (error) {
    loadError.value = extractApiErrorMessage(error, t('profile.authorizedApps.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function loadSessions(grantId: string | number): Promise<void> {
  const key = String(grantId)
  setMembership(sessionLoadingIds, key, true)
  sessionErrors.value = { ...sessionErrors.value, [key]: '' }
  try {
    sessionsByGrant.value = { ...sessionsByGrant.value, [key]: await listAuthorizationSessions(grantId) }
  } catch (error) {
    sessionErrors.value = { ...sessionErrors.value, [key]: extractApiErrorMessage(error, t('profile.authorizedApps.sessionsLoadFailed')) }
  } finally {
    setMembership(sessionLoadingIds, key, false)
  }
}

function toggleGrant(grant: AppAuthorizationGrant): void {
  const key = String(grant.id)
  const expanding = !expandedGrantIds.value.has(key)
  setMembership(expandedGrantIds, key, expanding)
  if (expanding && !sessionsByGrant.value[key]) void loadSessions(grant.id)
}

function beginRename(session: AppAuthorizationSession): void {
  editingSessionId.value = String(session.id)
  editingDeviceName.value = session.device_name
}

function cancelRename(): void {
  editingSessionId.value = ''
  editingDeviceName.value = ''
}

async function saveSessionName(session: AppAuthorizationSession): Promise<void> {
  const deviceName = editingDeviceName.value.trim()
  if (!deviceName) return
  try {
    await renameAuthorizationSession(session.id, deviceName)
    session.device_name = deviceName
    cancelRename()
    appStore.showSuccess(t('profile.authorizedApps.renameSuccess'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('profile.authorizedApps.renameFailed')))
  }
}

async function revokeGrant(): Promise<void> {
  const grant = grantToRevoke.value
  if (!grant) return
  try {
    await revokeAuthorizationGrant(grant.id)
    grants.value = grants.value.filter((item) => item.id !== grant.id)
    grantToRevoke.value = null
    appStore.showSuccess(t('profile.authorizedApps.revokeSuccess'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('profile.authorizedApps.revokeFailed')))
  }
}

async function revokeSession(): Promise<void> {
  const selected = sessionToRevoke.value
  if (!selected) return
  try {
    await revokeAuthorizationSession(selected.session.id)
    removeSession(selected.grant, selected.session.id)
    sessionToRevoke.value = null
    appStore.showSuccess(t('profile.authorizedApps.signOutSuccess'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('profile.authorizedApps.signOutFailed')))
  }
}

async function revokeOtherSessions(): Promise<void> {
  const selected = othersToRevoke.value
  if (!selected) return
  try {
    await revokeOtherAuthorizationSessions(selected.grant.id, selected.session.id)
    const key = String(selected.grant.id)
    sessionsByGrant.value = { ...sessionsByGrant.value, [key]: [selected.session] }
    selected.grant.session_count = 1
    othersToRevoke.value = null
    appStore.showSuccess(t('profile.authorizedApps.signOutOthersSuccess'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('profile.authorizedApps.signOutOthersFailed')))
  }
}

function removeSession(grant: AppAuthorizationGrant, sessionId: string | number): void {
  const key = String(grant.id)
  sessionsByGrant.value = {
    ...sessionsByGrant.value,
    [key]: (sessionsByGrant.value[key] || []).filter((session) => session.id !== sessionId)
  }
  grant.session_count = Math.max(0, grant.session_count - 1)
}

function sessionDisplayName(session?: AppAuthorizationSession): string {
  return session?.device_name || t('profile.authorizedApps.unnamedDevice')
}

onMounted(loadGrants)
</script>
