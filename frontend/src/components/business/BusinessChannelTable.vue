<template>
  <section class="panel">
    <div class="panel-title">
      <h2>{{ t('admin.business.channelComparison') }}</h2>
      <p>{{ t('admin.business.channelComparisonHint') }}</p>
    </div>
    <div class="table-scroll">
      <table>
        <thead><tr><th v-for="column in columns" :key="column.key"><button type="button" @click="toggleSort(column.key)">{{ t(`admin.business.${column.label}`) }}<Icon :name="column.key===sortKey&&sortOrder==='desc'?'chevronDown':'chevronUp'" size="xs"/></button></th></tr></thead>
        <tbody>
          <tr v-for="row in sortedRows" :key="row.channel">
            <td><button type="button" class="channel-link" @click="$emit('select', row.channel)">{{ t(`admin.business.channel_${row.channel}`) }}<Icon name="chevronRight" size="xs"/></button></td><td>{{ row.new_users }}</td><td>{{ row.activated_users }}</td><td>{{ row.first_paid_users }}</td>
            <td>{{ percent(row.paying_users ? row.repurchase_users * 100 / row.paying_users : 0) }}</td><td>{{ percent(row.d7_retention_rate) }}</td><td>{{ cny(row.net_paid_cny) }}</td><td>{{ cny(row.arppu_cny) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { computed, ref } from 'vue'
import type { BusinessChannelMetric } from '@/api/businessAnalytics'
import Icon from '@/components/icons/Icon.vue'

const props=defineProps<{ rows: BusinessChannelMetric[]; compact?: boolean }>()
defineEmits<{ (event: 'select', channel: string): void }>()
const { t } = useI18n()
type SortKey='channel'|'new_users'|'activated_users'|'first_paid_users'|'repurchase_rate'|'d7_retention_rate'|'net_paid_cny'|'arppu_cny'
const columns: Array<{key:SortKey;label:string}> = [{key:'channel',label:'channel'},{key:'new_users',label:'newUsers'},{key:'activated_users',label:'activatedUsers'},{key:'first_paid_users',label:'firstPaidUsers'},{key:'repurchase_rate',label:'repurchaseRate'},{key:'d7_retention_rate',label:'d7Retention'},{key:'net_paid_cny',label:'netPaid'},{key:'arppu_cny',label:'arppu'}]
const sortKey=ref<SortKey>('net_paid_cny');const sortOrder=ref<'asc'|'desc'>('desc')
const valueFor=(row:BusinessChannelMetric,key:SortKey)=>key==='repurchase_rate'?(row.paying_users?row.repurchase_users/row.paying_users:0):row[key]
const sortedRows=computed(()=>[...props.rows].sort((a,b)=>{const av=valueFor(a,sortKey.value),bv=valueFor(b,sortKey.value);const comparison=typeof av==='string'?av.localeCompare(String(bv)):Number(av)-Number(bv);return sortOrder.value==='asc'?comparison:-comparison}))
function toggleSort(key:SortKey){if(sortKey.value===key)sortOrder.value=sortOrder.value==='desc'?'asc':'desc';else{sortKey.value=key;sortOrder.value=key==='channel'?'asc':'desc'}}
const percent = (value: number) => `${value.toFixed(1)}%`
const cny = (value: number) => `￥${new Intl.NumberFormat(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(value)}`
</script>

<style scoped>
.panel{min-width:0;max-width:100%;border:1px solid var(--ui-border);border-radius:8px;background:var(--ui-surface);padding:16px}.panel-title{margin-bottom:13px}.panel-title h2{font-size:14px;font-weight:650}.panel-title p{margin-top:3px;color:var(--ui-text-muted);font-size:12px}.table-scroll{width:100%;max-width:100%;overflow-x:auto}table{width:100%;min-width:720px;border-collapse:collapse;font-size:12px}th,td{padding:11px 12px;border-bottom:1px solid var(--ui-border);text-align:right;font-variant-numeric:tabular-nums}th:first-child,td:first-child{text-align:left}th{color:var(--ui-text-muted);font-weight:600}th button{display:inline-flex;align-items:center;justify-content:flex-end;gap:3px;color:inherit;font:inherit}th:first-child button{justify-content:flex-start}.channel-link{display:inline-flex;min-height:32px;align-items:center;gap:4px;border-radius:5px;color:var(--ui-focus);font-weight:600}.channel-link:focus-visible{outline:2px solid var(--ui-focus);outline-offset:1px}tbody tr:hover{background:var(--ui-surface-subtle)}@media(max-width:767px){.panel{padding:16px}}
</style>
