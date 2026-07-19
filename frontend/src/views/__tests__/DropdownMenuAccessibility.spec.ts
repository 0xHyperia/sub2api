import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { parse } from 'vue/compiler-sfc'
import { describe, expect, it } from 'vitest'

interface AstExpression {
  content?: string
}

interface AstProp {
  type: number
  name: string
  value?: AstExpression
  arg?: AstExpression
}

interface AstNode {
  type: number
  tag?: string
  props?: AstProp[]
  children?: AstNode[]
  loc?: { start: { line: number } }
}

const scopedViews = [
  { path: 'src/views/admin/UsersView.vue', menus: 4, triggers: 5 },
  { path: 'src/views/admin/GroupsView.vue', menus: 1, triggers: 1 },
  { path: 'src/views/admin/ProxiesView.vue', menus: 1, triggers: 1 },
  { path: 'src/views/admin/SubscriptionsView.vue', menus: 1, triggers: 1 },
  { path: 'src/views/user/KeysView.vue', menus: 1, triggers: 1 },
  { path: 'src/views/user/UsageView.vue', menus: 1, triggers: 1 },
] as const

function readTemplate(relativePath: string): AstNode {
  const filename = resolve(process.cwd(), relativePath)
  const source = readFileSync(filename, 'utf8')
  const { descriptor, errors } = parse(source, { filename })
  expect(errors).toEqual([])
  expect(descriptor.template?.ast).toBeTruthy()
  return descriptor.template!.ast as AstNode
}

function collectElements(node: AstNode, result: AstNode[] = []): AstNode[] {
  if (node.type === 1) result.push(node)
  node.children?.forEach((child) => collectElements(child, result))
  return result
}

function staticAttribute(node: AstNode, name: string): string {
  const attribute = node.props?.find((prop) => prop.type === 6 && prop.name === name)
  return attribute?.value?.content || ''
}

function hasAttributeOrBinding(node: AstNode, name: string): boolean {
  return !!node.props?.some((prop) =>
    (prop.type === 6 && prop.name === name) ||
    (prop.type === 7 && prop.name === 'bind' && prop.arg?.content === name)
  )
}

function hasEvent(node: AstNode, name: string): boolean {
  return !!node.props?.some(
    (prop) => prop.type === 7 && prop.name === 'on' && prop.arg?.content === name
  )
}

function location(path: string, node: AstNode): string {
  return `${path}:${node.loc?.start.line ?? '?'}`
}

describe('dropdown menu accessibility contracts', () => {
  it.each(scopedViews)('$path connects all menus to $triggers keyboard-capable triggers', ({ path, menus, triggers }) => {
    const elements = collectElements(readTemplate(path))
    const menuNodes = elements.filter((node) => staticAttribute(node, 'role') === 'menu')
    const menuTriggers = elements.filter((node) => staticAttribute(node, 'aria-haspopup') === 'menu')

    expect(menuNodes, `${path} menu count`).toHaveLength(menus)
    expect(menuTriggers, `${path} trigger count`).toHaveLength(triggers)

    const invalidMenus = menuNodes
      .filter((node) =>
        !hasAttributeOrBinding(node, 'id') ||
        !hasAttributeOrBinding(node, 'ref') ||
        !hasAttributeOrBinding(node, 'aria-labelledby') ||
        !hasEvent(node, 'keydown')
      )
      .map((node) => location(path, node))

    const invalidTriggers = menuTriggers
      .filter((node) =>
        node.tag !== 'button' ||
        staticAttribute(node, 'type') !== 'button' ||
        !hasAttributeOrBinding(node, 'id') ||
        !hasAttributeOrBinding(node, 'aria-expanded') ||
        !hasAttributeOrBinding(node, 'aria-controls') ||
        !hasEvent(node, 'keydown')
      )
      .map((node) => location(path, node))

    expect(invalidMenus).toEqual([])
    expect(invalidTriggers).toEqual([])
  })
})
