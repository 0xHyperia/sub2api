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
  arg?: AstExpression
  exp?: AstExpression
  value?: AstExpression
}

interface AstNode {
  type: number
  tag?: string
  props?: AstProp[]
  children?: AstNode[]
  loc?: { start: { line: number } }
}

const targets = [
  {
    path: 'src/components/account/CreateAccountModal.vue',
    actions: ['moveTempUnschedRule', 'removeTempUnschedRule']
  },
  {
    path: 'src/components/account/EditAccountModal.vue',
    actions: ['moveTempUnschedRule', 'removeTempUnschedRule']
  },
  {
    path: 'src/components/account/BulkEditAccountModal.vue',
    actions: ['removeModelMapping']
  },
  {
    path: 'src/views/admin/SettingsView.vue',
    actions: ['model_whitelist!.splice']
  },
  {
    path: 'src/views/admin/ChannelsView.vue',
    actions: ['removeMappingEntry']
  }
] as const

function collectButtons(node: AstNode, result: AstNode[] = []): AstNode[] {
  if (node.type === 1 && node.tag === 'button') result.push(node)
  node.children?.forEach((child) => collectButtons(child, result))
  return result
}

function directiveExpression(node: AstNode, name: string, arg: string): string {
  const directive = node.props?.find(
    (prop) => prop.type === 7 && prop.name === name && prop.arg?.content === arg
  )
  return directive?.exp?.content || ''
}

function hasAccessibleName(node: AstNode): boolean {
  return Boolean(
    node.props?.some((prop) => {
      if (prop.type === 6) {
        return ['aria-label', 'aria-labelledby', 'title'].includes(prop.name)
      }
      return (
        prop.type === 7 &&
        prop.name === 'bind' &&
        ['aria-label', 'aria-labelledby', 'title'].includes(prop.arg?.content || '')
      )
    })
  )
}

describe.each(targets)('$path icon actions', ({ path, actions }) => {
  it('gives every destructive or reorder icon button an accessible name', () => {
    const filename = resolve(process.cwd(), path)
    const { descriptor, errors } = parse(readFileSync(filename, 'utf8'), { filename })
    expect(errors).toEqual([])

    const buttons = collectButtons(descriptor.template!.ast as AstNode)
    actions.forEach((action) => {
      const matching = buttons.filter((button) =>
        directiveExpression(button, 'on', 'click').includes(action)
      )
      expect(matching.length, `Missing button action ${action} in ${path}`).toBeGreaterThan(0)
      expect(
        matching
          .filter((button) => !hasAccessibleName(button))
          .map((button) => `${path}:${button.loc?.start.line ?? '?'}`)
      ).toEqual([])
    })
  })
})
