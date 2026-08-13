<template>
  <nav
    class="border-b border-outline"
    aria-label="代理中心导航"
  >
    <div
      class="distribution-tabs overflow-x-auto"
      :style="{ gridTemplateColumns: `repeat(${visibleItems.length}, minmax(0, 1fr))` }"
    >
      <RouterLink
        v-for="item in visibleItems"
        :key="item.path"
        :to="item.path"
        class="distribution-tab"
        :class="route.path === item.path ? 'distribution-tab-active' : ''"
        :aria-current="route.path === item.path ? 'page' : undefined"
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
  { path: "/distribution/team", label: "团队", icon: "userPlus", recruitmentOnly: true },
  { path: "/distribution/withdrawals", label: "结算", icon: "creditCard" },
] as const;

const visibleItems = computed(() =>
  items.filter((item) => !("recruitmentOnly" in item) || (access.value?.depth === 1 && access.value?.can_recruit_subagents)),
);

onMounted(() => {
  void loadDistributionAccess();
});
</script>

<style scoped>
.distribution-tab {
  display: flex;
  min-width: 0;
  height: 44px;
  align-items: center;
  justify-content: center;
  gap: 4px;
  border-bottom: 2px solid transparent;
  padding: 0 2px;
  font-size: 13px;
  color: var(--ui-text-muted);
  white-space: nowrap;
}
.distribution-tabs {
  display: grid;
  width: 100%;
  scrollbar-width: none;
  overscroll-behavior-inline: contain;
}
.distribution-tabs::-webkit-scrollbar { display: none; }
.distribution-tab:hover {
  color: var(--ui-text);
}
.distribution-tab-active {
  border-bottom-color: var(--ui-focus);
  color: var(--ui-text);
  font-weight: 600;
}
@media (min-width: 640px) {
  .distribution-tabs {
    display: flex;
    width: auto;
    gap: 4px;
    overflow-x: visible;
  }
  .distribution-tab {
    min-width: 58px;
    gap: 6px;
    padding: 0 12px;
    font-size: 14px;
  }
}
</style>
