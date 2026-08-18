<template>
  <div v-bind="rootAttrs" class="relative" ref="containerRef">
    <button
      v-bind="triggerAttrs"
      :id="triggerId"
      ref="triggerRef"
      type="button"
      @click="toggle"
      :disabled="disabled"
      :role="isSearchable ? undefined : 'combobox'"
      :aria-expanded="isOpen"
      aria-haspopup="listbox"
      :aria-controls="listboxId"
      :aria-activedescendant="!isSearchable ? focusedOptionId : undefined"
      :aria-label="triggerAriaLabel"
      :aria-describedby="ariaDescribedby"
      :class="[
        'select-trigger',
        isOpen && 'select-trigger-open',
        error && 'select-trigger-error',
        disabled && 'select-trigger-disabled',
        clearable && hasValue && !disabled && 'select-trigger-clearable',
      ]"
      @keydown="onTriggerKeyDown"
    >
      <span class="select-value">
        <slot name="selected" :option="selectedOption">
          {{ selectedLabel }}
        </slot>
      </span>
      <span class="select-icon">
        <Icon
          name="chevronDown"
          size="md"
          :class="['transition-transform duration-200', isOpen && 'rotate-180']"
        />
      </span>
    </button>
    <button
      v-if="clearable && hasValue && !disabled"
      type="button"
      class="select-clear"
      :aria-label="t('common.clear')"
      @click.stop="clearSelection"
      @mousedown.stop
    >
      <Icon name="x" size="sm" />
    </button>

    <!-- Teleport dropdown to body to escape stacking context -->
    <Teleport to="body">
      <Transition name="select-dropdown">
        <div
          v-if="isOpen"
          ref="dropdownRef"
          class="select-dropdown-portal"
          :class="[instanceId]"
          :style="dropdownStyle"
          @click.stop
          @mousedown.stop
          @keydown="onDropdownKeyDown"
        >
          <!-- Search input -->
          <div v-if="isSearchable" class="select-search">
            <Icon name="search" size="sm" class="text-foreground-subtle" />
            <input
              ref="searchInputRef"
              v-model="searchQuery"
              type="text"
              :placeholder="searchPlaceholderText"
              class="select-search-input"
              role="combobox"
              aria-autocomplete="list"
              aria-expanded="true"
              :aria-controls="listboxId"
              :aria-activedescendant="focusedOptionId"
              :aria-label="searchPlaceholderText"
              @click.stop
            />
          </div>

          <!-- Options list -->
          <div
            :id="listboxId"
            ref="optionsListRef"
            class="select-options"
            role="listbox"
            :aria-labelledby="triggerId"
          >
            <div
              v-for="(option, index) in filteredOptions"
              :key="`${typeof getOptionValue(option)}:${String(getOptionValue(option) ?? '')}`"
              :id="getOptionId(index)"
              :role="isGroupHeaderOption(option) ? 'presentation' : 'option'"
              :aria-selected="
                isGroupHeaderOption(option) ? undefined : isSelected(option)
              "
              :aria-disabled="
                isGroupHeaderOption(option)
                  ? undefined
                  : isOptionDisabled(option)
              "
              :tabindex="-1"
              @click.stop="isOptionFocusable(option) && selectOption(option)"
              @mouseenter="handleOptionMouseEnter(option, index)"
              :class="[
                'select-option',
                isGroupHeaderOption(option) && 'select-option-group',
                isSelected(option) && 'select-option-selected',
                isOptionDisabled(option) &&
                  !isGroupHeaderOption(option) &&
                  'select-option-disabled',
                focusedIndex === index &&
                  !isGroupHeaderOption(option) &&
                  'select-option-focused',
              ]"
            >
              <slot
                name="option"
                :option="option"
                :selected="isSelected(option)"
              >
                <Icon
                  v-if="option._creatable"
                  name="search"
                  size="sm"
                  class="flex-shrink-0 text-foreground-subtle"
                />
                <span
                  class="select-option-label"
                  :class="option._creatable && 'italic text-foreground-muted'"
                  >{{ getOptionLabel(option) }}</span
                >
                <Icon
                  v-if="isSelected(option)"
                  name="check"
                  size="sm"
                  class="text-foreground"
                  :stroke-width="2"
                />
              </slot>
            </div>

            <!-- Empty state -->
            <div v-if="filteredOptions.length === 0" class="select-empty">
              {{ props.loading ? t("common.loading") : emptyTextDisplay }}
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import {
  ref,
  computed,
  watch,
  onMounted,
  onUnmounted,
  nextTick,
  useAttrs,
} from "vue";
import { useI18n } from "vue-i18n";
import Icon from "@/components/icons/Icon.vue";

