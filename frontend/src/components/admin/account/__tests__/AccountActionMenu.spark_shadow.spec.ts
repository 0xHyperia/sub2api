import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import AccountActionMenu from '../AccountActionMenu.vue'
import type { Account } from '@/types'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

function makeAccount(overrides: Partial<Account>): Account {
  return {
    id: 1,
    name: 'test-account',
    platform: 'openai',
    type: 'oauth',
    proxy_id: null,
    concurrency: 3,
    priority: 50,
    status: 'active',
    error_message: null,
    last_used_at: null,
    expires_at: null,
    auto_pause_on_expired: false,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    schedulable: true,
    rate_limited_at: null,
    rate_limit_reset_at: null,
    overload_until: null,
    temp_unschedulable_until: null,
    temp_unschedulable_reason: null,
    session_window_start: null,
    session_window_end: null,
    session_window_status: null,
    ...overrides,
  }
}

const anchorRect = new DOMRect(100, 100, 24, 24)

// AccountActionMenu uses <Teleport to="body">; content is rendered in document.body, not in wrapper.
const getBodyText = () => document.body.textContent ?? ''
const getBodyButtons = () => Array.from(document.body.querySelectorAll('button'))

afterEach(() => {
  document.body.innerHTML = ''
})

describe('AccountActionMenu — spark shadow 按钮可见性', () => {
  it('普通账号显示「复制账号」按钮', () => {
    const account = makeAccount({ platform: 'anthropic', type: 'apikey', parent_account_id: null })
    const wrapper = mount(AccountActionMenu, {
      props: { show: true, account, anchorRect },
      attachTo: document.body,
    })
    expect(getBodyText()).toContain('admin.accounts.duplicateAccount')
    wrapper.unmount()
  })

  it('影子账号隐藏「复制账号」按钮', () => {
    const account = makeAccount({ platform: 'openai', type: 'oauth', parent_account_id: 42 })
    const wrapper = mount(AccountActionMenu, {
      props: { show: true, account, anchorRect },
      attachTo: document.body,
    })
    expect(getBodyText()).not.toContain('admin.accounts.duplicateAccount')
    wrapper.unmount()
  })

  it.each(['oauth', 'setup-token'] as const)('%s 账号隐藏「复制账号」按钮，避免共享可轮换令牌', (type) => {
    const account = makeAccount({ platform: 'openai', type, parent_account_id: null })
    const wrapper = mount(AccountActionMenu, {
      props: { show: true, account, anchorRect },
      attachTo: document.body,
    })
    expect(getBodyText()).not.toContain('admin.accounts.duplicateAccount')
    wrapper.unmount()
  })

  it('点击「复制账号」触发 duplicate 事件并携带 account', async () => {
    const account = makeAccount({ platform: 'anthropic', type: 'apikey', parent_account_id: null })
    const wrapper = mount(AccountActionMenu, {
      props: { show: true, account, anchorRect },
      attachTo: document.body,
    })

    const duplicateBtn = getBodyButtons().find(b => b.textContent?.includes('admin.accounts.duplicateAccount'))
    expect(duplicateBtn).toBeDefined()

    duplicateBtn!.click()
    await wrapper.vm.$nextTick()

    const emitted = wrapper.emitted('duplicate')
    expect(emitted).toBeTruthy()
    expect(emitted![0][0]).toMatchObject({ id: account.id, name: account.name })
    wrapper.unmount()
  })

  it('OpenAI OAuth 母账号（无 parent_account_id）显示「创建 spark 影子」按钮', () => {
    const account = makeAccount({ platform: 'openai', type: 'oauth', parent_account_id: null })
    const wrapper = mount(AccountActionMenu, {
      props: { show: true, account, anchorRect },
      attachTo: document.body,
    })
    expect(getBodyText()).toContain('admin.accounts.createSparkShadow')
    wrapper.unmount()
  })

  it('影子账号（parent_account_id 非 null）隐藏「创建 spark 影子」按钮', () => {
    const account = makeAccount({ platform: 'openai', type: 'oauth', parent_account_id: 42 })
    const wrapper = mount(AccountActionMenu, {
      props: { show: true, account, anchorRect },
      attachTo: document.body,
    })
    expect(getBodyText()).not.toContain('admin.accounts.createSparkShadow')
    wrapper.unmount()
  })

  it('非 OpenAI 账号隐藏「创建 spark 影子」按钮', () => {
    const account = makeAccount({ platform: 'antigravity', type: 'oauth', parent_account_id: null })
    const wrapper = mount(AccountActionMenu, {
      props: { show: true, account, anchorRect },
      attachTo: document.body,
    })
    expect(getBodyText()).not.toContain('admin.accounts.createSparkShadow')
    wrapper.unmount()
  })

  it('影子账号隐藏凭据/隐私类操作(重授权/刷新token/隐私)— 外审 G4', () => {
    const account = makeAccount({ platform: 'openai', type: 'oauth', parent_account_id: 42 })
    const wrapper = mount(AccountActionMenu, {
      props: { show: true, account, anchorRect },
      attachTo: document.body,
    })
    const body = getBodyText()
    expect(body).not.toContain('admin.accounts.reAuthorize')
    expect(body).not.toContain('admin.accounts.refreshToken')
    expect(body).not.toContain('admin.accounts.setPrivacy')
    wrapper.unmount()
  })

  it('普通 OpenAI OAuth 母账号仍显示凭据/隐私类操作', () => {
    const account = makeAccount({ platform: 'openai', type: 'oauth', parent_account_id: null })
    const wrapper = mount(AccountActionMenu, {
      props: { show: true, account, anchorRect },
      attachTo: document.body,
    })
    const body = getBodyText()
    expect(body).toContain('admin.accounts.reAuthorize')
    expect(body).toContain('admin.accounts.setPrivacy')
    wrapper.unmount()
  })

  it('点击按钮触发 create-spark-shadow 事件并携带 account', async () => {
    const account = makeAccount({ platform: 'openai', type: 'oauth', parent_account_id: null })
    const wrapper = mount(AccountActionMenu, {
      props: { show: true, account, anchorRect },
      attachTo: document.body,
    })

    // Content is teleported to body — find button by text there
    const sparkBtn = getBodyButtons().find(b => b.textContent?.includes('admin.accounts.createSparkShadow'))
    expect(sparkBtn).toBeDefined()

    sparkBtn!.click()
    await wrapper.vm.$nextTick()

    const emitted = wrapper.emitted('create-spark-shadow')
    expect(emitted).toBeTruthy()
    expect(emitted![0][0]).toMatchObject({ id: account.id, platform: 'openai' })

    wrapper.unmount()
  })

  it('提供 menu/menuitem 语义并支持方向键、Home、End 与 Escape 回焦', async () => {
    const opener = document.createElement('button')
    opener.textContent = 'open actions'
    document.body.appendChild(opener)
    opener.focus()

    const account = makeAccount({ platform: 'openai', type: 'oauth', parent_account_id: null })
    const wrapper = mount(AccountActionMenu, {
      props: { show: true, account, position },
      attachTo: document.body,
    })
    await wrapper.vm.$nextTick()
    await wrapper.vm.$nextTick()

    const menu = document.body.querySelector<HTMLElement>('[role="menu"]')
    const items = Array.from(document.body.querySelectorAll<HTMLButtonElement>('[role="menuitem"]'))
    expect(menu).not.toBeNull()
    expect(menu?.getAttribute('aria-label')).toContain(account.name)
    expect(items.length).toBeGreaterThan(3)
    expect(items.every(item => item.type === 'button' && item.tabIndex === -1)).toBe(true)
    expect(document.activeElement).toBe(items[0])

    items[0]!.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown', bubbles: true, cancelable: true }))
    expect(document.activeElement).toBe(items[1])

    items[1]!.dispatchEvent(new KeyboardEvent('keydown', { key: 'End', bubbles: true, cancelable: true }))
    expect(document.activeElement).toBe(items.at(-1))

    items.at(-1)!.dispatchEvent(new KeyboardEvent('keydown', { key: 'Home', bubbles: true, cancelable: true }))
    expect(document.activeElement).toBe(items[0])

    items[0]!.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true }))
    expect(wrapper.emitted('close')).toHaveLength(1)

    await wrapper.setProps({ show: false })
    await wrapper.vm.$nextTick()
    await wrapper.vm.$nextTick()
    expect(document.activeElement).toBe(opener)

    wrapper.unmount()
  })

  it('Tab 关闭菜单并把焦点移到触发器后的控件', async () => {
    const opener = document.createElement('button')
    const nextControl = document.createElement('button')
    opener.textContent = 'open actions'
    nextControl.textContent = 'next control'
    document.body.append(opener, nextControl)
    opener.focus()

    const wrapper = mount(AccountActionMenu, {
      props: {
        show: true,
        account: makeAccount({ platform: 'openai', type: 'oauth' }),
        position,
      },
      attachTo: document.body,
    })
    await wrapper.vm.$nextTick()
    await wrapper.vm.$nextTick()

    const firstItem = document.body.querySelector<HTMLButtonElement>('[role="menuitem"]')!
    firstItem.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', bubbles: true, cancelable: true }))
    await wrapper.vm.$nextTick()

    expect(wrapper.emitted('close')).toHaveLength(1)
    expect(document.activeElement).toBe(nextControl)
    wrapper.unmount()
  })
})
