import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

const { auth, userUnread, userList, adminUnread, adminList } = vi.hoisted(() => ({
  auth: { isAuthenticated: true, isAdmin: false },
  userUnread: vi.fn(),
  userList: vi.fn(),
  adminUnread: vi.fn(),
  adminList: vi.fn()
}))

vi.mock('@/stores/auth', () => ({ useAuthStore: () => auth }))
vi.mock('@/api/tickets', () => ({
  ticketsAPI: { unread: userUnread, list: userList }
}))
vi.mock('@/api/admin/tickets', () => ({
  adminTicketsAPI: { unread: adminUnread, list: adminList }
}))

import { useTicketNotificationStore } from '@/stores/ticketNotifications'

describe('useTicketNotificationStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.useFakeTimers()
    vi.clearAllMocks()
    auth.isAuthenticated = true
    auth.isAdmin = false
    vi.spyOn(document, 'visibilityState', 'get').mockReturnValue('visible')
    userUnread.mockResolvedValue(2)
    userList.mockResolvedValue({ items: [{ id: 1, number: '20260716-ABC234' }] })
    adminUnread.mockResolvedValue(5)
    adminList.mockResolvedValue({ items: [{ id: 2, number: '20260716-XYZ234' }] })
  })

  afterEach(() => {
    vi.restoreAllMocks()
    vi.useRealTimers()
  })

  it('loads the current user queue and the shared admin queue', async () => {
    const userStore = useTicketNotificationStore()

    await userStore.refresh()
    expect(userUnread).toHaveBeenCalledOnce()
    expect(userStore.unreadCount).toBe(2)

    setActivePinia(createPinia())
    auth.isAdmin = true
    const adminStore = useTicketNotificationStore()
    await adminStore.refresh()
    expect(adminUnread).toHaveBeenCalledOnce()
    expect(adminStore.unreadCount).toBe(5)
  })

  it('does not request notifications while the page is hidden', async () => {
    vi.spyOn(document, 'visibilityState', 'get').mockReturnValue('hidden')

    await useTicketNotificationStore().refresh()

    expect(userUnread).not.toHaveBeenCalled()
    expect(userList).not.toHaveBeenCalled()
  })

  it('polls every 30 seconds and clears the timer on stop', async () => {
    const store = useTicketNotificationStore()

    store.start()
    await vi.runOnlyPendingTimersAsync()
    expect(userUnread).toHaveBeenCalledTimes(2)

    store.stop()
    await vi.advanceTimersByTimeAsync(60_000)
    expect(userUnread).toHaveBeenCalledTimes(2)
  })
})
