import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../AppSidebar.vue')
const componentSource = readFileSync(componentPath, 'utf8')
const stylePath = resolve(dirname(fileURLToPath(import.meta.url)), '../../../style.css')
const styleSource = readFileSync(stylePath, 'utf8')

describe('AppSidebar custom SVG styles', () => {
  it('does not override uploaded SVG fill or stroke colors', () => {
    expect(componentSource).toContain('.sidebar-svg-icon {')
    expect(componentSource).toContain('color: currentColor;')
    expect(componentSource).toContain('display: block;')
    expect(componentSource).not.toContain('stroke: currentColor;')
    expect(componentSource).not.toContain('fill: none;')
  })
})

describe('AppSidebar scroll position persistence', () => {
  it('binds a template ref to the sidebar nav element', () => {
    expect(componentSource).toContain('ref="sidebarNavRef"')
    expect(componentSource).toContain('sidebar-nav')
  })

  it('declares sidebarNavRef in script setup', () => {
    expect(componentSource).toContain("const sidebarNavRef = ref<HTMLElement | null>(null)")
  })

  it('saves scroll position on beforeUnmount', () => {
    expect(componentSource).toContain('onBeforeUnmount')
    expect(componentSource).toContain('appStore.sidebarScrollTop')
    expect(componentSource).toContain('sidebarNavRef.value.scrollTop')
  })

  it('restores scroll position on mount', () => {
    expect(componentSource).toContain('onMounted')
    expect(componentSource).toContain('appStore.sidebarScrollTop')
    expect(componentSource).toContain('nextTick')
  })
})

describe('AppSidebar header styles', () => {
  it('does not clip the version badge dropdown', () => {
    const sidebarHeaderBlockMatch = styleSource.match(/\.sidebar-header\s*\{[\s\S]*?\n {2}\}/)
    const sidebarBrandBlockMatch = componentSource.match(/\.sidebar-brand\s*\{[\s\S]*?\n\}/)

    expect(sidebarHeaderBlockMatch).not.toBeNull()
    expect(sidebarBrandBlockMatch).not.toBeNull()
    expect(sidebarHeaderBlockMatch?.[0]).not.toContain('@apply overflow-hidden;')
    expect(sidebarBrandBlockMatch?.[0]).not.toContain('overflow: hidden;')
  })
})

describe('AppSidebar admin shell behavior', () => {
  it('mounts the configurable contact panel in the authenticated shell footer', () => {
    expect(componentSource).toContain("import ContactUsPanel from '@/components/layout/ContactUsPanel.vue'")
    expect(componentSource).toContain(':settings="appStore.cachedPublicSettings"')
    expect(componentSource).toContain(':collapsed="sidebarCollapsed"')
  })

  it('uses a quiet full-surface selection and avoids double-highlighting expanded groups', () => {
    expect(componentSource).toContain("'sidebar-group-active': isGroupActive(item) && !sidebarCollapsed")
    expect(componentSource).toContain("'sidebar-link-active': isGroupActive(item) && sidebarCollapsed")
    expect(componentSource).toContain('@apply bg-info-subtle font-semibold text-foreground hover:bg-info-subtle;')
    expect(componentSource).toContain('inset 0 0 0 1px rgb(var(--color-info) / 0.16)')
    expect(componentSource).not.toContain('inset 2px 0 0')
  })

  it('keeps collapsed navigation groups reachable by expanding the sidebar', () => {
    expect(componentSource).toContain("resolveGroupClickAction(sidebarCollapsed.value, item.expandOnly)")
    expect(componentSource).toContain("if (action === 'expand-sidebar')")
    expect(componentSource).toContain('appStore.setSidebarCollapsed(false)')
    expect(componentSource).toContain('expandedGroups.value.add(item.path)')
  })

  it('implements the mobile sidebar as an inert, dismissible drawer', () => {
    expect(componentSource).toContain(':inert="!isDesktopViewport && !mobileOpen"')
    expect(componentSource).toContain('class="drawer-backdrop"')
    expect(componentSource).toContain("event.key === 'Escape'")
    expect(componentSource).toContain("document.body.style.overflow = 'hidden'")
    expect(componentSource).toContain('previousActiveElement.focus()')
    expect(componentSource).toContain('@keydown="handleDrawerKeydown"')
  })

  it('organizes admin navigation by domain and removes the personal nav block', () => {
    for (const section of ['overview', 'resources', 'observability', 'commerce', 'communication', 'system']) {
      expect(componentSource).toContain(`id: '${section}'`)
    }
    expect(componentSource).not.toContain('personalNavItems')
  })

  it('keeps collapsed icon navigation named and removes the hidden brand link from focus', () => {
    expect(componentSource).toContain(':inert="sidebarCollapsed"')
    expect(componentSource.match(/:aria-label="sidebarCollapsed \? item\.label : undefined"/g)).toHaveLength(3)
    expect(componentSource).toContain('focus-visible:ring-offset-surface;')
    expect(componentSource).toContain('.sidebar-brand-title {')
    expect(componentSource).toContain('rounded-sm text-sm font-semibold')
  })
})
