import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { LEGACY_THEME_STORAGE_KEY, THEME_STORAGE_KEY } from '@/composables/useTheme'

const indexHtml = readFileSync(join(process.cwd(), 'index.html'), 'utf8')
const prepaintScript = indexHtml.match(
  /<script[^>]*data-theme-prepaint[^>]*>([\s\S]*?)<\/script>/
)?.[1]

type ThemeSnapshot = {
  dark: boolean
  dataTheme: string
  colorScheme: string
  requestedKeys: string[]
}

function runPrepaint(
  storedValues: Record<string, string | null>,
  systemDark = false
): ThemeSnapshot {
  if (!prepaintScript) throw new Error('Theme prepaint script is missing')

  const requestedKeys: string[] = []
  const root = {
    classList: {
      dark: false,
      toggle(name: string, enabled: boolean) {
        if (name === 'dark') this.dark = enabled
      },
    },
    dataset: {} as Record<string, string>,
    style: {} as Record<string, string>,
  }
  const storage = {
    getItem(key: string) {
      requestedKeys.push(key)
      return storedValues[key] ?? null
    },
  }

  new Function('window', 'document', 'localStorage', prepaintScript)(
    { matchMedia: () => ({ matches: systemDark }) },
    { documentElement: root },
    storage
  )

  return {
    dark: root.classList.dark,
    dataTheme: root.dataset.theme,
    colorScheme: root.style.colorScheme,
    requestedKeys,
  }
}

describe('theme prepaint', () => {
  it('uses the same current and legacy storage keys as useTheme', () => {
    const currentSnapshot = runPrepaint({ [THEME_STORAGE_KEY]: 'light' })
    const legacySnapshot = runPrepaint({ [LEGACY_THEME_STORAGE_KEY]: 'light' })

    expect(currentSnapshot.requestedKeys).toEqual([THEME_STORAGE_KEY])
    expect(legacySnapshot.requestedKeys).toEqual([THEME_STORAGE_KEY, LEGACY_THEME_STORAGE_KEY])
  })

  it('prefers the current stored theme over the legacy value', () => {
    const snapshot = runPrepaint({
      [THEME_STORAGE_KEY]: 'dark',
      [LEGACY_THEME_STORAGE_KEY]: 'light',
    })

    expect(snapshot).toMatchObject({ dark: true, dataTheme: 'dark', colorScheme: 'dark' })
  })

  it('falls back to a valid legacy preference', () => {
    const snapshot = runPrepaint({ [LEGACY_THEME_STORAGE_KEY]: 'dark' })

    expect(snapshot).toMatchObject({ dark: true, dataTheme: 'dark', colorScheme: 'dark' })
  })

  it('resolves system preference for missing or invalid stored values', () => {
    const snapshot = runPrepaint({ [THEME_STORAGE_KEY]: 'invalid' }, true)

    expect(snapshot).toMatchObject({ dark: true, dataTheme: 'dark', colorScheme: 'dark' })
  })
})
