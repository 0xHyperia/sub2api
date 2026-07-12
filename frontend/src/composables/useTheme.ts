import { computed, readonly, ref } from 'vue'

export type ThemePreference = 'light' | 'dark' | 'system'
export type ResolvedTheme = Exclude<ThemePreference, 'system'>

export const THEME_STORAGE_KEY = 'usa-zero-theme'
export const LEGACY_THEME_STORAGE_KEY = 'theme'

const preference = ref<ThemePreference>('system')
const resolvedTheme = ref<ResolvedTheme>('light')
let mediaQuery: MediaQueryList | null = null
let initialized = false

function isThemePreference(value: string | null): value is ThemePreference {
  return value === 'light' || value === 'dark' || value === 'system'
}

function getSystemTheme(): ResolvedTheme {
  if (typeof window === 'undefined') return 'light'
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function resolveTheme(value: ThemePreference): ResolvedTheme {
  return value === 'system' ? getSystemTheme() : value
}

function applyTheme(value: ThemePreference): void {
  const nextResolvedTheme = resolveTheme(value)
  resolvedTheme.value = nextResolvedTheme

  if (typeof document === 'undefined') return
  document.documentElement.classList.toggle('dark', nextResolvedTheme === 'dark')
  document.documentElement.dataset.theme = nextResolvedTheme
  document.documentElement.style.colorScheme = nextResolvedTheme
}

function readStoredTheme(): ThemePreference {
  if (typeof window === 'undefined') return 'system'

  const storedTheme = localStorage.getItem(THEME_STORAGE_KEY)
  if (isThemePreference(storedTheme)) return storedTheme

  const legacyTheme = localStorage.getItem(LEGACY_THEME_STORAGE_KEY)
  return isThemePreference(legacyTheme) ? legacyTheme : 'system'
}

function persistTheme(value: ThemePreference): void {
  if (typeof window === 'undefined') return

  localStorage.setItem(THEME_STORAGE_KEY, value)
  // Keep legacy consumers in sync until every route uses this composable.
  localStorage.setItem(LEGACY_THEME_STORAGE_KEY, value)
}

function handleSystemThemeChange(): void {
  if (preference.value === 'system') applyTheme('system')
}

function handleStorageChange(event: StorageEvent): void {
  if (event.key !== THEME_STORAGE_KEY && event.key !== LEGACY_THEME_STORAGE_KEY) return
  const nextPreference = readStoredTheme()
  preference.value = nextPreference
  applyTheme(nextPreference)
}

export function initializeTheme(): ResolvedTheme {
  const initialPreference = readStoredTheme()
  preference.value = initialPreference
  applyTheme(initialPreference)

  if (!initialized && typeof window !== 'undefined') {
    mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    if (typeof mediaQuery.addEventListener === 'function') {
      mediaQuery.addEventListener('change', handleSystemThemeChange)
    } else if (typeof mediaQuery.addListener === 'function') {
      mediaQuery.addListener(handleSystemThemeChange)
    }
    window.addEventListener('storage', handleStorageChange)
    initialized = true
  }

  return resolvedTheme.value
}

export function setTheme(value: ThemePreference): void {
  preference.value = value
  persistTheme(value)
  applyTheme(value)
}

export function toggleTheme(): void {
  setTheme(resolvedTheme.value === 'dark' ? 'light' : 'dark')
}

export function useTheme() {
  if (!initialized) initializeTheme()

  return {
    preference: readonly(preference),
    resolvedTheme: readonly(resolvedTheme),
    isDark: computed(() => resolvedTheme.value === 'dark'),
    setTheme,
    toggleTheme
  }
}
