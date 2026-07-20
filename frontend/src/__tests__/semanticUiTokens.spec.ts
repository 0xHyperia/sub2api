import { describe, expect, it } from 'vitest'
import { readdirSync, readFileSync } from 'node:fs'
import { join, relative } from 'node:path'

const sourceRoot = join(process.cwd(), 'src')
const sharedComponentsRoot = join(sourceRoot, 'components')
const tokensPath = join(sourceRoot, 'styles', 'tokens.css')
const tailwindConfigPath = join(process.cwd(), 'tailwind.config.js')
const legacyUtility =
  /(?:bg|text|border|divide|ring|shadow|from|to|via|outline|fill|stroke)-(?:red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose|slate|gray|zinc|neutral|stone|dark|primary)-\d+/g
const legacyDarkVariant =
  /dark:(?:bg|text|border|divide|ring|shadow|from|to|via|outline|fill|stroke)-/g
const hardcodedThemeBranch = /(?:isDark|isDarkMode)\.value\s*\?\s*['"]#/g
const unsafeSemanticPair =
  /(?:bg-brand\b[^'"\n]*\btext-white\b|text-white\b[^'"\n]*\bbg-brand\b|bg-surface-subtle\b[^'"\n]*\btext-white\b|text-white\b[^'"\n]*\bbg-surface-subtle\b|bg-(?:info|success|warning|danger)\b[^'"\n]*\btext-white\b|text-white\b[^'"\n]*\bbg-(?:info|success|warning|danger)\b|bg-foreground(?:\/\d+)?\b[^'"\n]*\btext-foreground-(?:muted|subtle)\b|text-foreground-(?:muted|subtle)\b[^'"\n]*\bbg-foreground(?:\/\d+)?\b)/g
const malformedUtility = /(?:^|\s)\/\d+\b|\bbg-[a-z-]+(?:subtle|foreground)\d+\b/gm

const collectVueFiles = (directory: string): string[] =>
  readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = join(directory, entry.name)
    if (entry.isDirectory()) return collectVueFiles(path)
    return entry.isFile() && entry.name.endsWith('.vue') ? [path] : []
  })

const parseThemeColors = (source: string, selector: ':root' | '.dark') => {
  const escapedSelector = selector.replace('.', '\\.')
  const block = source.match(new RegExp(`${escapedSelector}\\s*\\{([\\s\\S]*?)\\n\\}`))?.[1] ?? ''
  return Object.fromEntries(
    [...block.matchAll(/(--color-[\w-]+):\s*(\d+)\s+(\d+)\s+(\d+);/g)].map((match) => [
      match[1],
      match.slice(2, 5).map(Number),
    ])
  ) as Record<string, number[]>
}

const relativeLuminance = ([red, green, blue]: number[]) => {
  const [r, g, b] = [red, green, blue].map((channel) => {
    const value = channel / 255
    return value <= 0.03928 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4
  })
  return 0.2126 * r + 0.7152 * g + 0.0722 * b
}

const contrastRatio = (first: number[], second: number[]) => {
  const firstLuminance = relativeLuminance(first)
  const secondLuminance = relativeLuminance(second)
  return (Math.max(firstLuminance, secondLuminance) + 0.05)
    / (Math.min(firstLuminance, secondLuminance) + 0.05)
}

describe('semantic UI token contract', () => {
  it('does not reintroduce legacy palette utilities in Vue components', () => {
    const violations = collectVueFiles(sourceRoot).flatMap((file) => {
      const source = readFileSync(file, 'utf8')
      const matches = [
        ...source.matchAll(legacyUtility),
        ...source.matchAll(legacyDarkVariant),
        ...source.matchAll(hardcodedThemeBranch),
      ]
      return matches.map((match) => `${relative(sourceRoot, file)}: ${match[0]}`)
    })

    expect(violations).toEqual([])
  })

  it('keeps foregrounds paired with their semantic surfaces', () => {
    const violations = collectVueFiles(sourceRoot).flatMap((file) => {
      const source = readFileSync(file, 'utf8')
      return [...source.matchAll(unsafeSemanticPair)].map(
        (match) => `${relative(sourceRoot, file)}: ${match[0]}`
      )
    })

    expect(violations).toEqual([])
  })

  it('does not contain malformed utility fragments in shared components', () => {
    const violations = collectVueFiles(sharedComponentsRoot).flatMap((file) => {
      const source = readFileSync(file, 'utf8')
      return [...source.matchAll(malformedUtility)].map(
        (match) => `${relative(sourceRoot, file)}: ${match[0].trim()}`
      )
    })

    expect(violations).toEqual([])
  })

  it('defines reusable inverse, code, and solid status foreground tokens', () => {
    const tokens = readFileSync(tokensPath, 'utf8')
    const tailwindConfig = readFileSync(tailwindConfigPath, 'utf8')
    const themeTokens = [
      '--color-inverse',
      '--color-inverse-foreground',
      '--color-code',
      '--color-code-foreground',
      '--color-info-solid-foreground',
      '--color-success-solid-foreground',
      '--color-warning-solid-foreground',
      '--color-danger-solid-foreground',
    ]

    for (const token of themeTokens) {
      expect(tokens.match(new RegExp(`${token}:`, 'g'))).toHaveLength(2)
    }
    expect(tailwindConfig).toContain("inverse: {")
    expect(tailwindConfig).toContain("code: {")
    expect(tailwindConfig.match(/'solid-foreground':/g)).toHaveLength(4)
  })

  it('keeps semantic solid surfaces at AA text contrast in both themes', () => {
    const tokens = readFileSync(tokensPath, 'utf8')
    const pairs = [
      ['--color-brand', '--color-brand-foreground'],
      ['--color-inverse', '--color-inverse-foreground'],
      ['--color-code', '--color-code-foreground'],
      ['--color-info', '--color-info-solid-foreground'],
      ['--color-success', '--color-success-solid-foreground'],
      ['--color-warning', '--color-warning-solid-foreground'],
      ['--color-danger', '--color-danger-solid-foreground'],
    ]

    for (const selector of [':root', '.dark'] as const) {
      const colors = parseThemeColors(tokens, selector)
      for (const [surface, foreground] of pairs) {
        expect(
          contrastRatio(colors[surface], colors[foreground]),
          `${selector} ${surface} / ${foreground}`
        ).toBeGreaterThanOrEqual(4.5)
      }
    }
  })
})
