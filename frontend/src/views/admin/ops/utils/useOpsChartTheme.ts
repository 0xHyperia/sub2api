import { onMounted, onUnmounted, ref } from 'vue'

export function useOpsChartTheme() {
  const isDarkMode = ref(false)
  let observer: MutationObserver | null = null

  const syncTheme = () => {
    isDarkMode.value = typeof document !== 'undefined'
      && document.documentElement.classList.contains('dark')
  }

  onMounted(() => {
    syncTheme()
    if (typeof MutationObserver === 'undefined') return

    observer = new MutationObserver(syncTheme)
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })
  })

  onUnmounted(() => {
    observer?.disconnect()
    observer = null
  })

  return { isDarkMode }
}