const { t } = useI18n();

defineOptions({ inheritAttrs: false });

const attrs = useAttrs();
const rootAttrs = computed(() =>
  Object.fromEntries(
    Object.entries(attrs).filter(
      ([key]) => key === "class" || key === "style" || key.startsWith("data-"),
    ),
  ),
);
const triggerAttrs = computed(() =>
  Object.fromEntries(
    Object.entries(attrs).filter(
      ([key]) => key !== "class" && key !== "style" && !key.startsWith("data-"),
    ),
  ),
);

// Instance ID for unique click-outside detection
const instanceId = `select-${Math.random().toString(36).substring(2, 9)}`;
const triggerId = computed(() => {
  const id = props.id ?? attrs.id;
  return typeof id === "string" && id ? id : `${instanceId}-trigger`;
});
const listboxId = `${instanceId}-listbox`;

export interface SelectOption {
  value: string | number | boolean | null;
  label: string;
  disabled?: boolean;
  [key: string]: unknown;
}

interface Props {
  modelValue: string | number | boolean | null | undefined;
  options: SelectOption[] | Array<Record<string, unknown>>;
  placeholder?: string;
  disabled?: boolean;
  error?: boolean;
  searchable?: boolean | "auto";
  searchPlaceholder?: string;
  emptyText?: string;
  valueKey?: string;
  labelKey?: string;
  creatable?: boolean;
  creatablePrefix?: string;
  clearable?: boolean;
  id?: string;
  ariaLabel?: string;
  ariaDescribedby?: string;
  /** 远程搜索模式：输入不在本地过滤 options，而是防抖后 emit('search', query)，由父组件请求数据更新 options */
  remote?: boolean;
  /** 远程搜索模式下的加载态：options 为空时下拉显示 loading 文案 */
  loading?: boolean;
}

interface Emits {
  (e: "update:modelValue", value: string | number | boolean | null): void;
  (
    e: "change",
    value: string | number | boolean | null,
    option: SelectOption | null,
  ): void;
  (e: "search", query: string): void;
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  error: false,
  searchable: "auto",
  creatable: false,
  creatablePrefix: "",
  clearable: false,
  valueKey: "value",
  labelKey: "label",
  remote: false,
  loading: false,
});

const emit = defineEmits<Emits>();

const isOpen = ref(false);
const searchQuery = ref("");
const focusedIndex = ref(-1);
const containerRef = ref<HTMLElement | null>(null);
const triggerRef = ref<HTMLButtonElement | null>(null);
const searchInputRef = ref<HTMLInputElement | null>(null);
const dropdownRef = ref<HTMLElement | null>(null);
const optionsListRef = ref<HTMLElement | null>(null);
const dropdownPosition = ref<"bottom" | "top">("bottom");
const triggerRect = ref<DOMRect | null>(null);
let typeaheadBuffer = "";
let typeaheadTimer: number | null = null;

// i18n placeholders
const placeholderText = computed(
  () => props.placeholder ?? t("common.selectOption"),
);
const searchPlaceholderText = computed(
  () => props.searchPlaceholder ?? t("common.searchPlaceholder"),
);
const emptyTextDisplay = computed(
  () => props.emptyText ?? t("common.noOptionsFound"),
);

