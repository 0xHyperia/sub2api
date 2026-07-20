<template>
  <nav
    class="overflow-x-auto border-b border-outline"
    aria-label="代理中心导航"
  >
    <div class="flex min-w-max gap-1">
      <RouterLink
        v-for="item in visibleItems"
        :key="item.path"
        :to="item.path"
        class="distribution-tab"
        :class="route.path === item.path ? 'distribution-tab-active' : ''"
      >
        <Icon :name="item.icon" size="sm" />
        <span>{{ item.label }}</span>
      </RouterLink>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { computed, onMounted } from "vue";
import { useRoute } from "vue-router";
import Icon from "@/components/icons/Icon.vue";
import { useDistributionAccess } from "@/composables/useDistributionAccess";

const route = useRoute();
const { access, loadDistributionAccess } = useDistributionAccess();
const items = [
  { path: "/distribution", label: "概览", icon: "home" },
  { path: "/distribution/promotion", label: "推广", icon: "link" },
  { path: "/distribution/customers", label: "客户", icon: "users" },
  { path: "/distribution/commissions", label: "佣金", icon: "gift" },
  { path: "/distribution/team", label: "团队", icon: "userPlus", l1Only: true },
  { path: "/distribution/withdrawals", label: "结算", icon: "creditCard" },
] as const;

const visibleItems = computed(() =>
  items.filter((item) => !("l1Only" in item) || access.value?.depth === 1),
);

onMounted(() => {
  void loadDistributionAccess();
});
</script>

<style scoped>
.distribution-tab {
  display: flex;
  min-width: 54px;
  height: 42px;
  align-items: center;
  justify-content: center;
  gap: 2px;
  border-bottom: 2px solid transparent;
  padding: 0 4px;
  font-size: 14px;
  color: var(--ui-text-muted);
  white-space: nowrap;
}
.distribution-tab:hover {
  color: var(--ui-text);
}
.distribution-tab-active {
  border-bottom-color: var(--ui-primary);
  color: var(--ui-text);
  font-weight: 600;
}
@media (min-width: 640px) {
  .distribution-tab {
    gap: 6px;
    padding: 0 12px;
  }
}
</style>
