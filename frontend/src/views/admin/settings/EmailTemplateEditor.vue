<template>
  <div class="overflow-hidden rounded-panel border border-outline bg-surface shadow-card" :aria-busy="loadingList || editorBusy">
    <div
      class="flex flex-col gap-3 border-b border-outline px-4 py-4 sm:px-6 lg:flex-row lg:items-start lg:justify-between"
    >
      <div class="min-w-0">
        <h2 class="text-lg font-semibold text-foreground">
          {{ t("admin.settings.emailTemplates.title") }}
        </h2>
        <p class="mt-1 text-sm text-foreground-muted">
          {{ t("admin.settings.emailTemplates.description") }}
        </p>
        <p
          v-if="isDirty"
          id="email-template-unsaved-status"
          class="mt-2 text-sm font-medium text-warning-foreground"
          role="status"
        >
          {{ t("admin.settings.emailTemplates.unsavedChanges") }}
        </p>
      </div>
      <div class="grid w-full grid-cols-1 gap-2 sm:flex sm:w-auto sm:flex-wrap">
        <button
          type="button"
          data-testid="email-template-preview"
          class="btn btn-secondary btn-sm"
          :disabled="editorBusy || previewing || !canPreview"
          @click="refreshPreview"
        >
          {{ previewing ? t("admin.settings.emailTemplates.previewing") : t("admin.settings.emailTemplates.preview") }}
        </button>
        <button
          type="button"
          data-testid="email-template-restore"
          class="btn btn-secondary btn-sm"
          :disabled="editorBusy || !selectedEvent || !selectedLocale"
          @click="requestRestoreOfficial"
        >
          {{ restoring ? t("admin.settings.emailTemplates.restoring") : t("admin.settings.emailTemplates.restoreOfficial") }}
        </button>
        <button
          type="button"
          data-testid="email-template-save"
          class="btn btn-primary btn-sm"
          :disabled="editorBusy || !canSave"
          @click="saveTemplate"
        >
          {{ saving ? t("admin.settings.emailTemplates.saving") : t("admin.settings.emailTemplates.save") }}
        </button>
      </div>
    </div>

    <div class="space-y-6 p-4 sm:p-6">
      <div
        v-if="loadingList"
        class="flex items-center gap-2 text-sm text-foreground-subtle"
        role="status"
        aria-live="polite"
      >
        <span
          class="h-4 w-4 animate-spin rounded-full border-b-2 border-brand"
          aria-hidden="true"
        ></span>
        {{ t("common.loading") }}
      </div>

      <template v-else>
        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div>
            <label class="input-label" for="email-template-event">
              {{ t("admin.settings.emailTemplates.event") }}
            </label>
            <select
              id="email-template-event"
              :value="selectedEvent"
              class="input"
              :disabled="editorBusy || confirmDialog.show || eventOptions.length === 0"
              :aria-describedby="isDirty ? 'email-template-unsaved-status' : undefined"
              @change="handleEventChange"
            >
              <option
                v-for="option in eventOptions"
                :key="option.value"
                :value="option.value"
              >
                {{ formatEventOptionLabel(option) }}
              </option>
            </select>
          </div>
          <div>
            <label class="input-label" for="email-template-locale">
              {{ t("admin.settings.emailTemplates.locale") }}
            </label>
            <select
              id="email-template-locale"
              :value="selectedLocale"
              class="input"
              :disabled="editorBusy || confirmDialog.show || localeOptions.length === 0"
              :aria-describedby="isDirty ? 'email-template-unsaved-status' : undefined"
              @change="handleLocaleChange"
            >
              <option
                v-for="localeOption in localeOptions"
                :key="localeOption"
                :value="localeOption"
              >
                {{ formatLocale(localeOption) }}
              </option>
            </select>
          </div>
        </div>

        <div
          v-if="selectedEventMeta"
          class="rounded-panel border border-outline bg-surface-subtle p-4"
        >
          <div class="flex flex-wrap items-center gap-2">
            <div class="text-sm font-semibold text-foreground">
              {{ selectedEventMeta.label }}
            </div>
            <span
              class="rounded-control border border-outline bg-surface px-2.5 py-1 text-xs font-medium text-foreground-muted"
            >
              {{ selectedEventMeta.categoryLabel }}
            </span>
            <span
              class="rounded-control px-2.5 py-1 text-xs font-medium"
              :class="
                selectedEventMeta.optional
                  ? 'bg-warning-subtle text-warning-foreground'
                  : 'bg-success-subtle text-success-foreground'
              "
            >
              {{ selectedEventMeta.optional ? localText("可退订通知", "Optional") : localText("事务邮件", "Transactional") }}
            </span>
          </div>
          <p class="mt-2 text-sm leading-6 text-foreground-muted">
            {{ selectedEventMeta.timing }}
          </p>
          <p
            v-if="selectedEventDescription"
            class="mt-1 text-xs text-foreground-subtle"
          >
            {{ selectedEventDescription }}
          </p>
        </div>

        <div
          v-if="!eventOptions.length || !localeOptions.length"
          class="rounded-panel border border-warning/30 bg-warning-subtle p-4 text-sm text-warning-foreground"
        >
          {{ t("admin.settings.emailTemplates.empty") }}
        </div>

        <div
          v-else-if="templateLoadFailed"
          class="flex flex-col items-start gap-3 rounded-panel border border-danger/30 bg-danger-subtle p-4 text-sm text-danger-foreground sm:flex-row sm:items-center sm:justify-between"
          role="alert"
        >
          <span>{{ t("admin.settings.emailTemplates.loadFailed") }}</span>
          <button
            type="button"
            class="btn btn-secondary btn-sm w-full sm:w-auto"
            :disabled="editorBusy"
            @click="loadTemplate"
          >
            {{ t("common.retry") }}
          </button>
        </div>

        <div v-else class="grid min-w-0 grid-cols-1 gap-6 xl:grid-cols-2">
          <div class="min-w-0 space-y-4">
            <div>
              <label class="input-label" for="email-template-subject">
                {{ t("admin.settings.emailTemplates.subject") }}
              </label>
              <input
                id="email-template-subject"
                v-model="subject"
                type="text"
                class="input"
                :disabled="editorBusy"
                :aria-describedby="isDirty ? 'email-template-unsaved-status' : undefined"
                :placeholder="t('admin.settings.emailTemplates.subjectPlaceholder')"
              />
            </div>

            <div>
              <label class="input-label" for="email-template-html">
                {{ t("admin.settings.emailTemplates.html") }}
              </label>
              <textarea
                id="email-template-html"
                v-model="html"
                rows="18"
                class="input min-h-[24rem] w-full resize-y font-mono text-sm leading-6 sm:min-h-[28rem]"
                :disabled="editorBusy"
                :aria-describedby="isDirty ? 'email-template-unsaved-status' : undefined"
                :placeholder="t('admin.settings.emailTemplates.htmlPlaceholder')"
              ></textarea>
            </div>

            <div
              class="rounded-panel border border-outline bg-surface-subtle p-4"
            >
              <div class="text-sm font-medium text-foreground">
                {{ t("admin.settings.emailTemplates.placeholders") }}
              </div>
              <p class="mt-1 text-xs text-foreground-subtle">
                {{ t("admin.settings.emailTemplates.placeholdersHelp") }}
              </p>
              <div class="mt-3 flex flex-wrap gap-2">
                <button
                  v-for="placeholder in placeholderList"
                  :key="placeholder"
                  type="button"
                  class="max-w-full break-all rounded-control border border-outline bg-surface px-3 py-1 text-left font-mono text-xs text-foreground-muted transition-colors hover:border-brand hover:text-brand focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus"
                  @click="copyPlaceholder(placeholder)"
                >
                  {{ placeholder }}
                </button>
              </div>
            </div>
          </div>

          <div class="min-w-0 space-y-4">
            <div
              class="min-w-0 overflow-hidden rounded-panel border border-outline bg-surface"
            >
              <div
                class="flex min-w-0 items-start justify-between gap-3 border-b border-outline px-4 py-3"
              >
                <div class="min-w-0">
                  <div class="text-sm font-medium text-foreground">
                    {{ t("admin.settings.emailTemplates.livePreview") }}
                  </div>
                  <div class="mt-0.5 break-words text-xs text-foreground-subtle">
                    {{ previewSubject || t("admin.settings.emailTemplates.noPreview") }}
                  </div>
                </div>
                <span
                  v-if="isCustomTemplate"
                  class="flex-shrink-0 rounded-control bg-brand-subtle px-2.5 py-1 text-xs font-medium text-brand"
                >
                  {{ t("admin.settings.emailTemplates.customized") }}
                </span>
              </div>
              <div class="min-w-0 bg-canvas p-3">
                <iframe
                  class="h-[36rem] w-full min-w-0 max-w-full rounded-control border border-outline bg-white"
                  sandbox=""
                  :srcdoc="previewHtml"
                  :title="t('admin.settings.emailTemplates.livePreview')"
                ></iframe>
              </div>
            </div>

            <p class="text-xs text-foreground-subtle">
              {{ t("admin.settings.emailTemplates.previewSecurityHint") }}
            </p>
          </div>
        </div>
      </template>
    </div>

    <ConfirmDialog
      :show="confirmDialog.show"
      :title="confirmDialog.title"
      :message="confirmDialog.message"
      :confirm-text="confirmDialog.confirmText"
      :danger="confirmDialog.danger"
      @confirm="handleConfirm"
      @cancel="handleConfirmCancel"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { onBeforeRouteLeave, onBeforeRouteUpdate } from "vue-router";
