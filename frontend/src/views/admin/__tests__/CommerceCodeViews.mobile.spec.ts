import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { parse } from 'vue/compiler-sfc'

function readView(name: 'RedeemView' | 'PromoCodesView') {
  const filename = resolve(process.cwd(), `src/views/admin/${name}.vue`)
  const source = readFileSync(filename, 'utf8')
  const { descriptor, errors } = parse(source, { filename })
  expect(errors).toEqual([])
  return { source, template: descriptor.template?.content || '' }
}

describe('admin commerce code mobile cards', () => {
  it('presents redeem-code decisions on mobile while retaining desktop selection and batch flows', () => {
    const { template } = readView('RedeemView')

    expect(template).toMatch(/data-mobile-layout="redeem-cards"[^>]*md:hidden/)
    expect(template).toMatch(/data-desktop-layout="redeem-table"[^>]*hidden[^>]*md:block/)
    expect(template).toContain('data-mobile-select-code')
    expect(template).toContain('@change="toggleSelectRow(row.id, $event)"')
    expect(template).toContain('@change="toggleSelectAllVisible($event)"')
    expect(template).toContain('row.user?.email')
    expect(template).toContain('row.expires_at')
    expect(template).toContain("row.type === 'balance'")
    expect(template).toContain('@click="handleDelete(row)"')
    expect(template).toContain('data-test="batch-update-open"')
    expect(template).toContain('@click="showGenerateDialog = true"')
  })

  it('makes promo creation primary and separates frequent from low-frequency card actions', () => {
    const { source, template } = readView('PromoCodesView')

    expect(template).toMatch(/data-mobile-layout="promo-cards"[^>]*md:hidden/)
    expect(template).toMatch(/data-desktop-layout="promo-table"[^>]*hidden[^>]*md:block/)
    expect(template).toContain("localText('适用于新注册用户', 'For new registrations')")
    expect(template).toContain('row.bonus_amount.toFixed(2)')
    expect(template).toContain('row.used_count')
    expect(template).toContain('data-mobile-action="copy-register-link"')
    expect(template).toContain('data-mobile-action="edit"')

    const more = template.match(/<details class="group relative shrink-0">([\s\S]*?)<\/details>/)?.[1] || ''
    expect(more).toContain('@click="handleViewUsages(row)"')
    expect(more).toContain('@click="handleDelete(row)"')
    expect(source).toMatch(/showCreateDialog = true" class="btn btn-primary min-w-0 flex-1 sm:flex-none"/)
  })
})
