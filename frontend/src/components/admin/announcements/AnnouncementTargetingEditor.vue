<template>
  <div
    class="min-w-0 rounded-panel border border-outline bg-surface p-3 sm:p-4"
    :aria-invalid="validationError ? 'true' : undefined"
    :aria-describedby="validationError ? validationSummaryId : undefined"
  >
    <fieldset class="min-w-0 border-0 p-0">
      <legend id="announcement-targeting-mode-label" class="text-sm font-medium text-foreground">
        {{ t('admin.announcements.form.targetingMode') }}
      </legend>
      <p id="announcement-targeting-mode-description" class="input-hint">
        {{ mode === 'all' ? t('admin.announcements.form.targetingAll') : t('admin.announcements.form.targetingCustom') }}
      </p>

      <div class="mt-3 grid grid-cols-1 gap-2 sm:grid-cols-2">
        <label
          :class="[
            'flex min-h-touch cursor-pointer items-center gap-2 rounded-control border px-3 py-2 text-sm text-foreground transition-colors',
            mode === 'all'
              ? 'border-outline-strong bg-surface-subtle'
              : 'border-outline bg-surface hover:bg-surface-subtle'
          ]"
        >
          <input
            id="announcement-targeting-all"
            type="radio"
            name="announcement-targeting-mode"
            value="all"
            :checked="mode === 'all'"
            aria-describedby="announcement-targeting-mode-description"
            class="h-4 w-4 shrink-0 text-brand focus:ring-focus"
            @change="setMode('all')"
          />
          {{ t('admin.announcements.form.targetingAll') }}
        </label>
        <label
          :class="[
            'flex min-h-touch cursor-pointer items-center gap-2 rounded-control border px-3 py-2 text-sm text-foreground transition-colors',
            mode === 'custom'
              ? 'border-outline-strong bg-surface-subtle'
              : 'border-outline bg-surface hover:bg-surface-subtle'
          ]"
        >
          <input
            id="announcement-targeting-custom"
            type="radio"
            name="announcement-targeting-mode"
            value="custom"
            :checked="mode === 'custom'"
            aria-describedby="announcement-targeting-mode-description"
            class="h-4 w-4 shrink-0 text-brand focus:ring-focus"
            @change="setMode('custom')"
          />
          {{ t('admin.announcements.form.targetingCustom') }}
        </label>
      </div>
    </fieldset>

    <section
      v-if="mode === 'custom'"
      class="mt-4 min-w-0 border-t border-outline pt-4"
      aria-labelledby="announcement-targeting-rules-label"
    >
      <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <h3 id="announcement-targeting-rules-label" class="text-sm font-medium text-foreground">
          OR
          <span class="ml-1 text-xs font-normal text-foreground-subtle">
            ({{ anyOf.length }}/50)
          </span>
        </h3>
        <button
          type="button"
          class="btn btn-secondary w-full sm:w-auto"
          :disabled="anyOf.length >= 50"
          @click="addOrGroup"
        >
          <Icon name="plus" size="sm" aria-hidden="true" />
          {{ t('admin.announcements.form.addOrGroup') }}
        </button>
      </div>

      <div
        v-if="anyOf.length === 0"
        class="mt-3 rounded-panel border border-dashed border-outline-strong px-3 py-4 text-sm text-foreground-muted"
        role="status"
      >
        {{ t('admin.announcements.form.targetingCustom') }}: {{ t('admin.announcements.form.addOrGroup') }}
      </div>

      <div class="mt-3 divide-y divide-outline border-y border-outline">
        <section
          v-for="(group, groupIndex) in anyOf"
          :key="groupIndex"
          class="min-w-0 py-4"
          role="group"
          :aria-labelledby="`announcement-targeting-group-${groupIndex}-label`"
          :aria-describedby="groupError(group) ? `announcement-targeting-group-${groupIndex}-error` : `announcement-targeting-group-${groupIndex}-description`"
        >
          <div class="flex items-start justify-between gap-2">
            <div class="min-w-0">
              <h4
                :id="`announcement-targeting-group-${groupIndex}-label`"
                class="text-sm font-medium text-foreground"
              >
                {{ t('admin.announcements.form.targetingCustom') }} #{{ groupIndex + 1 }}
                <span class="ml-1 text-xs font-normal text-foreground-subtle">
                  AND ({{ group.all_of?.length || 0 }}/50)
                </span>
              </h4>
              <p
                :id="`announcement-targeting-group-${groupIndex}-description`"
                class="input-hint"
              >
                {{ t('admin.announcements.form.addAndCondition') }}
              </p>
            </div>

            <button
              type="button"
              class="btn btn-ghost btn-icon shrink-0 text-danger-foreground"
              :aria-label="`${t('common.delete')} ${t('admin.announcements.form.targetingCustom')} ${groupIndex + 1}`"
              :title="t('common.delete')"
              @click="removeOrGroup(groupIndex)"
            >
              <Icon name="trash" size="sm" aria-hidden="true" />
            </button>
          </div>

          <p
            v-if="groupError(group)"
            :id="`announcement-targeting-group-${groupIndex}-error`"
            class="input-error-text"
          >
            {{ groupError(group) }}
          </p>

          <div class="mt-3 divide-y divide-outline rounded-panel bg-surface-subtle px-3">
            <fieldset
              v-for="(cond, condIndex) in group.all_of || []"
              :key="condIndex"
              class="min-w-0 border-0 py-3"
              :aria-describedby="conditionError(cond) ? `announcement-targeting-condition-${groupIndex}-${condIndex}-error` : undefined"
            >
              <legend class="sr-only">
                {{ t('admin.announcements.form.conditionType') }} #{{ condIndex + 1 }}
              </legend>

              <div class="grid min-w-0 grid-cols-1 gap-3 md:grid-cols-[12rem_minmax(0,1fr)_2.5rem] md:items-end">
                <div class="min-w-0">
                  <span
                    :id="`announcement-targeting-condition-${groupIndex}-${condIndex}-type-label`"
                    class="input-label"
                  >
                    {{ t('admin.announcements.form.conditionType') }}
                  </span>
                  <Select
                    :model-value="cond.type"
                    :options="conditionTypeOptions"
                    :aria-labelledby="`announcement-targeting-condition-${groupIndex}-${condIndex}-type-label`"
                    @update:model-value="(value) => setConditionType(groupIndex, condIndex, value as AnnouncementConditionType)"
                  />
                </div>

                <div v-if="cond.type === 'subscription'" class="min-w-0">
                  <span
                    :id="`announcement-targeting-condition-${groupIndex}-${condIndex}-packages-label`"
                    class="sr-only"
                  >
                    {{ t('admin.announcements.form.selectPackages') }}
                  </span>
                  <GroupSelector
                    :model-value="cond.group_ids || []"
                    :groups="groups"
                    role="group"
                    :aria-labelledby="`announcement-targeting-condition-${groupIndex}-${condIndex}-packages-label`"
                    :aria-describedby="conditionError(cond) ? `announcement-targeting-condition-${groupIndex}-${condIndex}-error` : undefined"
                    :aria-invalid="conditionError(cond) ? 'true' : undefined"
                    @update:model-value="(value) => setSubscriptionGroups(groupIndex, condIndex, value)"
                  />
                  <p
                    v-if="conditionError(cond)"
                    :id="`announcement-targeting-condition-${groupIndex}-${condIndex}-error`"
                    class="input-error-text"
                  >
                    {{ conditionError(cond) }}
                  </p>
                </div>

                <div v-else class="grid min-w-0 grid-cols-1 gap-3 sm:grid-cols-2">
                  <div class="min-w-0">
                    <span
                      :id="`announcement-targeting-condition-${groupIndex}-${condIndex}-operator-label`"
                      class="input-label"
                    >
                      {{ t('admin.announcements.form.operator') }}
                    </span>
                    <Select
                      :model-value="cond.operator"
                      :options="balanceOperatorOptions"
                      :aria-labelledby="`announcement-targeting-condition-${groupIndex}-${condIndex}-operator-label`"
                      @update:model-value="(value) => setOperator(groupIndex, condIndex, value as AnnouncementOperator)"
                    />
                  </div>
                  <div class="min-w-0">
                    <label
                      :for="`announcement-targeting-condition-${groupIndex}-${condIndex}-value`"
                      class="input-label"
                    >
                      {{ t('admin.announcements.form.balanceValue') }}
                    </label>
                    <input
                      :id="`announcement-targeting-condition-${groupIndex}-${condIndex}-value`"
                      :value="String(cond.value ?? '')"
                      type="number"
                      step="any"
                      class="input"
                      @input="(event) => setBalanceValue(groupIndex, condIndex, (event.target as HTMLInputElement).value)"
                    />
                  </div>
                </div>

                <button
                  type="button"
                  class="btn btn-ghost btn-icon justify-self-end text-danger-foreground md:self-end"
                  :aria-label="`${t('common.delete')} ${t('admin.announcements.form.conditionType')} ${condIndex + 1}`"
                  :title="t('common.delete')"
                  @click="removeAndCondition(groupIndex, condIndex)"
                >
                  <Icon name="trash" size="sm" aria-hidden="true" />
                </button>
              </div>
            </fieldset>
          </div>

          <div class="mt-3 flex justify-stretch sm:justify-end">
            <button
              type="button"
              class="btn btn-secondary w-full sm:w-auto"
              :disabled="(group.all_of?.length || 0) >= 50"
              @click="addAndCondition(groupIndex)"
            >
              <Icon name="plus" size="sm" aria-hidden="true" />
              {{ t('admin.announcements.form.addAndCondition') }}
            </button>
          </div>
        </section>
      </div>

      <div
        v-if="validationError"
        :id="validationSummaryId"
        class="mt-3 rounded-panel border border-danger/20 bg-danger-subtle p-3 text-sm text-danger-foreground"
        role="alert"
      >
        {{ validationError }}
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type {
  AdminGroup,
  AnnouncementTargeting,
  AnnouncementCondition,
  AnnouncementConditionGroup,
  AnnouncementConditionType,
  AnnouncementOperator
} from '@/types'

