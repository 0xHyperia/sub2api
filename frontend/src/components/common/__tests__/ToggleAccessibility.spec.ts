import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'
import { parse } from 'vue/compiler-sfc'

interface AstExpression {
  content?: string
}

interface AstProp {
  type: number
  name: string
  value?: AstExpression
  arg?: AstExpression
  exp?: AstExpression
}

interface AstNode {
  type: number
  tag?: string
  props?: AstProp[]
  children?: AstNode[]
  loc?: { start: { line: number } }
}

const scopedFiles = [
  { path: 'src/views/admin/SettingsView.vue', expectedCount: 67 },
  { path: 'src/views/admin/ops/components/OpsSettingsDialog.vue', expectedCount: 14 },
  { path: 'src/components/admin/account/ScheduledTestsPanel.vue', expectedCount: 5 },
  { path: 'src/components/admin/monitor/MonitorFormDialog.vue', expectedCount: 1 }
] as const

function readTemplate(relativePath: string): { source: string; root: AstNode } {
  const filename = resolve(process.cwd(), relativePath)
  const source = readFileSync(filename, 'utf8')
  const { descriptor, errors } = parse(source, { filename })
  expect(errors).toEqual([])
  expect(descriptor.template?.ast).toBeTruthy()
  return { source, root: descriptor.template!.ast as AstNode }
}

function collectToggleNodes(node: AstNode, result: AstNode[] = []): AstNode[] {
  if (node.type === 1 && node.tag === 'Toggle') result.push(node)
  node.children?.forEach((child) => collectToggleNodes(child, result))
  return result
}

function boundExpression(node: AstNode, name: string): string {
  const prop = node.props?.find((candidate) => {
    if (candidate.type === 6) return candidate.name === name
    return candidate.type === 7 && candidate.name === 'bind' && candidate.arg?.content === name
  })
  if (!prop) return ''
  return (prop.type === 6 ? prop.value?.content : prop.exp?.content)?.trim() || ''
}

function modelExpression(node: AstNode): string {
  const model = node.props?.find((prop) => prop.type === 7 && prop.name === 'model')
  return (model?.exp?.content || boundExpression(node, 'model-value')).replace(/\s+/g, '')
}

function accessibleNameExpression(node: AstNode): string {
  return boundExpression(node, 'aria-label') || boundExpression(node, 'aria-labelledby')
}

function findToggleByModel(root: AstNode, expression: string): AstNode {
  const normalizedExpression = expression.replace(/\s+/g, '')
  const toggle = collectToggleNodes(root).find(
    (node) => modelExpression(node) === normalizedExpression
  )
  expect(toggle, `Missing Toggle for v-model ${expression}`).toBeTruthy()
  return toggle!
}

describe('Toggle accessible names in admin settings surfaces', () => {
  it.each(scopedFiles)('$path names all $expectedCount switches', ({ path, expectedCount }) => {
    const { root } = readTemplate(path)
    const toggles = collectToggleNodes(root)
    const unnamed = toggles
      .filter((toggle) => !accessibleNameExpression(toggle))
      .map((toggle) => `${path}:${toggle.loc?.start.line ?? '?'}`)

    expect(toggles).toHaveLength(expectedCount)
    expect(unnamed).toEqual([])
  })

  it('includes the repeated entity in every dynamic switch name', () => {
    const settings = readTemplate('src/views/admin/SettingsView.vue').root
    const scheduledTests = readTemplate(
      'src/components/admin/account/ScheduledTestsPanel.vue'
    ).root

    expect(
      accessibleNameExpression(
        findToggleByModel(
          settings,
          'authSourceDefaults[authSource.source].grant_on_signup'
        )
      )
    ).toContain('authSource.title')
    expect(
      accessibleNameExpression(
        findToggleByModel(
          settings,
          'authSourceDefaults[authSource.source].grant_on_first_bind'
        )
      )
    ).toContain('authSource.title')
    expect(
      accessibleNameExpression(findToggleByModel(settings, 'block.enabled'))
    ).toContain('index')
    expect(
      accessibleNameExpression(findToggleByModel(settings, 'block.cacheControlEnabled'))
    ).toContain('index')
    expect(
      accessibleNameExpression(findToggleByModel(scheduledTests, 'plan.enabled'))
    ).toContain('plan.model_id')
  })
})
