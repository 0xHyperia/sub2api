<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-[1440px]">
      <TablePageLayout>
        <template #actions>
          <header class="space-y-3">
            <div class="flex min-w-0 flex-wrap items-start justify-between gap-3">
              <div class="min-w-0">
                <h1 class="page-title">{{ t('admin.distribution.agents.title') }}</h1>
                <p class="page-description">{{ t('admin.distribution.agents.description') }}</p>
              </div>
              <div class="flex shrink-0 flex-wrap items-center justify-end gap-2">
              <button
                type="button"
                class="btn btn-primary h-10 px-3 sm:px-4"
                :title="t('admin.distribution.agents.addL1')"
                :aria-label="t('admin.distribution.agents.addL1')"
                @click="openAddDialog"
              >
                <Icon name="userPlus" size="sm" />
                <span class="hidden sm:inline">{{ t('admin.distribution.agents.addL1') }}</span>
              </button>
              <button
                type="button"
                class="btn btn-secondary h-10 px-3"
                :title="t('admin.distribution.agents.exportFiltered')"
                :aria-label="t('admin.distribution.agents.exportFiltered')"
                :disabled="exporting"
                @click="downloadExport"
              >
                <Icon name="download" size="sm" /><span class="hidden lg:inline"
                  >{{ t('admin.distribution.agents.export') }}</span
                >
              </button>
              <div class="relative">
                <button type="button" class="btn btn-secondary h-10 px-3" :title="t('admin.distribution.agents.columnSettings')" :aria-label="t('admin.distribution.agents.columnSettings')" :aria-expanded="columnMenuOpen" aria-haspopup="menu" @click="columnMenuOpen = !columnMenuOpen"><Icon name="cog" size="sm" /><span class="hidden lg:inline">{{ t('admin.distribution.agents.columnSettings') }}</span></button>
                <div v-if="columnMenuOpen" class="absolute right-0 top-full z-40 mt-1 w-52 rounded-panel border border-outline bg-surface-raised p-1 shadow-floating" role="menu" @click.stop>
                  <label v-for="column in configurableColumns" :key="column.key" class="dropdown-item cursor-pointer"><input v-model="visibleColumnKeys" type="checkbox" :value="column.key" class="h-4 w-4 rounded border-outline text-brand focus:ring-focus" /><span>{{ column.label }}</span></label>
                  <button type="button" class="dropdown-item mt-1 w-full border-t border-outline text-left text-brand" @click="resetColumns">{{ t('admin.distribution.agents.resetColumns') }}</button>
                </div>
              </div>
              <button
                type="button"
                class="btn btn-secondary btn-icon"
                :title="t('common.refresh')"
                :aria-label="t('common.refresh')"
                :disabled="loading"
                @click="load"
              >
                <Icon
                  name="refresh"
                  size="sm"
                  :class="loading ? 'animate-spin' : ''"
                />
              </button>
              </div>
            </div>
            <DistributionAnalyticsRange v-model="range" @change="onListRangeChange" />
          </header>
          <AdminDistributionNav class="mt-5" />
        </template>

        <template #filters>
          <div class="space-y-2">
          <div
            class="grid grid-cols-[minmax(0,1fr)_96px_96px] gap-2 sm:max-w-3xl sm:grid-cols-[minmax(280px,1fr)_140px_140px]"
          >
            <div class="relative min-w-0">
              <Icon
                name="search"
                size="sm"
                class="absolute left-3 top-1/2 -translate-y-1/2 text-foreground-subtle"
              />
              <input
                v-model="search"
                class="input pl-9"
                type="search"
                :placeholder="t('admin.distribution.agents.searchPlaceholder')"
                :aria-label="t('admin.distribution.agents.search')"
              />
            </div>
            <select v-model="depth" class="input" :aria-label="t('admin.distribution.agents.level')">
              <option value="">{{ t('admin.distribution.agents.allLevels') }}</option>
              <option value="1">{{ t('admin.distribution.agentAnalytics.l1') }}</option>
              <option value="2">{{ t('admin.distribution.agentAnalytics.l2') }}</option>
            </select>
            <select v-model="status" class="input" :aria-label="t('admin.distribution.agents.status')">
              <option value="">{{ t('admin.distribution.agents.allStatuses') }}</option>
              <option value="active">{{ t('admin.distribution.agents.active') }}</option>
              <option value="suspended">{{ t('admin.distribution.agents.suspended') }}</option>
              <option value="revoked">{{ t('admin.distribution.agents.revoked') }}</option>
            </select>
          </div>
          <div v-if="search || depth || status" class="flex flex-wrap items-center gap-2 text-xs">
            <span class="text-foreground-subtle">{{ t('admin.distribution.agents.currentFilters') }}</span>
            <button v-if="search" type="button" class="badge badge-gray" @click="search = ''">{{ t('admin.distribution.agents.searchChip', { query: search }) }} <Icon name="x" size="xs" /></button>
            <button v-if="depth" type="button" class="badge badge-gray" @click="depth = ''">{{ t(depth === '1' ? 'admin.distribution.agentAnalytics.l1' : 'admin.distribution.agentAnalytics.l2') }} <Icon name="x" size="xs" /></button>
            <button v-if="status" type="button" class="badge badge-gray" @click="status = ''">{{ statusText(status) }} <Icon name="x" size="xs" /></button>
            <button type="button" class="text-brand hover:underline" @click="clearAgentFilters">{{ t('admin.distribution.agents.clearAll') }}</button>
          </div>
          </div>
        </template>

        <template #table>
          <DataTable
            :columns="columns"
            :data="items"
            :loading="loading"
            row-key="id"
            :actions-count="1"
            clickable-rows
            :server-side-sort="true"
            default-sort-key="created_at"
            default-sort-order="desc"
            sort-storage-key="admin-distribution-agents"
            @sort="changeSort"
            @row-click="openManageDialog"
          >
            <template #cell-email="{ row }">
              <div class="flex min-w-52 max-w-64 items-center gap-3">
                <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-brand-subtle text-sm font-semibold text-brand">{{ agentInitial(row) }}</div>
                <div class="min-w-0">
                <p class="truncate font-medium text-foreground">
                  {{ row.email || t('admin.distribution.agentAnalytics.emailMissing') }}
                </p>
                <p class="mt-0.5 truncate text-xs text-foreground-subtle">
                  {{ row.username || t('admin.distribution.agentAnalytics.usernameMissing') }}
                </p>
                </div>
              </div>
            </template>
            <template #cell-relationship="{ row }">
              <div class="w-28 max-w-28">
                <span
                  class="badge"
                  :class="row.depth === 1 ? 'badge-primary' : 'badge-gray'"
                >
                  {{ t(row.depth === 1 ? 'admin.distribution.agentAnalytics.l1' : 'admin.distribution.agentAnalytics.l2') }}
                </span>
                <p class="mt-1 truncate text-xs text-foreground-subtle">
                  {{
                    row.depth === 1
                      ? t('admin.distribution.agents.platformDirect')
                      : t('admin.distribution.agents.parent', { email: row.parent_email || t('admin.distribution.agents.unknown') })
                  }}
                </p>
              </div>
            </template>
            <template #cell-effective_rate="{ row }">
              <div class="w-24 text-right">
                <p class="font-medium tabular-nums">
                  {{ formatPercent(row.effective_rate_bps) }}
                </p>
                <p class="mt-0.5 text-xs text-foreground-subtle">
                  {{
                    row.rate_override_bps == null ? t('admin.distribution.agents.systemDefault') : t('admin.distribution.agents.customRate')
                  }}
                </p>
              </div>
            </template>
            <template #cell-customers="{ row }"><div class="text-right"><p class="font-medium tabular-nums">{{ row.customer_count || 0 }}</p><p class="text-xs text-foreground-subtle">{{ t('admin.distribution.agents.payingCount', { count: row.paying_customer_count || 0 }) }}</p></div></template>
            <template #cell-period_customers="{ row }"><div class="text-right"><p class="font-medium tabular-nums">{{ periodNewCustomers(row) }} / {{ periodPayingCustomers(row) }}</p><p class="text-xs text-foreground-subtle">{{ t('admin.distribution.agents.newPayingRate', { rate: conversionRate(periodNewCustomers(row), periodPayingCustomers(row)) }) }}</p><p v-if="row.depth === 1" class="text-[11px] text-foreground-subtle">{{ t('admin.distribution.agents.directTeamCounts', { directNew: row.period_customer_count || 0, directPaying: row.period_paying_customers || 0, teamNew: row.team_customer_count || 0, teamPaying: row.team_paying_customers || 0 }) }}</p></div></template>
            <template #cell-customer_paid="{ row }"
              ><div class="text-right">
                <p class="font-medium tabular-nums">
                  {{ money(Number(row.period_customer_paid_cny) + Number(row.team_customer_paid_cny)) }}
                </p>
                <p class="text-xs text-foreground-subtle">
                  {{ t('admin.distribution.agents.directTeam', { direct: money(row.period_customer_paid_cny), team: money(row.team_customer_paid_cny) }) }}
                </p>
              </div></template
            >
            <template #cell-total_earned="{ row }"
              ><div class="text-right">
                <p class="font-semibold tabular-nums text-success-foreground">
                  {{ money(Number(row.period_commission_cny) + Number(row.team_commission_cny)) }}
                </p>
                <p class="text-xs text-foreground-subtle">
                  {{ t('admin.distribution.agents.directTeam', { direct: money(row.period_commission_cny), team: money(row.team_commission_cny) }) }}
                </p>
              </div></template
            >
            <template #cell-available="{ row }"
              ><div class="text-right">
                <p class="font-medium tabular-nums">
                  {{ money(row.available_cny) }}
                </p>
                <p class="text-xs text-foreground-subtle">
                  {{ t('admin.distribution.agents.frozenAmount', { amount: money(row.frozen_cny) }) }}
                </p>
              </div></template
            >
            <template #cell-status="{ row }">
              <span class="badge" :class="statusClass(row.status)">{{
                statusText(row.status)
              }}</span>
            </template>
            <template #cell-actions="{ row }">
              <div class="flex items-center justify-end gap-1">
                <button type="button" class="btn btn-ghost btn-icon btn-sm" :title="t('admin.distribution.agents.viewDetails')" :aria-label="t('admin.distribution.agents.viewDetails')" @click.stop="openManageDialog(row)"><Icon name="eye" size="sm" /></button>
                <DistributionAgentActionMenu :agent="row" @action="handleAgentAction(row, $event)" />
              </div>
            </template>

            <template #mobile-card="{ row }">
              <div class="space-y-3">
                <div class="flex min-w-0 items-start justify-between gap-3">
                  <div class="min-w-0">
                    <p class="truncate font-medium text-foreground">
                      {{ row.email || t('admin.distribution.agentAnalytics.emailMissing') }}
                    </p>
                    <p class="mt-0.5 truncate text-xs text-foreground-subtle">
                      {{ row.username || t('admin.distribution.agentAnalytics.usernameMissing') }}
                    </p>
                  </div>
                  <span
                    class="badge shrink-0"
                    :class="statusClass(row.status)"
                    >{{ statusText(row.status) }}</span
                  >
                </div>
                <div class="flex flex-wrap items-center gap-2">
                  <span
                    class="badge"
                    :class="row.depth === 1 ? 'badge-primary' : 'badge-gray'"
                    >{{ t(row.depth === 1 ? 'admin.distribution.agentAnalytics.l1' : 'admin.distribution.agentAnalytics.l2') }}</span
                  >
                  <span class="font-mono text-xs text-foreground-muted">{{
                    row.promotion_code
                  }}</span>
                </div>
                <dl
                  class="grid grid-cols-2 gap-3 border-t border-outline pt-3 text-xs"
                >
                  <div>
                    <dt class="text-foreground-subtle">{{ t('admin.distribution.agentAnalytics.periodRate') }}</dt>
                    <dd class="mt-1 font-semibold">
                      {{ periodNewCustomers(row) }} / {{ periodPayingCustomers(row) }} · {{ conversionRate(periodNewCustomers(row), periodPayingCustomers(row)) }}
                    </dd>
                    <p v-if="row.depth === 1" class="mt-0.5 text-[11px] text-foreground-subtle">{{ t('admin.distribution.agents.directTeamCounts', { directNew: row.period_customer_count || 0, directPaying: row.period_paying_customers || 0, teamNew: row.team_customer_count || 0, teamPaying: row.team_paying_customers || 0 }) }}</p>
                  </div>
                  <div>
                    <dt class="text-foreground-subtle">{{ t('admin.distribution.agentAnalytics.periodPaid') }}</dt>
                    <dd class="mt-1 font-semibold">
                      {{ money(Number(row.period_customer_paid_cny) + Number(row.team_customer_paid_cny)) }}
                    </dd>
                    <p v-if="row.depth === 1" class="mt-0.5 text-[11px] text-foreground-subtle">{{ t('admin.distribution.agentAnalytics.directTeam', { direct: money(row.period_customer_paid_cny), team: money(row.team_customer_paid_cny) }) }}</p>
                  </div>
                  <div>
                    <dt class="text-foreground-subtle">{{ t('admin.distribution.agentAnalytics.periodCommission') }}</dt>
                    <dd class="mt-1 font-semibold text-success-foreground">
                      {{ money(Number(row.period_commission_cny) + Number(row.team_commission_cny)) }}
                    </dd>
                    <p v-if="row.depth === 1" class="mt-0.5 text-[11px] text-foreground-subtle">{{ t('admin.distribution.agentAnalytics.directTeam', { direct: money(row.period_commission_cny), team: money(row.team_commission_cny) }) }}</p>
                  </div>
                  <div>
                    <dt class="text-foreground-subtle">{{ t('admin.distribution.agentAnalytics.rate') }}</dt>
                    <dd class="mt-1 font-semibold">
                      {{ formatPercent(row.effective_rate_bps) }}
                    </dd>
                  </div>
                  <div>
                    <dt class="text-foreground-subtle">{{ t('admin.distribution.agentAnalytics.balance') }}</dt>
                    <dd class="mt-1 font-semibold">
                      {{ money(row.available_cny) }} /
                      {{ money(row.frozen_cny) }}
                    </dd>
                  </div>
                </dl>
                <button
                  type="button"
                  class="btn btn-secondary w-full"
                  @click="openManageDialog(row)"
                >
                  <Icon name="eye" size="sm" />{{ t('admin.distribution.agentAnalytics.details') }}
                </button>
              </div>
            </template>

            <template #empty>
              <div class="py-10 text-center">
                <p class="font-medium text-foreground">{{ t('admin.distribution.agents.noMatch') }}</p>
                <p class="mt-1 text-sm text-foreground-subtle">
                  {{ t('admin.distribution.agents.noMatchHint') }}
                </p>
              </div>
            </template>
          </DataTable>
        </template>

        <template #pagination>
          <Pagination
            v-if="pagination.total > 0"
            :page="pagination.page"
            :total="pagination.total"
            :page-size="pagination.page_size"
            @update:page="changePage"
            @update:pageSize="changePageSize"
          />
        </template>
      </TablePageLayout>
    </div>

    <BaseDialog
      :show="addDialog"
      :title="agentDialogTitle"
      width="normal"
      @close="closeAddDialog"
    >
      <form
        id="add-distribution-agent-form"
        class="space-y-5"
        @submit.prevent="submitAgent"
      >
        <ol class="grid grid-cols-3 gap-2" :aria-label="t('admin.distribution.agents.stepLabel')">
          <li v-for="(step, index) in agentSteps" :key="step" class="flex min-w-0 items-center gap-2 text-xs" :class="agentStep >= index + 1 ? 'text-brand' : 'text-foreground-subtle'"><span class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full border text-[11px] font-semibold" :class="agentStep >= index + 1 ? 'border-brand bg-brand-subtle' : 'border-outline'">{{ index + 1 }}</span><span class="truncate">{{ step }}</span></li>
        </ol>
        <template v-if="agentStep === 1">
        <RemoteEntityCombobox
          v-model="selectedUser"
          input-id="distribution-agent-user"
          :label="t('admin.distribution.agents.selectUser')"
          :placeholder="t('admin.distribution.agents.userPlaceholder')"
          :search="searchAgentCandidates"
        />
        <p class="text-sm text-foreground-subtle">{{ t('admin.distribution.agents.selectUserHint') }}</p>
        </template>
        <template v-else-if="agentStep === 2">
        <div class="grid gap-4 sm:grid-cols-2">
          <div>
            <label for="distribution-agent-depth" class="input-label">{{ t('admin.distribution.agents.level') }}</label>
            <select id="distribution-agent-depth" v-model="agentDepth" class="input">
              <option :value="1">{{ t('admin.distribution.agentAnalytics.l1') }}</option>
              <option :value="2">{{ t('admin.distribution.agentAnalytics.l2') }}</option>
            </select>
          </div>
          <div v-if="agentDepth === 2">
            <label for="distribution-agent-parent" class="input-label">{{ t('admin.distribution.agents.parentL1') }}</label>
            <select id="distribution-agent-parent" v-model="parentAgentId" class="input">
              <option :value="null">{{ t('admin.distribution.agents.selectParent') }}</option>
              <option v-for="parent in l1Agents" :key="parent.agent_id" :value="parent.agent_id">
                {{ parent.email || parent.username }}
              </option>
            </select>
          </div>
        </div>
        <p v-if="isCustomerUpgrade" class="rounded-md border border-warning/30 bg-warning/10 px-3 py-2 text-sm text-warning">
          {{ t('admin.distribution.agents.upgradeHint', { level: agentDepth === 1 ? 1 : 2 }) }}
        </p>
        </template>
        <template v-else>
        <div>
          <label for="distribution-agent-rate" class="input-label"
            >{{ t('admin.distribution.agents.rate') }}</label
          >
          <div class="relative">
            <input
              id="distribution-agent-rate"
              v-model="ratePercent"
              type="number"
              class="input pr-9"
              min="0"
              max="100"
              step="0.01"
              :placeholder="defaultRateLabel"
            />
            <span class="input-suffix">%</span>
          </div>
          <p class="input-hint">{{ t('admin.distribution.agents.defaultRateHint', { rate: defaultRateLabel }) }}</p>
        </div>

        <div class="border-t border-outline pt-4">
          <button
            type="button"
            class="flex items-center gap-2 text-sm font-medium text-foreground-muted hover:text-foreground"
            :aria-expanded="showAdvanced"
            @click="showAdvanced = !showAdvanced"
          >
            <Icon
              name="chevronDown"
              size="sm"
              :class="showAdvanced ? 'rotate-180' : ''"
            />
            {{ t('admin.distribution.agents.advanced') }}
          </button>
          <div v-if="showAdvanced" class="mt-4">
            <label for="distribution-agent-code" class="input-label"
              >{{ t('admin.distribution.agents.customCode') }}</label
            >
            <input
              id="distribution-agent-code"
              v-model="promotionCode"
              type="text"
              class="input font-mono uppercase"
              maxlength="32"
              :placeholder="t('admin.distribution.agents.autoCode')"
            />
          </div>
        </div>
        <div class="rounded-panel border border-outline bg-surface-subtle p-4 text-sm"><p class="font-medium">{{ t('admin.distribution.agents.confirmConfig') }}</p><dl class="mt-3 grid grid-cols-2 gap-3 text-xs"><div><dt class="text-foreground-subtle">{{ t('admin.distribution.agents.user') }}</dt><dd class="mt-1 truncate font-medium">{{ selectedUser?.email || selectedUser?.username }}</dd></div><div><dt class="text-foreground-subtle">{{ t('admin.distribution.agents.level') }}</dt><dd class="mt-1 font-medium">{{ t(agentDepth === 1 ? 'admin.distribution.agentAnalytics.l1' : 'admin.distribution.agentAnalytics.l2') }}</dd></div><div><dt class="text-foreground-subtle">{{ t('admin.distribution.agents.rate') }}</dt><dd class="mt-1 font-medium">{{ String(ratePercent || '').trim() ? `${ratePercent}%` : `${t('admin.distribution.agents.systemDefault')} ${defaultRateLabel}%` }}</dd></div><div><dt class="text-foreground-subtle">{{ t('admin.distribution.agents.promotionCode') }}</dt><dd class="mt-1 font-mono">{{ promotionCode || t('admin.distribution.agents.autoGenerate') }}</dd></div></dl></div>
        </template>
      </form>
      <template #footer>
        <div class="flex w-full justify-between gap-2">
          <button
            type="button"
            class="btn btn-secondary"
            @click="agentStep === 1 ? closeAddDialog() : agentStep--"
          >
            {{ t(agentStep === 1 ? 'admin.distribution.agents.cancel' : 'admin.distribution.agents.previous') }}
          </button>
          <button
            v-if="agentStep < 3"
            type="button"
            class="btn btn-primary"
            :disabled="agentStep === 1 ? !selectedUser : agentDepth === 2 && !parentAgentId"
            @click="agentStep++"
          >{{ t('admin.distribution.agents.next') }}<Icon name="arrowRight" size="sm" /></button>
          <button
            v-else
            type="submit"
            form="add-distribution-agent-form"
            class="btn btn-primary"
            :disabled="submitting || !selectedUser"
          >
            <Icon name="userPlus" size="sm" />{{
              submitting ? t('admin.distribution.agents.adding') : t('admin.distribution.agents.confirmAdd')
            }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <BaseDrawer
      :show="manageDialog"
      :title="t('admin.distribution.agents.detailTitle')"
      :description="t('admin.distribution.agents.detailDescription')"
      @close="closeManageDialog"
    >
      <div v-if="managedAgent" class="space-y-5">
        <div class="flex flex-wrap items-start justify-between gap-4 border-b border-outline pb-4">
          <div class="flex min-w-0 items-center gap-3">
            <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-full bg-brand-subtle text-brand"><Icon name="user" size="md" /></div>
            <div class="min-w-0">
              <p class="truncate text-base font-semibold text-foreground">{{ managedAgent.email || t('admin.distribution.agentAnalytics.emailMissing') }}</p>
              <p class="mt-1 truncate text-sm text-foreground-subtle">{{ managedAgent.username || t('admin.distribution.agentAnalytics.usernameMissing') }}</p>
            </div>
          </div>
          <div class="flex shrink-0 items-center gap-2">
            <span class="badge" :class="managedAgent.depth === 1 ? 'badge-primary' : 'badge-gray'">{{ t(managedAgent.depth === 1 ? 'admin.distribution.agentAnalytics.l1' : 'admin.distribution.agentAnalytics.l2') }}</span>
            <span class="badge" :class="statusClass(managedAgent.status)">{{ statusText(managedAgent.status) }}</span>
          </div>
        </div>
        <DistributionAnalyticsRange v-model="detailRange" @change="loadManagedAnalytics" />
        <div class="flex gap-1 overflow-x-auto border-b border-outline" role="tablist" :aria-label="t('admin.distribution.agents.detailTitle')">
          <button v-for="tab in manageTabs" :key="tab.key" type="button" role="tab" :aria-selected="manageTab === tab.key" class="shrink-0 border-b-2 px-3 py-2 text-sm transition-colors" :class="manageTab === tab.key ? 'border-brand font-semibold text-foreground' : 'border-transparent text-foreground-subtle hover:text-foreground'" @click="manageTab = tab.key">{{ tab.label }}</button>
        </div>
        <section v-if="analyticsLoading && manageTab !== 'audit'" class="flex min-h-48 items-center justify-center"><LoadingSpinner /></section>
        <section v-else-if="manageTab === 'overview' && agentAnalytics" class="space-y-4">
          <div class="grid grid-cols-2 gap-px overflow-hidden rounded-panel border border-outline bg-outline sm:grid-cols-4"><article v-for="item in detailKpis" :key="item.label" class="bg-surface p-4"><p class="text-xs text-foreground-subtle">{{ item.label }}</p><strong class="mt-2 block text-lg tabular-nums">{{ item.value }}</strong><small class="mt-1 block text-xs text-foreground-muted">{{ item.hint }}</small></article></div>
          <DistributionBusinessChart :direct="agentAnalytics.analytics.daily_direct" :team="agentAnalytics.analytics.daily_team" :resolution="agentAnalytics.analytics.trend_resolution" :date-from="agentAnalytics.analytics.date_from" :date-to="agentAnalytics.analytics.date_to" :show-scope="managedAgent.depth === 1" :title="t('common.distributionAnalytics.trend')" :subtitle="detailRangeLabel" />
        </section>
        <section v-else-if="manageTab === 'customers' && agentAnalytics" class="space-y-4">
          <div class="grid gap-4 sm:grid-cols-2"><div class="rounded-panel border border-outline p-4"><p class="text-sm font-semibold">{{ t('admin.distribution.agents.directBusiness') }}</p><dl class="mt-4 grid grid-cols-2 gap-4 text-sm"><div><dt class="text-foreground-subtle">{{ t('admin.distribution.agents.newPaying') }}</dt><dd class="mt-1 font-semibold">{{ agentAnalytics.analytics.direct.current.new_customers }} / {{ agentAnalytics.analytics.direct.current.paying_customers }}</dd></div><div><dt class="text-foreground-subtle">{{ t('admin.distribution.agents.conversion') }}</dt><dd class="mt-1 font-semibold">{{ formatAnalyticsPercent(agentAnalytics.analytics.direct.current.conversion_rate) }}</dd></div><div><dt class="text-foreground-subtle">{{ t('admin.distribution.agents.customerPaid') }}</dt><dd class="mt-1 font-semibold">{{ money(agentAnalytics.analytics.direct.current.customer_paid_cny) }}</dd></div><div><dt class="text-foreground-subtle">{{ t('admin.distribution.agents.directCommission') }}</dt><dd class="mt-1 font-semibold">{{ money(agentAnalytics.analytics.direct.current.commission_cny) }}</dd></div></dl></div><div v-if="managedAgent.depth === 1" class="rounded-panel border border-outline p-4"><p class="text-sm font-semibold">{{ t('admin.distribution.agents.teamBusiness') }}</p><dl class="mt-4 grid grid-cols-2 gap-4 text-sm"><div><dt class="text-foreground-subtle">{{ t('admin.distribution.agents.newPaying') }}</dt><dd class="mt-1 font-semibold">{{ agentAnalytics.analytics.team.current.new_customers }} / {{ agentAnalytics.analytics.team.current.paying_customers }}</dd></div><div><dt class="text-foreground-subtle">{{ t('admin.distribution.agents.activeChildren') }}</dt><dd class="mt-1 font-semibold">{{ agentAnalytics.analytics.active_agents }}</dd></div><div><dt class="text-foreground-subtle">{{ t('admin.distribution.agents.teamPaid') }}</dt><dd class="mt-1 font-semibold">{{ money(agentAnalytics.analytics.team.current.customer_paid_cny) }}</dd></div><div><dt class="text-foreground-subtle">{{ t('admin.distribution.agents.teamCommission') }}</dt><dd class="mt-1 font-semibold">{{ money(agentAnalytics.analytics.team.current.commission_cny) }}</dd></div></dl></div></div>
          <div v-if="agentAnalytics.ranking.length" class="rounded-panel border border-outline"><div class="border-b border-outline px-4 py-3 text-sm font-semibold">{{ t('admin.distribution.agents.childRanking') }}</div><div class="divide-y divide-outline"><div v-for="(child,index) in agentAnalytics.ranking" :key="child.agent_id" class="flex min-h-14 items-center gap-3 px-4"><span class="text-xs text-foreground-subtle">{{ index + 1 }}</span><span class="min-w-0 flex-1 truncate text-sm">{{ child.username || child.email }}</span><span class="text-right text-sm"><strong class="block tabular-nums">{{ money(child.customer_paid_cny) }}</strong><small class="text-foreground-subtle">{{ t('admin.distribution.agents.acquiredPaid', { newCount: child.new_customers, payingCount: child.paying_customers }) }}</small></span></div></div></div>
        </section>
        <section v-else-if="manageTab === 'commission' && agentAnalytics" class="space-y-4"><dl class="grid grid-cols-2 gap-4 rounded-panel border border-outline p-4 text-sm sm:grid-cols-4"><div><dt class="text-foreground-subtle">{{ t('admin.distribution.agents.availableCommission') }}</dt><dd class="mt-1 font-semibold text-success-foreground">{{ money(managedAgent.available_cny) }}</dd></div><div><dt class="text-foreground-subtle">{{ t('admin.distribution.agents.frozenCommission') }}</dt><dd class="mt-1 font-semibold">{{ money(managedAgent.frozen_cny) }}</dd></div><div><dt class="text-foreground-subtle">{{ t('admin.distribution.agents.reservedCommission') }}</dt><dd class="mt-1 font-semibold">{{ money(managedAgent.reserved_cny) }}</dd></div><div><dt class="text-foreground-subtle">{{ t('admin.distribution.agents.refundDebt') }}</dt><dd class="mt-1 font-semibold" :class="Number(managedAgent.debt_cny) ? 'text-danger-foreground' : ''">{{ money(managedAgent.debt_cny) }}</dd></div><div><dt class="text-foreground-subtle">{{ t('admin.distribution.agents.periodCommission') }}</dt><dd class="mt-1 font-semibold">{{ money(agentAnalytics.analytics.total.current.commission_cny) }}</dd></div><div><dt class="text-foreground-subtle">{{ t('admin.distribution.agents.cumulativeCommission') }}</dt><dd class="mt-1 font-semibold">{{ money(managedAgent.total_earned_cny) }}</dd></div><div><dt class="text-foreground-subtle">{{ t('admin.distribution.agents.cumulativeWithdrawn') }}</dt><dd class="mt-1 font-semibold">{{ money(managedAgent.total_withdrawn_cny) }}</dd></div><div><dt class="text-foreground-subtle">{{ t('admin.distribution.agents.convertedBalance') }}</dt><dd class="mt-1 font-semibold">{{ money(managedAgent.total_converted_cny) }}</dd></div></dl></section>
        <section v-else-if="manageTab === 'strategy'" class="rounded-panel border border-outline p-4"><dl class="grid grid-cols-2 gap-4 text-sm sm:grid-cols-4"><div><dt class="text-foreground-subtle">{{ t('admin.distribution.agents.promotionCode') }}</dt><dd class="mt-1 truncate font-mono">{{ managedAgent.promotion_code }}</dd></div><div><dt class="text-foreground-subtle">{{ t('admin.distribution.agents.rate') }}</dt><dd class="mt-1 font-semibold">{{ formatPercent(managedAgent.effective_rate_bps) }}</dd></div><div><dt class="text-foreground-subtle">{{ t('admin.distribution.agents.recruitment') }}</dt><dd class="mt-1"><span class="badge" :class="managedAgent.can_recruit_subagents ? 'badge-success':'badge-gray'">{{ t(managedAgent.can_recruit_subagents ? 'admin.distribution.agents.authorized' : 'admin.distribution.agents.unauthorized') }}</span></dd></div><div><dt class="text-foreground-subtle">{{ t('admin.distribution.agents.promotionStats') }}</dt><dd class="mt-1"><span class="badge" :class="managedAgent.can_view_promotion_stats ? 'badge-success':'badge-gray'">{{ t(managedAgent.can_view_promotion_stats ? 'admin.distribution.agents.viewable' : 'admin.distribution.agents.noPermission') }}</span></dd></div></dl></section>
        <section v-if="manageTab === 'audit'">
          <div class="flex items-center justify-between gap-3">
            <h3 class="text-sm font-semibold">{{ t('admin.distribution.agents.changeLog') }}</h3>
            <span class="text-xs text-foreground-subtle"
              >{{ t('admin.distribution.agents.recentEvents', { count: agentEvents.length }) }}</span
            >
          </div>
          <div v-if="eventsLoading" class="mt-3 flex justify-center py-4">
            <span class="text-sm text-foreground-subtle">{{ t('common.loading') }}</span>
          </div>
          <ol
            v-else-if="agentEvents.length"
            class="mt-3 space-y-3 border-l border-outline pl-4"
          >
            <li v-for="event in agentEvents" :key="event.id" class="relative">
              <span
                class="absolute -left-[21px] top-1.5 h-2.5 w-2.5 rounded-full border-2 border-surface bg-brand"
              ></span>
              <div class="flex flex-wrap items-center justify-between gap-2">
                <p class="text-sm font-medium">{{ agentEventTitle(event) }}</p>
                <time class="text-xs text-foreground-subtle">{{
                  new Date(event.created_at).toLocaleString()
                }}</time>
              </div>
              <p class="mt-1 text-xs text-foreground-subtle">
                {{ event.actor_email || t('admin.distribution.agents.system') }} ·
                {{ event.reason || t('admin.distribution.agents.noNote') }}
              </p>
            </li>
          </ol>
          <p v-else class="mt-3 text-sm text-foreground-subtle">{{ t('admin.distribution.agents.noEvents') }}</p>
        </section>
        <div v-if="manageTab !== 'audit'" class="flex flex-wrap justify-end gap-2 border-t border-outline pt-4">
          <button type="button" class="btn btn-secondary" @click="openRewardRuleDialog"><Icon name="gift" size="sm" />{{ t('common.distributionRewards.adminTitle') }}</button>
          <button type="button" class="btn btn-secondary" @click="openPromotionStatsPermissionDialog"><Icon name="chart" size="sm" />{{ t(managedAgent.can_view_promotion_stats ? 'admin.distribution.agents.closeStats' : 'admin.distribution.agents.openStats') }}</button>
          <button
            v-if="managedAgent.depth === 1"
            type="button"
            class="btn btn-secondary"
            @click="openRecruitmentPermissionDialog"
          >
            <Icon name="users" size="sm" />{{
              managedAgent.can_recruit_subagents
                ? t('admin.distribution.agents.revokeRecruitment')
                : t('admin.distribution.agents.grantRecruitment')
            }}
          </button>
          <button
            type="button"
            class="btn btn-secondary"
            @click="openRateDialog"
          >
            <Icon name="edit" size="sm" />{{ t('admin.distribution.agents.editRate') }}
          </button>
          <button
            v-if="managedAgent.status === 'active'"
            type="button"
            class="btn btn-secondary"
            @click="openStatusDialog('suspended')"
          >
            <Icon name="ban" size="sm" />{{ t('admin.distribution.agents.suspendAgent') }}
          </button>
          <button
            v-else
            type="button"
            class="btn btn-primary"
            @click="openStatusDialog('active')"
          >
            <Icon name="check" size="sm" />{{
              t(managedAgent.status === "revoked" ? 'admin.distribution.agents.reactivateAgent' : 'admin.distribution.agents.restoreAgent')
            }}
          </button>
          <button
            v-if="managedAgent.status !== 'revoked'"
            type="button"
            class="btn btn-danger"
            @click="openStatusDialog('revoked')"
          >
            <Icon name="trash" size="sm" />{{ t('admin.distribution.agents.revokeAgent') }}
          </button>
        </div>
      </div>
    </BaseDrawer>

    <BaseDialog :show="rewardRuleDialog" :title="t('common.distributionRewards.adminTitle')" width="normal" @close="rewardRuleDialog = false">
      <form id="admin-reward-rule-form" class="space-y-4" @submit.prevent="saveRewardRule">
        <p class="rounded-control bg-surface-subtle px-3 py-2 text-sm text-foreground-subtle">{{ t('common.distributionRewards.adminHint') }}</p>
        <div class="flex items-center justify-between"><span class="form-label">{{ t('common.distributionRewards.registration') }}</span><Toggle v-model="rewardRule.registration_enabled" :aria-label="t('common.distributionRewards.registration')" /></div>
        <label class="form-field"><span class="form-label">{{ t('common.distributionRewards.registrationTotal') }}</span><div class="relative"><input v-model="rewardRule.registration_reward_cny" class="input pl-8" type="number" min="0" step="0.01"><span class="input-prefix">¥</span></div></label>
        <div class="flex items-center justify-between"><span class="form-label">{{ t('common.distributionRewards.recharge') }}</span><Toggle v-model="rewardRule.recharge_enabled" :aria-label="t('common.distributionRewards.recharge')" /></div>
        <label class="form-field"><span class="form-label">{{ t('common.distributionRewards.rechargeThreshold') }}</span><div class="relative"><input v-model="rewardRule.recharge_threshold_cny" class="input pl-8" type="number" min="0" step="0.01"><span class="input-prefix">¥</span></div></label>
        <label class="form-field"><span class="form-label">{{ t('common.distributionRewards.rechargeTotal') }}</span><div class="relative"><input v-model="rewardRule.recharge_reward_cny" class="input pl-8" type="number" min="0" step="0.01"><span class="input-prefix">¥</span></div></label>
      </form>
      <template #footer><div class="flex justify-end gap-2"><button type="button" class="btn btn-secondary" @click="rewardRuleDialog = false">{{ t('common.distributionRewards.cancel') }}</button><button type="submit" form="admin-reward-rule-form" class="btn btn-primary" :disabled="rewardRuleSaving">{{ rewardRuleSaving ? t('common.distributionRewards.saving') : t('common.distributionRewards.saveTotal') }}</button></div></template>
    </BaseDialog>

    <BaseDialog :show="promotionStatsPermissionDialog" :title="t(managedAgent?.can_view_promotion_stats ? 'admin.distribution.agents.statsCloseTitle' : 'admin.distribution.agents.statsOpenTitle')" width="narrow" @close="promotionStatsPermissionDialog = false">
      <div class="space-y-4"><p class="text-sm text-foreground-subtle">{{ t(managedAgent?.can_view_promotion_stats ? 'admin.distribution.agents.statsCloseHint' : 'admin.distribution.agents.statsOpenHint') }}</p><div><label for="distribution-promotion-stats-reason" class="input-label">{{ t('admin.distribution.agents.operationReason') }}</label><textarea id="distribution-promotion-stats-reason" v-model="promotionStatsPermissionReason" class="input min-h-24 resize-y" maxlength="200" :placeholder="t('admin.distribution.agents.statsReasonPlaceholder')"></textarea></div></div>
      <template #footer><div class="flex justify-end gap-2"><button class="btn btn-secondary" @click="promotionStatsPermissionDialog = false">{{ t('admin.distribution.agents.cancel') }}</button><button class="btn btn-primary" :disabled="promotionStatsPermissionSaving || !promotionStatsPermissionReason.trim()" @click="savePromotionStatsPermission">{{ t(promotionStatsPermissionSaving ? 'admin.distribution.agents.saving' : 'admin.distribution.agents.confirm') }}</button></div></template>
    </BaseDialog>

    <BaseDialog
      :show="recruitmentPermissionDialog"
      :title="
        t(managedAgent?.can_recruit_subagents ? 'admin.distribution.agents.revokeRecruitment' : 'admin.distribution.agents.grantRecruitment')
      "
      width="narrow"
      @close="recruitmentPermissionDialog = false"
    >
      <div class="space-y-4">
        <p class="text-sm text-foreground-subtle">
          {{
            managedAgent?.can_recruit_subagents
              ? t('admin.distribution.agents.revokeRecruitmentHint')
              : t('admin.distribution.agents.grantRecruitmentHint')
          }}
        </p>
        <div>
          <label for="distribution-recruitment-reason" class="input-label"
            >{{ t('admin.distribution.agents.operationReason') }}</label
          >
          <textarea
            id="distribution-recruitment-reason"
            v-model="recruitmentPermissionReason"
            class="input min-h-24 resize-y"
            maxlength="200"
            :placeholder="t('admin.distribution.agents.recruitmentReasonPlaceholder')"
          ></textarea>
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button
            class="btn btn-secondary"
            @click="recruitmentPermissionDialog = false"
          >
            {{ t('admin.distribution.agents.cancel') }}
          </button>
          <button
            class="btn btn-primary"
            :disabled="
              recruitmentPermissionSaving || !recruitmentPermissionReason.trim()
            "
            @click="saveRecruitmentPermission"
          >
            {{ t(recruitmentPermissionSaving ? 'admin.distribution.agents.saving' : 'admin.distribution.agents.confirm') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <BaseDialog
      :show="rateDialog"
      :title="t('admin.distribution.agents.editRate')"
      width="narrow"
      @close="rateDialog = false"
    >
      <div v-if="managedAgent" class="space-y-4">
        <div>
          <p class="text-sm font-medium">{{ managedAgent.email }}</p>
          <p class="mt-1 text-xs text-foreground-subtle">
            {{ t('admin.distribution.agents.currentEffective', { rate: formatPercent(managedAgent.effective_rate_bps), mode: t(managedAgent.rate_override_bps == null ? 'admin.distribution.agents.inheritDefault' : 'admin.distribution.agents.customRate') }) }}
          </p>
        </div>
        <div
          class="inline-flex w-full rounded-control border border-outline bg-surface-subtle p-1"
          role="group"
          :aria-label="t('admin.distribution.agents.ratePolicy')"
        >
          <button
            type="button"
            class="btn flex-1"
            :class="rateUseDefault ? 'btn-primary' : 'btn-ghost'"
            @click="rateUseDefault = true"
          >
            {{ t('admin.distribution.agents.systemDefault') }}</button
          ><button
            type="button"
            class="btn flex-1"
            :class="!rateUseDefault ? 'btn-primary' : 'btn-ghost'"
            @click="rateUseDefault = false"
          >
            {{ t('admin.distribution.agents.customRate') }}
          </button>
        </div>
        <div v-if="!rateUseDefault">
          <label for="distribution-edit-rate" class="input-label"
            >{{ t('admin.distribution.agents.rate') }}</label
          >
          <div class="relative">
            <input
              id="distribution-edit-rate"
              v-model="editRatePercent"
              class="input pr-9"
              type="number"
              min="0"
              max="100"
              step="0.01"
            /><span class="input-suffix">%</span>
          </div>
          <p class="input-hint">{{ t('admin.distribution.agents.defaultRateValue', { rate: managedDefaultRate }) }}</p>
        </div>
        <div>
          <label for="distribution-rate-reason" class="input-label"
            >{{ t('admin.distribution.agents.changeReason') }}</label
          ><textarea
            id="distribution-rate-reason"
            v-model="rateReason"
            class="input min-h-24 resize-y"
            maxlength="200"
            :placeholder="t('admin.distribution.agents.changeReasonPlaceholder')"
          ></textarea>
          <p class="input-hint">{{ t('admin.distribution.agents.auditHint') }}</p>
        </div>
      </div>
      <template #footer
        ><div class="flex justify-end gap-2">
          <button class="btn btn-secondary" @click="rateDialog = false">
            {{ t('admin.distribution.agents.cancel') }}</button
          ><button
            class="btn btn-primary"
            :disabled="rateSaving || !canSaveRate"
            @click="saveRate"
          >
            {{ t(rateSaving ? 'admin.distribution.agents.saving' : 'admin.distribution.agents.saveRate') }}
          </button>
        </div></template
      >
    </BaseDialog>

    <BaseDialog
      :show="statusDialog"
      :title="statusDialogTitle"
      width="narrow"
      @close="statusDialog = false"
    >
      <div class="space-y-4">
        <p class="text-sm text-foreground-subtle">{{ statusDialogMessage }}</p>
        <div>
          <label for="distribution-agent-status-reason" class="input-label">
            {{ t('admin.distribution.agents.operationReason') }}
          </label>
          <textarea
            id="distribution-agent-status-reason"
            v-model="statusReason"
            class="input min-h-24 resize-y"
            maxlength="200"
            :placeholder="t('admin.distribution.agents.statusReasonPlaceholder')"
          ></textarea>
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button
            type="button"
            class="btn btn-secondary"
            @click="statusDialog = false"
          >
            {{ t('admin.distribution.agents.cancel') }}
          </button>
          <button
            type="button"
            class="btn"
            :class="pendingStatus === 'revoked' ? 'btn-danger' : 'btn-primary'"
            :disabled="
              statusSaving || (statusReasonRequired && !statusReason.trim())
            "
            @click="confirmStatusChange"
          >
            {{ t(statusSaving ? 'admin.distribution.agents.processing' : 'admin.distribution.agents.confirm') }}
          </button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import Toggle from "@/components/common/Toggle.vue";
import { useRoute, useRouter } from "vue-router";
import AppLayout from "@/components/layout/AppLayout.vue";
import TablePageLayout from "@/components/layout/TablePageLayout.vue";
import DataTable from "@/components/common/DataTable.vue";
import Pagination from "@/components/common/Pagination.vue";
import BaseDialog from "@/components/common/BaseDialog.vue";
import BaseDrawer from "@/components/common/BaseDrawer.vue";
import LoadingSpinner from "@/components/common/LoadingSpinner.vue";
import DistributionBusinessChart from "@/components/distribution/DistributionBusinessChart.vue";
import DistributionAnalyticsRange from "@/components/distribution/DistributionAnalyticsRange.vue";
import { analyticsRangeParams, defaultDistributionAnalyticsRange, formatAnalyticsRangeLabel, type DistributionAnalyticsRangeValue } from "@/components/distribution/distributionAnalyticsRange";
import Icon from "@/components/icons/Icon.vue";
import RemoteEntityCombobox from "@/components/admin/distribution/RemoteEntityCombobox.vue";
import DistributionAgentActionMenu from "@/components/admin/distribution/DistributionAgentActionMenu.vue";
import AdminDistributionNav from "@/components/admin/distribution/AdminDistributionNav.vue";
import type { DistributionPickerOption } from "@/components/admin/distribution/types";
import type { Column } from "@/components/common/types";
import type { DistributionAgent, DistributionAgentAnalytics } from "@/api/distribution";
import {
  getSettings,
  getAgentAnalytics,
  grantAgent,
  exportAgents,
  listAgentEvents,
  listAgents,
  lookupAgentCandidates,
  lookupAgents,
  updateAgentRate,
  updateAgentRecruitmentPermission,
  updateAgentPromotionStatsPermission,
  updateAgentStatus,
  type DistributionAgentEvent,
  type DistributionSettings,
  getAgentRewardRule,
  updateAgentRewardRule,
} from "@/api/admin/distribution";
import { useAppStore } from "@/stores/app";
import { extractApiErrorMessage, extractI18nErrorMessage } from "@/utils/apiError";
import { saveDistributionExport } from "@/utils/distributionExport";

const app = useAppStore();
const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const items = ref<DistributionAgent[]>([]);
const loading = ref(false);
const exporting = ref(false);
const search = ref(String(route.query.search || ""));
const status = ref("");
const depth = ref("");
const pagination = ref({ page: 1, page_size: 20, total: 0, pages: 1 });
const range = ref(defaultDistributionAnalyticsRange());
let searchTimer: number | null = null;

const addDialog = ref(false);
const agentStep = ref(1);
const agentSteps = computed(() => [t('admin.distribution.agents.stepUser'), t('admin.distribution.agents.stepRelation'), t('admin.distribution.agents.stepBusiness')]);
const selectedUser = ref<DistributionPickerOption | null>(null);
const agentDepth = ref<1 | 2>(1);
const parentAgentId = ref<number | null>(null);
const l1Agents = ref<Awaited<ReturnType<typeof lookupAgents>>>([]);
const ratePercent = ref("");
const promotionCode = ref("");
const showAdvanced = ref(false);
const submitting = ref(false);
const settings = ref<DistributionSettings | null>(null);
const rewardRuleDialog = ref(false);
const rewardRuleSaving = ref(false);
const rewardRule = ref({ registration_enabled: false, registration_reward_cny: "0", recharge_enabled: false, recharge_threshold_cny: "0", recharge_reward_cny: "0" });
const isCustomerUpgrade = computed(() => selectedUser.value?.reason === t('admin.distribution.agents.distributionCustomer'));
const agentDialogTitle = computed(() =>
  isCustomerUpgrade.value ? t('admin.distribution.agents.upgrade') : t(agentDepth.value === 1 ? 'admin.distribution.agents.addL1' : 'admin.distribution.agents.addL2'),
);

const manageDialog = ref(false);
const managedAgent = ref<DistributionAgent | null>(null);
const manageTab = ref<'overview' | 'customers' | 'commission' | 'strategy' | 'audit'>('overview');
const manageTabs = computed(() => [
  { key: 'overview' as const, label: t('admin.distribution.agents.tabOverview') },
  { key: 'customers' as const, label: t('admin.distribution.agents.tabCustomers') },
  { key: 'commission' as const, label: t('admin.distribution.agents.tabCommission') },
  { key: 'strategy' as const, label: t('admin.distribution.agents.tabStrategy') },
  { key: 'audit' as const, label: t('admin.distribution.agents.tabAudit') },
]);
const detailRange = ref(defaultDistributionAnalyticsRange());
const detailRangeLabel = computed(() => formatAnalyticsRangeLabel(detailRange.value));
const agentAnalytics = ref<DistributionAgentAnalytics | null>(null);
const analyticsLoading = ref(false);
const statusDialog = ref(false);
const pendingStatus = ref<"active" | "suspended" | "revoked">("suspended");
const statusReason = ref("");
const statusSaving = ref(false);
const agentEvents = ref<DistributionAgentEvent[]>([]);
const eventsLoading = ref(false);
const sortBy = ref("created_at");
const sortOrder = ref<"asc" | "desc">("desc");
const rateDialog = ref(false);
const rateUseDefault = ref(true);
const editRatePercent = ref("");
const rateReason = ref("");
const rateSaving = ref(false);
const recruitmentPermissionDialog = ref(false);
const recruitmentPermissionReason = ref("");
const recruitmentPermissionSaving = ref(false);
const promotionStatsPermissionDialog = ref(false);
const promotionStatsPermissionReason = ref("");
const promotionStatsPermissionSaving = ref(false);

const baseColumns = computed<Column[]>(() => [
  { key: "email", label: t('admin.distribution.agents.colUser'), sortable: true },
  { key: "relationship", label: t('admin.distribution.agents.colRelation') },
  { key: "customers", label: t('admin.distribution.agents.colCustomers'), class: "text-right" },
  { key: "period_customers", label: t('admin.distribution.agents.colPeriodCustomers'), class: "text-right" },
  {
    key: "effective_rate",
    label: t('admin.distribution.agents.colRate'),
    sortable: true,
    class: "text-right",
  },
  {
    key: "customer_paid",
    label: t('admin.distribution.agents.colPaid'),
    sortable: true,
    class: "text-right",
  },
  {
    key: "total_earned",
    label: t('admin.distribution.agents.colCommission'),
    sortable: true,
    class: "text-right",
  },
  {
    key: "available",
    label: t('admin.distribution.agents.colBalance'),
    sortable: true,
    class: "text-right",
  },
  {
    key: "status",
    label: t('admin.distribution.agents.colStatus'),
    sortable: true,
    class: "text-center",
  },
  { key: "actions", label: t('admin.distribution.agents.colActions'), class: "w-16 text-center" },
]);
const configurableColumns = computed(() => baseColumns.value.filter((column) => !['email', 'actions'].includes(column.key)));
const defaultVisibleColumnKeys = ['email', 'relationship', 'period_customers', 'customer_paid', 'total_earned', 'status', 'actions'];
function loadVisibleColumns() {
  try {
    const stored = JSON.parse(localStorage.getItem('admin-distribution-agent-columns') || 'null');
    return Array.isArray(stored) ? stored.filter((key): key is string => typeof key === 'string') : [...defaultVisibleColumnKeys];
  } catch { return [...defaultVisibleColumnKeys]; }
}
const visibleColumnKeys = ref<string[]>(loadVisibleColumns());
const columns = computed(() => baseColumns.value.filter((column) => visibleColumnKeys.value.includes(column.key)));
const columnMenuOpen = ref(false);
function resetColumns() { visibleColumnKeys.value = [...defaultVisibleColumnKeys]; }
watch(visibleColumnKeys, (value) => localStorage.setItem('admin-distribution-agent-columns', JSON.stringify(value)), { deep: true });
function agentInitial(agent: DistributionAgent) { return (agent.username || agent.email || '?').trim().charAt(0).toUpperCase(); }

const defaultRateLabel = computed(() =>
  ((settings.value?.l1_default_rate_bps ?? 0) / 100).toFixed(2),
);
const isReactivatingRevoked = computed(
  () =>
    pendingStatus.value === "active" &&
    managedAgent.value?.status === "revoked",
);
const statusReasonRequired = computed(() => true);
const statusDialogTitle = computed(() =>
  isReactivatingRevoked.value
    ? t('admin.distribution.agents.statusReenableTitle')
    : { active: t('admin.distribution.agents.statusRestoreTitle'), suspended: t('admin.distribution.agents.statusSuspendTitle'), revoked: t('admin.distribution.agents.statusRevokeTitle') }[
        pendingStatus.value
      ],
);
const statusDialogMessage = computed(() =>
  pendingStatus.value === "revoked"
    ? t('admin.distribution.agents.revokeMessage')
    : pendingStatus.value === "suspended"
      ? t('admin.distribution.agents.suspendMessage')
      : isReactivatingRevoked.value
        ? t('admin.distribution.agents.reenableMessage')
        : t('admin.distribution.agents.restoreMessage'),
);
const managedDefaultRate = computed(() => {
  if (!managedAgent.value || !settings.value) return "0.00";
  return (
    (managedAgent.value.depth === 1
      ? settings.value.l1_default_rate_bps
      : settings.value.l2_default_rate_bps) / 100
  ).toFixed(2);
});
const canSaveRate = computed(() => {
  if (!rateReason.value.trim()) return false;
  if (rateUseDefault.value) return true;
  const value = Number(editRatePercent.value);
  return Number.isFinite(value) && value >= 0 && value <= 100;
});

function statusText(value: string) {
  return (
    { active: t('admin.distribution.agents.active'), suspended: t('admin.distribution.agents.suspended'), revoked: t('admin.distribution.agents.revoked') }[value] || value
  );
}

function statusClass(value: string) {
  return value === "active"
    ? "badge-success"
    : value === "suspended"
      ? "badge-warning"
      : "badge-gray";
}

function formatPercent(bps: number) {
  return `${(bps / 100).toFixed(2)}%`;
}

function money(value: string | number) {
  return new Intl.NumberFormat(undefined, { style: "currency", currency: "CNY", minimumFractionDigits: 2 }).format(Number(value || 0));
}

function conversionRate(customers: number, paying: number) {
  return customers > 0 ? `${(paying / customers * 100).toFixed(1)}%` : "—";
}
function periodNewCustomers(row: DistributionAgent) {
  return Number(row.period_customer_count || 0) + (row.depth === 1 ? Number(row.team_customer_count || 0) : 0)
}
function periodPayingCustomers(row: DistributionAgent) {
  return Number(row.period_paying_customers || 0) + (row.depth === 1 ? Number(row.team_paying_customers || 0) : 0)
}

function formatAnalyticsPercent(value: string | number) {
  return `${Number(value || 0).toFixed(1)}%`;
}

const detailKpis = computed(() => agentAnalytics.value ? [
  { label: t('common.distributionAnalytics.currentPaid'), value: money(agentAnalytics.value.analytics.total.current.customer_paid_cny), hint: managedAgent.value?.depth === 1 ? `${t('common.distributionAnalytics.team')} ${money(agentAnalytics.value.analytics.team.current.customer_paid_cny)}` : t('common.distributionAnalytics.direct') },
  { label: t('common.distributionAnalytics.currentCommission'), value: money(agentAnalytics.value.analytics.total.current.commission_cny), hint: managedAgent.value?.depth === 1 ? `${t('common.distributionAnalytics.team')} ${money(agentAnalytics.value.analytics.team.current.commission_cny)}` : t('common.distributionAnalytics.direct') },
  { label: t('common.distributionAnalytics.acquisitionTotal'), value: `${agentAnalytics.value.analytics.total.current.new_customers} / ${agentAnalytics.value.analytics.total.current.paying_customers}`, hint: t('common.distributionAnalytics.range') },
  { label: t('common.distributionAnalytics.conversionRate'), value: formatAnalyticsPercent(agentAnalytics.value.analytics.total.current.conversion_rate), hint: t('common.distributionAnalytics.cohortHint') },
] : []);

async function load() {
  loading.value = true;
  try {
    const result = await listAgents({
      page: pagination.value.page,
      page_size: pagination.value.page_size,
      search: search.value.trim() || undefined,
      status: status.value || undefined,
      depth: depth.value ? Number(depth.value) : undefined,
      sort_by: sortBy.value,
      sort_order: sortOrder.value,
      ...analyticsRangeParams(range.value),
      ...analyticsRangeParams(range.value),
    });
    items.value = result.items;
    pagination.value = {
      page: result.page,
      page_size: result.page_size,
      total: result.total,
      pages: result.pages,
    };
    const requestedAgentID = Number(route.query.agent_id || 0);
    if (requestedAgentID > 0 && !manageDialog.value) {
      const requestedAgent = items.value.find((item) => item.id === requestedAgentID);
      if (requestedAgent) {
        void router.replace({ query: { ...route.query, agent_id: undefined } });
        void openManageDialog(requestedAgent);
      }
    }
  } catch (error) {
    app.showError(extractApiErrorMessage(error, t('admin.distribution.agents.loadFailed')));
  } finally {
    loading.value = false;
  }
}

async function downloadExport() {
  exporting.value = true;
  try {
    const blob = await exportAgents({
      search: search.value.trim() || undefined,
      status: status.value || undefined,
      depth: depth.value ? Number(depth.value) : undefined,
      sort_by: sortBy.value,
      sort_order: sortOrder.value,
    });
    saveDistributionExport(blob, "distribution-agents");
    app.showSuccess(t('admin.distribution.agents.exportSuccess'));
  } catch (error) {
    app.showError(extractApiErrorMessage(error, t('admin.distribution.agents.exportFailed')));
  } finally {
    exporting.value = false;
  }
}

function scheduleLoad() {
  pagination.value.page = 1;
  if (searchTimer) window.clearTimeout(searchTimer);
  searchTimer = window.setTimeout(load, 300);
}

function changePage(page: number) {
  pagination.value.page = page;
  void load();
}

function changePageSize(pageSize: number) {
  pagination.value.page = 1;
  pagination.value.page_size = pageSize;
  void load();
}

function changeSort(key: string, order: "asc" | "desc") {
  sortBy.value = key;
  sortOrder.value = order;
  pagination.value.page = 1;
  void load();
}

async function searchAgentCandidates(
  query: string,
): Promise<DistributionPickerOption[]> {
  const reasons: Record<string, string> = {
    user_inactive: t('admin.distribution.agents.userInactive'),
    already_agent: t('admin.distribution.agents.alreadyAgent'),
    distribution_customer: t('admin.distribution.agents.distributionCustomer'),
    affiliate_invitee: t('admin.distribution.agents.affiliateInvitee'),
  };
  const result = await lookupAgentCandidates(query);
  return result.map((user) => ({
    id: user.user_id,
    email: user.email,
    username: user.username,
    meta: user.status === "active" ? t('admin.distribution.agents.userNormal') : user.status,
    selectable: user.selectable,
    reason: user.unavailable_reason
      ? reasons[user.unavailable_reason] || t('admin.distribution.agents.unavailable')
      : undefined,
  }));
}

async function openAddDialog() {
  addDialog.value = true;
  agentStep.value = 1;
  agentDepth.value = 1;
  parentAgentId.value = null;
  try {
    l1Agents.value = (await lookupAgents("", { include_inactive: false })).filter(
      (agent) => agent.depth === 1 && agent.status === "active",
    );
  } catch {
    l1Agents.value = [];
  }
  if (!settings.value) {
    try {
      settings.value = await getSettings();
    } catch {
      app.showError(t('admin.distribution.agents.loadDefaultsFailed'));
    }
  }
}

function closeAddDialog() {
  addDialog.value = false;
  agentStep.value = 1;
  selectedUser.value = null;
  agentDepth.value = 1;
  parentAgentId.value = null;
  ratePercent.value = "";
  promotionCode.value = "";
  showAdvanced.value = false;
}

async function submitAgent() {
  if (!selectedUser.value) return;
  if (agentDepth.value === 2 && !parentAgentId.value) {
    app.showError(t('admin.distribution.agents.selectParentError'));
    return;
  }
  // Native number inputs can provide a numeric value at runtime even when the
  // ref was initialized with a string. Normalize before validating/submitting.
  const rateInput = String(ratePercent.value ?? "").trim();
  const parsedRate = rateInput === "" ? undefined : Number(rateInput);
  if (
    parsedRate !== undefined &&
    (!Number.isFinite(parsedRate) || parsedRate < 0 || parsedRate > 100)
  ) {
    app.showError(t('admin.distribution.agents.rateRangeError'));
    return;
  }
  submitting.value = true;
  try {
    await grantAgent({
      user_id: selectedUser.value.id,
      depth: agentDepth.value,
      parent_agent_id: agentDepth.value === 2 ? parentAgentId.value ?? undefined : undefined,
      rate_override_bps:
        parsedRate === undefined ? undefined : Math.round(parsedRate * 100),
      promotion_code: promotionCode.value.trim().toUpperCase() || undefined,
      upgrade_customer: isCustomerUpgrade.value || undefined,
    });
    app.showSuccess(isCustomerUpgrade.value ? t('admin.distribution.agents.upgraded') : t('admin.distribution.agents.added', { level: agentDepth.value === 1 ? 1 : 2 }));
    closeAddDialog();
    pagination.value.page = 1;
    await load();
  } catch (error) {
    app.showError(extractApiErrorMessage(error, t('admin.distribution.agents.addFailed')));
  } finally {
    submitting.value = false;
  }
}

async function openManageDialog(agent: DistributionAgent) {
  managedAgent.value = agent;
  manageTab.value = 'overview';
  manageDialog.value = true;
  agentEvents.value = [];
  agentAnalytics.value = null;
  detailRange.value = { ...range.value };
  eventsLoading.value = true;
  analyticsLoading.value = true;
  if (!settings.value)
    void getSettings()
      .then((value) => {
        settings.value = value;
      })
      .catch(() => app.showError(t('admin.distribution.agents.loadRateFailed')));
  try {
    const [events, analytics] = await Promise.all([listAgentEvents(agent.id), getAgentAnalytics(agent.id, analyticsRangeParams(detailRange.value))]);
    agentEvents.value = events;
    agentAnalytics.value = analytics;
  } catch (error) {
    app.showError(extractApiErrorMessage(error, t('admin.distribution.agents.loadEventsFailed')));
  } finally {
    eventsLoading.value = false;
    analyticsLoading.value = false;
  }
}

function handleAgentAction(agent: DistributionAgent, action: string) {
  if (action === 'customers') {
    return void router.push({ path: '/admin/distribution/customers', query: { agent_id: String(agent.id) } })
  }
  if (action === 'view') return openManageDialog(agent)
  if (action === 'rewards') {
    managedAgent.value = agent
    return void openRewardRuleDialog()
  }
  if (action === 'rate') {
    managedAgent.value = agent
    return openRateDialog()
  }
  if (action === 'permissions') {
    managedAgent.value = agent
    return openPromotionStatsPermissionDialog()
  }
  if (action === 'suspend') {
    managedAgent.value = agent
    return openStatusDialog('suspended')
  }
  if (action === 'activate') {
    managedAgent.value = agent
    return openStatusDialog('active')
  }
  if (action === 'revoke') {
    managedAgent.value = agent
    return openStatusDialog('revoked')
  }
}

function clearAgentFilters() {
  search.value = ''
  depth.value = ''
  status.value = ''
}

async function openRewardRuleDialog() {
  if (!managedAgent.value || managedAgent.value.depth !== 1) return;
  try {
    const rule = await getAgentRewardRule(managedAgent.value.id);
    rewardRule.value = { registration_enabled: rule.registration_enabled, registration_reward_cny: rule.registration_reward_cny, recharge_enabled: rule.recharge_enabled, recharge_threshold_cny: rule.recharge_threshold_cny, recharge_reward_cny: rule.recharge_reward_cny };
    rewardRuleDialog.value = true;
  } catch (error) {
    app.showError(extractApiErrorMessage(error, t('common.distributionRewards.loadTotalFailed')));
  }
}
async function saveRewardRule() {
  if (!managedAgent.value) return;
  rewardRuleSaving.value = true;
  try {
    await updateAgentRewardRule(managedAgent.value.id, rewardRule.value);
    app.showSuccess(t('common.distributionRewards.savedTotal'));
    rewardRuleDialog.value = false;
  } catch (error) {
    app.showError(extractI18nErrorMessage(error, t, "common.errors", t('common.distributionRewards.saveTotalFailed')));
  } finally { rewardRuleSaving.value = false; }
}

function closeManageDialog() {
  manageDialog.value = false;
  managedAgent.value = null;
  manageTab.value = 'overview';
  agentEvents.value = [];
}

function agentEventTitle(event: DistributionAgentEvent) {
  if (event.event_type === "created") return t('admin.distribution.agents.createdEvent');
  if (event.event_type === "permission_changed")
    return event.new_status?.startsWith("promotion_stats:")
      ? t(event.new_status.endsWith("true") ? 'admin.distribution.agents.statsGrantedEvent' : 'admin.distribution.agents.statsRevokedEvent')
      : t(event.new_status === "true" ? 'admin.distribution.agents.recruitmentGrantedEvent' : 'admin.distribution.agents.recruitmentRevokedEvent');
  if (event.event_type === "status_changed")
    return `${statusText(event.old_status || "")} → ${statusText(event.new_status || "")}`;
  const before =
    event.old_effective_rate_bps == null
      ? "-"
      : formatPercent(event.old_effective_rate_bps);
  const after =
    event.new_effective_rate_bps == null
      ? "-"
      : formatPercent(event.new_effective_rate_bps);
  return t('admin.distribution.agents.rateEvent', { before, after });
}

function openRecruitmentPermissionDialog() {
  recruitmentPermissionReason.value = "";
  recruitmentPermissionDialog.value = true;
}

async function saveRecruitmentPermission() {
  if (!managedAgent.value || !recruitmentPermissionReason.value.trim()) return;
  recruitmentPermissionSaving.value = true;
  try {
    const enabled = !managedAgent.value.can_recruit_subagents;
    await updateAgentRecruitmentPermission(managedAgent.value.id, {
      enabled,
      reason: recruitmentPermissionReason.value.trim(),
    });
    app.showSuccess(t(enabled ? 'admin.distribution.agents.recruitmentGranted' : 'admin.distribution.agents.recruitmentRevoked'));
    recruitmentPermissionDialog.value = false;
    manageDialog.value = false;
    await load();
  } catch (error) {
    app.showError(extractApiErrorMessage(error, t('admin.distribution.agents.recruitmentUpdateFailed')));
  } finally {
    recruitmentPermissionSaving.value = false;
  }
}

function openPromotionStatsPermissionDialog() {
  promotionStatsPermissionReason.value = "";
  promotionStatsPermissionDialog.value = true;
}

async function savePromotionStatsPermission() {
  if (!managedAgent.value || !promotionStatsPermissionReason.value.trim()) return;
  promotionStatsPermissionSaving.value = true;
  try {
    const enabled = !managedAgent.value.can_view_promotion_stats;
    await updateAgentPromotionStatsPermission(managedAgent.value.id, { enabled, reason: promotionStatsPermissionReason.value.trim() });
    app.showSuccess(t(enabled ? 'admin.distribution.agents.statsGranted' : 'admin.distribution.agents.statsRevoked'));
    promotionStatsPermissionDialog.value = false;
    manageDialog.value = false;
    await load();
  } catch (error) {
    app.showError(extractApiErrorMessage(error, t('admin.distribution.agents.statsUpdateFailed')));
  } finally {
    promotionStatsPermissionSaving.value = false;
  }
}

function openStatusDialog(nextStatus: "active" | "suspended" | "revoked") {
  pendingStatus.value = nextStatus;
  statusReason.value = "";
  statusDialog.value = true;
}

async function confirmStatusChange() {
  if (!managedAgent.value) return;
  if (statusReasonRequired.value && !statusReason.value.trim()) return;
  statusSaving.value = true;
  try {
    await updateAgentStatus(
      managedAgent.value.id,
      pendingStatus.value,
      statusReason.value.trim(),
    );
    app.showSuccess(t('admin.distribution.agents.statusUpdated'));
    statusDialog.value = false;
    closeManageDialog();
    await load();
  } catch (error) {
    app.showError(extractApiErrorMessage(error, t('admin.distribution.agents.statusUpdateFailed')));
  } finally {
    statusSaving.value = false;
  }
}

function openRateDialog() {
  if (!managedAgent.value) return;
  rateUseDefault.value = managedAgent.value.rate_override_bps == null;
  editRatePercent.value = (
    managedAgent.value.rate_override_bps == null
      ? managedAgent.value.effective_rate_bps
      : managedAgent.value.rate_override_bps
  ).toString();
  editRatePercent.value = (Number(editRatePercent.value) / 100).toFixed(2);
  rateReason.value = "";
  rateDialog.value = true;
}

async function saveRate() {
  if (!managedAgent.value || !canSaveRate.value) return;
  rateSaving.value = true;
  try {
    await updateAgentRate(managedAgent.value.id, {
      use_default: rateUseDefault.value,
      rate_override_bps: rateUseDefault.value
        ? undefined
        : Math.round(Number(editRatePercent.value) * 100),
      reason: rateReason.value.trim(),
    });
    app.showSuccess(t('admin.distribution.agents.rateUpdated'));
    rateDialog.value = false;
    manageDialog.value = false;
    await load();
  } catch (error) {
    app.showError(extractApiErrorMessage(error, t('admin.distribution.agents.rateUpdateFailed')));
  } finally {
    rateSaving.value = false;
  }
}

watch([search, status, depth], scheduleLoad);
function onListRangeChange() { pagination.value.page = 1; void load() }
async function loadManagedAnalytics(_range?: DistributionAnalyticsRangeValue) {
  if (!managedAgent.value || !manageDialog.value) return;
  analyticsLoading.value = true;
  try { agentAnalytics.value = await getAgentAnalytics(managedAgent.value.id, analyticsRangeParams(detailRange.value)); }
  catch (error) { app.showError(extractApiErrorMessage(error, t('admin.distribution.agents.loadAnalyticsFailed'))); }
  finally { analyticsLoading.value = false; }
}
onMounted(load);
</script>

<style scoped>
.input-suffix {
  position: absolute;
  right: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  color: var(--ui-text-muted);
  font-size: 0.875rem;
}
.input-prefix {
  position: absolute;
  left: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  color: var(--ui-text-muted);
  font-size: 0.875rem;
  pointer-events: none;
}
</style>
