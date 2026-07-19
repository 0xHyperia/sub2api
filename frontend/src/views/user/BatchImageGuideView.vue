<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-col gap-3">
          <div class="space-y-2 lg:hidden">
            <div class="flex min-w-0 items-center gap-2">
              <SearchInput
                v-model="filters.taskName"
                placeholder="搜索任务名称"
                class="min-w-0 flex-1"
                @search="applyFilters"
              />
              <button
                type="button"
                class="btn btn-secondary relative h-10 shrink-0 px-3"
                :aria-expanded="showMobileFilterModal"
                aria-controls="batch-image-mobile-filters"
                @click="openMobileFilterModal"
              >
                <Icon name="filter" size="sm" />
                <span>筛选</span>
                <span v-if="activeBatchFilterCount" class="flex h-4 min-w-4 items-center justify-center rounded-full bg-brand px-1 text-[9px] font-semibold text-brand-foreground">
                  {{ activeBatchFilterCount }}
                </span>
              </button>
            </div>
            <div class="grid grid-cols-[40px_40px_minmax(0,1fr)] gap-2">
              <button type="button" class="btn btn-secondary btn-icon" :disabled="loadingKeys || loadingJobs" title="刷新" aria-label="刷新" @click="refreshPage">
                <Icon name="refresh" size="sm" :class="loadingKeys || loadingJobs ? 'animate-spin' : ''" />
              </button>
              <button type="button" class="btn btn-secondary btn-icon" title="使用说明" aria-label="使用说明" @click="showGuideModal = true">
                <Icon name="book" size="sm" />
              </button>
              <button type="button" class="btn btn-primary min-w-0" @click="openCreateModal">
                <Icon name="plus" size="sm" />
                <span class="truncate">创建批量任务</span>
              </button>
            </div>
          </div>

          <div class="hidden flex-col gap-3 lg:flex 2xl:flex-row 2xl:items-center 2xl:justify-between">
            <div class="grid w-full grid-cols-[260px_160px_144px_152px] gap-3 2xl:w-auto">
              <div class="min-w-0">
                <SearchInput
                  v-model="filters.taskName"
                  placeholder="搜索任务名称"
                  class="w-full"
                  @search="applyFilters"
                />
              </div>
              <Select v-model="filters.apiKeyId" :options="apiKeyFilterOptions" class="w-full" @change="applyFilters" />
              <Select v-model="filters.status" :options="statusFilterOptions" class="w-full" @change="applyFilters" />
              <Select v-model="filters.downloaded" :options="downloadFilterOptions" class="w-full" @change="applyFilters" />
            </div>
            <div class="flex flex-wrap items-center justify-start gap-2 sm:justify-end 2xl:flex-shrink-0">
              <button type="button" class="btn btn-secondary" :disabled="loadingJobs" @click="resetFilters">
                重置
              </button>
              <button type="button" class="btn btn-secondary" :disabled="loadingKeys || loadingJobs" :title="'刷新'" @click="refreshPage">
                <Icon name="refresh" size="md" :class="loadingKeys || loadingJobs ? 'animate-spin' : ''" />
              </button>
              <button type="button" class="btn btn-secondary" @click="showGuideModal = true">
                <Icon name="book" size="md" class="mr-2" />
                使用说明
              </button>
              <button type="button" class="btn btn-primary" @click="openCreateModal">
                <Icon name="plus" size="md" class="mr-2" />
                创建批量任务
              </button>
            </div>
          </div>

          <div
            v-if="selectedJobIds.size"
            class="flex flex-wrap items-center justify-between gap-3 rounded-panel border border-outline bg-surface px-3 py-2 shadow-card"
          >
            <span class="text-sm text-foreground-muted">
              已选择 <span class="font-medium text-foreground">{{ selectedJobIds.size }}</span> 个任务
            </span>
            <div class="flex flex-wrap items-center gap-2">
              <button
                type="button"
                class="btn btn-secondary btn-sm"
                :disabled="bulkDownloading || selectedDownloadableRows.length === 0"
                @click="downloadSelectedJobs"
              >
                <Icon :name="bulkDownloading ? 'refresh' : 'download'" size="sm" class="mr-1.5" :class="bulkDownloading ? 'animate-spin' : ''" />
                下载选中
              </button>
              <button
                type="button"
                class="btn btn-secondary btn-sm text-danger-foreground hover:bg-danger-subtle"
                :disabled="bulkDeleting"
                @click="deleteSelectedJobs"
              >
                <Icon :name="bulkDeleting ? 'refresh' : 'trash'" size="sm" class="mr-1.5" :class="bulkDeleting ? 'animate-spin' : ''" />
                删除记录
              </button>
            </div>
          </div>
        </div>
      </template>

      <template #table>
        <DataTable
          :columns="columns"
          :data="visibleBatchJobs"
          :loading="loadingKeys || loadingJobs"
          :expandable-actions="false"
          row-key="id"
        >
          <template #mobile-card="{ row }">
            <article class="min-w-0">
              <header class="flex min-w-0 items-start gap-3">
                <input
                  type="checkbox"
                  class="mt-0.5 h-4 w-4 shrink-0 rounded border-outline-strong text-brand focus:ring-focus"
                  :aria-label="batchJobSelectionLabel(row)"
                  :checked="selectedJobIds.has(row.id)"
                  @change="toggleJobSelection(row.id, ($event.target as HTMLInputElement).checked)"
                  @click.stop
                />
                <div class="min-w-0 flex-1">
                  <div class="flex min-w-0 items-center gap-2">
                    <h3 class="truncate text-sm font-semibold text-foreground" :title="row.task_name || defaultTaskName(row.created_at)">
                      {{ row.task_name || defaultTaskName(row.created_at) }}
                    </h3>
                    <span v-if="row.is_child" class="badge badge-warning shrink-0 font-normal">子任务</span>
                  </div>
                  <p class="mt-1 flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1 text-xs text-foreground-subtle">
                    <span class="max-w-40 truncate" :title="row.model">{{ row.model }}</span>
                    <span aria-hidden="true">·</span>
                    <span>{{ formatDate(row.created_at) }}</span>
                  </p>
                </div>
                <span :class="statusBadgeClass(displayJob(row))" class="badge shrink-0 whitespace-nowrap">
                  {{ statusLabel(displayJob(row)) }}
                </span>
              </header>

              <button
                v-if="row.child_count > 0 && !row.is_child"
                type="button"
                class="mt-3 flex min-h-9 w-full items-center justify-between rounded-control bg-surface-subtle px-3 text-xs font-medium text-foreground-muted"
                :aria-expanded="expandedParentIds.has(row.id)"
                @click.stop="toggleChildRows(row.id)"
              >
                <span>{{ row.child_count }} 个子任务</span>
                <Icon :name="expandedParentIds.has(row.id) ? 'chevronUp' : 'chevronDown'" size="xs" />
              </button>

              <dl class="mt-3 grid grid-cols-2 gap-2 rounded-control bg-surface-subtle px-3 py-2.5">
                <div>
                  <dt class="text-[10px] text-foreground-subtle">成功 / 失败</dt>
                  <dd class="mt-0.5 flex items-baseline gap-1.5 text-sm font-semibold tabular-nums">
                    <span class="text-success-foreground">{{ displayJob(row).success_count }}</span>
                    <span class="text-foreground-subtle">/</span>
                    <span :class="displayJob(row).fail_count > 0 ? 'text-danger-foreground' : 'text-foreground-muted'">{{ displayJob(row).fail_count }}</span>
                    <span class="text-[10px] font-normal text-foreground-subtle">共 {{ displayJob(row).item_count }}</span>
                  </dd>
                </div>
                <div class="text-right">
                  <dt class="text-[10px] text-foreground-subtle">费用</dt>
                  <dd class="mt-0.5 text-sm font-semibold tabular-nums text-foreground">{{ costLabel(displayJob(row)) }}</dd>
                </div>
              </dl>

              <div class="mt-3 grid grid-cols-[minmax(0,1fr)_minmax(0,1fr)_40px] gap-2 border-t border-outline pt-3">
                <button type="button" class="btn btn-primary btn-sm min-w-0" @click.stop="selectJob(row.id)">
                  <Icon name="eye" size="sm" />
                  <span>查看</span>
                </button>
                <button
                  type="button"
                  class="btn btn-secondary btn-sm min-w-0"
                  :disabled="!canDownload(row) || downloading"
                  @click.stop="downloadJob(row)"
                >
                  <Icon :name="isDownloadingJob(row.id) ? 'refresh' : 'download'" size="sm" :class="isDownloadingJob(row.id) ? 'animate-spin' : ''" />
                  <span>下载</span>
                </button>
                <button
                  v-if="canRetry(row) || canDeleteRecord(row)"
                  type="button"
                  :id="moreMenuTriggerId(row.id)"
                  class="btn btn-secondary btn-icon h-9 w-10"
                  :class="{ 'bg-surface-subtle text-foreground': openMoreJobId === row.id }"
                  title="更多操作"
                  aria-haspopup="menu"
                  :aria-controls="moreMenuId(row.id)"
                  :aria-expanded="openMoreJobId === row.id"
                  @click.stop="toggleMoreMenu(row, $event)"
                >
                  <Icon name="more" size="sm" />
                </button>
                <span v-else class="h-9 w-10" aria-hidden="true"></span>
              </div>
            </article>
          </template>

          <template #header-select>
            <input
              type="checkbox"
              class="h-4 w-4 rounded border-outline-strong text-brand focus:ring-focus"
              :aria-label="batchImageText('selectAllJobs')"
              :checked="allVisibleSelected"
              :indeterminate="someVisibleSelected"
              @change="toggleAllVisible(($event.target as HTMLInputElement).checked)"
            />
          </template>

          <template #cell-select="{ row }">
            <input
              type="checkbox"
              class="h-4 w-4 rounded border-outline-strong text-brand focus:ring-focus"
              :aria-label="batchJobSelectionLabel(row)"
              :checked="selectedJobIds.has(row.id)"
              @change="toggleJobSelection(row.id, ($event.target as HTMLInputElement).checked)"
              @click.stop
            />
          </template>

          <template #cell-id="{ row }">
            <div class="flex min-w-0 items-start gap-1 lg:w-[220px]" :class="row.is_child ? 'pl-6' : ''">
	              <button
	                v-if="row.child_count > 0 && !row.is_child"
	                type="button"
	                class="mt-1 flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-control text-foreground-subtle transition-colors hover:bg-surface-subtle hover:text-foreground focus:outline-none focus-visible:ring-2 focus-visible:ring-focus/30"
	                :title="expandedParentIds.has(row.id) ? '收起子任务' : `展开 ${row.child_count} 个子任务`"
	                @click.stop="toggleChildRows(row.id)"
	              >
	                <Icon :name="expandedParentIds.has(row.id) ? 'chevronDown' : 'chevronRight'" size="xs" />
	              </button>
	              <span v-else class="w-6 flex-shrink-0" />
	              <button type="button" class="min-w-0 flex-1 rounded-control py-1 text-left transition-colors hover:bg-surface-subtle focus:outline-none focus-visible:ring-2 focus-visible:ring-focus/30" @click="selectJob(row.id)">
	                <span
	                  class="flex min-w-0 items-center gap-2 text-sm font-medium"
	                  :class="row.task_name ? 'text-foreground' : 'text-foreground-muted'"
                >
                  <span class="min-w-0 truncate">{{ row.task_name || defaultTaskName(row.created_at) }}</span>
                  <span v-if="row.child_count > 0 && !row.is_child" class="badge badge-gray flex-shrink-0 font-normal">
                    {{ row.child_count }} 子任务
                  </span>
                  <span v-if="row.is_child" class="badge badge-warning flex-shrink-0 font-normal">
                    子任务
                  </span>
	                </span>
	                <span class="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1 text-xs text-foreground-muted">
	                  <span>{{ formatDate(row.created_at) }}</span>
	                </span>
	              </button>
	            </div>
	          </template>

          <template #cell-model="{ row }">
            <div class="mx-auto min-w-0 text-center lg:max-w-[180px]">
	              <p class="truncate text-sm text-foreground-muted" :title="row.model">{{ row.model }}</p>
	            </div>
	          </template>

          <template #cell-api_key_name="{ value }">
            <span class="block truncate text-center text-sm text-foreground-muted">
              {{ value || '未记录' }}
            </span>
          </template>

          <template #cell-status="{ row }">
            <div class="flex justify-center">
              <span :class="statusBadgeClass(displayJob(row))" class="badge">
                {{ statusLabel(displayJob(row)) }}
              </span>
            </div>
          </template>

          <template #cell-counts="{ row }">
            <div class="flex items-center justify-center gap-2 text-sm tabular-nums">
              <span class="text-success-foreground">{{ displayJob(row).success_count }}</span>
              <span class="text-foreground-subtle">/</span>
              <span :class="displayJob(row).fail_count > 0 ? 'text-danger-foreground' : 'text-foreground-subtle'">{{ displayJob(row).fail_count }}</span>
              <span class="text-xs text-foreground-subtle">共 {{ displayJob(row).item_count }}</span>
            </div>
          </template>

          <template #cell-cost="{ row }">
            <span class="block text-center text-sm text-foreground-muted">
              {{ costLabel(displayJob(row)) }}
            </span>
          </template>

          <template #cell-downloaded="{ row }">
            <span class="block text-center text-sm" :class="row.downloaded_at ? 'text-success-foreground' : 'text-foreground-muted'">
              {{ row.downloaded_at ? formatDate(row.downloaded_at) : '未下载' }}
            </span>
          </template>

	          <template #cell-actions="{ row }">
	            <div class="flex items-center justify-center gap-1">
              <button
                type="button"
                class="batch-row-action flex flex-col items-center gap-0.5 rounded-control p-1.5 text-foreground-subtle transition-colors hover:bg-surface-subtle hover:text-brand focus:outline-none focus-visible:ring-2 focus-visible:ring-focus/30"
                title="查看详情"
                @click="selectJob(row.id)"
              >
                <Icon name="eye" size="sm" />
                <span class="text-xs">查看</span>
              </button>
              <button
                type="button"
                class="batch-row-action flex flex-col items-center gap-0.5 rounded-control p-1.5 transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-focus/30"
                :class="canDownload(row) ? 'text-foreground-subtle hover:bg-success-subtle hover:text-success-foreground' : 'text-foreground-subtle'"
                :disabled="!canDownload(row) || downloading"
                title="下载 ZIP"
                @click="downloadJob(row)"
              >
                <Icon
                  :name="isDownloadingJob(row.id) ? 'refresh' : 'download'"
	                  size="sm"
	                  :class="isDownloadingJob(row.id) ? 'animate-spin' : ''"
	                />
                <span class="text-xs">下载</span>
	              </button>
              <div v-if="canRetry(row) || canDeleteRecord(row)">
                <button
                  type="button"
                  :id="moreMenuTriggerId(row.id)"
                  class="batch-row-action flex flex-col items-center gap-0.5 rounded-control p-1.5 text-foreground-subtle transition-colors hover:bg-surface-subtle hover:text-foreground focus:outline-none focus-visible:ring-2 focus-visible:ring-focus/30"
                  :class="{ 'bg-surface-subtle text-foreground': openMoreJobId === row.id }"
                  title="更多操作"
                  aria-haspopup="menu"
                  :aria-controls="moreMenuId(row.id)"
                  :aria-expanded="openMoreJobId === row.id"
                  @click.stop="toggleMoreMenu(row, $event)"
                >
                  <Icon name="more" size="sm" />
                  <span class="text-xs">更多</span>
                </button>
              </div>
	            </div>
	          </template>

          <template #empty>
            <div class="flex min-h-[260px] flex-col items-center justify-center py-6 md:min-h-[300px]">
              <Icon name="sparkles" size="xl" class="mb-4 h-12 w-12 text-foreground-subtle" />
              <p class="text-lg font-medium text-foreground">暂无批量任务</p>
              <p class="mt-1 text-sm text-foreground-muted">
                点击右上角创建批量任务。
              </p>
            </div>
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <div
          v-if="visibleBatchJobs.length > 0 || pagination.page > 1"
          class="flex flex-col gap-3 border-t border-outline bg-surface px-4 py-3 sm:flex-row sm:items-center sm:justify-between sm:px-6"
        >
          <div class="flex flex-wrap items-center gap-3 text-sm text-foreground-muted">
            <span>
              第 <span class="font-medium">{{ pagination.page }}</span> 页
            </span>
            <span>
              本页 <span class="font-medium">{{ visibleBatchJobs.length }}</span> 条
            </span>
            <div class="flex items-center gap-2">
              <span>每页</span>
              <Select
                v-model="pagination.page_size"
                :options="batchPageSizeOptions"
                class="w-24"
                @change="handlePageSizeChange"
              />
            </div>
          </div>
          <div class="flex items-center justify-end gap-2">
            <button
              type="button"
              class="btn btn-secondary btn-sm"
              :disabled="pagination.page <= 1 || loadingJobs"
              @click="handlePageChange(pagination.page - 1)"
            >
              <Icon name="chevronLeft" size="sm" class="mr-1" />
              上一页
            </button>
            <button
              type="button"
              class="btn btn-secondary btn-sm"
              :disabled="!pagination.has_more || loadingJobs"
              @click="handlePageChange(pagination.page + 1)"
            >
              下一页
              <Icon name="chevronRight" size="sm" class="ml-1" />
            </button>
          </div>
        </div>
      </template>
    </TablePageLayout>

    <Teleport to="body">
      <div
        v-if="openMoreJobId"
        :id="moreMenuId(openMoreJobId)"
        ref="moreMenuRef"
        role="menu"
        :aria-labelledby="moreMenuTriggerId(openMoreJobId)"
        class="fixed z-[9999] w-44 overflow-hidden rounded-panel border border-outline bg-surface py-1 text-sm shadow-floating"
        :style="moreMenuStyle"
        @click.stop
        @keydown="handleMoreMenuKeydown"
      >
        <template v-for="job in batchJobs" :key="job.id">
          <template v-if="job.id === openMoreJobId">
            <button
              v-if="canRetry(job)"
              type="button"
              role="menuitem"
              tabindex="-1"
              class="flex w-full items-center gap-2 px-3 py-2 text-left text-foreground-muted transition-colors hover:bg-warning-subtle hover:text-warning-foreground focus:outline-none focus-visible:bg-warning-subtle focus-visible:text-warning-foreground focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-focus/40 disabled:opacity-60"
              :disabled="retryingBatchId === job.id"
              @click="retryFailedJob(job)"
            >
              <Icon name="refresh" size="sm" :class="retryingBatchId === job.id ? 'animate-spin' : ''" />
              重试失败项
            </button>
            <button
              v-if="canDeleteRecord(job)"
              type="button"
              role="menuitem"
              tabindex="-1"
              class="flex w-full items-center gap-2 px-3 py-2 text-left text-danger-foreground transition-colors hover:bg-danger-subtle focus:outline-none focus-visible:bg-danger-subtle focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-focus/40 disabled:opacity-60"
              :disabled="deletingBatchId === job.id"
              @click="deleteJob(job)"
            >
              <Icon :name="deletingBatchId === job.id ? 'refresh' : 'trash'" size="sm" :class="deletingBatchId === job.id ? 'animate-spin' : ''" />
              删除记录
            </button>
          </template>
        </template>
      </div>
    </Teleport>

    <Teleport to="body">
      <div
        v-if="promptPopover.visible"
        class="batch-prompt-popover fixed z-[9999] rounded-panel border border-outline bg-surface-raised p-3 text-sm text-foreground shadow-floating"
        :style="promptPopover.style"
        @mouseenter="cancelPromptPopoverClose"
        @mouseleave="schedulePromptPopoverClose"
      >
        <div class="mb-2 flex items-center justify-between gap-3">
          <span class="text-xs font-medium text-foreground-muted">完整 Prompt</span>
          <button
            type="button"
            class="rounded-control px-2 py-1 text-xs font-medium text-brand transition-colors hover:bg-brand-subtle focus:outline-none focus-visible:ring-2 focus-visible:ring-focus/30"
            @click="copyPromptPopover"
          >
            复制
          </button>
        </div>
        <p class="max-h-48 overflow-y-auto whitespace-pre-wrap break-words leading-6 selection:bg-brand-subtle selection:text-foreground">
          {{ promptPopover.text }}
        </p>
      </div>
    </Teleport>

    <BaseDialog
      :show="showMobileFilterModal"
      title="筛选任务"
      width="narrow"
      @close="closeMobileFilterModal"
    >
      <div id="batch-image-mobile-filters" class="space-y-4">
        <label class="block text-sm font-medium text-foreground-muted">
          API Key
          <Select v-model="mobileFilterDraft.apiKeyId" :options="apiKeyFilterOptions" class="mt-1.5 w-full" />
        </label>
        <label class="block text-sm font-medium text-foreground-muted">
          状态
          <Select v-model="mobileFilterDraft.status" :options="statusFilterOptions" class="mt-1.5 w-full" />
        </label>
        <label class="block text-sm font-medium text-foreground-muted">
          下载状态
          <Select v-model="mobileFilterDraft.downloaded" :options="downloadFilterOptions" class="mt-1.5 w-full" />
        </label>
      </div>
      <template #footer>
        <button type="button" class="btn btn-secondary" @click="resetMobileFilterDraft">重置</button>
        <button type="button" class="btn btn-primary" @click="applyMobileFilterDraft">查看结果</button>
      </template>
    </BaseDialog>

    <BaseDialog :show="!!currentJob" title="任务详情" width="extra-wide" @close="closeDetail">
      <div v-if="currentJob" class="space-y-4">
        <div class="rounded-panel border border-outline bg-surface-subtle px-4 py-3">
          <div class="grid gap-x-6 gap-y-3 sm:grid-cols-2 lg:grid-cols-4">
            <div class="min-w-0 text-center">
              <p class="text-xs text-foreground-muted">状态</p>
              <div class="mt-1 flex justify-center">
                <span :class="statusBadgeClass(currentDisplayJob || currentJob)" class="badge whitespace-nowrap">
                  {{ statusLabel(currentDisplayJob || currentJob) }}
                </span>
              </div>
            </div>
            <div class="min-w-0 text-center">
              <p class="text-xs text-foreground-muted">{{ hasChildJobs(currentJob.id) ? '汇总结果' : '结果' }}</p>
              <p class="mt-1 flex items-center justify-center gap-2 font-medium tabular-nums">
              <span class="text-success-foreground">{{ (currentDisplayJob || currentJob).success_count }}</span>
              <span class="text-foreground-subtle">/</span>
              <span :class="(currentDisplayJob || currentJob).fail_count > 0 ? 'text-danger-foreground' : 'text-foreground-subtle'">{{ (currentDisplayJob || currentJob).fail_count }}</span>
            </p>
            </div>
            <div class="min-w-0 text-center">
              <p class="text-xs text-foreground-muted">费用</p>
              <p class="mt-1 truncate font-medium text-foreground">{{ costLabel(currentDisplayJob || currentJob) }}</p>
            </div>
            <div class="min-w-0 text-center">
              <p class="text-xs text-foreground-muted">下载状态</p>
              <p class="mt-1 truncate font-medium text-foreground">
              {{ currentJob.downloaded_at ? formatDate(currentJob.downloaded_at) : '未下载' }}
            </p>
            </div>
          </div>
        </div>

        <div class="flex flex-wrap items-center justify-between gap-3">
          <h3 class="text-sm font-semibold text-foreground">明细</h3>
          <button type="button" class="btn btn-secondary btn-sm" :disabled="refreshing || loadingItems" @click="refreshDetail">
            <Icon name="refresh" size="sm" class="mr-1.5" :class="refreshing || loadingItems ? 'animate-spin' : ''" />
            刷新
          </button>
        </div>

        <template v-if="items.length">
          <div class="grid gap-3 lg:hidden">
            <article
              v-for="item in items"
              :key="itemPreviewKey(item)"
              class="min-w-0 rounded-panel border border-outline bg-surface p-3"
              :class="detailItemRowClass(item)"
            >
              <header class="flex min-w-0 items-start justify-between gap-3">
                <div class="min-w-0">
                  <p class="text-[10px] font-medium uppercase text-foreground-subtle">Custom ID</p>
                  <h4 class="mt-0.5 truncate font-mono text-sm font-semibold" :class="isRecoveredOriginalFailure(item) ? 'text-foreground-subtle' : 'text-foreground'" :title="item.custom_id">
                    {{ item.custom_id }}
                  </h4>
                </div>
                <span :class="itemDisplayStatusBadgeClass(item)" class="badge max-w-36 shrink-0 truncate whitespace-nowrap" :title="itemDisplayStatusLabel(item)">
                  {{ itemDisplayStatusLabel(item) }}
                </span>
              </header>

              <div class="mt-3 grid grid-cols-[84px_minmax(0,1fr)] gap-3">
                <div class="aspect-square overflow-hidden rounded-control border border-outline bg-surface-subtle">
                  <button
                    v-if="itemPreviewUrls[itemPreviewKey(item)] && !previewErrorIds.has(itemPreviewKey(item))"
                    type="button"
                    class="block h-full w-full overflow-hidden"
                    :title="`放大压缩预览 ${item.custom_id}`"
                    @click="openImagePreview(item)"
                  >
                    <img :src="itemPreviewUrls[itemPreviewKey(item)]" class="h-full w-full object-cover" alt="" @error="handlePreviewError(itemPreviewKey(item))" />
                  </button>
                  <button
                    v-else-if="canLoadItemPreview(item)"
                    type="button"
                    class="flex h-full w-full flex-col items-center justify-center gap-1 text-foreground-subtle transition-colors hover:bg-surface-subtle hover:text-brand disabled:cursor-wait disabled:opacity-70"
                    :disabled="previewLoadingIds.has(itemPreviewKey(item))"
                    :title="previewErrorIds.has(itemPreviewKey(item)) ? '重新加载压缩预览' : '加载压缩预览'"
                    @click="loadItemPreview(item)"
                  >
                    <Icon :name="previewLoadingIds.has(itemPreviewKey(item)) ? 'refresh' : 'eye'" size="sm" :class="previewLoadingIds.has(itemPreviewKey(item)) ? 'animate-spin' : ''" />
                    <span class="text-[10px]">{{ previewErrorIds.has(itemPreviewKey(item)) ? '重试' : '预览' }}</span>
                  </button>
                  <div v-else class="flex h-full w-full flex-col items-center justify-center gap-1 text-foreground-subtle">
                    <Icon name="document" size="sm" />
                    <span class="text-[10px]">无图片</span>
                  </div>
                </div>

                <div class="min-w-0">
                  <p class="text-[10px] font-medium text-foreground-subtle">Prompt</p>
                  <p class="mt-1 line-clamp-3 text-xs leading-5" :class="isRecoveredOriginalFailure(item) ? 'text-foreground-subtle' : 'text-foreground-muted'">
                    {{ item.prompt_preview || '-' }}
                  </p>
                  <span
                    class="mt-2 inline-flex max-w-full items-center truncate rounded-control px-2 py-1 text-[11px] font-medium ring-1 ring-inset"
                    :class="itemResultClass(item)"
                    :title="itemResultLabel(item)"
                  >
                    {{ itemResultLabel(item) }}
                  </span>
                </div>
              </div>
            </article>
          </div>

          <div class="hidden overflow-x-auto rounded-panel border border-outline bg-surface lg:block">
          <table class="w-full min-w-[860px] table-fixed divide-y divide-outline text-sm">
            <colgroup>
              <col class="w-[18%]" />
              <col class="w-[34%]" />
              <col class="w-[12%]" />
              <col class="w-[10%]" />
              <col class="w-[26%]" />
            </colgroup>
            <thead class="bg-surface-subtle">
              <tr>
                <th class="px-3 py-3 text-center text-sm font-medium text-foreground-muted">Custom ID</th>
                <th class="px-3 py-3 text-left text-sm font-medium text-foreground-muted">Prompt</th>
                <th class="px-3 py-3 text-center text-sm font-medium text-foreground-muted">状态</th>
                <th class="px-3 py-3 text-center text-sm font-medium text-foreground-muted">预览</th>
                <th class="px-3 py-3 text-center text-sm font-medium text-foreground-muted">结果</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-outline">
              <tr
                v-for="item in items"
                :key="itemPreviewKey(item)"
                class="align-middle"
                :class="detailItemRowClass(item)"
              >
                <td class="px-3 py-2.5 text-center">
                  <span
                    class="block min-w-0 truncate font-mono text-sm"
                    :class="isRecoveredOriginalFailure(item) ? 'text-foreground-subtle' : 'text-foreground'"
                    :title="item.custom_id"
                  >
                    {{ item.custom_id }}
                  </span>
                </td>
                <td class="px-3 py-2.5 text-left" :class="isRecoveredOriginalFailure(item) ? 'text-foreground-subtle' : 'text-foreground-muted'">
                  <div
                    class="batch-prompt-trigger cursor-default truncate rounded px-1 text-sm leading-6 focus:outline-none focus-visible:ring-2 focus-visible:ring-focus/40"
                    tabindex="0"
                    @pointerenter="schedulePromptPopoverOpen($event, item.prompt_preview || '-')"
                    @pointerleave="schedulePromptPopoverClose"
                    @mouseenter="schedulePromptPopoverOpen($event, item.prompt_preview || '-')"
                    @mouseleave="schedulePromptPopoverClose"
                    @click="showPromptPopover($event, item.prompt_preview || '-')"
                    @focus="showPromptPopover($event, item.prompt_preview || '-')"
                    @focusin="showPromptPopover($event, item.prompt_preview || '-')"
                    @blur="schedulePromptPopoverClose"
                  >
                    {{ item.prompt_preview || '-' }}
                  </div>
                </td>
                <td class="px-3 py-2.5 text-center">
                  <span :class="itemDisplayStatusBadgeClass(item)" class="badge max-w-full truncate whitespace-nowrap" :title="itemDisplayStatusLabel(item)">
                    {{ itemDisplayStatusLabel(item) }}
                  </span>
                </td>
                <td class="px-3 py-2.5 text-center">
                  <div class="mx-auto h-12 w-12 overflow-hidden rounded-control border border-outline bg-surface-subtle">
                    <button
                      v-if="itemPreviewUrls[itemPreviewKey(item)] && !previewErrorIds.has(itemPreviewKey(item))"
                      type="button"
                      class="block h-full w-full overflow-hidden"
                      :title="`放大压缩预览 ${item.custom_id}`"
                      @click="openImagePreview(item)"
                    >
                      <img
                        :src="itemPreviewUrls[itemPreviewKey(item)]"
                        class="h-full w-full object-cover"
                        alt=""
                        @error="handlePreviewError(itemPreviewKey(item))"
                      />
                    </button>
                    <button
                      v-else-if="canLoadItemPreview(item)"
                      type="button"
                      class="flex h-full w-full items-center justify-center text-foreground-subtle transition-colors hover:bg-surface-subtle hover:text-brand disabled:cursor-wait disabled:opacity-70"
                      :disabled="previewLoadingIds.has(itemPreviewKey(item))"
                      :title="previewErrorIds.has(itemPreviewKey(item)) ? '重新加载压缩预览' : '加载压缩预览'"
                      @click="loadItemPreview(item)"
                    >
                      <Icon :name="previewLoadingIds.has(itemPreviewKey(item)) ? 'refresh' : 'eye'" size="sm" :class="previewLoadingIds.has(itemPreviewKey(item)) ? 'animate-spin' : ''" />
                    </button>
                    <div v-else class="flex h-full w-full items-center justify-center text-foreground-subtle" :title="item.image_count > 0 ? '不可预览' : '无图片'">
                      <Icon name="document" size="sm" />
                    </div>
                  </div>
                </td>
                <td class="px-3 py-2.5 text-center">
                  <span
                    class="inline-flex max-w-full items-center justify-center truncate rounded-control px-2.5 py-1 text-xs font-medium leading-5 ring-1 ring-inset"
                    :class="itemResultClass(item)"
                    :title="itemResultLabel(item)"
                  >
                    {{ itemResultLabel(item) }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
          </div>
        </template>
        <div v-else class="rounded-panel border border-dashed border-outline py-10 text-center">
          <Icon name="refresh" size="lg" class="mx-auto mb-3 text-foreground-subtle" :class="loadingItems ? 'animate-spin' : ''" />
          <p class="text-sm font-medium text-foreground-muted">
            {{ loadingItems ? '正在加载明细...' : '暂无明细' }}
          </p>
          <p v-if="!loadingItems" class="mt-1 text-sm text-foreground-muted">
            排队或生成中的任务会先显示已提交的 prompt，结果整理完成后会更新图片状态。
          </p>
        </div>
      </div>

      <template #footer>
        <div class="batch-detail-actions flex w-full flex-col-reverse gap-2 sm:flex-row sm:flex-wrap sm:justify-end">
	          <button type="button" class="btn btn-secondary w-full sm:w-auto" :disabled="!currentJob || !canCancel(currentJob) || cancelling" @click="cancelSelected">
	            <Icon v-if="cancelling" name="refresh" size="sm" class="mr-2 animate-spin" />
	            取消任务
	          </button>
	          <button
	            v-if="currentJob && currentDisplayJob && canRetry(currentDisplayJob)"
	            type="button"
	            class="btn btn-secondary inline-flex w-full min-w-[116px] items-center justify-center sm:w-auto"
	            :disabled="retryingBatchId === currentJob.id"
	            @click="retrySelected"
	          >
	            <Icon name="refresh" size="sm" class="mr-2" :class="currentJob && retryingBatchId === currentJob.id ? 'animate-spin' : ''" />
	            重试失败项
	          </button>
	          <button
            type="button"
            class="btn btn-primary inline-flex w-full min-w-[112px] items-center justify-center sm:w-auto"
            :disabled="!currentJob || !canDownload(currentJob) || downloading"
            @click="downloadSelected"
          >
            <Icon
              :name="currentJob && isDownloadingJob(currentJob.id) ? 'refresh' : 'download'"
              size="sm"
              class="mr-2"
              :class="currentJob && isDownloadingJob(currentJob.id) ? 'animate-spin' : ''"
            />
            下载 ZIP
          </button>
        </div>
      </template>
    </BaseDialog>

    <BaseDialog :show="!!previewImageItem" :title="previewImageItem?.custom_id || '图片预览'" width="extra-wide" :z-index="60" @close="closeImagePreview">
      <div class="space-y-3">
        <div class="rounded-panel border border-warning/30 bg-warning-subtle px-3 py-2 text-sm text-warning-foreground">
          当前显示的是浏览器本地缓存的压缩缩略图，清晰度会有影响；需要查看原图请下载 ZIP。
        </div>
        <div class="flex min-h-[420px] items-center justify-center rounded-panel bg-canvas p-4">
          <img
            v-if="previewImageUrl"
            :src="previewImageUrl"
            class="max-h-[70vh] max-w-full rounded-control object-contain"
            :alt="previewImageItem?.custom_id || ''"
          />
        </div>
      </div>
    </BaseDialog>

    <BaseDialog :show="showCreateModal" title="创建批量任务" width="wide" @close="closeCreateModal">
      <form class="space-y-5" @submit.prevent="submitJob">
        <div class="grid gap-4 md:grid-cols-2">
          <div class="md:col-span-2">
            <label class="input-label">任务名称</label>
            <input
              v-model="form.taskName"
              type="text"
              maxlength="255"
              class="input"
              placeholder="不填写则默认使用当前时间"
            />
          </div>

          <div class="md:col-span-2">
            <label class="input-label">API Key</label>
            <select v-model.number="form.apiKeyId" class="input" :disabled="loadingKeys">
              <option :value="0">{{ loadingKeys ? '加载 API Key 中...' : '请选择 Gemini API Key' }}</option>
              <option v-for="key in geminiApiKeys" :key="key.id" :value="key.id">
                {{ key.name }} · {{ key.group?.name || 'Gemini' }}
              </option>
            </select>
            <p v-if="!loadingKeys && geminiApiKeys.length === 0" class="input-hint text-warning-foreground">
              当前没有可用于批量生图的 Gemini API Key。请先创建并绑定已开启批量生图的 Gemini 分组。
            </p>
          </div>

          <div>
            <label class="input-label">模型</label>
            <select v-model="form.model" class="input" :disabled="loadingModels || availableBatchImageModels.length === 0">
              <option v-if="loadingModels" value="">{{ batchImageText('loadingModels') }}</option>
              <option v-else-if="availableBatchImageModels.length === 0" value="">{{ batchImageText('noModels') }}</option>
              <option v-for="model in availableBatchImageModels" :key="model.value" :value="model.value">
                {{ model.label }}
              </option>
            </select>
            <p v-if="modelLoadError" class="input-hint text-warning-foreground">
              {{ modelLoadError }}
            </p>
            <p v-else-if="selectedApiKey && !loadingModels && availableBatchImageModels.length === 0" class="input-hint text-warning-foreground">
              {{ batchImageText('noModelsHint') }}
            </p>
          </div>

          <div>
            <label class="input-label">图片尺寸</label>
            <div class="input flex items-center bg-surface-subtle text-foreground-muted">
              1K
            </div>
            <p class="input-hint">当前批量任务固定按 1K 图片提交。</p>
          </div>

          <div>
            <label class="input-label">输出格式</label>
            <select v-model="form.responseMimeType" class="input">
              <option value="image/png">PNG</option>
              <option value="image/jpeg">JPEG</option>
              <option value="image/webp">WebP</option>
            </select>
          </div>

          <div>
            <label class="input-label">预计生成</label>
            <div class="input flex items-center bg-surface-subtle text-foreground-muted">
              {{ estimatedOutputCount }} 张 / {{ promptRows.length }} 条
            </div>
          </div>
        </div>

        <div class="space-y-3">
          <div class="flex items-center justify-between gap-3">
            <label class="input-label mb-0">Prompt</label>
            <span class="text-xs text-foreground-muted">已添加 {{ promptRows.length }} 条</span>
          </div>
          <div class="rounded-panel border border-outline bg-surface-subtle p-3">
            <textarea
              v-model="promptDraft"
              rows="3"
              class="h-[76px] w-full resize-y rounded-control border border-outline-strong bg-surface px-3 py-2 text-sm leading-5 text-foreground outline-none focus:border-focus focus:ring-2 focus:ring-focus/20"
              placeholder="粘贴 prompt，添加后进入下方列表"
            />
            <div class="mt-2 grid gap-2 md:grid-cols-[minmax(0,1fr)_112px_132px_112px] md:items-center">
              <input
                v-model="customIdDraft"
                type="text"
                maxlength="255"
                class="input h-9 text-sm"
                placeholder="Custom ID 可选"
              />
              <select
                v-model.number="outputCountDraft"
                class="batch-output-count-select input h-9 text-sm"
                title="每条生成张数"
                aria-label="每条生成张数"
              >
                <option v-for="count in outputCountOptions" :key="count" :value="count">
                  {{ count }} 张
                </option>
              </select>
              <label
                class="btn btn-secondary h-9 cursor-pointer justify-center text-sm"
                :class="referenceImageDrafts.length >= selectedModelReferenceLimit ? 'pointer-events-none opacity-60' : ''"
              >
                <Icon name="upload" size="sm" class="mr-1.5" />
                参考图
                <input
                  type="file"
                  accept="image/png,image/jpeg,image/webp"
                  multiple
                  class="hidden"
                  :disabled="referenceImageDrafts.length >= selectedModelReferenceLimit"
                  @change="handleReferenceImageFiles"
                />
              </label>
              <button type="button" class="btn btn-secondary h-9 justify-center whitespace-nowrap px-4 text-sm" :disabled="!promptDraft.trim()" @click="addPromptRow">
                <Icon name="plus" size="sm" class="mr-1.5" />
                添加
              </button>
            </div>
            <div v-if="referenceImageDrafts.length" class="mt-3 flex flex-wrap gap-2">
              <span
                v-for="(ref, refIndex) in referenceImageDrafts"
                :key="`${ref.name}-${refIndex}`"
                class="inline-flex max-w-full items-center gap-1 rounded-control border border-outline bg-surface px-2 py-1 text-xs text-foreground-muted"
              >
                <span class="max-w-[180px] truncate">{{ ref.name }}</span>
                <button type="button" class="text-foreground-subtle hover:text-danger-foreground" title="移除参考图" @click="removeReferenceImageDraft(refIndex)">
                  <Icon name="x" size="xs" />
                </button>
              </span>
            </div>
            <p class="mt-2 text-xs text-foreground-muted">
              每条最多 {{ BATCH_IMAGE_MAX_OUTPUTS_PER_ITEM }} 张，整组最多 {{ BATCH_IMAGE_MAX_OUTPUTS_PER_JOB }} 张；当前模型每条最多 {{ selectedModelReferenceLimit }} 张参考图，参考图按生成张数重复消耗输入 token。
            </p>
          </div>
          <div v-if="promptRows.length" class="overflow-hidden rounded-panel border border-outline bg-surface">
            <div
              v-for="(row, index) in promptRows"
              :key="row.localId"
              class="flex items-center gap-3 border-b border-outline px-3 py-2 last:border-b-0"
            >
              <span class="w-20 flex-shrink-0 font-mono text-xs text-foreground-muted">{{ row.custom_id }}</span>
              <p class="min-w-0 flex-1 truncate text-sm text-foreground">{{ row.prompt }}</p>
              <span v-if="row.output_count > 1" class="flex-shrink-0 text-xs text-foreground-muted">
                x{{ row.output_count }}
              </span>
              <span v-if="row.reference_images.length" class="flex-shrink-0 text-xs text-foreground-muted">
                {{ row.reference_images.length }} 参考图
              </span>
              <button type="button" class="btn-ghost btn-icon flex-shrink-0 text-danger-foreground hover:bg-danger-subtle" title="删除" @click="removePromptRow(index)">
                <Icon name="trash" size="sm" />
              </button>
            </div>
          </div>
          <div v-else class="rounded-panel border border-dashed border-outline px-3 py-6 text-center text-sm text-foreground-subtle">
            还没有添加 prompt。
          </div>
        </div>

	        <div class="rounded-panel border border-warning/30 bg-warning-subtle p-3 text-sm leading-6 text-warning-foreground">
	          取消任务会请求上游取消；已被系统索引为成功的图片仍会按成功项结算扣费，其余冻结金额会释放。
	        </div>
	        <div v-if="submitting" class="rounded-panel border border-info/30 bg-info-subtle p-3 text-sm leading-6 text-info-foreground">
	          正在创建上游批量任务，通常需要几秒，请不要重复提交。
	        </div>
	      </form>

      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" :disabled="submitting" @click="closeCreateModal">取消</button>
	          <button type="button" class="btn btn-primary inline-flex min-w-[120px] justify-center" :disabled="submitting || loadingModels || (parsedItems.length === 0 && !promptDraft.trim()) || !selectedApiKey || !form.model" @click="submitJob">
            <Icon v-if="submitting" name="refresh" size="sm" class="mr-2 animate-spin" />
            {{ submitting ? '提交中...' : '提交任务' }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <BaseDialog :show="showGuideModal" title="批量生图使用说明" width="wide" @close="showGuideModal = false">
	      <div class="space-y-5">
	        <section class="space-y-3">
	          <h3 class="text-sm font-semibold text-foreground">当前界面如何使用</h3>
	          <div class="rounded-panel border border-outline bg-surface-subtle p-3 text-sm leading-6 text-foreground-muted">
	            <p>1. 选择已开启批量生图的 Gemini API Key，模型列表会按该 Key 所属分组可用模型展示。</p>
	            <p>2. 任务名称可以留空，提交时会自动使用当前时间；Prompt 需要一条条添加到列表里，每条 Prompt 可附参考图，也可以设置重复生成张数。</p>
	            <p>3. 提交后任务会先排队，明细会展示已提交的 Prompt；图片预览默认不加载，点击明细里的预览按钮才会加载单张图。</p>
	            <p>4. 完成后可以下载 ZIP；部分失败时，更多菜单里可以只重试失败项。当前结算仍按成功输出图张数计算，不单独对参考图加价。</p>
	          </div>
	        </section>
	        <section class="space-y-3">
	          <div class="flex flex-wrap items-center justify-between gap-3">
	            <h3 class="text-sm font-semibold text-foreground">给 Codex 的 Skill 说明</h3>
	            <p class="text-xs text-foreground-muted">用于告诉 Codex 如何代替用户整理 prompt、提交任务和下载结果。</p>
	          </div>
	        <textarea
	          :value="agentInstruction"
	          readonly
	          class="min-h-[420px] w-full resize-y rounded-control border border-outline-strong bg-canvas p-4 font-mono text-sm leading-6 text-foreground outline-none focus:border-focus focus:ring-2 focus:ring-focus/20"
	        />
	        </section>
	      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" @click="showGuideModal = false">关闭</button>
          <button type="button" class="btn btn-primary" @click="copyInstruction">
            <Icon name="copy" size="sm" class="mr-2" />
            复制说明
          </button>
        </div>
      </template>
    </BaseDialog>

    <ConfirmDialog
      :show="!!pendingConfirmation"
      :title="confirmationDialogTitle"
      :message="confirmationDialogMessage"
      :confirm-text="confirmationDialogConfirmText"
      :cancel-text="batchImageText('confirmCancel')"
      :danger="pendingConfirmation?.type !== 'cancel'"
      @confirm="confirmPendingAction"
      @cancel="closeConfirmationDialog"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import SearchInput from '@/components/common/SearchInput.vue'
import Icon from '@/components/icons/Icon.vue'
import { useClipboard } from '@/composables/useClipboard'
import { getPersistedPageSize, setPersistedPageSize } from '@/composables/usePersistedPageSize'
import { useAppStore } from '@/stores/app'
import { keysAPI } from '@/api'
import {
  cancelBatchImageJob,
  deleteBatchImageJobRecord,
  downloadBatchImageZip,
  getBatchImageItemContent,
  getBatchImageJob,
  listBatchImageJobs,
  listBatchImageItems,
  listBatchImageModels,
  saveBlob,
  submitBatchImageJob,
  type BatchImageItem,
  type BatchImageJob,
  type BatchImageJobsListOptions,
  type BatchImageReferenceImage,
  type BatchImageStatus,
  type BatchImageSubmitItem,
} from '@/api/batchImage'
import type { ApiKey } from '@/types'
import type { Column } from '@/components/common/types'

type BatchImageJobRow = Pick<BatchImageJob, 'id' | 'task_name' | 'parent_batch_id' | 'status' | 'model' | 'provider' | 'item_count' | 'success_count' | 'fail_count' | 'estimated_cost' | 'hold_amount' | 'actual_cost' | 'created_at' | 'downloaded_at'> & {
  api_key_id: number
  api_key_name: string
  child_count: number
  is_child?: boolean
}

type BatchImageDetailItem = BatchImageItem & {
  batch_id: string
  source_task_name: string
}

type PromptRow = {
  localId: string
  custom_id: string
  prompt: string
  output_count: number
  reference_images: BatchImageReferenceImage[]
}

type PendingConfirmation =
  | { type: 'cancel'; jobId: string }
  | { type: 'delete'; job: BatchImageJobRow }
  | { type: 'delete-selected'; jobIds: string[] }

type ReferenceImageDraft = BatchImageReferenceImage & {
  name: string
  size: number
}

type PreviewCacheRecord = {
  key: string
  blob: Blob
  size: number
  createdAt: number
  lastAccessedAt: number
}

type PreviewImageSource = ImageBitmap | HTMLImageElement

const TERMINAL_STATUSES = new Set(['completed', 'failed', 'cancelled', 'output_deleted'])
const PREVIEW_CACHE_DB_NAME = 'sub2api-batch-image-preview-cache'
const PREVIEW_CACHE_STORE_NAME = 'thumbnails'
const PREVIEW_THUMBNAIL_MAX_EDGE = 360
const PREVIEW_THUMBNAIL_QUALITY = 0.72
const PREVIEW_CACHE_MAX_AGE_MS = 3 * 24 * 60 * 60 * 1000
const PREVIEW_CACHE_MAX_ENTRIES = 120
const PREVIEW_CACHE_MAX_BYTES = 48 * 1024 * 1024
const BATCH_IMAGE_MAX_OUTPUTS_PER_ITEM = 4
const BATCH_IMAGE_MAX_OUTPUTS_PER_JOB = 200
const outputCountOptions = Array.from({ length: BATCH_IMAGE_MAX_OUTPUTS_PER_ITEM }, (_, index) => index + 1)
const batchPageSizeOptions: SelectOption[] = [20, 50, 100].map(size => ({ value: size, label: String(size) }))

const appStore = useAppStore()
const { copyToClipboard } = useClipboard()
const { locale } = useI18n()

const columns: Column[] = [
  { key: 'select', label: '', sortable: false, class: 'w-12 text-center' },
  { key: 'id', label: '任务名称', sortable: false, class: 'w-[240px] max-w-[240px]' },
  { key: 'model', label: '模型', sortable: false, class: 'w-[180px] max-w-[180px] text-center' },
  { key: 'api_key_name', label: '提交密钥', sortable: false, class: 'w-40 max-w-40 text-center' },
  { key: 'status', label: '状态', sortable: false, class: 'w-28 text-center' },
  { key: 'counts', label: '结果', sortable: false, class: 'w-32 text-center' },
  { key: 'cost', label: '费用', sortable: false, class: 'w-36 text-center' },
  { key: 'downloaded', label: '下载状态', sortable: false, class: 'w-40 text-center' },
  { key: 'actions', label: '操作', sortable: false, class: 'w-40 text-center' },
]

const statusFilterOptions: SelectOption[] = [
  { value: '', label: '全部状态' },
  { value: 'queued', label: '排队中' },
  { value: 'running', label: '生成中' },
  { value: 'processing_results', label: '整理结果' },
  { value: 'settling', label: '结算中' },
  { value: 'completed', label: '已完成' },
  { value: 'failed', label: '失败' },
  { value: 'cancelled', label: '已取消' },
  { value: 'output_deleted', label: '结果已删除' },
]

const downloadFilterOptions: SelectOption[] = [
  { value: '', label: '全部下载状态' },
  { value: 'true', label: '已下载' },
  { value: 'false', label: '未下载' },
]

const form = reactive({
  apiKeyId: 0,
  taskName: '',
  model: '',
  responseMimeType: 'image/png',
})

const filters = reactive({
  taskName: '',
  apiKeyId: '',
  status: '',
  downloaded: '',
})

const showMobileFilterModal = ref(false)
const mobileFilterDraft = reactive({
  apiKeyId: '',
  status: '',
  downloaded: '',
})
const activeBatchFilterCount = computed(() =>
  [filters.apiKeyId, filters.status, filters.downloaded].filter(Boolean).length
)

const pagination = reactive({
  page: 1,
  page_size: Math.min(getPersistedPageSize(20), 100),
  has_more: false,
})

const apiKeys = ref<ApiKey[]>([])
const loadingKeys = ref(false)
const loadingJobs = ref(false)
const submitting = ref(false)
const refreshing = ref(false)
const cancelling = ref(false)
const downloading = ref(false)
const downloadingBatchId = ref('')
const retryingBatchId = ref('')
const bulkDownloading = ref(false)
const bulkDeleting = ref(false)
const deletingBatchId = ref('')
const loadingItems = ref(false)
const loadingModels = ref(false)
const showCreateModal = ref(false)
const showGuideModal = ref(false)
const currentJob = ref<BatchImageJob | null>(null)
const selectedBatchId = ref('')
const selectedBatchApiKeyId = ref(0)
const items = ref<BatchImageDetailItem[]>([])
const batchJobs = ref<BatchImageJobRow[]>([])
const selectedJobIds = ref(new Set<string>())
const expandedParentIds = ref(new Set<string>())
const promptRows = ref<PromptRow[]>([])
const promptDraft = ref('')
const customIdDraft = ref('')
const outputCountDraft = ref(1)
const referenceImageDrafts = ref<ReferenceImageDraft[]>([])
const itemPreviewUrls = reactive<Record<string, string>>({})
const previewLoadingIds = ref(new Set<string>())
const previewErrorIds = ref(new Set<string>())
const previewImageItem = ref<BatchImageItem | null>(null)
const availableBatchImageModels = ref<Array<{ value: string; label: string }>>([])
const modelLoadError = ref('')
const openMoreJobId = ref('')
const moreMenuRef = ref<HTMLElement | null>(null)
const moreMenuStyle = ref<Record<string, string>>({})
const pendingConfirmation = ref<PendingConfirmation | null>(null)
const promptPopover = reactive({
  visible: false,
  text: '',
  style: {} as Record<string, string>,
})
let modelRequestSeq = 0
let pollTimer: ReturnType<typeof setInterval> | null = null
let previewCacheDBPromise: Promise<IDBDatabase | null> | null = null
let previewCacheCleanupTimer: ReturnType<typeof setInterval> | null = null
let promptPopoverCloseTimer: ReturnType<typeof setTimeout> | null = null
let promptPopoverOpenTimer: ReturnType<typeof setTimeout> | null = null
let activePromptPopoverTarget: HTMLElement | null = null
let moreMenuTrigger: HTMLElement | null = null

const geminiApiKeys = computed(() =>
  apiKeys.value.filter((key) =>
    key.status === 'active' &&
    key.group?.platform === 'gemini' &&
    key.group?.allow_batch_image_generation === true,
  ),
)

const selectedApiKey = computed(() =>
  geminiApiKeys.value.find((key) => key.id === Number(form.apiKeyId)) || null,
)

const filteredApiKeys = computed(() => {
  const selectedFilterID = Number(filters.apiKeyId || 0)
  if (!selectedFilterID) return geminiApiKeys.value
  return geminiApiKeys.value.filter(key => key.id === selectedFilterID)
})

const apiKeyFilterOptions = computed<SelectOption[]>(() => [
  { value: '', label: '全部 API Key' },
  ...geminiApiKeys.value.map(key => ({
    value: String(key.id),
    label: key.name || `API Key #${key.id}`,
  })),
])

const selectedRows = computed(() =>
  batchJobs.value.filter(job => selectedJobIds.value.has(job.id)),
)

const confirmationDialogTitle = computed(() => {
  if (pendingConfirmation.value?.type === 'cancel') return batchImageText('cancelTitle')
  if (pendingConfirmation.value?.type === 'delete-selected') return batchImageText('deleteSelectedTitle')
  return batchImageText('deleteTitle')
})

const confirmationDialogMessage = computed(() => {
  if (pendingConfirmation.value?.type === 'cancel') return batchImageText('cancelConfirm')
  if (pendingConfirmation.value?.type === 'delete-selected') return batchImageText('deleteSelectedConfirm')
  return batchImageText('deleteConfirm')
})

const confirmationDialogConfirmText = computed(() =>
  pendingConfirmation.value?.type === 'cancel'
    ? batchImageText('cancelAction')
    : batchImageText('deleteAction')
)

const childrenByParent = computed(() => {
  const groups = new Map<string, BatchImageJobRow[]>()
  for (const job of batchJobs.value) {
    if (!job.parent_batch_id) continue
    const rows = groups.get(job.parent_batch_id) || []
    rows.push(job)
    groups.set(job.parent_batch_id, rows)
  }
  for (const rows of groups.values()) {
    rows.sort((a, b) => a.created_at - b.created_at)
  }
  return groups
})

const visibleBatchJobs = computed(() => {
  const rows: BatchImageJobRow[] = []
  for (const job of batchJobs.value.filter(item => !item.parent_batch_id)) {
    rows.push(job)
    if (expandedParentIds.value.has(job.id)) {
      rows.push(...(childrenByParent.value.get(job.id) || []).map(child => ({ ...child, is_child: true })))
    }
  }
  return rows
})

const selectedDownloadableRows = computed(() =>
  selectedRows.value.filter(job => canDownload(job)),
)

const allVisibleSelected = computed(() =>
  visibleBatchJobs.value.length > 0 && visibleBatchJobs.value.every(job => selectedJobIds.value.has(job.id)),
)

const someVisibleSelected = computed(() =>
  visibleBatchJobs.value.some(job => selectedJobIds.value.has(job.id)) && !allVisibleSelected.value,
)

const previewImageUrl = computed(() => {
  const item = previewImageItem.value
  if (!item) return ''
  return itemPreviewUrls[itemPreviewKey(item)] || ''
})

const recoveredOriginalCustomIds = computed(() => {
  const rootBatchId = detailRootBatchId()
  if (!rootBatchId) return new Set<string>()
  const ids = new Set<string>()
  for (const item of items.value) {
    if (!isChildDetailItem(item) || !isSuccessfulImageItem(item)) continue
    const sourceCustomID = retrySourceCustomID(item.custom_id)
    if (sourceCustomID) ids.add(sourceCustomID)
  }
  return ids
})

const currentDisplayJob = computed(() => {
  if (!currentJob.value) return null
  return displayJob(currentJob.value)
})

const endpointBase = computed(() => {
  const configured = appStore.apiBaseUrl?.trim()
  if (configured) return configured.replace(/\/+$/, '')
  if (typeof window !== 'undefined') return window.location.origin.replace(/\/+$/, '')
  return '<你的 Sub2API API 端点>'
})

const selectedModelReferenceLimit = computed(() => referenceImageLimitForModel(form.model))

const estimatedOutputCount = computed(() =>
  promptRows.value.reduce((sum, row) => sum + normalizeOutputCount(row.output_count), 0),
)

const parsedItems = computed<BatchImageSubmitItem[]>(() => {
  const used = new Set<string>()
  return promptRows.value
    .map((row, index) => {
      const customID = uniqueCustomID(row.custom_id || `img_${String(index + 1).padStart(3, '0')}`, used, index)
      const item: BatchImageSubmitItem = { custom_id: customID, prompt: row.prompt.trim() }
      const outputCount = normalizeOutputCount(row.output_count)
      if (outputCount > 1) {
        item.output_count = outputCount
      }
      if (row.reference_images.length) {
        item.reference_images = row.reference_images
      }
      return item
    })
    .filter(item => item.prompt)
})

function referenceImageLimitForModel(model: string) {
  const normalized = String(model || '').toLowerCase()
  if (normalized.includes('pro-image')) return 14
  if (normalized.includes('flash-image')) return 3
  return 0
}

const agentInstruction = computed(() => `---
name: sub2api-batch-image
description: 当用户希望用 Gemini/Vertex 批量生成图片、批量跑提示词、下载批量生图结果、重试失败图片时使用。
---

你是 Codex 中的批量生图执行 Agent。用户不需要手动填写页面表单；你应从当前聊天、用户给的文件、目录或上下文中整理任务名称、prompt 列表和输出目录，只有缺少关键决策时才向用户提问。

默认端点：
${endpointBase.value}

你需要自己完成：
1. 从用户聊天或附件中提取 prompt。每条 prompt 保留完整文本，按顺序生成稳定 custom_id，例如 img_001、img_002。
2. 从用户要求或上下文推断任务名称；没有明确名称时用当前时间生成任务名。
3. 从用户要求或上下文推断输出目录；如果用户没有说保存到哪里，才询问用户。
4. 提交前必须先计算 expected_output_count = 所有 item 的 output_count 之和。单个批量任务硬性最多 200 张输出图；超过 200 张必须拆成多组任务，不能提交一个超大任务，也不能把参考图附件上限当成生成张数上限。
5. 如果用户提供参考图，把参考图按用途绑定到具体 item。参考图只是输入附件，不是输出图数量。模型单条限制必须按模型执行：Gemini 2.5 Flash Image 每条最多 3 张参考图；Gemini 3 Pro Image 每条最多 14 张参考图。不要把后端附件风控理解成 Pro 单条能力：按 output_count 展开后，所有 item 的参考图附件总数还有内部保护阈值 1000 个，inline base64 参考图解码后总量最多 128MB。这个 1000 只是服务器拒绝异常请求的保护阈值，不是推荐规模；参考图很多或总请求体较大时应主动拆分任务。
6. 参考图会按 output_count 重复消耗输入 token；大量任务、重复复用同一张参考图或参考图总体积较大时，优先使用 gs:// file_uri 或拆分成多组任务。
7. 选择 API Key 和模型：先获取当前可用的批量生图 Key/模型；如果用户指定模型且该 Key 支持，则使用用户指定模型；否则使用该 Key 可用模型中的默认/第一个。不要展示或询问内部 provider 名称。
8. 调用批量生图 API 提交、轮询、下载，不要求用户去页面里手填。

API 调用规范：
- 模型：GET ${joinEndpointPath(endpointBase.value, '/v1/images/batches/models')}
- 提交：POST ${joinEndpointPath(endpointBase.value, '/v1/images/batches')}
- 查询：GET ${joinEndpointPath(endpointBase.value, '/v1/images/batches/{id}')}
- 明细：GET ${joinEndpointPath(endpointBase.value, '/v1/images/batches/{id}/items')}
- 下载：GET ${joinEndpointPath(endpointBase.value, '/v1/images/batches/{id}/download')}
- 取消：POST ${joinEndpointPath(endpointBase.value, '/v1/images/batches/{id}/cancel')}

提交请求体：
{
  "model": "<按所选 Key 可用模型填写>",
  "task_name": "<从聊天推断；为空则用当前时间>",
  "image_size": "1K",
  "response_mime_type": "image/png",
  "items": [
    {
      "custom_id": "img_001",
      "prompt": "<第一条完整 prompt>",
      "output_count": 1,
      "reference_images": [
        {
          "id": "face",
          "type": "subject",
          "mime_type": "image/png",
          "data": "<base64，不含 data:image/png;base64, 前缀>"
        }
      ]
    }
  ]
}

必须遵守：
- 不要把 API Key 写入仓库、日志、提交记录或最终回复。
- 不要把参考图 base64 写入最终回复、日志或公开文件。恢复记录中只保存参考图文件名、用途、数量和请求 JSON 文件路径；若请求 JSON 文件包含 base64，应保存在用户指定输出目录且不要提交到仓库。
- output_count 表示同一 prompt 和参考图重复生成几张，默认 1，每条最多 4；这不是依赖 Gemini 单次请求返回多图，而是系统展开成多个真实任务项。提交前必须确认预计输出图总数不超过 200，超过就拆分成多组任务。绝不能因为参考图附件有更高的内部保护阈值，就提交会生成超过 200 张图的任务。
- 当前对用户的批量生图计费仍按成功输出图片数量结算，不单独对参考图加价。可以向用户说明：参考图会产生少量上游输入 token 和临时存储成本，且会随 output_count 重复计算；页面显示的冻结/结算金额按输出图片数量计算。
- 提交成功后，必须立刻在输出目录写入本地恢复记录，例如 batch-image-resume.json。不要在恢复记录里保存 API Key。
- 恢复记录至少包含：endpoint、task_name、batch_id、model、output_dir、request_file、submitted_at、last_status、status_url、items_url、download_url、prompt_count、expected_output_count，以及可用于失败重试的 custom_id 到 prompt 映射或请求 JSON 文件路径。
- 每次查询状态后更新恢复记录，写入 last_checked_at、last_status、成功数、失败数、实际扣费和失败摘要。会话中断或暂停后，下次必须能凭该文件继续查询、下载或重试。
- 不要高频轮询。首次查询等待约 20 到 30 秒；queued 状态每 60 到 120 秒查询一次；如果连续 3 次仍是 queued，就先停止主动查询，告诉用户任务仍在排队，并保留恢复记录，之后可继续其他任务或等待用户稍后让你恢复。
- running 状态每约 60 秒查询一次，服务器压力大或大批量任务时可以更久；processing_results 等接近完成的状态可每 20 到 45 秒查询一次。
- 任务完成后报告任务名、任务 id、成功数、失败数、实际扣费和保存路径。
- 只下载成功图片。部分失败时，先展示失败 custom_id、错误码、错误来源和简要原因。
- 重试只能重试失败项，不能重复提交已成功项。若历史任务没有保存失败项 prompt，必须告诉用户无法自动重试，并询问用户是否提供原 prompt。
- 取消任务前必须提醒：已被系统索引为成功的图片仍会按成功项结算扣费，其余冻结金额会释放。
- 图片预览按需加载；不要为了查看列表自动批量加载图片内容。`)

function joinEndpointPath(base: string, path: string): string {
  return `${base.replace(/\/+$/, '')}/${path.replace(/^\/+/, '')}`
}

function uniqueCustomID(raw: string, used: Set<string>, index: number): string {
  const base = raw.replace(/[^\w.-]+/g, '_').replace(/^_+|_+$/g, '') || `img_${String(index + 1).padStart(3, '0')}`
  let candidate = base
  let suffix = 2
  while (used.has(candidate)) {
    candidate = `${base}_${suffix}`
    suffix += 1
  }
  used.add(candidate)
  return candidate
}

function normalizeOutputCount(value: unknown): number {
  const parsed = Math.floor(Number(value || 1))
  if (!Number.isFinite(parsed)) return 1
  return Math.min(BATCH_IMAGE_MAX_OUTPUTS_PER_ITEM, Math.max(1, parsed))
}

function addPromptRow() {
  const prompt = promptDraft.value.trim()
  if (!prompt) return
  const outputCount = normalizeOutputCount(outputCountDraft.value)
  const used = new Set(promptRows.value.map(row => row.custom_id))
  const customID = uniqueCustomID(customIdDraft.value || `img_${String(promptRows.value.length + 1).padStart(3, '0')}`, used, promptRows.value.length)
  promptRows.value = [
    ...promptRows.value,
    {
      localId: `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
      custom_id: customID,
      prompt,
      output_count: outputCount,
      reference_images: referenceImageDrafts.value.map(({ name: _name, size: _size, ...ref }) => ref),
    },
  ]
  promptDraft.value = ''
  customIdDraft.value = ''
  outputCountDraft.value = 1
  referenceImageDrafts.value = []
}

function removePromptRow(index: number) {
  promptRows.value = promptRows.value.filter((_, currentIndex) => currentIndex !== index)
}

function removeReferenceImageDraft(index: number) {
  referenceImageDrafts.value = referenceImageDrafts.value.filter((_, currentIndex) => currentIndex !== index)
}

async function handleReferenceImageFiles(event: Event) {
  const input = event.target as HTMLInputElement
  const files = Array.from(input.files || [])
  input.value = ''
  if (files.length === 0) return
  const limit = selectedModelReferenceLimit.value
  if (limit <= 0) {
    appStore.showError('当前模型不支持参考图。')
    return
  }
  const slots = Math.max(0, limit - referenceImageDrafts.value.length)
  if (slots <= 0) {
    appStore.showError(`当前模型每条最多 ${limit} 张参考图。`)
    return
  }
  const accepted = files.slice(0, slots)
  if (accepted.length < files.length) {
    appStore.showError(`当前模型每条最多 ${limit} 张参考图，已忽略超出的文件。`)
  }
  const next: ReferenceImageDraft[] = []
  for (const file of accepted) {
    if (!['image/png', 'image/jpeg', 'image/webp'].includes(file.type)) {
      appStore.showError('参考图仅支持 PNG、JPEG 或 WebP。')
      continue
    }
    if (file.size > 10 * 1024 * 1024) {
      appStore.showError(`${file.name} 超过 10MB，已忽略。`)
      continue
    }
    const data = await readFileAsBase64(file)
    next.push({
      id: file.name,
      type: 'reference',
      mime_type: file.type,
      data,
      name: file.name,
      size: file.size,
    })
  }
  referenceImageDrafts.value = [...referenceImageDrafts.value, ...next]
}

function readFileAsBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onerror = () => reject(reader.error || new Error('Failed to read file'))
    reader.onload = () => {
      const result = String(reader.result || '')
      resolve(result.includes(',') ? result.slice(result.indexOf(',') + 1) : result)
    }
    reader.readAsDataURL(file)
  })
}

async function loadApiKeys() {
  loadingKeys.value = true
  try {
    const response = await keysAPI.list(1, 100, { status: 'active', sort_by: 'created_at', sort_order: 'desc' })
    apiKeys.value = response.items || []
    if (!selectedApiKey.value && geminiApiKeys.value.length > 0) {
      form.apiKeyId = geminiApiKeys.value[0].id
    }
    if (filters.apiKeyId && !geminiApiKeys.value.some(key => String(key.id) === filters.apiKeyId)) {
      filters.apiKeyId = ''
    }
    if (!selectedApiKey.value) {
      availableBatchImageModels.value = []
      form.model = ''
    }
  } catch (error: any) {
    appStore.showError(batchImageErrorMessage(error, batchImageText('loadKeysFailed')))
  } finally {
    loadingKeys.value = false
  }
}

async function loadAvailableModels() {
  const key = selectedApiKey.value
  const requestID = ++modelRequestSeq
  modelLoadError.value = ''
  availableBatchImageModels.value = []
  form.model = ''
  if (!key) return

  loadingModels.value = true
  try {
    const result = await listBatchImageModels(key.key)
    if (requestID !== modelRequestSeq) return
    const seen = new Set<string>()
    availableBatchImageModels.value = (result.data || [])
      .map(model => String(model.id || '').trim())
      .filter((model) => {
        if (!model || seen.has(model)) return false
        seen.add(model)
        return true
      })
      .map(model => ({ value: model, label: model }))
    form.model = availableBatchImageModels.value[0]?.value || ''
  } catch (error: any) {
    if (requestID !== modelRequestSeq) return
    modelLoadError.value = batchImageErrorMessage(error, batchImageText('loadModelsFailed'))
  } finally {
    if (requestID === modelRequestSeq) {
      loadingModels.value = false
    }
  }
}

async function refreshPage() {
  await loadApiKeys()
  await loadBatchJobs()
}

function applyFilters() {
  pagination.page = 1
  selectedJobIds.value = new Set()
  void loadBatchJobs()
}

function resetFilters() {
  filters.taskName = ''
  filters.apiKeyId = ''
  filters.status = ''
  filters.downloaded = ''
  applyFilters()
}

function openMobileFilterModal() {
  mobileFilterDraft.apiKeyId = filters.apiKeyId
  mobileFilterDraft.status = filters.status
  mobileFilterDraft.downloaded = filters.downloaded
  showMobileFilterModal.value = true
}

function closeMobileFilterModal() {
  showMobileFilterModal.value = false
}

function resetMobileFilterDraft() {
  mobileFilterDraft.apiKeyId = ''
  mobileFilterDraft.status = ''
  mobileFilterDraft.downloaded = ''
}

function applyMobileFilterDraft() {
  filters.apiKeyId = mobileFilterDraft.apiKeyId
  filters.status = mobileFilterDraft.status
  filters.downloaded = mobileFilterDraft.downloaded
  closeMobileFilterModal()
  applyFilters()
}

function listOptions(): BatchImageJobsListOptions {
  const options: BatchImageJobsListOptions = {
    limit: pagination.page_size,
    cursor: String((pagination.page - 1) * pagination.page_size),
  }
  if (filters.taskName.trim()) options.taskName = filters.taskName.trim()
  if (filters.status) options.status = filters.status
  if (filters.downloaded) options.downloaded = filters.downloaded
  return options
}

function toJobRow(job: BatchImageJob, key = selectedApiKey.value): BatchImageJobRow {
  return {
    id: job.id,
    task_name: job.task_name || defaultTaskName(job.created_at),
    parent_batch_id: job.parent_batch_id || null,
    status: job.status,
    model: job.model,
    provider: job.provider,
    item_count: job.item_count,
    success_count: job.success_count,
    fail_count: job.fail_count,
    estimated_cost: job.estimated_cost,
    hold_amount: job.hold_amount,
    actual_cost: job.actual_cost,
    created_at: job.created_at,
    downloaded_at: job.downloaded_at,
    api_key_id: key?.id || 0,
    api_key_name: key?.name || '',
    child_count: 0,
  }
}

function applyChildCounts(rows: BatchImageJobRow[]) {
  const counts = new Map<string, number>()
  for (const row of rows) {
    if (!row.parent_batch_id) continue
    counts.set(row.parent_batch_id, (counts.get(row.parent_batch_id) || 0) + 1)
  }
  return rows.map(row => ({ ...row, child_count: counts.get(row.id) || 0 }))
}

function displayJob<T extends Pick<BatchImageJob, 'id' | 'parent_batch_id' | 'status' | 'item_count' | 'success_count' | 'fail_count' | 'estimated_cost' | 'hold_amount' | 'actual_cost'>>(job: T): T {
  if (job.parent_batch_id) return job
  const children = childrenByParent.value.get(job.id) || []
  if (!children.length) return job

  const childSuccess = children.reduce((sum, child) => sum + child.success_count, 0)
  const childEstimated = children.reduce((sum, child) => sum + child.estimated_cost, 0)
  const childHold = children.reduce((sum, child) => sum + child.hold_amount, 0)
  const childActual = children.reduce((sum, child) => sum + (child.actual_cost || 0), 0)
  const childActualReady = children.every(child => child.actual_cost !== null)
  const successCount = Math.min(job.item_count, job.success_count + childSuccess)
  const failCount = Math.max(0, job.item_count - successCount)
  const actualCost = job.actual_cost === null
    ? (childActualReady ? childActual : null)
    : job.actual_cost + childActual

  return {
    ...job,
    success_count: successCount,
    fail_count: failCount,
    status: failCount === 0 && TERMINAL_STATUSES.has(job.status) ? 'completed' : job.status,
    estimated_cost: job.estimated_cost + childEstimated,
    hold_amount: job.hold_amount + childHold,
    actual_cost: actualCost,
  }
}

function hasChildJobs(batchId: string) {
  return (childrenByParent.value.get(batchId) || []).length > 0
}

function toggleChildRows(batchId: string) {
  const next = new Set(expandedParentIds.value)
  if (next.has(batchId)) next.delete(batchId)
  else next.add(batchId)
  expandedParentIds.value = next
}

function menuDomSuffix(batchId: string): string {
  return String(batchId).replace(/[^a-zA-Z0-9_-]/g, '-')
}

function moreMenuId(batchId: string): string {
  return `batch-job-menu-${menuDomSuffix(batchId)}`
}

function moreMenuTriggerId(batchId: string): string {
  return `batch-job-menu-trigger-${menuDomSuffix(batchId)}`
}

function batchJobSelectionLabel(job: BatchImageJobRow): string {
  const name = job.task_name || defaultTaskName(job.created_at)
  return isZhLocale() ? `选择任务：${name}` : `Select job: ${name}`
}

function closeMoreMenu(restoreFocus = false) {
  const trigger = moreMenuTrigger
  openMoreJobId.value = ''
  moreMenuStyle.value = {}
  moreMenuTrigger = null
  if (restoreFocus && trigger?.isConnected) {
    void nextTick(() => trigger.focus())
  }
}

function dismissMoreMenu(): void {
  closeMoreMenu(false)
}

async function toggleMoreMenu(job: BatchImageJobRow, event: MouseEvent) {
  if (openMoreJobId.value === job.id) {
    closeMoreMenu(true)
    return
  }
  const trigger = event.currentTarget as HTMLElement | null
  const rect = trigger?.getBoundingClientRect()
  if (!rect) return
  const menuWidth = 176
  const margin = 8
  const left = Math.max(margin, Math.min(rect.right - menuWidth, window.innerWidth - menuWidth - margin))
  const top = Math.min(rect.bottom + margin, window.innerHeight - 96)
  moreMenuStyle.value = {
    left: `${left}px`,
    top: `${Math.max(margin, top)}px`,
  }
  moreMenuTrigger = trigger
  openMoreJobId.value = job.id
  await nextTick()
  moreMenuRef.value?.querySelector<HTMLElement>('[role="menuitem"]:not([disabled])')?.focus()
}

function handleMoreMenuKeydown(event: KeyboardEvent): void {
  const menu = moreMenuRef.value
  if (!menu) return
  const items = Array.from(menu.querySelectorAll<HTMLElement>('[role="menuitem"]:not([disabled])'))

  if (event.key === 'Escape') {
    event.preventDefault()
    event.stopPropagation()
    closeMoreMenu(true)
    return
  }

  if (event.key === 'Tab') {
    closeMoreMenu(false)
    return
  }

  if (!['ArrowDown', 'ArrowUp', 'Home', 'End'].includes(event.key) || items.length === 0) return
  event.preventDefault()
  const currentIndex = items.indexOf(document.activeElement as HTMLElement)
  if (event.key === 'Home') {
    items[0].focus()
  } else if (event.key === 'End') {
    items[items.length - 1].focus()
  } else if (event.key === 'ArrowDown') {
    items[(currentIndex + 1 + items.length) % items.length].focus()
  } else {
    items[(currentIndex - 1 + items.length) % items.length].focus()
  }
}

function handleMoreMenuDocumentKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape' && openMoreJobId.value) {
    event.preventDefault()
    closeMoreMenu(true)
  }
}

function cancelPromptPopoverClose() {
  if (!promptPopoverCloseTimer) return
  clearTimeout(promptPopoverCloseTimer)
  promptPopoverCloseTimer = null
}

function cancelPromptPopoverOpen() {
  if (!promptPopoverOpenTimer) return
  clearTimeout(promptPopoverOpenTimer)
  promptPopoverOpenTimer = null
}

function closePromptPopover() {
  cancelPromptPopoverOpen()
  cancelPromptPopoverClose()
  promptPopover.visible = false
  promptPopover.text = ''
  promptPopover.style = {}
  activePromptPopoverTarget = null
}

function schedulePromptPopoverClose() {
  cancelPromptPopoverOpen()
  cancelPromptPopoverClose()
  promptPopoverCloseTimer = setTimeout(() => {
    closePromptPopover()
  }, 180)
}

function schedulePromptPopoverOpen(event: MouseEvent | PointerEvent, text: string) {
  const target = event.currentTarget as HTMLElement | null
  if (!target) return
  const value = String(text || '').trim()
  if (!value || value === '-') return
  activePromptPopoverTarget = target
  cancelPromptPopoverOpen()
  cancelPromptPopoverClose()
  promptPopoverOpenTimer = setTimeout(() => {
    if (activePromptPopoverTarget !== target || !document.body.contains(target)) return
    openPromptPopover(target, value)
  }, 520)
}

function showPromptPopover(event: MouseEvent | FocusEvent, text: string) {
  const value = String(text || '').trim()
  if (!value || value === '-') return
  const target = event.currentTarget as HTMLElement | null
  cancelPromptPopoverClose()
  cancelPromptPopoverOpen()
  if (!target) return
  activePromptPopoverTarget = target
  openPromptPopover(target, value)
}

function openPromptPopover(target: HTMLElement, value: string) {
  const rect = target.getBoundingClientRect()
  if (!rect) return
  const viewportWidth = window.innerWidth || 1280
  const viewportHeight = window.innerHeight || 720
  const horizontalMargin = 8
  const width = Math.min(440, Math.max(0, viewportWidth - horizontalMargin * 2))
  const left = Math.max(horizontalMargin, Math.min(rect.left, viewportWidth - width - horizontalMargin))
  const estimatedHeight = 178
  const preferredTop = rect.bottom + 8
  const top = preferredTop + estimatedHeight > viewportHeight
    ? Math.max(16, rect.top - estimatedHeight - 8)
    : preferredTop
  promptPopover.text = value
  promptPopover.style = {
    left: `${left}px`,
    top: `${top}px`,
    width: `${width}px`,
    maxWidth: 'calc(100vw - 16px)',
  }
  promptPopover.visible = true
}

function copyPromptPopover() {
  if (!promptPopover.text) return
  void copyToClipboard(promptPopover.text, 'Prompt 已复制')
}

async function loadBatchJobs() {
  const keys = filteredApiKeys.value
  if (!keys.length) {
    batchJobs.value = []
    pagination.has_more = false
    return
  }
  loadingJobs.value = true
  closeMoreMenu()
  try {
    const options = listOptions()
    const results = await Promise.all(keys.map(async (key) => {
      const result = await listBatchImageJobs(key.key, options)
      return {
        hasMore: Boolean(result.has_more),
        rows: (result.data || []).map(job => toJobRow(job, key)),
      }
    }))
    batchJobs.value = applyChildCounts(results
      .flatMap(result => result.rows)
      .sort((a, b) => b.created_at - a.created_at)
      .slice(0, pagination.page_size))
    pagination.has_more = results.some(result => result.hasMore)
    selectedJobIds.value = new Set([...selectedJobIds.value].filter(id => visibleBatchJobs.value.some(job => job.id === id)))
  } catch (error: any) {
    appStore.showError(batchImageErrorMessage(error, batchImageText('loadJobsFailed')))
  } finally {
    loadingJobs.value = false
  }
}

function upsertJob(job: BatchImageJob) {
  const next = toJobRow(job)
  const index = batchJobs.value.findIndex(item => item.id === job.id)
  if (index >= 0) {
    const rows = [...batchJobs.value]
    rows[index] = { ...next, is_child: rows[index].is_child }
    batchJobs.value = applyChildCounts(rows)
    return
  }
  batchJobs.value = applyChildCounts([next, ...batchJobs.value].slice(0, pagination.page_size))
}

function handlePageChange(page: number) {
  if (page < 1 || page === pagination.page) return
  pagination.page = page
  selectedJobIds.value = new Set()
  void loadBatchJobs()
}

function handlePageSizeChange(value: string | number | boolean | null) {
  if (value === null || typeof value === 'boolean') return
  const nextSize = Math.min(Math.max(Number(value) || 20, 1), 100)
  pagination.page_size = nextSize
  pagination.page = 1
  setPersistedPageSize(nextSize)
  selectedJobIds.value = new Set()
  void loadBatchJobs()
}

function openCreateModal() {
  showCreateModal.value = true
  if (!apiKeys.value.length) {
    void loadApiKeys()
  }
}

function closeCreateModal() {
  if (submitting.value) return
  showCreateModal.value = false
  resetCreateDraft()
}

function resetCreateDraft() {
  form.taskName = ''
  form.responseMimeType = 'image/png'
  promptRows.value = []
  promptDraft.value = ''
  customIdDraft.value = ''
  outputCountDraft.value = 1
  referenceImageDrafts.value = []
}

function closeDetail() {
  closePromptPopover()
  currentJob.value = null
  selectedBatchId.value = ''
  selectedBatchApiKeyId.value = 0
  items.value = []
  clearItemPreviews()
}

function keyForSelectedBatch(): ApiKey | null {
  if (selectedBatchApiKeyId.value) {
    const key = geminiApiKeys.value.find(item => item.id === selectedBatchApiKeyId.value)
    if (key) return key
  }
  return selectedApiKey.value
}

function requireApiKey(): ApiKey | null {
  if (!selectedApiKey.value) {
    appStore.showError(batchImageText('selectApiKey'))
    return null
  }
  return selectedApiKey.value
}

function validateForm(): boolean {
  if (!requireApiKey()) return false
  if (!form.model) {
    appStore.showError(availableBatchImageModels.value.length === 0 ? batchImageText('noModelsForKey') : batchImageText('selectModel'))
    return false
  }
  if (parsedItems.value.length === 0) {
    appStore.showError(batchImageText('promptRequired'))
    return false
  }
  if (estimatedOutputCount.value > BATCH_IMAGE_MAX_OUTPUTS_PER_JOB) {
    appStore.showError(batchImageText('tooManyOutputImages'))
    return false
  }
  const refLimit = selectedModelReferenceLimit.value
  if (promptRows.value.some(row => row.reference_images.length > refLimit)) {
    appStore.showError(batchImageText('tooManyReferenceImages'))
    return false
  }
  return true
}

async function submitJob() {
  if (submitting.value) return
  if (promptDraft.value.trim()) addPromptRow()
  if (!validateForm()) return
  const key = requireApiKey()
  if (!key) return
	  submitting.value = true
	  try {
	    const job = await submitBatchImageJob(
	      key.key,
	      {
	        model: form.model,
        task_name: form.taskName.trim() || defaultTaskName(),
        image_size: '1K',
        response_mime_type: form.responseMimeType,
        items: parsedItems.value,
	      },
	      `sub2api-ui-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`,
	    )
	    currentJob.value = job
	    selectedBatchId.value = job.id
	    selectedBatchApiKeyId.value = key.id
	    items.value = []
	    upsertJob(job)
	    showCreateModal.value = false
	    resetCreateDraft()
	    appStore.showSuccess(batchImageText('submitted'))
	    void loadItems()
	    startPolling()
  } catch (error: any) {
    appStore.showError(batchImageErrorMessage(error, batchImageText('submitFailed')))
  } finally {
    submitting.value = false
  }
}

async function refreshSelected() {
  if (!selectedBatchId.value) return
  const key = keyForSelectedBatch() || requireApiKey()
  if (!key) return
  refreshing.value = true
  try {
    const job = await getBatchImageJob(key.key, selectedBatchId.value)
    currentJob.value = job
    upsertJob(job)
    if (TERMINAL_STATUSES.has(job.status)) stopPolling()
  } catch (error: any) {
    appStore.showError(batchImageErrorMessage(error, batchImageText('refreshFailed')))
  } finally {
    refreshing.value = false
  }
}

async function refreshDetail() {
  await Promise.all([
    refreshSelected(),
    loadItems(),
  ])
}

function selectJob(batchId: string) {
  const row = batchJobs.value.find(job => job.id === batchId)
  if (row?.api_key_id && geminiApiKeys.value.some(key => key.id === row.api_key_id)) {
    form.apiKeyId = row.api_key_id
    selectedBatchApiKeyId.value = row.api_key_id
  } else {
    selectedBatchApiKeyId.value = 0
  }
  selectedBatchId.value = batchId
  currentJob.value = null
  items.value = []
  void refreshSelected()
  void loadItems()
}

function startPolling() {
  stopPolling()
  pollTimer = setInterval(() => {
    if (!currentJob.value || TERMINAL_STATUSES.has(currentJob.value.status)) {
      stopPolling()
      return
    }
    void refreshSelected()
  }, 8000)
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

function canCancel(job: Pick<BatchImageJob, 'status'>) {
  return !TERMINAL_STATUSES.has(job.status)
}

function canDownload(job: Pick<BatchImageJob, 'status' | 'success_count'>) {
  return job.status === 'completed' && job.success_count > 0
}

function canRetry(job: Pick<BatchImageJob, 'status' | 'fail_count'>) {
  const display = 'id' in job ? displayJob(job as BatchImageJob) : job
  return TERMINAL_STATUSES.has(display.status) && display.fail_count > 0
}

function isDownloadingJob(batchId: string) {
  return downloading.value && downloadingBatchId.value === batchId
}

function applyJobApiKey(job: BatchImageJobRow | Pick<BatchImageJob, 'id'>) {
  if ('api_key_id' in job && job.api_key_id && geminiApiKeys.value.some(key => key.id === job.api_key_id)) {
    form.apiKeyId = job.api_key_id
  }
}

function apiKeyForJob(job: BatchImageJobRow | Pick<BatchImageJob, 'id'>): ApiKey | null {
  if ('api_key_id' in job && job.api_key_id) {
    return geminiApiKeys.value.find(key => key.id === job.api_key_id) || null
  }
  return selectedApiKey.value
}

function toggleJobSelection(batchId: string, checked: boolean) {
  const next = new Set(selectedJobIds.value)
  if (checked) next.add(batchId)
  else next.delete(batchId)
  selectedJobIds.value = next
}

function toggleAllVisible(checked: boolean) {
  const next = new Set(selectedJobIds.value)
  for (const job of visibleBatchJobs.value) {
    if (checked) next.add(job.id)
    else next.delete(job.id)
  }
  selectedJobIds.value = next
}

function canDeleteRecord(job: Pick<BatchImageJob, 'status'>) {
  return TERMINAL_STATUSES.has(job.status)
}

function cancelSelected(): void {
  if (!currentJob.value) return
  if (!(keyForSelectedBatch() || requireApiKey())) return
  pendingConfirmation.value = { type: 'cancel', jobId: currentJob.value.id }
}

async function performCancelSelected(jobId: string) {
  if (!currentJob.value || currentJob.value.id !== jobId) return
  const key = keyForSelectedBatch() || requireApiKey()
  if (!key) return
  cancelling.value = true
  try {
    const job = await cancelBatchImageJob(key.key, currentJob.value.id)
    currentJob.value = job
    upsertJob(job)
    appStore.showSuccess(batchImageText('cancelled'))
  } catch (error: any) {
    appStore.showError(batchImageErrorMessage(error, batchImageText('cancelFailed')))
  } finally {
    cancelling.value = false
  }
}

async function downloadSelected() {
  if (!currentJob.value) return
  await downloadJob(currentJob.value)
}

async function retrySelected() {
  if (!currentJob.value) return
  await retryFailedJob(currentJob.value)
}

async function retryFailedJob(job: BatchImageJobRow | BatchImageJob) {
  if (!canRetry(job) || retryingBatchId.value) return
  closeMoreMenu()
  const key = apiKeyForJob(job) || keyForSelectedBatch() || requireApiKey()
  if (!key) return
  retryingBatchId.value = job.id
  try {
    const sourceItems = await ensureItemsForRetry(key.key, job.id)
    const failedItems = sourceItems
      .filter(item => item.status === 'failed')
      .map(item => ({ custom_id: retryCustomID(item.custom_id), prompt: String(item.prompt_preview || '').trim() }))
      .filter(item => item.prompt)
    if (failedItems.length === 0) {
      appStore.showError(batchImageText('retryMissingPrompts'))
      return
    }
    const retryJob = await submitBatchImageJob(
      key.key,
      {
        model: job.model,
        task_name: `${job.task_name || defaultTaskName()} 重试失败项`,
        parent_batch_id: rootBatchIdForRetry(job),
        provider: job.provider,
        image_size: '1K',
        response_mime_type: form.responseMimeType,
        items: failedItems,
      },
      `sub2api-ui-retry-${job.id}-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`,
    )
    currentJob.value = retryJob
    selectedBatchId.value = retryJob.id
    selectedBatchApiKeyId.value = key.id
    items.value = []
    upsertJob(retryJob)
    if (retryJob.parent_batch_id) {
      expandedParentIds.value = new Set([...expandedParentIds.value, retryJob.parent_batch_id])
    }
    appStore.showSuccess(batchImageText('retrySubmitted'))
    void loadItems()
    startPolling()
  } catch (error: any) {
    appStore.showError(batchImageErrorMessage(error, batchImageText('retryFailed')))
  } finally {
    retryingBatchId.value = ''
  }
}

async function ensureItemsForRetry(apiKey: string, batchId: string) {
  if (selectedBatchId.value === batchId && items.value.length > 0) {
    return items.value
  }
  const result = await listBatchImageItems(apiKey, batchId)
  return result.data || []
}

function retryCustomID(customID: string) {
  const base = String(customID || 'item').replace(/[^\w.-]+/g, '_').replace(/^_+|_+$/g, '') || 'item'
  return `${base}_retry_${Date.now().toString(36)}`
}

function rootBatchIdForRetry(job: BatchImageJobRow | BatchImageJob) {
  return job.parent_batch_id || job.id
}

async function downloadJob(job: (BatchImageJobRow | Pick<BatchImageJob, 'id'>)) {
  if (downloading.value) return
  closeMoreMenu()
  applyJobApiKey(job)
  const key = apiKeyForJob(job) || requireApiKey()
  if (!key) return
  downloading.value = true
  downloadingBatchId.value = job.id
  try {
    const blob = await downloadBatchImageZip(key.key, job.id)
    saveBlob(blob, `${job.id}.zip`)
    markJobDownloaded(job.id)
  } catch (error: any) {
    appStore.showError(batchImageErrorMessage(error, batchImageText('downloadFailed')))
  } finally {
    downloading.value = false
    downloadingBatchId.value = ''
  }
}

async function downloadSelectedJobs() {
  if (bulkDownloading.value || selectedDownloadableRows.value.length === 0) return
  bulkDownloading.value = true
  try {
    for (const row of selectedDownloadableRows.value) {
      const key = apiKeyForJob(row)
      if (!key) continue
      downloading.value = true
      downloadingBatchId.value = row.id
      const blob = await downloadBatchImageZip(key.key, row.id)
      saveBlob(blob, `${row.id}.zip`)
      markJobDownloaded(row.id)
    }
    appStore.showSuccess(batchImageText('batchDownloadStarted'))
  } catch (error: any) {
    appStore.showError(batchImageErrorMessage(error, batchImageText('downloadFailed')))
  } finally {
    bulkDownloading.value = false
    downloading.value = false
    downloadingBatchId.value = ''
  }
}

function deleteJob(job: BatchImageJobRow): void {
  if (!canDeleteRecord(job) || deletingBatchId.value) return
  closeMoreMenu()
  const key = apiKeyForJob(job)
  if (!key) return
  pendingConfirmation.value = { type: 'delete', job }
}

async function performDeleteJob(job: BatchImageJobRow) {
  const key = apiKeyForJob(job)
  if (!key || !canDeleteRecord(job) || deletingBatchId.value) return
  deletingBatchId.value = job.id
  try {
    await deleteBatchImageJobRecord(key.key, job.id)
    removeJobFromList(job.id)
    appStore.showSuccess(batchImageText('deleted'))
  } catch (error: any) {
    appStore.showError(batchImageErrorMessage(error, batchImageText('deleteFailed')))
  } finally {
    deletingBatchId.value = ''
  }
}

function deleteSelectedJobs(): void {
  const rows = selectedRows.value.filter(job => canDeleteRecord(job))
  if (bulkDeleting.value || rows.length === 0) return
  pendingConfirmation.value = { type: 'delete-selected', jobIds: rows.map(row => row.id) }
}

async function performDeleteSelectedJobs(jobIds: string[]) {
  const selectedIds = new Set(jobIds)
  const rows = batchJobs.value.filter(job => selectedIds.has(job.id) && canDeleteRecord(job))
  if (bulkDeleting.value || rows.length === 0) return
  bulkDeleting.value = true
  try {
    for (const row of rows) {
      const key = apiKeyForJob(row)
      if (!key) continue
      deletingBatchId.value = row.id
      await deleteBatchImageJobRecord(key.key, row.id)
      removeJobFromList(row.id)
    }
    appStore.showSuccess(batchImageText('deleted'))
  } catch (error: any) {
    appStore.showError(batchImageErrorMessage(error, batchImageText('deleteFailed')))
  } finally {
    bulkDeleting.value = false
    deletingBatchId.value = ''
  }
}

function closeConfirmationDialog(): void {
  pendingConfirmation.value = null
}

async function confirmPendingAction(): Promise<void> {
  const confirmation = pendingConfirmation.value
  pendingConfirmation.value = null
  if (!confirmation) return

  if (confirmation.type === 'cancel') {
    await performCancelSelected(confirmation.jobId)
  } else if (confirmation.type === 'delete') {
    await performDeleteJob(confirmation.job)
  } else {
    await performDeleteSelectedJobs(confirmation.jobIds)
  }
}

function markJobDownloaded(batchId: string) {
  const downloadedAt = Math.floor(Date.now() / 1000)
  batchJobs.value = batchJobs.value.map(job => job.id === batchId ? { ...job, downloaded_at: job.downloaded_at || downloadedAt } : job)
  if (currentJob.value?.id === batchId && !currentJob.value.downloaded_at) {
    currentJob.value = { ...currentJob.value, downloaded_at: downloadedAt }
  }
}

function removeJobFromList(batchId: string) {
  batchJobs.value = batchJobs.value.filter(job => job.id !== batchId)
  toggleJobSelection(batchId, false)
  if (currentJob.value?.id === batchId) closeDetail()
}

function canLoadItemPreview(item: BatchImageItem) {
  return (item.status === 'succeeded' || item.status === 'success') && item.image_count > 0
}

function isSuccessfulImageItem(item: Pick<BatchImageItem, 'status' | 'image_count'>) {
  return (item.status === 'succeeded' || item.status === 'success') && item.image_count > 0
}

function detailRootBatchId() {
  return currentJob.value?.parent_batch_id || selectedBatchId.value || currentJob.value?.id || ''
}

function isChildDetailItem(item: Pick<BatchImageDetailItem, 'batch_id'>) {
  const rootBatchId = detailRootBatchId()
  return Boolean(rootBatchId && item.batch_id && item.batch_id !== rootBatchId)
}

function retrySourceCustomID(customID: string) {
  return String(customID || '').replace(/(?:_retry_[a-z0-9]+)+$/i, '')
}

function isRecoveredOriginalFailure(item: BatchImageDetailItem) {
  const rootBatchId = detailRootBatchId()
  return Boolean(
    rootBatchId
    && item.batch_id === rootBatchId
    && item.status === 'failed'
    && recoveredOriginalCustomIds.value.has(item.custom_id),
  )
}

function detailItemRowClass(item: BatchImageDetailItem) {
  if (isRecoveredOriginalFailure(item)) {
    return 'bg-surface-subtle text-foreground-subtle hover:bg-surface'
  }
  return 'hover:bg-surface-subtle'
}

function previewCacheSupported() {
  return typeof window !== 'undefined' && 'indexedDB' in window
}

function previewCacheKey(batchId: string, customID: string, imageIndex = 0) {
  return [batchId, customID, imageIndex].map(part => encodeURIComponent(String(part))).join(':')
}

function itemPreviewKey(item: Pick<BatchImageItem, 'batch_id' | 'custom_id'>) {
  return previewCacheKey(item.batch_id || selectedBatchId.value || currentJob.value?.id || '', item.custom_id, 0)
}

function idbRequest<T>(request: IDBRequest<T>): Promise<T> {
  return new Promise((resolve, reject) => {
    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error)
  })
}

function openPreviewCacheDB(): Promise<IDBDatabase | null> {
  if (!previewCacheSupported()) return Promise.resolve(null)
  if (previewCacheDBPromise) return previewCacheDBPromise

  previewCacheDBPromise = new Promise((resolve) => {
    const request = window.indexedDB.open(PREVIEW_CACHE_DB_NAME, 1)
    request.onupgradeneeded = () => {
      const db = request.result
      if (!db.objectStoreNames.contains(PREVIEW_CACHE_STORE_NAME)) {
        const store = db.createObjectStore(PREVIEW_CACHE_STORE_NAME, { keyPath: 'key' })
        store.createIndex('lastAccessedAt', 'lastAccessedAt', { unique: false })
      }
    }
    request.onsuccess = () => resolve(request.result)
    request.onerror = () => resolve(null)
    request.onblocked = () => resolve(null)
  })
  return previewCacheDBPromise
}

async function getCachedPreviewBlob(cacheKey: string): Promise<Blob | null> {
  const db = await openPreviewCacheDB()
  if (!db) return null
  const record = await idbRequest<PreviewCacheRecord | undefined>(
    db.transaction(PREVIEW_CACHE_STORE_NAME, 'readonly').objectStore(PREVIEW_CACHE_STORE_NAME).get(cacheKey),
  ).catch(() => undefined)
  if (!record?.blob) return null

  const now = Date.now()
  if (now - record.createdAt > PREVIEW_CACHE_MAX_AGE_MS) {
    void deleteCachedPreview(cacheKey)
    return null
  }
  void touchCachedPreview(cacheKey, now)
  return record.blob
}

async function hydrateCachedItemPreviews(detailItems: BatchImageDetailItem[]) {
  const previewableItems = detailItems.filter(item => canLoadItemPreview(item))
  if (!previewableItems.length || !previewCacheSupported()) return

  await Promise.all(previewableItems.map(async (item) => {
    const batchId = item.batch_id || selectedBatchId.value || currentJob.value?.id || ''
    const previewKey = itemPreviewKey(item)
    if (!batchId || itemPreviewUrls[previewKey] || previewErrorIds.value.has(previewKey)) return
    const cached = await getCachedPreviewBlob(previewCacheKey(batchId, item.custom_id, 0)).catch(() => null)
    if (!cached || itemPreviewUrls[previewKey]) return
    itemPreviewUrls[previewKey] = URL.createObjectURL(cached)
  }))
}

async function putCachedPreviewBlob(cacheKey: string, blob: Blob) {
  const db = await openPreviewCacheDB()
  if (!db) return
  const now = Date.now()
  const record: PreviewCacheRecord = {
    key: cacheKey,
    blob,
    size: blob.size,
    createdAt: now,
    lastAccessedAt: now,
  }
  await idbRequest(db.transaction(PREVIEW_CACHE_STORE_NAME, 'readwrite').objectStore(PREVIEW_CACHE_STORE_NAME).put(record)).catch(() => null)
  void cleanupPreviewCache()
}

async function touchCachedPreview(cacheKey: string, lastAccessedAt: number) {
  const db = await openPreviewCacheDB()
  if (!db) return
  const record = await idbRequest<PreviewCacheRecord | undefined>(
    db.transaction(PREVIEW_CACHE_STORE_NAME, 'readonly').objectStore(PREVIEW_CACHE_STORE_NAME).get(cacheKey),
  ).catch(() => undefined)
  if (!record) return
  record.lastAccessedAt = lastAccessedAt
  await idbRequest(db.transaction(PREVIEW_CACHE_STORE_NAME, 'readwrite').objectStore(PREVIEW_CACHE_STORE_NAME).put(record)).catch(() => null)
}

async function deleteCachedPreview(cacheKey: string) {
  const db = await openPreviewCacheDB()
  if (!db) return
  await idbRequest(db.transaction(PREVIEW_CACHE_STORE_NAME, 'readwrite').objectStore(PREVIEW_CACHE_STORE_NAME).delete(cacheKey)).catch(() => null)
}

async function cleanupPreviewCache() {
  const db = await openPreviewCacheDB()
  if (!db) return
  const records = await idbRequest<PreviewCacheRecord[]>(
    db.transaction(PREVIEW_CACHE_STORE_NAME, 'readonly').objectStore(PREVIEW_CACHE_STORE_NAME).getAll(),
  ).catch(() => [])
  if (!records.length) return

  const now = Date.now()
  const sorted = [...records].sort((a, b) => a.lastAccessedAt - b.lastAccessedAt)
  const deleteKeys = new Set<string>()
  let totalBytes = 0
  let keptCount = 0

  for (const record of sorted) {
    if (now - record.createdAt > PREVIEW_CACHE_MAX_AGE_MS) {
      deleteKeys.add(record.key)
      continue
    }
    totalBytes += record.size || record.blob?.size || 0
    keptCount += 1
  }

  for (const record of sorted) {
    if (deleteKeys.has(record.key)) continue
    if (keptCount <= PREVIEW_CACHE_MAX_ENTRIES && totalBytes <= PREVIEW_CACHE_MAX_BYTES) break
    deleteKeys.add(record.key)
    totalBytes -= record.size || record.blob?.size || 0
    keptCount -= 1
  }

  if (!deleteKeys.size) return
  const store = db.transaction(PREVIEW_CACHE_STORE_NAME, 'readwrite').objectStore(PREVIEW_CACHE_STORE_NAME)
  for (const key of deleteKeys) {
    store.delete(key)
  }
}

async function createThumbnailBlob(blob: Blob): Promise<Blob> {
  const source = await loadPreviewImageSource(blob)
  const width = source.width
  const height = source.height
  const scale = Math.min(1, PREVIEW_THUMBNAIL_MAX_EDGE / Math.max(width, height))
  const targetWidth = Math.max(1, Math.round(width * scale))
  const targetHeight = Math.max(1, Math.round(height * scale))
  const canvas = document.createElement('canvas')
  canvas.width = targetWidth
  canvas.height = targetHeight
  const ctx = canvas.getContext('2d')
  if (!ctx) throw new Error('canvas unavailable')
  ctx.drawImage(source.image, 0, 0, targetWidth, targetHeight)
  source.close()
  return await new Promise<Blob>((resolve, reject) => {
    canvas.toBlob((thumbnail) => {
      if (thumbnail) resolve(thumbnail)
      else reject(new Error('thumbnail unavailable'))
    }, 'image/webp', PREVIEW_THUMBNAIL_QUALITY)
  })
}

async function loadPreviewImageSource(blob: Blob): Promise<{ image: PreviewImageSource, width: number, height: number, close: () => void }> {
  if ('createImageBitmap' in window) {
    const bitmap = await window.createImageBitmap(blob)
    return {
      image: bitmap,
      width: bitmap.width,
      height: bitmap.height,
      close: () => bitmap.close(),
    }
  }

  const url = URL.createObjectURL(blob)
  try {
    const image = await new Promise<HTMLImageElement>((resolve, reject) => {
      const img = new Image()
      img.onload = () => resolve(img)
      img.onerror = () => reject(new Error('image unavailable'))
      img.src = url
    })
    return {
      image,
      width: image.naturalWidth || image.width,
      height: image.naturalHeight || image.height,
      close: () => URL.revokeObjectURL(url),
    }
  } catch (error) {
    URL.revokeObjectURL(url)
    throw error
  }
}

async function loadItems() {
  const batchId = selectedBatchId.value || currentJob.value?.id || ''
  if (!batchId) return
  const key = keyForSelectedBatch() || requireApiKey()
  if (!key) return
  loadingItems.value = true
  try {
    clearItemPreviews()
    const jobs = detailJobsForBatch(batchId)
    const results = await Promise.all(jobs.map(async (job) => {
      const result = await listBatchImageItems(key.key, job.id)
      return (result.data || []).map(item => ({
        ...item,
        batch_id: job.id,
        source_task_name: detailSourceName(job, batchId),
      }))
    }))
    const detailItems = results.flat()
    items.value = detailItems
    void hydrateCachedItemPreviews(detailItems)
  } catch (error: any) {
    appStore.showError(batchImageErrorMessage(error, batchImageText('loadItemsFailed')))
  } finally {
    loadingItems.value = false
  }
}

function detailJobsForBatch(batchId: string): BatchImageJobRow[] {
  const row = batchJobs.value.find(job => job.id === batchId)
  const base = row || (currentJob.value && currentJob.value.id === batchId ? toJobRow(currentJob.value, keyForSelectedBatch() || selectedApiKey.value) : null)
  if (!base) return []
  if (base.parent_batch_id) return [base]
  return [base, ...(childrenByParent.value.get(base.id) || [])]
}

function detailSourceName(job: Pick<BatchImageJobRow, 'id' | 'task_name' | 'parent_batch_id'>, rootBatchId: string) {
  const name = job.task_name || job.id
  if (job.id === rootBatchId) return `主任务：${name}`
  return `子任务：${name}`
}

async function loadItemPreview(item: BatchImageItem) {
  const batchId = item.batch_id || selectedBatchId.value || currentJob.value?.id || ''
  const previewKey = itemPreviewKey(item)
  if (!batchId || !canLoadItemPreview(item) || (itemPreviewUrls[previewKey] && !previewErrorIds.value.has(previewKey))) return
  const key = keyForSelectedBatch() || requireApiKey()
  if (!key) return
  const cacheKey = previewCacheKey(batchId, item.custom_id, 0)
  previewLoadingIds.value = new Set([...previewLoadingIds.value, previewKey])
  try {
    previewErrorIds.value = new Set([...previewErrorIds.value].filter(id => id !== previewKey))
    if (itemPreviewUrls[previewKey]) {
      URL.revokeObjectURL(itemPreviewUrls[previewKey])
      delete itemPreviewUrls[previewKey]
    }
    const cached = await getCachedPreviewBlob(cacheKey)
    if (cached) {
      itemPreviewUrls[previewKey] = URL.createObjectURL(cached)
      return
    }
    const blob = await getBatchImageItemContent(key.key, batchId, item.custom_id, 0)
    const thumbnail = await createThumbnailBlob(blob).catch(() => blob)
    itemPreviewUrls[previewKey] = URL.createObjectURL(thumbnail)
    if (thumbnail !== blob || thumbnail.size <= 1024 * 1024) {
      void putCachedPreviewBlob(cacheKey, thumbnail)
    }
  } catch (error: any) {
    previewErrorIds.value = new Set([...previewErrorIds.value, previewKey])
    appStore.showError(batchImageErrorMessage(error, batchImageText('loadPreviewFailed')))
  } finally {
    const next = new Set(previewLoadingIds.value)
    next.delete(previewKey)
    previewLoadingIds.value = next
  }
}

function openImagePreview(item: BatchImageItem) {
  const previewKey = itemPreviewKey(item)
  if (!itemPreviewUrls[previewKey] || previewErrorIds.value.has(previewKey)) return
  previewImageItem.value = item
}

function closeImagePreview() {
  previewImageItem.value = null
}

function handlePreviewError(customID: string) {
  if (itemPreviewUrls[customID]) {
    URL.revokeObjectURL(itemPreviewUrls[customID])
    delete itemPreviewUrls[customID]
  }
  previewErrorIds.value = new Set([...previewErrorIds.value, customID])
}

function clearItemPreviews() {
  closePromptPopover()
  for (const url of Object.values(itemPreviewUrls)) {
    if (url) URL.revokeObjectURL(url)
  }
  for (const key of Object.keys(itemPreviewUrls)) {
    delete itemPreviewUrls[key]
  }
  previewLoadingIds.value = new Set()
  previewErrorIds.value = new Set()
  previewImageItem.value = null
}

function copyInstruction() {
  void copyToClipboard(agentInstruction.value, batchImageText('copiedInstruction'))
}

function statusLabel(jobOrStatus: BatchImageStatus | Pick<BatchImageJob, 'status' | 'success_count' | 'fail_count'>) {
  const status = typeof jobOrStatus === 'string' ? jobOrStatus : jobOrStatus.status
  if (typeof jobOrStatus !== 'string' && status === 'completed' && jobOrStatus.fail_count > 0) {
    if (jobOrStatus.success_count > 0) return '部分成功'
    return '全部失败'
  }
  const labels: Record<string, string> = {
    queued: '排队中',
    running: '生成中',
    indexing: '整理结果',
    processing_results: '整理结果',
    settling: '结算中',
    completed: '已完成',
    failed: '失败',
    cancelled: '已取消',
    output_deleted: '结果已删除',
  }
  return labels[status] || status
}

function statusBadgeClass(jobOrStatus: BatchImageStatus | Pick<BatchImageJob, 'status' | 'success_count' | 'fail_count'>) {
  const status = typeof jobOrStatus === 'string' ? jobOrStatus : jobOrStatus.status
  if (typeof jobOrStatus !== 'string' && status === 'completed' && jobOrStatus.fail_count > 0) {
    if (jobOrStatus.success_count > 0) return 'badge-warning'
    return 'badge-danger'
  }
  if (status === 'completed') return 'badge-success'
  if (status === 'failed' || status === 'cancelled') return 'badge-danger'
  if (status === 'output_deleted') return 'badge-gray'
  return 'badge-primary'
}

function itemStatusLabel(status: string) {
  const labels: Record<string, string> = {
    pending: '排队中',
    succeeded: '成功',
    success: '成功',
    failed: '失败',
    cancelled: '已取消',
  }
  return labels[status] || status
}

function itemDisplayStatusLabel(item: BatchImageDetailItem) {
  if (isRecoveredOriginalFailure(item)) return '已补成功'
  return itemStatusLabel(item.status)
}

function itemStatusBadgeClass(status: string) {
  if (status === 'succeeded' || status === 'success') return 'badge-success'
  if (status === 'failed' || status === 'cancelled') return 'badge-danger'
  return 'badge-primary'
}

function itemDisplayStatusBadgeClass(item: BatchImageDetailItem) {
  if (isRecoveredOriginalFailure(item)) return 'badge-gray'
  return itemStatusBadgeClass(item.status)
}

function itemResultLabel(item: BatchImageDetailItem) {
  if (isRecoveredOriginalFailure(item)) return '旧失败已由重试子任务补成功'
  if (item.error) return friendlyItemError(item.error)
  if (item.status === 'succeeded' || item.status === 'success') {
    return itemPreviewUrls[itemPreviewKey(item)] ? '图片已生成，可点击预览' : '图片已生成，可下载'
  }
  if (item.status === 'failed') return '未生成可用图片'
  if (item.status === 'cancelled') return '任务已取消'
  return '等待生成结果'
}

function itemResultClass(item: BatchImageDetailItem) {
  if (isRecoveredOriginalFailure(item)) return 'bg-surface-subtle text-foreground-subtle ring-outline'
  if (item.error || item.status === 'failed' || item.status === 'cancelled') return 'bg-danger-subtle text-danger-foreground ring-danger/20'
  if (item.status === 'succeeded' || item.status === 'success') return 'ring-success/20 bg-success-subtle text-success-foreground'
  return 'bg-surface-subtle text-foreground-subtle ring-outline'
}

function friendlyItemError(error: BatchImageItem['error']) {
  if (!error) return '-'
  if (error.code === 'EMPTY_IMAGE_OUTPUT') return '上游返回了结果，但这条没有图片内容。通常是 Gemini/Vertex 单条生成失败或被安全策略拦截。'
  if (error.code === 'PROVIDER_ITEM_FAILED') return '上游返回的这条结果没有可用图片。'
  return error.message || error.code || '-'
}

function formatMoney(value: number | null | undefined) {
  if (value === null || value === undefined || Number.isNaN(Number(value))) return '$0.00'
  return `$${Number(value).toFixed(2)}`
}

function terminalZeroCost(job: Pick<BatchImageJob, 'status' | 'actual_cost'>) {
  return job.actual_cost === null && (job.status === 'failed' || job.status === 'cancelled')
}

function costLabel(job: Pick<BatchImageJob, 'status' | 'hold_amount' | 'actual_cost'>) {
  if (job.actual_cost !== null) return formatMoney(job.actual_cost)
  if (terminalZeroCost(job)) return formatMoney(0)
  return `冻结 ${formatMoney(job.hold_amount)}`
}

type BatchImageTextKey =
  | 'loadKeysFailed'
  | 'loadModelsFailed'
  | 'loadJobsFailed'
  | 'selectApiKey'
  | 'noModelsForKey'
  | 'selectModel'
  | 'promptRequired'
  | 'submitted'
  | 'submitFailed'
  | 'refreshFailed'
  | 'selectAllJobs'
  | 'cancelTitle'
  | 'cancelAction'
  | 'cancelConfirm'
  | 'cancelled'
  | 'cancelFailed'
  | 'batchDownloadStarted'
	  | 'downloadFailed'
	  | 'retrySubmitted'
	  | 'retryFailed'
	  | 'retryMissingPrompts'
  | 'deleteTitle'
  | 'deleteSelectedTitle'
  | 'deleteAction'
  | 'deleteConfirm'
  | 'deleteSelectedConfirm'
  | 'deleted'
  | 'deleteFailed'
	  | 'loadItemsFailed'
	  | 'loadPreviewFailed'
  | 'copiedInstruction'
  | 'loadingModels'
  | 'noModels'
  | 'noModelsHint'
  | 'noCompatibleAccount'
  | 'unsupportedProvider'
  | 'providerSubmitFailed'
  | 'vertexGcsBucketMissing'
  | 'queueFailed'
  | 'billingHoldFailed'
  | 'groupDisabled'
  | 'pricingMissing'
  | 'insufficientBalance'
  | 'invalidModel'
  | 'invalidItems'
  | 'duplicateCustomId'
  | 'promptTooLong'
  | 'invalidReferenceImage'
  | 'tooManyReferenceImages'
  | 'referenceImagesTooLarge'
  | 'tooManyOutputImages'
  | 'idempotencyConflict'
  | 'notReady'
  | 'outputDeleted'
  | 'resultMissing'
  | 'itemFailed'
  | 'itemImageIndexOutOfRange'
  | 'downloadLimited'
  | 'downloadTooLarge'
  | 'deleteNotReady'
  | 'disabled'
  | 'authRequired'
  | 'adminReference'
  | 'confirmCancel'
  | 'errorReference'

function isZhLocale() {
  return String(locale.value || '').toLowerCase().startsWith('zh')
}

function batchImageText(key: BatchImageTextKey) {
  const zh: Record<BatchImageTextKey, string> = {
    loadKeysFailed: '加载 API Key 失败',
    loadModelsFailed: '加载可用模型失败',
    loadJobsFailed: '加载批量任务失败',
    selectApiKey: '请选择可用的 Gemini API Key',
    noModelsForKey: '当前密钥没有可用的批量生图模型',
    selectModel: '请选择模型',
    promptRequired: '请至少填写一条 prompt',
    submitted: '批量任务已提交',
    submitFailed: '提交失败',
    refreshFailed: '刷新失败',
    selectAllJobs: '选择本页全部任务',
    cancelTitle: '取消批量任务',
    cancelAction: '确认取消',
    cancelConfirm: '取消会请求上游取消；已被系统索引为成功的图片仍会按成功项结算扣费，其余冻结金额会释放。确定取消吗？',
    cancelled: '已请求取消任务',
    cancelFailed: '取消失败',
    batchDownloadStarted: '已开始下载选中的任务',
	    downloadFailed: '下载失败',
	    retrySubmitted: '已提交失败项重试任务',
	    retryFailed: '重试失败项失败',
	    retryMissingPrompts: '这个任务没有保存失败项 prompt，无法自动重试。请复制原 prompt 后重新创建任务。',
    deleteTitle: '删除任务记录',
    deleteSelectedTitle: '删除选中的任务记录',
    deleteAction: '确认删除',
    deleteConfirm: '删除后这个任务会从你的列表隐藏，但账务记录仍会保留。确定删除吗？',
    deleteSelectedConfirm: '删除后选中的任务会从你的列表隐藏，但账务记录仍会保留。确定删除吗？',
    deleted: '任务记录已删除',
    deleteFailed: '删除任务记录失败',
	    loadItemsFailed: '加载明细失败',
	    loadPreviewFailed: '加载图片预览失败',
    copiedInstruction: '已复制批量生图说明',
    loadingModels: '加载可用模型中...',
    noModels: '无可用模型',
    noModelsHint: '当前密钥所属分组没有配置可用于批量生图的模型。',
    noCompatibleAccount: '当前密钥所属分组没有可用的批量生图上游账号。请联系管理员检查：该分组是否绑定了可调度的 Gemini API Key 或 Vertex 服务账号，以及账号是否支持所选模型。',
    unsupportedProvider: '这个任务使用的批量生图通道当前不可用。请联系管理员检查批量生图通道配置。',
    providerSubmitFailed: '上游批量生图任务提交失败。请联系管理员检查上游账号状态、模型权限或服务状态。',
    vertexGcsBucketMissing: 'Vertex 批量生图缺少托管 GCS 存储桶配置。请联系管理员配置 BATCH_IMAGE_VERTEX_MANAGED_GCS_BUCKET 后再提交。',
    queueFailed: '任务队列暂时不可用，批量任务没有成功入队。请联系管理员检查队列服务。',
    billingHoldFailed: '费用冻结失败，批量任务没有成功提交。请联系管理员检查余额冻结或计费服务。',
    groupDisabled: '当前密钥所属分组没有开启批量生图。你可以换一个已开启批量生图的密钥，或联系管理员开启。',
    pricingMissing: '所选模型还没有配置批量生图价格。请联系管理员补充价格配置。',
    insufficientBalance: '余额不足，无法冻结本次批量生图费用。',
    invalidModel: '请选择一个可用于当前密钥的批量生图模型。',
    invalidItems: 'Prompt 列表格式不正确，请检查是否为空、是否超过数量限制，或图片尺寸是否仍为 1K。',
    duplicateCustomId: 'Prompt 列表里的 custom_id 不能重复。',
    promptTooLong: '单条 prompt 过长，请缩短后重试。',
    invalidReferenceImage: '参考图格式不正确，请使用 10MB 以内的 PNG、JPEG 或 WebP。',
    tooManyReferenceImages: '参考图数量超过限制：Flash Image 每条最多 3 张，Pro Image 每条最多 14 张，整组最多 1000 张。',
    referenceImagesTooLarge: '参考图总量过大。inline 参考图整组最多 128MB；大量参考图请改用 gs:// file_uri 或拆分任务。',
    tooManyOutputImages: '预计生成张数超过限制：每条最多 4 张，整组最多 200 张。',
    idempotencyConflict: '这次提交和之前的请求标识冲突，请刷新页面后重新提交。',
    notReady: '任务还没有完成，完成后才能下载。',
    outputDeleted: '这个任务的结果文件已经被清理，无法下载。',
    resultMissing: '结果文件不可用，可能是上游结果文件已清理、存储权限异常，或管理员迁移过存储配置。请联系管理员检查结果文件。',
    itemFailed: '这条明细没有成功图片，无法预览。',
    itemImageIndexOutOfRange: '这条明细没有可预览的图片。',
    downloadLimited: '当前下载请求太多，请稍后再试。',
    downloadTooLarge: '这个 ZIP 太大，已超过单次下载限制。请减少单次下载数量，或联系管理员调整批量下载上限。',
    deleteNotReady: '任务结束后才能删除记录。正在生成或结算中的任务请先等待完成。',
    disabled: '批量生图功能当前未开启。',
    authRequired: '当前 API Key 不可用或已失效，请重新选择密钥。',
    adminReference: '请把错误码和请求 ID 发给管理员排查。',
    confirmCancel: '取消',
    errorReference: '错误信息',
  }
  const en: Record<BatchImageTextKey, string> = {
    loadKeysFailed: 'Failed to load API keys.',
    loadModelsFailed: 'Failed to load available models.',
    loadJobsFailed: 'Failed to load batch jobs.',
    selectApiKey: 'Select an available Gemini API key.',
    noModelsForKey: 'This key has no available batch image models.',
    selectModel: 'Select a model.',
    promptRequired: 'Enter at least one prompt.',
    submitted: 'Batch job submitted.',
    submitFailed: 'Failed to submit the batch job.',
    refreshFailed: 'Failed to refresh the job.',
    selectAllJobs: 'Select all jobs on this page',
    cancelTitle: 'Cancel batch job',
    cancelAction: 'Cancel job',
    cancelConfirm: 'Cancellation will be sent upstream. Images already indexed as successful will still be billed, and the remaining hold will be released. Continue?',
    cancelled: 'Cancellation requested.',
    cancelFailed: 'Failed to cancel the job.',
    batchDownloadStarted: 'Downloads for the selected jobs have started.',
	    downloadFailed: 'Failed to download the result.',
	    retrySubmitted: 'Retry job submitted for failed items.',
	    retryFailed: 'Failed to retry failed items.',
	    retryMissingPrompts: 'This job does not have saved prompts for failed items, so it cannot be retried automatically. Recreate it with the original prompt.',
    deleteTitle: 'Delete job record',
    deleteSelectedTitle: 'Delete selected job records',
    deleteAction: 'Delete',
    deleteConfirm: 'This hides the job from your list while keeping billing records. Delete it?',
    deleteSelectedConfirm: 'This hides the selected jobs from your list while keeping billing records. Delete them?',
    deleted: 'Job record deleted.',
    deleteFailed: 'Failed to delete the job record.',
	    loadItemsFailed: 'Failed to load item details.',
	    loadPreviewFailed: 'Failed to load the image preview.',
    copiedInstruction: 'Batch image instructions copied.',
    loadingModels: 'Loading available models...',
    noModels: 'No available models',
    noModelsHint: 'This key’s group has no models configured for batch image generation.',
    noCompatibleAccount: 'No usable upstream batch image account is available for this key’s group. Contact an administrator to check the group’s schedulable Gemini API key or Vertex service account and model support.',
    unsupportedProvider: 'The batch image provider for this job is not available. Contact an administrator to check the batch image provider configuration.',
    providerSubmitFailed: 'The upstream batch image job failed to submit. Contact an administrator to check the upstream account, model permission, or provider status.',
    vertexGcsBucketMissing: 'Vertex batch image generation is missing the managed GCS bucket configuration. Contact an administrator to configure BATCH_IMAGE_VERTEX_MANAGED_GCS_BUCKET before submitting again.',
    queueFailed: 'The task queue is temporarily unavailable, so the batch job was not queued. Contact an administrator to check the queue service.',
    billingHoldFailed: 'The cost hold failed, so the batch job was not submitted. Contact an administrator to check billing or balance hold service.',
    groupDisabled: 'Batch image generation is not enabled for this key’s group. Choose another enabled key or contact an administrator.',
    pricingMissing: 'The selected model does not have batch image pricing configured. Contact an administrator to add pricing first.',
    insufficientBalance: 'Insufficient balance to hold the estimated batch image cost.',
    invalidModel: 'Select a batch image model available for the current key.',
    invalidItems: 'The prompt list is invalid. Check that it is not empty, within the item limit, and still using 1K image size.',
    duplicateCustomId: 'Custom IDs in the prompt list must be unique.',
    promptTooLong: 'One prompt is too long. Shorten it and try again.',
    invalidReferenceImage: 'A reference image is invalid. Use PNG, JPEG, or WebP under 10 MB.',
    tooManyReferenceImages: 'Too many reference images. Flash Image allows up to 3 per item, Pro Image allows up to 14, and each job allows up to 1000 total.',
    referenceImagesTooLarge: 'Reference images are too large. Inline reference images are limited to 128 MB per job; use gs:// file_uri or split the job for large batches.',
    tooManyOutputImages: 'Too many expected output images. Each prompt can request up to 4 images, and each job can generate up to 200 images.',
    idempotencyConflict: 'This submission conflicts with a previous request ID. Refresh the page and submit again.',
    notReady: 'The job is not complete yet. Download will be available after completion.',
    outputDeleted: 'The result files for this job have already been cleaned up.',
    resultMissing: 'The result file is unavailable. It may have been cleaned up, storage permissions may be broken, or storage settings may have changed. Contact an administrator to check the result file.',
    itemFailed: 'This item has no successful image to preview.',
    itemImageIndexOutOfRange: 'This item has no previewable image.',
    downloadLimited: 'Too many download requests are active. Please try again later.',
    downloadTooLarge: 'This ZIP is too large for a single download. Download fewer items at once or ask an administrator to raise the batch download limit.',
    deleteNotReady: 'Job records can only be deleted after the job finishes.',
    disabled: 'Batch image generation is currently disabled.',
    authRequired: 'The current API key is unavailable or expired. Select the key again.',
    adminReference: 'Send the error code and request ID to an administrator for troubleshooting.',
    confirmCancel: 'Cancel',
    errorReference: 'Error detail',
  }
  return (isZhLocale() ? zh : en)[key]
}

function batchImageErrorReference(error: any) {
  const parts: string[] = []
  const code = String(error?.code || '').trim()
  const requestId = String(error?.requestId || '').trim()
  const status = String(error?.status || '').trim()
  if (code) parts.push(isZhLocale() ? `错误码：${code}` : `code: ${code}`)
  if (requestId) parts.push(isZhLocale() ? `请求 ID：${requestId}` : `request ID: ${requestId}`)
  if (!code && status) parts.push(isZhLocale() ? `HTTP 状态：${status}` : `HTTP status: ${status}`)
  return parts.length ? `（${parts.join('，')}）` : ''
}

function batchImageAdminError(base: string, error: any) {
  const reference = batchImageErrorReference(error)
  return `${base}${reference ? ` ${reference}` : ''} ${batchImageText('adminReference')}`
}

function batchImagePlainError(base: string) {
  return base
}

function batchImageErrorMessage(error: any, fallback: string) {
  const code = String(error?.code || '').trim()
  const message = String(error?.message || '').trim()
  if (code === 'API_KEY_REQUIRED' || code === '401') {
    return batchImagePlainError(batchImageText('authRequired'))
  }
  if (code === 'BATCH_IMAGE_NO_ACCOUNT_AVAILABLE' || /no compatible batch image account/i.test(message)) {
    return batchImageAdminError(batchImageText('noCompatibleAccount'), error)
  }
  if (code === 'BATCH_IMAGE_UNSUPPORTED_PROVIDER' || /unsupported batch image provider/i.test(message)) {
    return batchImageAdminError(batchImageText('unsupportedProvider'), error)
  }
  if (code === 'BATCH_IMAGE_VERTEX_GCS_BUCKET_MISSING' || code === 'VERTEX_MANAGED_GCS_BUCKET_MISSING') {
    return batchImageAdminError(batchImageText('vertexGcsBucketMissing'), error)
  }
  if (
    code === 'BATCH_IMAGE_PROVIDER_SUBMIT_FAILED' ||
    code === 'BATCH_IMAGE_PROVIDER_MISSING_API_KEY' ||
    code === 'BATCH_IMAGE_PROVIDER_MISSING_SERVICE_ACCOUNT' ||
    code === 'BATCH_IMAGE_PROVIDER_UNSUPPORTED_ACCOUNT'
  ) {
    return batchImageAdminError(batchImageText('providerSubmitFailed'), error)
  }
  if (code === 'BATCH_IMAGE_QUEUE_FAILED' || code === 'BATCH_IMAGE_QUEUE_NOT_CONFIGURED') {
    return batchImageAdminError(batchImageText('queueFailed'), error)
  }
  if (code === 'BATCH_IMAGE_BILLING_HOLD_FAILED') {
    return batchImageAdminError(batchImageText('billingHoldFailed'), error)
  }
  if (code === 'BATCH_IMAGE_GROUP_DISABLED') {
    return batchImagePlainError(batchImageText('groupDisabled'))
  }
  if (code === 'BATCH_IMAGE_SETTLEMENT_PRICING_MISSING') {
    return batchImageAdminError(batchImageText('pricingMissing'), error)
  }
  if (code === 'BATCH_IMAGE_INSUFFICIENT_BALANCE') {
    return batchImagePlainError(batchImageText('insufficientBalance'))
  }
  if (code === 'BATCH_IMAGE_INVALID_MODEL') {
    return batchImageText('invalidModel')
  }
  if (code === 'BATCH_IMAGE_INVALID_ITEMS') {
    return batchImageText('invalidItems')
  }
  if (code === 'BATCH_IMAGE_DUPLICATE_CUSTOM_ID') {
    return batchImageText('duplicateCustomId')
  }
  if (code === 'BATCH_IMAGE_PROMPT_TOO_LONG') {
    return batchImageText('promptTooLong')
  }
  if (code === 'BATCH_IMAGE_INVALID_REFERENCE_IMAGE') {
    return batchImageText('invalidReferenceImage')
  }
  if (code === 'BATCH_IMAGE_TOO_MANY_REFERENCE_IMAGES') {
    return batchImageText('tooManyReferenceImages')
  }
  if (code === 'BATCH_IMAGE_REFERENCE_IMAGES_TOO_LARGE') {
    return batchImageText('referenceImagesTooLarge')
  }
  if (code === 'BATCH_IMAGE_TOO_MANY_OUTPUT_IMAGES') {
    return batchImageText('tooManyOutputImages')
  }
  if (code === 'BATCH_IMAGE_IDEMPOTENCY_CONFLICT') {
    return batchImagePlainError(batchImageText('idempotencyConflict'))
  }
  if (code === 'BATCH_IMAGE_NOT_READY') {
    return batchImageText('notReady')
  }
  if (code === 'BATCH_IMAGE_OUTPUT_DELETED') {
    return batchImageText('outputDeleted')
  }
  if (code === 'BATCH_IMAGE_RESULT_MISSING') {
    return batchImageAdminError(batchImageText('resultMissing'), error)
  }
  if (code === 'BATCH_IMAGE_ITEM_FAILED') {
    return batchImagePlainError(batchImageText('itemFailed'))
  }
  if (code === 'BATCH_IMAGE_ITEM_IMAGE_INDEX_OUT_OF_RANGE') {
    return batchImagePlainError(batchImageText('itemImageIndexOutOfRange'))
  }
  if (code === 'BATCH_IMAGE_DOWNLOAD_LIMITED') {
    return batchImageText('downloadLimited')
  }
  if (code === 'BATCH_IMAGE_DOWNLOAD_TOO_LARGE') {
    return batchImageText('downloadTooLarge')
  }
  if (code === 'BATCH_IMAGE_RECORD_DELETE_NOT_READY') {
    return batchImagePlainError(batchImageText('deleteNotReady'))
  }
  if (code === 'BATCH_IMAGE_DISABLED') {
    return batchImageAdminError(batchImageText('disabled'), error)
  }
  if (code === 'INTERNAL_ERROR' || code === '500') {
    return batchImageAdminError(fallback, error)
  }
  if (isZhLocale()) {
    const detail = message ? `${batchImageText('errorReference')}：${message}` : batchImageText('adminReference')
    return `${fallback}。${detail} ${batchImageErrorReference(error)}`
  }
  return message || fallback
}

function formatDate(timestamp: number) {
  if (!timestamp) return ''
  return new Date(timestamp * 1000).toLocaleString()
}

function defaultTaskName(timestamp?: number) {
  const date = timestamp ? new Date(timestamp * 1000) : new Date()
  return date.toLocaleString()
}

onMounted(() => {
  void appStore.fetchPublicSettings()
  void refreshPage()
  void cleanupPreviewCache()
  previewCacheCleanupTimer = setInterval(() => {
    void cleanupPreviewCache()
  }, 60 * 60 * 1000)
  document.addEventListener('click', dismissMoreMenu)
  window.addEventListener('resize', dismissMoreMenu)
  window.addEventListener('scroll', dismissMoreMenu, true)
  document.addEventListener('keydown', handleMoreMenuDocumentKeydown)
  window.addEventListener('resize', closePromptPopover)
  window.addEventListener('scroll', closePromptPopover, true)
})

watch(
  () => form.apiKeyId,
  () => {
    void loadAvailableModels()
  },
)

watch(
  () => form.model,
  () => {
    const limit = selectedModelReferenceLimit.value
    if (limit <= 0) {
      referenceImageDrafts.value = []
      return
    }
    if (referenceImageDrafts.value.length > limit) {
      referenceImageDrafts.value = referenceImageDrafts.value.slice(0, limit)
    }
  },
)

onBeforeUnmount(() => {
  stopPolling()
  if (previewCacheCleanupTimer) {
    clearInterval(previewCacheCleanupTimer)
    previewCacheCleanupTimer = null
  }
  clearItemPreviews()
  closeMoreMenu(false)
  closePromptPopover()
  document.removeEventListener('click', dismissMoreMenu)
  window.removeEventListener('resize', dismissMoreMenu)
  window.removeEventListener('scroll', dismissMoreMenu, true)
  document.removeEventListener('keydown', handleMoreMenuDocumentKeydown)
  window.removeEventListener('resize', closePromptPopover)
  window.removeEventListener('scroll', closePromptPopover, true)
})
</script>

<style scoped>
.batch-row-action {
  display: flex !important;
  flex-direction: column !important;
  align-items: center !important;
  justify-content: center !important;
  min-width: 42px;
  line-height: 1;
  outline: none;
}

.batch-row-action:focus {
  outline: none;
}

.batch-row-action :deep(svg) {
  margin-right: 0 !important;
}

.batch-prompt-popover {
  max-width: calc(100vw - 16px);
  user-select: text;
}

.batch-prompt-popover p {
  scrollbar-width: thin;
}

.batch-output-count-select {
  height: 36px;
  min-height: 36px;
  padding-top: 0;
  padding-bottom: 0;
  padding-left: 14px;
  padding-right: 34px;
  line-height: 36px;
}
</style>