// 远程搜索的防抖间隔（对齐 OpenAIFastPolicyUserSelector 的 300ms 惯例）。
const REMOTE_SEARCH_DEBOUNCE_MS = 300;
let remoteSearchTimer: ReturnType<typeof setTimeout> | null = null;

const isSearchable = computed(() => {
  // 远程搜索模式始终显示搜索框（选项只是服务端结果的一页）。
  if (props.remote) return true;
  if (props.searchable === "auto") return props.options.length > 5;
  return props.searchable;
});

// Computed style for teleported dropdown
const dropdownStyle = computed(() => {
  if (!triggerRect.value) return {};

  const rect = triggerRect.value;
  const viewportPadding = 8;
  const dropdownMinimumWidth = 200;
  const viewportRight = Math.max(
    viewportPadding,
    window.innerWidth - viewportPadding,
  );
  const left = Math.min(Math.max(viewportPadding, rect.left), viewportRight);
  const availableWidth = Math.max(0, viewportRight - left);
  const preferredMinWidth = Math.max(dropdownMinimumWidth, rect.width);
  const minWidth = Math.min(preferredMinWidth, availableWidth);
  const style: Record<string, string> = {
    position: "fixed",
    left: `${left}px`,
    minWidth: `${minWidth}px`,
    maxWidth: `${availableWidth}px`,
    zIndex: "100000020",
  };

  if (dropdownPosition.value === "top") {
    style.bottom = `${window.innerHeight - rect.top + 4}px`;
  } else {
    style.top = `${rect.bottom + 4}px`;
  }

  return style;
});

const getOptionValue = (option: any): any => {
  if (typeof option === "object" && option !== null) {
    return option[props.valueKey];
  }
  return option;
};

const getOptionLabel = (option: any): string => {
  if (typeof option === "object" && option !== null) {
    return String(option[props.labelKey] ?? "");
  }
  return String(option ?? "");
};

const isOptionDisabled = (option: any): boolean => {
  if (typeof option === "object" && option !== null) {
    return !!option.disabled;
  }
  return false;
};

const isGroupHeaderOption = (option: any): boolean => {
  if (typeof option === "object" && option !== null) {
    return option.kind === "group";
  }
  return false;
};

const isOptionFocusable = (option: any): boolean => {
  return !isOptionDisabled(option) && !isGroupHeaderOption(option);
};

const selectedOption = computed(() => {
  return (
    props.options.find((opt) => getOptionValue(opt) === props.modelValue) ||
    null
  );
});

const selectedLabel = computed(() => {
  if (selectedOption.value) {
    return getOptionLabel(selectedOption.value);
  }
  // In creatable mode, show the raw value if no matching option
  if (props.creatable && props.modelValue) {
    return String(props.modelValue);
  }
  return placeholderText.value;
});

const hasValue = computed(
  () =>
    props.modelValue !== null &&
    props.modelValue !== undefined &&
    props.modelValue !== "",
);

const triggerAriaLabel = computed(() => {
  if (attrs["aria-labelledby"]) return undefined;
  if (props.ariaLabel) return props.ariaLabel;
  if (typeof attrs["aria-label"] === "string" && attrs["aria-label"])
    return attrs["aria-label"];
  return hasValue.value
    ? `${placeholderText.value}: ${selectedLabel.value}`
    : placeholderText.value;
});

const filteredOptions = computed(() => {
  let opts = props.options as any[];
  // 远程搜索模式不在本地过滤（选项即服务端搜索结果的一页）。
  if (isSearchable.value && searchQuery.value && !props.remote) {
    const query = searchQuery.value.toLowerCase();
    opts = opts.filter((opt) => {
      // Match label
      if (getOptionLabel(opt).toLowerCase().includes(query)) return true;
      // Also match description if present
      if (
        opt.description &&
        String(opt.description).toLowerCase().includes(query)
      )
        return true;
      return false;
    });
    // In creatable mode, always prepend a fuzzy search option
    if (props.creatable && searchQuery.value.trim()) {
      const trimmed = searchQuery.value.trim();
      const prefix = props.creatablePrefix || t("common.search");
      opts = [
        {
          [props.valueKey]: trimmed,
          [props.labelKey]: `${prefix} "${trimmed}"`,
          _creatable: true,
        },
        ...opts,
      ];
    }
  }
  return opts;
});