import { adminAPI } from "@/api";
import type {
  EmailTemplateEventOption,
  EmailTemplateOption,
} from "@/api/admin/settings";
import ConfirmDialog from "@/components/common/ConfirmDialog.vue";
import { useAppStore } from "@/stores";
import { extractApiErrorMessage } from "@/utils/apiError";

const { t, locale } = useI18n();
const appStore = useAppStore();

const fallbackPlaceholders = [
  "{{site_name}}",
  "{{recipient_name}}",
  "{{recipient_email}}",
  "{{verification_code}}",
  "{{expires_in_minutes}}",
  "{{reset_url}}",
  "{{subscription_group}}",
  "{{subscription_days}}",
  "{{expiry_time}}",
  "{{days_remaining}}",
  "{{current_balance}}",
  "{{threshold}}",
  "{{recharge_url}}",
  "{{recharge_amount}}",
  "{{order_id}}",
  "{{unsubscribe_url}}",
  "{{account_id}}",
  "{{account_name}}",
  "{{platform}}",
  "{{quota_dimension}}",
  "{{quota_used}}",
  "{{quota_limit}}",
  "{{quota_remaining}}",
  "{{quota_threshold}}",
  "{{triggered_at}}",
  "{{group_name}}",
  "{{moderation_category}}",
  "{{moderation_score}}",
  "{{violation_count}}",
  "{{ban_threshold}}",
  "{{rule_name}}",
  "{{severity}}",
  "{{alert_status}}",
  "{{metric_type}}",
  "{{operator}}",
  "{{metric_value}}",
  "{{threshold_value}}",
  "{{alert_description}}",
  "{{report_name}}",
  "{{report_type}}",
  "{{report_start_time}}",
  "{{report_end_time}}",
  "{{report_html}}",
];

