<template>
  <section class="analytics-panel">
    <div class="panel-heading">
      <div>
        <h2>{{ title }}</h2>
        <p>{{ subtitle }} · {{ granularityLabel }}</p>
      </div>
      <div
        class="segments"
        role="group"
        :aria-label="t('admin.business.metric')"
      >
        <button
          v-for="option in visibleOptions"
          :key="option.key"
          type="button"
          :aria-pressed="selected.includes(option.key)"
          @click="toggle(option.key)"
        >
          {{ option.label }}
        </button>
      </div>
    </div>
    <div class="totals">
      <div v-for="item in selectedTotals" :key="item.key">
        <span><i :style="{ background: item.color }" />{{ item.label }}</span
        ><strong>{{ formatValue(item.value, item.key) }}</strong
        ><small
          v-if="item.change !== null"
          :class="item.change >= 0 ? 'positive' : 'negative'"
          >{{
            t("admin.business.vsPrevious", {
              value: `${item.change >= 0 ? "+" : ""}${item.change.toFixed(1)}%`,
            })
          }}</small
        ><small v-else class="neutral">{{
          t("admin.business.noComparison")
        }}</small>
      </div>
    </div>
    <div
      v-if="chartData"
      ref="chartWrap"
      class="chart-wrap"
      tabindex="0"
      @keydown="onKeydown"
    >
      <Line ref="chartRef" :data="chartData" :options="chartOptions" />
    </div>
    <div v-else class="empty">{{ t("admin.business.noData") }}</div>
  </section>
</template>
<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { Line } from "vue-chartjs";
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler,
  type ChartOptions,
  type Chart,
} from "chart.js";
import { useI18n } from "vue-i18n";
import { useChartTheme } from "@/composables/useChartTheme";
import type { BusinessTrendPoint } from "@/api/businessAnalytics";
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler,
);
type Metric =
  | "net_paid_cny"
  | "gross_paid_cny"
  | "refunded_cny"
  | "new_users"
  | "activated_users"
  | "first_paid_users"
  | "paying_users"
  | "repurchase_users"
  | "active_users"
  | "consumed_revenue"
  | "supplier_cost";
const props = defineProps<{
  points: BusinessTrendPoint[];
  previousPoints?: BusinessTrendPoint[];
  title: string;
  subtitle: string;
  granularity: "hour" | "day" | "week" | "month";
  defaultMetrics: readonly Metric[];
  availableMetrics: readonly Metric[];
  totals?: Partial<Record<Metric, { value: number; previous: number }>>;
}>();
const { t } = useI18n();
const { chartTheme } = useChartTheme();
const selected = ref<Metric[]>([...props.defaultMetrics]);
const chartRef = ref<InstanceType<typeof Line> | null>(null);
const definitions = computed(
  () =>
    ({
      net_paid_cny: t("admin.business.netPaid"),
      gross_paid_cny: t("admin.business.metric_gross_paid"),
      refunded_cny: t("admin.business.metric_refunded"),
      new_users: t("admin.business.newUsers"),
      activated_users: t("admin.business.activatedUsers"),
      first_paid_users: t("admin.business.firstPaidUsers"),
      paying_users: t("admin.business.metric_paying_users"),
      repurchase_users: t("admin.business.metric_repurchase_users"),
      active_users: t("admin.business.activeUsers"),
      consumed_revenue: t("admin.business.consumedRevenue"),
      supplier_cost: t("admin.business.supplierCost"),
    }) as Record<Metric, string>,
);
const granularityLabel = computed(() =>
  t(`admin.business.granularity_${props.granularity}`),
);
const palette = computed(() => [
  chartTheme.value.info,
  chartTheme.value.success,
  chartTheme.value.warning,
  chartTheme.value.danger,
  chartTheme.value.brand,
]);
const visibleOptions = computed(() =>
  props.availableMetrics.map((key) => ({ key, label: definitions.value[key] })),
);
watch(
  () => props.defaultMetrics,
  (value) => {
    selected.value = [...value];
  },
);
function unit(key: Metric) {
  if (cnyMetrics.has(key)) return "cny";
  if (usdMetrics.has(key)) return "usd";
  return "count";
}
function toggle(key: Metric) {
  if (selected.value.includes(key)) {
    if (selected.value.length > 1)
      selected.value = selected.value.filter((x) => x !== key);
    return;
  }
  const compatible = selected.value.filter((item) => unit(item) === unit(key));
  selected.value = compatible.length < 3 ? [...compatible, key] : compatible;
}
const cnyMetrics = new Set<Metric>([
  "net_paid_cny",
  "gross_paid_cny",
  "refunded_cny",
]);
const usdMetrics = new Set<Metric>(["consumed_revenue", "supplier_cost"]);
const money = (key: Metric) => cnyMetrics.has(key) || usdMetrics.has(key);
function fixedCurrency(symbol: string, value: number) {
  return `${symbol}${new Intl.NumberFormat(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(value)}`;
}
function formatValue(value: number, key: Metric) {
  if (cnyMetrics.has(key)) return fixedCurrency("￥", value);
  if (usdMetrics.has(key)) return fixedCurrency("$", value);
  return new Intl.NumberFormat().format(Math.round(value));
}
function bucketLabel(value: string, full = false) {
  const d = new Date(value.includes("T") ? value : `${value}T00:00:00+08:00`);
  if (props.granularity === "hour")
    return new Intl.DateTimeFormat(
      undefined,
      full
        ? {
            year: "numeric",
            month: "2-digit",
            day: "2-digit",
            hour: "2-digit",
            minute: "2-digit",
          }
        : { hour: "2-digit", minute: "2-digit", hour12: false },
    ).format(d);
  if (props.granularity === "month")
    return new Intl.DateTimeFormat(undefined, {
      year: "numeric",
      month: "short",
    }).format(d);
  return new Intl.DateTimeFormat(
    undefined,
    full
      ? { year: "numeric", month: "short", day: "numeric" }
      : { month: "numeric", day: "numeric" },
  ).format(d);
}
const sum = (points: BusinessTrendPoint[], key: Metric) =>
  points.reduce((total, p) => total + Number(p[key] || 0), 0);