const getOptionId = (index: number): string => `${instanceId}-option-${index}`;

const focusedOptionId = computed(() => {
  if (
    !isOpen.value ||
    focusedIndex.value < 0 ||
    focusedIndex.value >= filteredOptions.value.length
  ) {
    return undefined;
  }
  const option = filteredOptions.value[focusedIndex.value];
  return isOptionFocusable(option)
    ? getOptionId(focusedIndex.value)
    : undefined;
});

const isSelected = (option: any): boolean => {
  return getOptionValue(option) === props.modelValue;
};

const findNextEnabledIndex = (startIndex: number): number => {
  const opts = filteredOptions.value;
  if (opts.length === 0) return -1;
  for (let offset = 0; offset < opts.length; offset++) {
    const idx = (startIndex + offset) % opts.length;
    if (isOptionFocusable(opts[idx])) return idx;
  }
  return -1;
};

const findPrevEnabledIndex = (startIndex: number): number => {
  const opts = filteredOptions.value;
  if (opts.length === 0) return -1;
  for (let offset = 0; offset < opts.length; offset++) {
    const idx = (startIndex - offset + opts.length) % opts.length;
    if (isOptionFocusable(opts[idx])) return idx;
  }
  return -1;
};

const handleOptionMouseEnter = (option: any, index: number) => {
  if (isOptionDisabled(option) || isGroupHeaderOption(option)) return;
  focusedIndex.value = index;
};

// Update trigger rect periodically while open to follow scroll/resize
const updateTriggerRect = () => {
  if (containerRef.value) {
    triggerRect.value = containerRef.value.getBoundingClientRect();
  }
};

const calculateDropdownPosition = () => {
  if (!containerRef.value) return;
  updateTriggerRect();

  nextTick(() => {
    if (!dropdownRef.value || !triggerRect.value) return;
    const dropdownHeight = dropdownRef.value.offsetHeight || 240;
    const spaceBelow = window.innerHeight - triggerRect.value.bottom;
    const spaceAbove = triggerRect.value.top;

    if (spaceBelow < dropdownHeight && spaceAbove > dropdownHeight) {
      dropdownPosition.value = "top";
    } else {
      dropdownPosition.value = "bottom";
    }
  });
};

const toggle = () => {
  if (props.disabled) return;
  isOpen.value = !isOpen.value;
};

watch(isOpen, (open) => {
  if (open) {
    calculateDropdownPosition();
    // Reset focused index to current selection or first item
    if (filteredOptions.value.length === 0) {
      focusedIndex.value = -1;
    } else {
      const selectedIdx = filteredOptions.value.findIndex(isSelected);
      const initialIdx = selectedIdx >= 0 ? selectedIdx : 0;
      focusedIndex.value = !isOptionFocusable(filteredOptions.value[initialIdx])
        ? findNextEnabledIndex(initialIdx + 1)
        : initialIdx;
    }

    if (isSearchable.value) {
      nextTick(() => searchInputRef.value?.focus());
    }
    // Add scroll listener to update position
    window.addEventListener("scroll", updateTriggerRect, {
      capture: true,
      passive: true,
    });
    window.addEventListener("resize", calculateDropdownPosition);
  } else {
    searchQuery.value = "";
    focusedIndex.value = -1;
    // 关闭时取消仍在排队的远程搜索（避免关闭后尾随 emit 一次 search(''))。
    if (remoteSearchTimer) {
      clearTimeout(remoteSearchTimer);
      remoteSearchTimer = null;
    }
    window.removeEventListener("scroll", updateTriggerRect, { capture: true });
    window.removeEventListener("resize", calculateDropdownPosition);
  }
});

