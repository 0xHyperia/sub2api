<template>
  <div class="table-page-layout" :class="{ 'mobile-mode': isMobile }">
    <!-- 固定区域：操作按钮 -->
    <div v-if="$slots.actions" class="layout-section-fixed">
      <slot name="actions" />
    </div>

    <!-- 固定区域：搜索和过滤器 -->
    <div v-if="$slots.filters" class="layout-section-fixed">
      <slot name="filters" />
    </div>

    <!-- 滚动区域：表格 -->
    <div class="layout-section-scrollable">
      <div class="card table-scroll-container">
        <slot name="table" />
      </div>
    </div>

    <!-- 固定区域：分页器 -->
    <div v-if="$slots.pagination" class="layout-section-fixed">
      <slot name="pagination" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'

const desktopViewportQuery = '(min-width: 1024px)'
const isMobile = ref(
  typeof window === 'undefined' ? false : !window.matchMedia(desktopViewportQuery).matches
)
let viewportMediaQuery: MediaQueryList | null = null
let viewportListener: ((event: MediaQueryListEvent) => void) | null = null

onMounted(() => {
  viewportMediaQuery = window.matchMedia(desktopViewportQuery)
  isMobile.value = !viewportMediaQuery.matches
  viewportListener = (event: MediaQueryListEvent) => {
    isMobile.value = !event.matches
  }

  if (typeof viewportMediaQuery.addEventListener === 'function') {
    viewportMediaQuery.addEventListener('change', viewportListener)
  } else {
    viewportMediaQuery.addListener(viewportListener)
  }
})

onUnmounted(() => {
  if (viewportMediaQuery && viewportListener) {
    if (typeof viewportMediaQuery.removeEventListener === 'function') {
      viewportMediaQuery.removeEventListener('change', viewportListener)
    } else {
      viewportMediaQuery.removeListener(viewportListener)
    }
  }
  viewportMediaQuery = null
  viewportListener = null
})
</script>

<style scoped>
/* 桌面端：Flexbox 布局 */
.table-page-layout {
  @apply flex flex-col;
  gap: 16px;
  width: 100%;
  min-width: 0;
  max-width: 100%;
  height: calc(100dvh - var(--app-header-height, 64px) - (var(--page-gutter, 32px) * 2));
}

.layout-section-fixed {
  @apply flex-shrink-0;
  min-width: 0;
  max-width: 100%;
}

.layout-section-scrollable {
  @apply flex-1 min-h-0 flex flex-col;
  min-width: 0;
  max-width: 100%;
}

/* 表格滚动容器 - 增强版表体滚动方案 */
.table-scroll-container {
  @apply flex h-full flex-col overflow-hidden;
  width: 100%;
  min-width: 0;
  max-width: 100%;
  border: 1px solid var(--ui-border, #dbe3ee);
  border-radius: 8px;
  background: var(--ui-surface, #fff);
  box-shadow: var(--ui-shadow-xs, 0 1px 2px rgba(15, 23, 42, 0.04));
}

.table-scroll-container :deep(.table-wrapper) {
  @apply flex-1 overflow-x-auto overflow-y-auto;
  /* 确保横向滚动条显示在最底部 */
  scrollbar-gutter: stable;
}

.table-scroll-container :deep(table) {
  @apply w-full;
  min-width: max-content; /* 关键：确保表格宽度根据内容撑开，从而触发横向滚动 */
  display: table; /* 使用标准 table 布局以支持 sticky 列 */
}

.table-scroll-container :deep(thead) {
  background: var(--ui-surface-subtle, #f4f7fb);
}

.table-scroll-container :deep(tbody) {
  /* 保持默认 table-row-group 显示，不使用 block */
}

.table-scroll-container :deep(th) {
  @apply px-4 py-3 text-left text-xs font-semibold;
  border-bottom: 1px solid var(--ui-border, #dbe3ee);
  color: var(--ui-text-muted, #667085);
}

.table-scroll-container :deep(td) {
  @apply px-4 py-3 text-sm;
  border-bottom: 1px solid var(--ui-border, #dbe3ee);
  color: var(--ui-text, #0f172a);
}

/* 移动端：恢复正常滚动 */
.table-page-layout.mobile-mode {
  height: auto;
  min-height: 0;
  gap: 1rem;
}

.table-page-layout.mobile-mode .table-scroll-container {
  @apply h-auto overflow-visible border-none shadow-none bg-transparent;
  width: 100%;
  min-width: 0;
  max-width: 100%;
}

.table-page-layout.mobile-mode .layout-section-scrollable {
  @apply flex-none min-h-fit;
}

.table-page-layout.mobile-mode .table-scroll-container :deep(table) {
  @apply flex-none;
  display: table;
  min-width: 100%;
}
</style>
