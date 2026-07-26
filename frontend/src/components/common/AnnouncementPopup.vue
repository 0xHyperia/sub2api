<template>
  <BaseDialog
    :show="!!displayedAnnouncement"
    :title="displayedAnnouncement?.title || t('announcements.title')"
    width="wide"
    :z-index="120"
    :close-on-escape="preview"
    :show-close-button="preview"
    @close="handleDismiss"
  >
    <div v-if="displayedAnnouncement" class="min-w-0">
      <div class="mb-4 flex items-center gap-2 text-xs text-foreground-muted">
        <Icon name="bell" size="sm" aria-hidden="true" />
        <span class="rounded-control bg-warning-subtle px-2 py-1 font-semibold text-warning-foreground">
          {{ t('announcements.unread') }}
        </span>
        <time>{{ formatRelativeWithDateTime(displayedAnnouncement.created_at) }}</time>
      </div>
      <div
        class="announcement-popup-content markdown-body max-h-[55vh] overflow-y-auto pr-2"
        v-html="renderedContent"
      ></div>
    </div>

    <template #footer>
      <div class="flex justify-end">
        <button
          type="button"
          class="btn btn-primary"
          data-testid="announcement-popup-dismiss"
          @click="handleDismiss"
        >
          <Icon :name="preview ? 'x' : 'check'" size="sm" />
          {{ preview ? t('common.close') : t('announcements.markRead') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { useAnnouncementStore } from '@/stores/announcements'
import { formatRelativeWithDateTime } from '@/utils/format'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import type { Announcement, UserAnnouncement } from '@/types'
import '@/styles/announcement-markdown.css'

type PreviewAnnouncement = Pick<Announcement | UserAnnouncement, 'title' | 'content' | 'created_at'>

const props = withDefaults(defineProps<{
  announcement?: PreviewAnnouncement | null
  preview?: boolean
}>(), {
  announcement: null,
  preview: false,
})

const emit = defineEmits<{
  close: []
}>()

const { t } = useI18n()
// Admin preview is self-contained and must not require the user announcement store.
const announcementStore = props.preview ? null : useAnnouncementStore()
const displayedAnnouncement = computed(() => (
  props.preview ? props.announcement : announcementStore?.currentPopup ?? null
))

marked.setOptions({
  breaks: true,
  gfm: true,
})

const renderedContent = computed(() => {
  const content = displayedAnnouncement.value?.content
  if (!content) return ''
  const html = marked.parse(content) as string
  return DOMPurify.sanitize(html)
})

function handleDismiss() {
  if (props.preview) {
    emit('close')
    return
  }
  announcementStore?.dismissPopup()
}
</script>

<style scoped>
.announcement-popup-content :deep(h1),
.announcement-popup-content :deep(h2),
.announcement-popup-content :deep(h3) {
  margin: 1.25rem 0 0.5rem;
  color: var(--ui-text);
  font-weight: 700;
  letter-spacing: 0;
}

.announcement-popup-content :deep(h1) {
  font-size: 1.25rem;
}

.announcement-popup-content :deep(h2) {
  font-size: 1.125rem;
}

.announcement-popup-content :deep(h3) {
  font-size: 1rem;
}

.announcement-popup-content :deep(p),
.announcement-popup-content :deep(ul),
.announcement-popup-content :deep(ol) {
  margin-bottom: 0.875rem;
  color: var(--ui-text-muted);
  font-size: 0.875rem;
  line-height: 1.7;
}

.announcement-popup-content :deep(ul),
.announcement-popup-content :deep(ol) {
  padding-left: 1.25rem;
}

.announcement-popup-content :deep(a) {
  color: rgb(var(--color-info));
  text-decoration: underline;
  text-underline-offset: 2px;
}

.announcement-popup-content :deep(pre) {
  max-width: 100%;
  overflow-x: auto;
  padding: 0.875rem;
  border: 1px solid var(--ui-border);
  border-radius: 8px;
  background: var(--ui-surface-subtle);
  font-size: 0.75rem;
}
</style>