watch(filteredOptions, () => {
  if (!isOpen.value) return;
  focusedIndex.value = findNextEnabledIndex(0);
  if (focusedIndex.value >= 0) scrollToFocused();
});

// 远程搜索：输入防抖后交给父组件请求（!isOpen 抑制关闭重置 searchQuery 触发的空 query）。
watch(searchQuery, (query) => {
  if (!props.remote || !isOpen.value) return;
  if (remoteSearchTimer) clearTimeout(remoteSearchTimer);
  remoteSearchTimer = setTimeout(() => {
    remoteSearchTimer = null;
    emit("search", query.trim());
  }, REMOTE_SEARCH_DEBOUNCE_MS);
});

const selectOption = (option: any) => {
  if (!isOptionFocusable(option)) return;
  const value = getOptionValue(option) ?? null;
  emit("update:modelValue", value);
  emit("change", value, option);
  isOpen.value = false;
  triggerRef.value?.focus();
};

const clearSelection = () => {
  if (props.disabled) return;
  emit("update:modelValue", null);
  emit("change", null, null);
};

const focusFirstOption = () => {
  focusedIndex.value = findNextEnabledIndex(0);
  if (focusedIndex.value >= 0) scrollToFocused();
};

const focusLastOption = () => {
  focusedIndex.value = findPrevEnabledIndex(filteredOptions.value.length - 1);
  if (focusedIndex.value >= 0) scrollToFocused();
};

const moveFocusedOption = (direction: 1 | -1) => {
  focusedIndex.value =
    direction === 1
      ? findNextEnabledIndex(focusedIndex.value + 1)
      : findPrevEnabledIndex(focusedIndex.value - 1);
  if (focusedIndex.value >= 0) scrollToFocused();
};

const selectFocusedOption = () => {
  if (
    focusedIndex.value < 0 ||
    focusedIndex.value >= filteredOptions.value.length
  )
    return;
  selectOption(filteredOptions.value[focusedIndex.value]);
};

const handleTypeahead = (key: string) => {
  if (typeaheadTimer) window.clearTimeout(typeaheadTimer);
  typeaheadBuffer += key.toLocaleLowerCase();
  typeaheadTimer = window.setTimeout(() => {
    typeaheadBuffer = "";
    typeaheadTimer = null;
  }, 500);

  const options = filteredOptions.value;
  if (options.length === 0) return;
  const start = focusedIndex.value >= 0 ? focusedIndex.value + 1 : 0;
  for (let offset = 0; offset < options.length; offset += 1) {
    const index = (start + offset) % options.length;
    const option = options[index];
    if (!isOptionFocusable(option)) continue;
    if (
      getOptionLabel(option).toLocaleLowerCase().startsWith(typeaheadBuffer)
    ) {
      focusedIndex.value = index;
      scrollToFocused();
      return;
    }
  }
};

// Select-only combobox keyboard interaction keeps focus on the trigger.
const onTriggerKeyDown = (event: KeyboardEvent) => {
  if (props.disabled) return;

  switch (event.key) {
    case "ArrowDown":
      event.preventDefault();
      if (!isOpen.value) isOpen.value = true;
      else moveFocusedOption(1);
      return;
    case "ArrowUp":
      event.preventDefault();
      if (!isOpen.value) isOpen.value = true;
      else moveFocusedOption(-1);
      return;
    case "Home":
      if (!isOpen.value) return;
      event.preventDefault();
      focusFirstOption();
      return;
    case "End":
      if (!isOpen.value) return;
      event.preventDefault();
      focusLastOption();
      return;
    case "Enter":
    case " ":
      event.preventDefault();
      if (!isOpen.value) isOpen.value = true;
      else selectFocusedOption();
      return;
    case "Escape":
      if (!isOpen.value) return;
      event.preventDefault();
      isOpen.value = false;
      return;
    case "Tab":
      isOpen.value = false;
      return;
    case "Backspace":
    case "Delete":
      if (!isOpen.value && props.clearable && hasValue.value) {
        event.preventDefault();
        clearSelection();
      }
      return;
    default:
      if (
        !isSearchable.value &&
        event.key.length === 1 &&
        !event.ctrlKey &&
        !event.metaKey &&
        !event.altKey
      ) {
        event.preventDefault();
        if (!isOpen.value) isOpen.value = true;
        nextTick(() => handleTypeahead(event.key));
      }
  }
};