import Select from '@/components/common/Select.vue'
import GroupSelector from '@/components/common/GroupSelector.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

const props = defineProps<{
  modelValue: AnnouncementTargeting
  groups: AdminGroup[]
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: AnnouncementTargeting): void
}>()

const validationSummaryId = 'announcement-targeting-validation-summary'
const anyOf = computed(() => props.modelValue?.any_of ?? [])

type Mode = 'all' | 'custom'
const mode = computed<Mode>(() => (anyOf.value.length === 0 ? 'all' : 'custom'))

const conditionTypeOptions = computed(() => [
  { value: 'subscription', label: t('admin.announcements.form.conditionSubscription') },
  { value: 'balance', label: t('admin.announcements.form.conditionBalance') }
])

const balanceOperatorOptions = computed(() => [
  { value: 'gt', label: t('admin.announcements.operators.gt') },
  { value: 'gte', label: t('admin.announcements.operators.gte') },
  { value: 'lt', label: t('admin.announcements.operators.lt') },
  { value: 'lte', label: t('admin.announcements.operators.lte') },
  { value: 'eq', label: t('admin.announcements.operators.eq') }
])

function setMode(next: Mode) {
  if (next === 'all') {
    emit('update:modelValue', { any_of: [] })
    return
  }
  if (anyOf.value.length === 0) {
    emit('update:modelValue', { any_of: [{ all_of: [defaultSubscriptionCondition()] }] })
  }
}