const selectedTotals = computed(() =>
  selected.value.map((key, index) => {
    const value = props.totals?.[key]?.value ?? sum(props.points, key),
      previous =
        props.totals?.[key]?.previous ?? sum(props.previousPoints || [], key);
    return {
      key,
      label: definitions.value[key],
      value,
      change: previous ? ((value - previous) / previous) * 100 : null,
      color: palette.value[index],
    };
  }),
);
const chartData = computed(() =>
  props.points.length
    ? {
        labels: props.points.map((p) => bucketLabel(p.bucket)),
        datasets: selected.value.flatMap((key, index) => [
          {
            label: definitions.value[key],
            data: props.points.map((p) => Number(p[key] || 0)),
            borderColor: palette.value[index],
            backgroundColor: "transparent",
            pointRadius: props.points.length < 4 ? 4 : 1.5,
            pointHoverRadius: 5,
            borderWidth: 2,
            tension: 0.25,
          },
          {
            label: `${definitions.value[key]} · ${t("admin.business.previous")}`,
            data: (props.previousPoints || []).map((p) => Number(p[key] || 0)),
            borderColor: palette.value[index],
            backgroundColor: "transparent",
            borderDash: [5, 4],
            pointRadius: 0,
            borderWidth: 1.5,
            tension: 0.25,
          },
        ]),
      }
    : null,
);
const chartOptions = computed<ChartOptions<"line">>(() => ({
  responsive: true,
  maintainAspectRatio: false,
  animation:
    typeof matchMedia !== "undefined" &&
    matchMedia("(prefers-reduced-motion: reduce)").matches
      ? false
      : { duration: 180 },
  interaction: { mode: "index", intersect: false },
  plugins: {
    legend: {
      display: true,
      labels: {
        color: chartTheme.value.foregroundMuted,
        usePointStyle: true,
        boxWidth: 8,
        font: { size: 11 },
        filter: (item) =>
          !String(item.text).includes(`· ${t("admin.business.previous")}`),
      },
    },
    tooltip: {
      callbacks: {
        title: (items) => {
          const point = props.points[items[0]?.dataIndex || 0];
          if (!point) return "";
          if (props.granularity === "hour") {
            const start = new Date(point.bucket),
              end = new Date(start.getTime() + 3599999);
            return `${bucketLabel(point.bucket, true)}–${new Intl.DateTimeFormat(undefined, { hour: "2-digit", minute: "2-digit" }).format(end)}`;
          }
          return bucketLabel(point.bucket, true);
        },
        label: (ctx) =>
          `${ctx.dataset.label}: ${formatValue(Number(ctx.raw || 0), selected.value[Math.floor(ctx.datasetIndex / 2)] || selected.value[0])}`,
        footer: (items) => {
          const index = items[0]?.dataIndex ?? 0;
          return selected.value.map((key) => {
            const current = Number(props.points[index]?.[key] || 0),
              previous = Number(props.previousPoints?.[index]?.[key] || 0),
              delta = current - previous,
              change = previous ? (delta * 100) / previous : null;
            const comparison = change === null
              ? t("admin.business.noComparison")
              : `${change >= 0 ? "+" : ""}${change.toFixed(1)}%`;
            return `${definitions.value[key]} ${t("admin.business.difference")}: ${formatValue(delta, key)} (${comparison})`;
          });
        },
      },
    },
  },
  scales: {
    x: {
      grid: { display: false },
      ticks: {
        color: chartTheme.value.foregroundSubtle,
        maxTicksLimit: props.granularity === "hour" ? 7 : 8,
        maxRotation: 0,
        font: { size: 10 },
      },
    },
    y: {
      beginAtZero: true,
      grid: { color: chartTheme.value.outline },
      ticks: {
        color: chartTheme.value.foregroundMuted,
        font: { size: 10 },
        callback: (value) => {
          const n = Number(value);
          return money(selected.value[0])
            ? new Intl.NumberFormat(undefined, {
                notation: "compact",
                maximumFractionDigits: 1,
              }).format(n)
            : new Intl.NumberFormat(undefined, { notation: "compact" }).format(
                n,
              );
        },
      },
    },
  },
}));
function onKeydown(event: KeyboardEvent) {
  const chart = (chartRef.value as unknown as { chart?: Chart<"line"> } | null)
    ?.chart;
  if (!["ArrowLeft", "ArrowRight", "Escape"].includes(event.key) || !chart)
    return;
  event.preventDefault();
  if (event.key === "Escape") {
    chart.tooltip?.setActiveElements([], { x: 0, y: 0 });
    chart.update();
    return;
  }
  const current =
    chart.tooltip?.getActiveElements()[0]?.index ??
    (event.key === "ArrowRight" ? -1 : props.points.length);
  const next = Math.max(
    0,
    Math.min(
      props.points.length - 1,
      current + (event.key === "ArrowRight" ? 1 : -1),
    ),
  );
  chart.tooltip?.setActiveElements([{ datasetIndex: 0, index: next }], {
    x: chart.scales.x.getPixelForValue(next),
    y: chart.scales.y.getPixelForValue(
      Number(props.points[next]?.[selected.value[0]] || 0),
    ),
  });
  chart.update();
}
</script>
<style scoped>
.analytics-panel {
  min-width: 0;
  border: 1px solid var(--ui-border);
  border-radius: 8px;
  background: var(--ui-surface);
  padding: 16px;
}
.panel-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}
.panel-heading h2 {
  font-size: 14px;
  font-weight: 650;
}
.panel-heading p {
  margin-top: 3px;
  color: var(--ui-text-muted);
  font-size: 12px;
}
.segments {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 3px;
}
.segments button {
  min-height: 32px;
  border: 1px solid transparent;
  border-radius: 6px;
  padding: 0 9px;
  color: var(--ui-text-muted);
  font-size: 11px;
}
.segments button[aria-pressed="true"] {
  border-color: var(--ui-border);
  background: var(--ui-surface-raised);
  color: var(--ui-text);
  font-weight: 650;
}
.totals {
  display: flex;
  flex-wrap: wrap;
  gap: 26px;
  margin-top: 15px;
}
.totals div {
  display: grid;
  grid-template-columns: auto auto;
  align-items: baseline;
  gap: 2px 10px;
}
.totals span {
  grid-column: 1/-1;
  color: var(--ui-text-muted);
  font-size: 11px;
}
.totals i {
  display: inline-block;
  width: 7px;
  height: 7px;
  margin-right: 5px;
  border-radius: 50%;
}
.totals strong {
  font-size: 19px;
  font-variant-numeric: tabular-nums;
}
.totals small {
  font-size: 10px;
}
.positive {
  color: rgb(var(--color-success-foreground));
}
.negative {
  color: rgb(var(--color-danger-foreground));
}
.neutral {
  color: var(--ui-text-subtle);
}
.chart-wrap {
  height: 280px;
  margin-top: 10px;
}
.chart-wrap:focus-visible {
  outline: 2px solid var(--ui-focus);
  outline-offset: 3px;
}
.empty {
  display: grid;
  min-height: 250px;
  place-items: center;
  color: var(--ui-text-muted);
  font-size: 13px;
}
@media (max-width: 767px) {
  .panel-heading {
    flex-direction: column;
  }
  .segments {
    width: 100%;
    justify-content: flex-start;
  }
  .segments button {
    min-height: 40px;
    flex: 1;
  }
  .totals {
    gap: 14px;
  }
  .chart-wrap {
    height: 250px;
  }
}
</style>