const loadingList = ref(true);
const loadingTemplate = ref(false);
const saving = ref(false);
const previewing = ref(false);
const restoring = ref(false);
const eventOptions = ref<EmailTemplateOption[]>([]);
const localeOptions = ref<string[]>([]);
const selectedEvent = ref("");
const selectedLocale = ref("");
const subject = ref("");
const html = ref("");
const isCustomTemplate = ref(false);
const placeholders = ref<string[]>([]);
const previewSubject = ref("");
const previewHtml = ref("");
const initializingSelection = ref(false);
const templateLoadFailed = ref(false);

interface PersistedTemplate {
  subject: string;
  html: string;
}

type PendingConfirmation =
  | { kind: "discard"; resolve: (confirmed: boolean) => void }
  | { kind: "restore" };

const persistedTemplate = ref<PersistedTemplate | null>(null);
const confirmDialog = reactive({
  show: false,
  title: "",
  message: "",
  confirmText: "",
  danger: true,
});

let pendingConfirmation: PendingConfirmation | null = null;
let listRequestId = 0;
let templateRequestId = 0;
let previewRequestId = 0;
let saveRequestId = 0;
let restoreRequestId = 0;

interface EventDisplayMeta {
  label: string;
  timing: string;
  categoryLabel: string;
}

function localText(zh: string, en: string): string {
  return locale.value.toLowerCase().startsWith("zh") ? zh : en;
}