const onDropdownKeyDown = (e: KeyboardEvent) => {
  switch (e.key) {
    case "ArrowDown":
      e.preventDefault();
      moveFocusedOption(1);
      break;
    case "ArrowUp":
      e.preventDefault();
      moveFocusedOption(-1);
      break;
    case "Home":
      e.preventDefault();
      focusFirstOption();
      break;
    case "End":
      e.preventDefault();
      focusLastOption();
      break;
    case "Enter":
      e.preventDefault();
      selectFocusedOption();
      break;
    case "Escape":
      e.preventDefault();
      isOpen.value = false;
      triggerRef.value?.focus();
      break;
    case "Tab":
      isOpen.value = false;
      break;
  }
};

const scrollToFocused = () => {
  nextTick(() => {
    const list = optionsListRef.value;
    if (!list) return;
    const focusedEl = list.children[focusedIndex.value] as HTMLElement;
    if (!focusedEl) return;

    if (focusedEl.offsetTop < list.scrollTop) {
      list.scrollTop = focusedEl.offsetTop;
    } else if (
      focusedEl.offsetTop + focusedEl.offsetHeight >
      list.scrollTop + list.offsetHeight
    ) {
      list.scrollTop =
        focusedEl.offsetTop + focusedEl.offsetHeight - list.offsetHeight;
    }
  });
};

const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as HTMLElement;
  // Check if click is inside THIS specific instance's dropdown or trigger
  const isInDropdown = !!target.closest(`.${instanceId}`);
  const isInTrigger = containerRef.value?.contains(target);

  if (!isInDropdown && !isInTrigger && isOpen.value) {
    isOpen.value = false;
  }
};