function defaultSubscriptionCondition(): AnnouncementCondition {
  return {
    type: 'subscription' as AnnouncementConditionType,
    operator: 'in' as AnnouncementOperator,
    group_ids: []
  }
}

function defaultBalanceCondition(): AnnouncementCondition {
  return {
    type: 'balance' as AnnouncementConditionType,
    operator: 'gte' as AnnouncementOperator,
    value: 0
  }
}

type TargetingDraft = {
  any_of: AnnouncementConditionGroup[]
}

function updateTargeting(mutator: (draft: TargetingDraft) => void) {
  const draft: TargetingDraft = JSON.parse(JSON.stringify(props.modelValue ?? { any_of: [] }))
  if (!draft.any_of) draft.any_of = []
  mutator(draft)
  emit('update:modelValue', draft)
}

function addOrGroup() {
  updateTargeting((draft) => {
    if (draft.any_of.length >= 50) return
    draft.any_of.push({ all_of: [defaultSubscriptionCondition()] })
  })
}

function removeOrGroup(groupIndex: number) {
  updateTargeting((draft) => {
    draft.any_of.splice(groupIndex, 1)
  })
}

function addAndCondition(groupIndex: number) {
  updateTargeting((draft) => {
    const group = draft.any_of[groupIndex]
    if (!group.all_of) group.all_of = []
    if (group.all_of.length >= 50) return
    group.all_of.push(defaultSubscriptionCondition())
  })
}

