import { describe, expect, it } from 'vitest'
import { readdirSync, readFileSync } from 'node:fs'
import { join, relative } from 'node:path'

const sourceRoot = join(process.cwd(), 'src')
const legacyUtility =
  /(?:bg|text|border|divide|ring|shadow|from|to|via|outline|fill|stroke)-(?:red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose|slate|gray|zinc|neutral|stone|dark|primary)-\d+/g
const legacyDarkVariant =
  /dark:(?:bg|text|border|divide|ring|shadow|from|to|via|outline|fill|stroke)-/g
const hardcodedThemeBranch = /(?:isDark|isDarkMode)\.value\s*\?\s*['"]#/g

const collectVueFiles = (directory: string): string[] =>
  readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = join(directory, entry.name)
    if (entry.isDirectory()) return collectVueFiles(path)
    return entry.isFile() && entry.name.endsWith('.vue') ? [path] : []
  })

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
})