const eventDisplayMeta: Record<string, EventDisplayMeta> = {
  "auth.verify_code": {
    label: "邮箱验证码",
    timing: "注册、绑定邮箱、OAuth 补全邮箱或 TOTP 邮箱校验时发送。",
    categoryLabel: "认证安全",
  },
  "auth.password_reset": {
    label: "密码重置",
    timing: "用户请求密码重置链接时发送。",
    categoryLabel: "认证安全",
  },
  "notification_email.verify_code": {
    label: "通知邮箱验证码",
    timing: "用户添加并验证额外通知邮箱时发送。",
    categoryLabel: "认证安全",
  },
  "subscription.purchase_success": {
    label: "订阅开通成功",
    timing: "订阅订单完成支付并成功开通或续期后发送。",
    categoryLabel: "订阅",
  },
  "subscription.expiry_reminder": {
    label: "订阅到期提醒",
    timing: "后台任务在订阅仍有效且距离到期剩余 7 天、3 天、1 天时各发送一次，可通过邮件设置中的开关关闭。",
    categoryLabel: "订阅",
  },
  "balance.low": {
    label: "余额不足提醒",
    timing: "用户余额低于全局或个人配置的提醒阈值时发送。",
    categoryLabel: "计费",
  },
  "balance.recharge_success": {
    label: "余额充值成功",
    timing: "余额充值订单支付完成并入账后发送。",
    categoryLabel: "计费",
  },
  "account.quota_alert": {
    label: "账号限额告警",
    timing: "上游账号的用量达到配置的额度告警阈值时发送给管理员通知邮箱。",
    categoryLabel: "管理告警",
  },
  "content_moderation.violation_notice": {
    label: "内容审计违规提醒",
    timing: "用户请求命中内容审计或风控规则、但尚未被禁用时发送。",
    categoryLabel: "风控",
  },
  "content_moderation.account_disabled": {
    label: "内容审计禁用账号",
    timing: "内容审计违规次数达到封禁阈值并自动禁用用户账号时发送。",
    categoryLabel: "风控",
  },
  "ops.alert": {
    label: "运维告警",
    timing: "运维监控规则触发告警并满足邮件通知配置时发送给运维收件人。",
    categoryLabel: "运维",
  },
  "ops.scheduled_report": {
    label: "运维定时报表",
    timing: "运维日报、周报、错误摘要或账号健康报表到达配置的发送时间时发送。",
    categoryLabel: "运维",
  },
};