function removeAndCondition(groupIndex: number, condIndex: number) {
  updateTargeting((draft) => {
    const group = draft.any_of[groupIndex]
    if (!group?.all_of) return
    group.all_of.splice(condIndex, 1)
  })
}

function setConditionType(groupIndex: number, condIndex: number, nextType: AnnouncementConditionType) {
  updateTargeting((draft) => {
    const group = draft.any_of[groupIndex]
    if (!group?.all_of) return

    if (nextType === 'subscription') {
      group.all_of[condIndex] = defaultSubscriptionCondition()
    } else {
      group.all_of[condIndex] = defaultBalanceCondition()
    }
  })
}

function setOperator(groupIndex: number, condIndex: number, op: AnnouncementOperator) {
  updateTargeting((draft) => {
    const group = draft.any_of[groupIndex]
    if (!group?.all_of) return

    const cond = group.all_of[condIndex]
    if (!cond) return

    cond.operator = op
  })
}

function setSubscriptionGroups(groupIndex: number, condIndex: number, groupIds: number[]) {
  updateTargeting((draft) => {
    const condition = draft.any_of[groupIndex]?.all_of?.[condIndex]
    if (!condition || condition.type !== 'subscription') return

    condition.operator = 'in'
    condition.group_ids = groupIds.slice()
  })
}

function setBalanceValue(groupIndex: number, condIndex: number, raw: string) {
  const n = raw === '' ? 0 : Number(raw)
  updateTargeting((draft) => {
    const group = draft.any_of[groupIndex]
    if (!group?.all_of) return

    const cond = group.all_of[condIndex]
    if (!cond) return

    cond.value = Number.isFinite(n) ? n : 0
  })
}

function groupError(group: AnnouncementConditionGroup) {
  return (group.all_of?.length || 0) === 0
    ? t('admin.announcements.form.addAndCondition')
    : ''
}

function conditionError(condition: AnnouncementCondition) {
  return condition.type === 'subscription' && (!condition.group_ids || condition.group_ids.length === 0)
    ? t('admin.announcements.form.selectPackages')
    : ''
}

const validationError = computed(() => {
  if (mode.value !== 'custom') return ''

  const groups = anyOf.value
  if (groups.length === 0) return t('admin.announcements.form.addOrGroup')

  if (groups.length > 50) return 'any_of > 50'

  for (const g of groups) {
    const allOf = g?.all_of ?? []
    if (allOf.length === 0) return t('admin.announcements.form.addAndCondition')
    if (allOf.length > 50) return 'all_of > 50'

    for (const c of allOf) {
      if (c.type === 'subscription') {
        if (!c.group_ids || c.group_ids.length === 0) return t('admin.announcements.form.selectPackages')
      }
    }
  }

  return ''
})
</script>
