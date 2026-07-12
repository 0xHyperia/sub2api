import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const testDir = dirname(fileURLToPath(import.meta.url))
const headerSource = readFileSync(resolve(testDir, '../AppHeader.vue'), 'utf8')
const layoutSource = readFileSync(resolve(testDir, '../AppLayout.vue'), 'utf8')

describe('admin shell structure', () => {
  it('keeps the route title visible at every viewport size', () => {
    expect(headerSource).toContain('<h1 class="truncate text-base')
    expect(headerSource).not.toContain('<div class="hidden lg:block">')
  })

  it('keeps personal balance and subscription noise out of the admin header', () => {
    expect(headerSource).toContain('v-if="user && !authStore.isAdmin"')
    expect(headerSource).toContain('isAdminContext ? \'/dashboard\' : \'/admin/dashboard\'')
    expect(headerSource).toContain("import { useTheme } from '@/composables/useTheme'")
  })

  it('uses the shared accessible keyboard model for the user menu', () => {
    expect(headerSource).toContain("useDropdownMenu('header-user-menu')")
    expect(headerSource).toContain('aria-haspopup="menu"')
    expect(headerSource).toContain('role="menu"')
    expect(headerSource).toContain('@keydown="handleMenuKeydown"')
    expect(headerSource).toContain('role="menuitem"')
  })

  it('uses a neutral shell and exposes a skip target', () => {
    expect(layoutSource).toContain('href="#app-main-content"')
    expect(layoutSource).toContain('id="app-main-content"')
    expect(layoutSource).not.toContain('bg-mesh-gradient')
  })
})
