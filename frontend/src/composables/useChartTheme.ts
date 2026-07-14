import { computed, onMounted, onUnmounted, ref } from 'vue'

const LIGHT_FALLBACKS = {
  brand: '15 23 42',
  info: '37 99 235',
  success: '5 150 105',
  warning: '217 119 6',
  danger: '220 38 38',
  surface: '255 255 255',
  surfaceRaised: '255 255 255',
  foreground: '15 23 42',
  foregroundMuted: '71 85 105',
  foregroundSubtle: '100 116 139',
  outline: '226 232 240',
} as const

const DARK_FALLBACKS = {
  brand: '241 245 249',
  info: '37 99 235',
  success: '5 150 105',
  warning: '180 83 9',
  danger: '220 38 38',
  surface: '22 28 34',
  surfaceRaised: '29 36 43',
  foreground: '238 244 251',
  foregroundMuted: '190 202 214',
  foregroundSubtle: '143 158 174',
  outline: '45 57 70',
} as const

const toCssColor = (variable: string, fallback: string, alpha = 1): string => {
  const value = typeof document === 'undefined'
    ? fallback
    : getComputedStyle(document.documentElement).getPropertyValue(variable).trim() || fallback
  const channels = value.split(/\s+/).map(Number)
  if (channels.length !== 3 || channels.some((channel) => !Number.isFinite(channel))) return value
  return alpha === 1
    ? `rgb(${channels.join(', ')})`
    : `rgba(${channels.join(', ')}, ${alpha})`
}

export function useChartTheme() {
  const isDarkMode = ref(false)
  const themeRevision = ref(0)
  let observer: MutationObserver | null = null

  const syncTheme = () => {
    isDarkMode.value = typeof document !== 'undefined'
      && document.documentElement.classList.contains('dark')
    themeRevision.value += 1
  }

  const chartTheme = computed(() => {
    themeRevision.value
    const fallback = isDarkMode.value ? DARK_FALLBACKS : LIGHT_FALLBACKS
    const color = (name: keyof typeof fallback, variable: string, alpha = 1) =>
      toCssColor(variable, fallback[name], alpha)

    return {
      brand: color('brand', '--color-brand'),
      brandAlpha: color('brand', '--color-brand', 0.14),
      info: color('info', '--color-info'),
      infoAlpha: color('info', '--color-info', 0.14),
      success: color('success', '--color-success'),
      successAlpha: color('success', '--color-success', 0.14),
      warning: color('warning', '--color-warning'),
      warningAlpha: color('warning', '--color-warning', 0.14),
      danger: color('danger', '--color-danger'),
      dangerAlpha: color('danger', '--color-danger', 0.14),
      surface: color('surface', '--color-surface'),
      surfaceRaised: color('surfaceRaised', '--color-surface-raised'),
      foreground: color('foreground', '--color-foreground'),
      foregroundMuted: color('foregroundMuted', '--color-foreground-muted'),
      foregroundSubtle: color('foregroundSubtle', '--color-foreground-subtle'),
      outline: color('outline', '--color-border'),
    }
  })

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

  return { isDarkMode, chartTheme }
}
