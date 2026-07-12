import { nextTick, ref } from 'vue'

let dropdownMenuSequence = 0

type MenuFocusTarget = 'first' | 'last'

const menuItemSelector = [
  '[role="menuitem"]',
  '[role="menuitemcheckbox"]',
  '[role="menuitemradio"]'
].map((selector) => `${selector}:not([aria-disabled="true"]):not([disabled])`).join(',')

export function useDropdownMenu(idPrefix = 'dropdown-menu') {
  const instanceId = ++dropdownMenuSequence
  const open = ref(false)
  const triggerRef = ref<HTMLButtonElement | null>(null)
  const menuRef = ref<HTMLElement | null>(null)
  const triggerId = `${idPrefix}-trigger-${instanceId}`
  const menuId = `${idPrefix}-${instanceId}`

  function getMenuItems(): HTMLElement[] {
    if (!menuRef.value) return []
    return Array.from(menuRef.value.querySelectorAll<HTMLElement>(menuItemSelector))
  }

  function focusMenuItem(target: MenuFocusTarget): void {
    const items = getMenuItems()
    const item = target === 'first' ? items[0] : items[items.length - 1]
    item?.focus()
  }

  async function openMenu(focusTarget: MenuFocusTarget = 'first'): Promise<void> {
    open.value = true
    await nextTick()
    if (open.value) focusMenuItem(focusTarget)
  }

  async function closeMenu(restoreTriggerFocus = false): Promise<void> {
    open.value = false
    if (!restoreTriggerFocus) return
    await nextTick()
    triggerRef.value?.focus()
  }

  function toggleMenu(): void {
    if (open.value) {
      open.value = false
      return
    }
    void openMenu('first')
  }

  function handleTriggerKeydown(event: KeyboardEvent): void {
    if (event.key === 'ArrowDown') {
      event.preventDefault()
      void openMenu('first')
    } else if (event.key === 'ArrowUp') {
      event.preventDefault()
      void openMenu('last')
    } else if (event.key === 'Escape' && open.value) {
      event.preventDefault()
      void closeMenu(true)
    }
  }

  function handleMenuKeydown(event: KeyboardEvent): void {
    const items = getMenuItems()
    if (items.length === 0) return

    const target = event.target instanceof Element
      ? event.target.closest<HTMLElement>(menuItemSelector)
      : null
    const currentIndex = target ? items.indexOf(target) : -1

    switch (event.key) {
      case 'ArrowDown':
        event.preventDefault()
        items[(currentIndex + 1 + items.length) % items.length]?.focus()
        break
      case 'ArrowUp':
        event.preventDefault()
        items[(currentIndex - 1 + items.length) % items.length]?.focus()
        break
      case 'Home':
        event.preventDefault()
        items[0]?.focus()
        break
      case 'End':
        event.preventDefault()
        items[items.length - 1]?.focus()
        break
      case 'Escape':
        event.preventDefault()
        event.stopPropagation()
        void closeMenu(true)
        break
      case 'Tab':
        open.value = false
        break
    }
  }

  return {
    open,
    triggerRef,
    menuRef,
    triggerId,
    menuId,
    openMenu,
    closeMenu,
    toggleMenu,
    handleTriggerKeydown,
    handleMenuKeydown
  }
}