const eventDisplayMetaEn: Record<string, EventDisplayMeta> = {
  "auth.verify_code": {
    label: "Email Verification Code",
    timing: "Sent for registration, email binding, OAuth pending email completion, or TOTP email verification.",
    categoryLabel: "Auth",
  },
  "auth.password_reset": {
    label: "Password Reset",
    timing: "Sent when a user requests a password reset link.",
    categoryLabel: "Auth",
  },
  "notification_email.verify_code": {
    label: "Notification Email Verification",
    timing: "Sent when a user adds and verifies an extra notification email address.",
    categoryLabel: "Auth",
  },
  "subscription.purchase_success": {
    label: "Subscription Activated",
    timing: "Sent after a subscription order is paid and the subscription is activated or extended.",
    categoryLabel: "Subscription",
  },
  "subscription.expiry_reminder": {
    label: "Subscription Expiry Reminder",
    timing: "Sent by the background job when an active subscription has 7, 3, or 1 day remaining. It can be disabled in Email settings.",
    categoryLabel: "Subscription",
  },
  "balance.low": {
    label: "Low Balance Alert",
    timing: "Sent when a user's balance drops below the global or personal reminder threshold.",
    categoryLabel: "Billing",
  },
  "balance.recharge_success": {
    label: "Balance Recharge Success",
    timing: "Sent after a balance recharge order is paid and credited.",
    categoryLabel: "Billing",
  },
  "account.quota_alert": {
    label: "Account Quota Alert",
    timing: "Sent to admin notification emails when an upstream account reaches the configured quota alert threshold.",
    categoryLabel: "Admin",
  },
  "content_moderation.violation_notice": {
    label: "Risk Control Violation Notice",
    timing: "Sent when a user request triggers content moderation or risk-control rules but the account is not disabled yet.",
    categoryLabel: "Risk Control",
  },
  "content_moderation.account_disabled": {
    label: "Risk Control Account Disabled",
    timing: "Sent when content moderation reaches the ban threshold and automatically disables the user account.",
    categoryLabel: "Risk Control",
  },
  "ops.alert": {
    label: "Ops Alert",
    timing: "Sent to ops recipients when an ops monitoring rule fires and email notification settings allow it.",
    categoryLabel: "Ops",
  },
  "ops.scheduled_report": {
    label: "Ops Scheduled Report",
    timing: "Sent when a configured daily, weekly, error digest, or account health report reaches its scheduled send time.",
    categoryLabel: "Ops",
  },
};

function normalizeEventOption(option: EmailTemplateEventOption): EmailTemplateOption {
  if (typeof option === "string") {
    return { value: option };
  }
  return option;
}

function eventMetaFor(option?: EmailTemplateOption | null) {
  if (!option) return null;
  const displayMeta = (
    locale.value.toLowerCase().startsWith("zh")
      ? eventDisplayMeta
      : eventDisplayMetaEn
  )[option.value];
  const label = displayMeta?.label || option.label || option.value;
  const timing = displayMeta?.timing || option.description || "";
  const categoryLabel =
    displayMeta?.categoryLabel || formatCategory(option.category || "");
  return {
    label,
    timing,
    categoryLabel,
    optional: option.optional === true,
  };
}

function formatEventOptionLabel(option: EmailTemplateOption): string {
  const meta = eventMetaFor(option);
  if (!meta) return option.label || option.value;
  return meta.label;
}

function formatCategory(category: string): string {
  const normalized = category.trim().toLowerCase();
  if (!normalized) return localText("通知", "Notification");
  const labels: Record<string, { zh: string; en: string }> = {
    auth: { zh: "认证安全", en: "Auth" },
    subscription: { zh: "订阅", en: "Subscription" },
    billing: { zh: "计费", en: "Billing" },
    admin: { zh: "管理告警", en: "Admin" },
    risk_control: { zh: "风控", en: "Risk Control" },
    ops: { zh: "运维", en: "Ops" },
  };
  const item = labels[normalized];
  return item ? localText(item.zh, item.en) : category;
}

const selectedEventOption = computed(() => {
  return (
    eventOptions.value.find((option) => option.value === selectedEvent.value) ||
    null
  );
});

const selectedEventMeta = computed(() => eventMetaFor(selectedEventOption.value));

const selectedEventDescription = computed(() => {
  return (
    selectedEventOption.value?.description || ""
  );
});

