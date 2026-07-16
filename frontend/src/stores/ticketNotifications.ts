import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { useAuthStore } from './auth'
import { adminTicketsAPI } from '@/api/admin/tickets'
import { ticketsAPI } from '@/api/tickets'
import type { Ticket } from '@/types'

export const useTicketNotificationStore = defineStore('ticketNotifications', () => {
  const unreadCount = ref(0)
  const items = ref<Ticket[]>([])
  const loading = ref(false)
  let timer: ReturnType<typeof setInterval> | null = null
  const auth = useAuthStore()
  const isAdmin = computed(() => auth.isAdmin)

  async function refresh() {
    if (!auth.isAuthenticated || document.visibilityState === 'hidden') return

    loading.value = true
    try {
      const api = isAdmin.value ? adminTicketsAPI : ticketsAPI
      const [count, list] = await Promise.all([
        api.unread(),
        api.list({ page: 1, page_size: 8, unread_only: true })
      ])
      unreadCount.value = count
      items.value = list.items
    } catch (error) {
      console.error('Failed to fetch ticket notifications', error)
    } finally {
      loading.value = false
    }
  }

  function start() {
    stop()
    void refresh()
    timer = setInterval(() => void refresh(), 30_000)
  }

  function stop() {
    if (timer) {
      clearInterval(timer)
      timer = null
    }
  }

  function reset() {
    stop()
    unreadCount.value = 0
    items.value = []
  }

  return { unreadCount, items, loading, refresh, start, stop, reset }
})