onMounted(() => {
  document.addEventListener("click", handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener("click", handleClickOutside);
  window.removeEventListener("scroll", updateTriggerRect, { capture: true });
  window.removeEventListener("resize", calculateDropdownPosition);
  if (typeaheadTimer) window.clearTimeout(typeaheadTimer);
  if (remoteSearchTimer) {
    clearTimeout(remoteSearchTimer);
    remoteSearchTimer = null;
  }
});
</script>

<style scoped>
.select-trigger {
  display: flex;
  width: 100%;
  min-height: var(--control-height, 40px);
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 7px 11px;
  border: 1px solid var(--ui-border, #dbe3ee);
  border-radius: 6px;
  background: var(--ui-surface, #fff);
  color: var(--ui-text, #0f172a);
  cursor: pointer;
  font-size: 13px;
  outline: none;
  transition:
    border-color 150ms ease,
    box-shadow 150ms ease,
    background-color 150ms ease;
}

.select-trigger:hover {
  border-color: var(--ui-border-strong, #c7d3e2);
}

.select-trigger:focus-visible {
  border-color: var(--ui-focus, #2563eb);
  box-shadow: 0 0 0 3px
    color-mix(in srgb, var(--ui-focus, #2563eb) 16%, transparent);
}

.select-trigger-open {
  border-color: var(--ui-focus, #2563eb);
  box-shadow: 0 0 0 3px
    color-mix(in srgb, var(--ui-focus, #2563eb) 16%, transparent);
}

.select-trigger-error {
  border-color: var(--ui-danger, #dc2626);
}

.select-trigger-disabled {
  background: var(--ui-surface-subtle, #f4f7fb);
  cursor: not-allowed;
  opacity: 0.56;
}

.select-trigger-clearable {
  padding-right: 45px;
}

.select-value {
  @apply flex-1 truncate text-left;
}

.select-icon {
  flex-shrink: 0;
  color: var(--ui-text-subtle, #8793a3);
}

.select-clear {
  position: absolute;
  top: 50%;
  right: 31px;
  z-index: 1;
  display: inline-flex;
  width: 28px;
  height: 28px;
  align-items: center;
  justify-content: center;
  padding: 0;
  border: 0;
  border-radius: 4px;
  background: transparent;
  color: var(--ui-text-subtle, #8793a3);
  cursor: pointer;
  transform: translateY(-50%);
  transition:
    background-color 150ms ease,
    color 150ms ease;
}

.select-clear:hover {
  background: var(--ui-surface-subtle, #f4f7fb);
  color: var(--ui-text, #0f172a);
}

.select-clear:focus-visible {
  outline: 2px solid var(--ui-focus, #2563eb);
  outline-offset: 1px;
}

@media (max-width: 639px) {
  .select-trigger {
    min-height: var(--control-height-lg, 44px);
  }
}
</style>

<style>
.select-dropdown-portal {
  width: max-content;
  min-width: 200px;
  overflow: hidden;
  padding: 4px;
  border: 1px solid var(--ui-border, #dbe3ee);
  border-radius: 8px;
  background: var(--ui-surface-raised, #fff);
  box-shadow: var(--ui-shadow-floating, 0 14px 34px rgba(15, 23, 42, 0.14));
  pointer-events: auto !important;
}

.select-dropdown-portal .select-search {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 0 4px;
  padding: 7px 8px 8px;
  border-bottom: 1px solid var(--ui-border, #dbe3ee);
}

.select-dropdown-portal .select-search-input {
  min-width: 0;
  flex: 1;
  border: 0;
  background: transparent;
  color: var(--ui-text, #0f172a);
  font-size: 13px;
  outline: none;
}

.select-dropdown-portal .select-search-input::placeholder {
  color: var(--ui-text-subtle, #8793a3);
}

.select-dropdown-portal .select-options {
  max-height: 20rem;
  overflow-y: auto;
  outline: none;
}

.select-dropdown-portal .select-option {
  display: flex;
  min-height: 36px;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 7px 10px;
  border-radius: 6px;
  color: var(--ui-text-muted, #667085);
  cursor: pointer;
  font-size: 13px;
  transition:
    background-color 120ms ease,
    color 120ms ease;
  pointer-events: auto !important;
}

.select-dropdown-portal .select-option:hover {
  background: var(--ui-surface-subtle, #f4f7fb);
  color: var(--ui-text, #0f172a);
}

.select-dropdown-portal .select-option-selected {
  background: var(--ui-surface-subtle, #f4f7fb);
  color: var(--ui-text, #0f172a);
  font-weight: 650;
}

.select-dropdown-portal .select-option-focused {
  background: color-mix(
    in srgb,
    var(--ui-focus, #2563eb) 10%,
    var(--ui-surface, #fff)
  );
  color: var(--ui-text, #0f172a);
}

.select-dropdown-portal .select-option-disabled {
  @apply cursor-not-allowed opacity-40;
}

.select-dropdown-portal .select-option-group {
  min-height: 30px;
  background: transparent;
  color: var(--ui-text-subtle, #8793a3);
  cursor: default;
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.select-dropdown-portal .select-option-group:hover {
  background: transparent;
}

.select-dropdown-portal .select-option-label {
  @apply flex-1 min-w-0 truncate text-left;
}

.select-dropdown-portal .select-empty {
  padding: 24px 12px;
  color: var(--ui-text-muted, #667085);
  text-align: center;
  font-size: 13px;
}

.select-dropdown-enter-active,
.select-dropdown-leave-active {
  transition:
    opacity 140ms ease,
    transform 140ms ease;
}

.select-dropdown-enter-from,
.select-dropdown-leave-to {
  opacity: 0;
  transform: translateY(-4px) scale(0.99);
}

@media (max-width: 639px) {
  .select-dropdown-portal .select-option {
    min-height: 44px;
  }
}
</style>