const placeholderList = computed(() => {
  const combined = [...placeholders.value, ...fallbackPlaceholders];
  return Array.from(
    new Set(
      combined
        .map((item) => formatPlaceholder(item))
        .filter((item) => item.length > 0),
    ),
  );
});

function formatPlaceholder(placeholder: string): string {
  const trimmed = placeholder.trim();
  if (!trimmed) return "";
  if (trimmed.startsWith("{{") && trimmed.endsWith("}}")) return trimmed;
  return `{{${trimmed}}}`;
}

const canSave = computed(
  () =>
    Boolean(selectedEvent.value && selectedLocale.value) &&
    !templateLoadFailed.value &&
    subject.value.trim().length > 0 &&
    html.value.trim().length > 0,
);

const canPreview = computed(
  () =>
    Boolean(selectedEvent.value && selectedLocale.value) &&
    !templateLoadFailed.value &&
    html.value.trim().length > 0,
);

const editorBusy = computed(
  () => loadingTemplate.value || saving.value || restoring.value,
);

const isDirty = computed(() => {
  const persisted = persistedTemplate.value;
  if (!persisted) return false;
  return subject.value !== persisted.subject || html.value !== persisted.html;
});

function formatLocale(locale: string): string {
  const lower = locale.toLowerCase();
  if (lower === "zh" || lower.startsWith("zh-")) {
    return t("admin.settings.emailTemplates.localeZh");
  }
  if (lower === "en" || lower.startsWith("en-")) {
    return t("admin.settings.emailTemplates.localeEn");
  }
  return locale;
}

function selectInitialLocale(locales: string[]): string {
  const currentLocale = locale.value.toLowerCase();
  const exactMatch = locales.find(
    (availableLocale) => availableLocale.toLowerCase() === currentLocale,
  );
  if (exactMatch) return exactMatch;

  const currentLanguage = currentLocale.split("-")[0];
  const languageMatch = locales.find(
    (availableLocale) => availableLocale.toLowerCase().split("-")[0] === currentLanguage,
  );
  if (languageMatch) return languageMatch;

  return locales[0] || "";
}

function applyTemplate(template: {
  subject: string;
  html: string;
  is_custom?: boolean;
  placeholders?: string[];
}, replaceEditor = true) {
  persistedTemplate.value = {
    subject: template.subject,
    html: template.html,
  };
  if (replaceEditor) {
    subject.value = template.subject;
    html.value = template.html;
  }
  isCustomTemplate.value = template.is_custom === true;
  placeholders.value = template.placeholders || [];
  templateLoadFailed.value = false;
}

function resetEditorForLoad() {
  persistedTemplate.value = null;
  subject.value = "";
  html.value = "";
  isCustomTemplate.value = false;
  previewSubject.value = "";
  previewHtml.value = "";
  previewRequestId += 1;
  previewing.value = false;
}

function discardLocalChanges() {
  const persisted = persistedTemplate.value;
  if (!persisted) return;
  subject.value = persisted.subject;
  html.value = persisted.html;
}

async function loadTemplate() {
  if (!selectedEvent.value || !selectedLocale.value) return;

  const requestId = ++templateRequestId;
  const eventValue = selectedEvent.value;
  const localeValue = selectedLocale.value;
  loadingTemplate.value = true;
  templateLoadFailed.value = false;
  resetEditorForLoad();

  try {
    const template = await adminAPI.settings.getEmailTemplate(
      eventValue,
      localeValue,
    );
    if (
      requestId !== templateRequestId ||
      selectedEvent.value !== eventValue ||
      selectedLocale.value !== localeValue
    ) {
      return;
    }
    applyTemplate(template);
    await refreshPreview();
  } catch (err: unknown) {
    if (
      requestId !== templateRequestId ||
      selectedEvent.value !== eventValue ||
      selectedLocale.value !== localeValue
    ) {
      return;
    }
    templateLoadFailed.value = true;
    appStore.showError(extractApiErrorMessage(err, t("common.error")));
  } finally {
    if (requestId === templateRequestId) {
      loadingTemplate.value = false;
    }
  }
}

async function loadTemplateList() {
  const requestId = ++listRequestId;
  loadingList.value = true;
  try {
    const response = await adminAPI.settings.getEmailTemplates();
    if (requestId !== listRequestId) return;
    eventOptions.value = response.events.map(normalizeEventOption);
    localeOptions.value = response.locales;
    placeholders.value = response.placeholders || [];
    initializingSelection.value = true;
    selectedEvent.value = eventOptions.value[0]?.value || "";
    selectedLocale.value = selectInitialLocale(response.locales);
    await loadTemplate();
    initializingSelection.value = false;
  } catch (err: unknown) {
    if (requestId !== listRequestId) return;
    initializingSelection.value = false;
    appStore.showError(extractApiErrorMessage(err, t("common.error")));
  } finally {
    if (requestId === listRequestId) {
      loadingList.value = false;
    }
  }
}

async function saveTemplate() {
  if (!canSave.value) {
    appStore.showError(t("admin.settings.emailTemplates.validationRequired"));
    return;
  }

  const requestId = ++saveRequestId;
  const eventValue = selectedEvent.value;
  const localeValue = selectedLocale.value;
  const submittedTemplate = {
    subject: subject.value,
    html: html.value,
  };
  saving.value = true;

  try {
    const template = await adminAPI.settings.updateEmailTemplate(
      eventValue,
      localeValue,
      submittedTemplate,
    );
    if (
      requestId !== saveRequestId ||
      selectedEvent.value !== eventValue ||
      selectedLocale.value !== localeValue
    ) {
      return;
    }

    const editorStillMatchesSubmission =
      subject.value === submittedTemplate.subject &&
      html.value === submittedTemplate.html;
    applyTemplate(template, editorStillMatchesSubmission);
    await refreshPreview();
    appStore.showSuccess(t("admin.settings.emailTemplates.saveSuccess"));
  } catch (err: unknown) {
    if (requestId !== saveRequestId) return;
    appStore.showError(extractApiErrorMessage(err, t("common.error")));
  } finally {
    if (requestId === saveRequestId) {
      saving.value = false;
    }
  }
}

async function refreshPreview() {
  const requestId = ++previewRequestId;
  if (!canPreview.value) {
    previewSubject.value = "";
    previewHtml.value = "";
    previewing.value = false;
    return;
  }

  const previewRequest = {
    event: selectedEvent.value,
    locale: selectedLocale.value,
    subject: subject.value,
    html: html.value,
  };
  previewing.value = true;

  try {
    const preview = await adminAPI.settings.previewEmailTemplate(previewRequest);
    if (
      requestId !== previewRequestId ||
      selectedEvent.value !== previewRequest.event ||
      selectedLocale.value !== previewRequest.locale ||
      subject.value !== previewRequest.subject ||
      html.value !== previewRequest.html
    ) {
      return;
    }
    previewSubject.value = preview.subject;
    previewHtml.value = preview.html;
  } catch (err: unknown) {
    if (
      requestId !== previewRequestId ||
      selectedEvent.value !== previewRequest.event ||
      selectedLocale.value !== previewRequest.locale ||
      subject.value !== previewRequest.subject ||
      html.value !== previewRequest.html
    ) {
      return;
    }
    appStore.showError(extractApiErrorMessage(err, t("common.error")));
  } finally {
    if (requestId === previewRequestId) {
      previewing.value = false;
    }
  }
}

function requestRestoreOfficial() {
  if (!selectedEvent.value || !selectedLocale.value) return;
  if (confirmDialog.show) return;

  pendingConfirmation = { kind: "restore" };
  confirmDialog.title = t("admin.settings.emailTemplates.restoreOfficial");
  confirmDialog.message = t("admin.settings.emailTemplates.restoreConfirm");
  confirmDialog.confirmText = t("admin.settings.emailTemplates.restoreOfficial");
  confirmDialog.danger = true;
  confirmDialog.show = true;
}

async function restoreOfficial() {
  const requestId = ++restoreRequestId;
  const eventValue = selectedEvent.value;
  const localeValue = selectedLocale.value;

  restoring.value = true;
  try {
    const template = await adminAPI.settings.restoreOfficialEmailTemplate(
      eventValue,
      localeValue,
    );
    if (
      requestId !== restoreRequestId ||
      selectedEvent.value !== eventValue ||
      selectedLocale.value !== localeValue
    ) {
      return;
    }
    applyTemplate(template);
    await refreshPreview();
    appStore.showSuccess(t("admin.settings.emailTemplates.restoreSuccess"));
  } catch (err: unknown) {
    if (requestId !== restoreRequestId) return;
    appStore.showError(extractApiErrorMessage(err, t("common.error")));
  } finally {
    if (requestId === restoreRequestId) {
      restoring.value = false;
    }
  }
}

function requestDiscardChanges(): Promise<boolean> {
  if (!isDirty.value) return Promise.resolve(true);
  if (confirmDialog.show || pendingConfirmation) return Promise.resolve(false);

  return new Promise((resolve) => {
    pendingConfirmation = { kind: "discard", resolve };
    confirmDialog.title = t("admin.settings.emailTemplates.discardTitle");
    confirmDialog.message = t("admin.settings.emailTemplates.discardConfirm");
    confirmDialog.confirmText = t("admin.settings.emailTemplates.discardChanges");
    confirmDialog.danger = true;
    confirmDialog.show = true;
  });
}

function handleConfirm() {
  const confirmation = pendingConfirmation;
  pendingConfirmation = null;
  confirmDialog.show = false;

  if (confirmation?.kind === "discard") {
    discardLocalChanges();
    confirmation.resolve(true);
    return;
  }

  if (confirmation?.kind === "restore") {
    void restoreOfficial();
  }
}

function handleConfirmCancel() {
  const confirmation = pendingConfirmation;
  pendingConfirmation = null;
  confirmDialog.show = false;
  if (confirmation?.kind === "discard") {
    confirmation.resolve(false);
  }
}

function changeSelection(kind: "event" | "locale", nextValue: string) {
  const currentValue = kind === "event" ? selectedEvent.value : selectedLocale.value;
  if (!nextValue || nextValue === currentValue) return;

  const applySelection = () => {
    if (kind === "event") {
      selectedEvent.value = nextValue;
    } else {
      selectedLocale.value = nextValue;
    }
  };

  if (!isDirty.value) {
    applySelection();
    return;
  }

  void requestDiscardChanges().then((confirmed) => {
    if (confirmed) applySelection();
  });
}

function handleEventChange(event: Event) {
  changeSelection("event", (event.target as HTMLSelectElement).value);
}

function handleLocaleChange(event: Event) {
  changeSelection("locale", (event.target as HTMLSelectElement).value);
}

async function copyPlaceholder(placeholder: string) {
  try {
    await navigator.clipboard.writeText(placeholder);
    appStore.showSuccess(t("admin.settings.emailTemplates.placeholderCopied"));
  } catch {
    appStore.showError(t("common.error"));
  }
}

watch(
  [selectedEvent, selectedLocale],
  ([eventValue, localeValue], [oldEvent, oldLocale]) => {
    if (initializingSelection.value) return;
    if (!eventValue || !localeValue) return;
    if (eventValue === oldEvent && localeValue === oldLocale) return;
    void loadTemplate();
  },
);

onBeforeRouteLeave(() => requestDiscardChanges());
onBeforeRouteUpdate(() => requestDiscardChanges());

onMounted(() => {
  void loadTemplateList();
});

onUnmounted(() => {
  listRequestId += 1;
  templateRequestId += 1;
  previewRequestId += 1;
  saveRequestId += 1;
  restoreRequestId += 1;
  if (pendingConfirmation?.kind === "discard") {
    pendingConfirmation.resolve(false);
  }
  pendingConfirmation = null;
});

defineExpose({
  isDirty,
  requestDiscardChanges,
});
</script>
