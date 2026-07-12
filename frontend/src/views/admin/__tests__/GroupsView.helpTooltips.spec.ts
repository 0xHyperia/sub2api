import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import { parse } from 'vue/compiler-sfc'
import { afterEach, describe, expect, it, vi } from 'vitest'

import HelpTooltip from '@/components/common/HelpTooltip.vue'

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
  loc?: { source?: string; start: { line: number } }
}

const filename = resolve(process.cwd(), 'src/views/admin/GroupsView.vue')
const groupsViewSource = readFileSync(filename, 'utf8')

function readTemplate(): AstNode {
  const { descriptor, errors } = parse(groupsViewSource, { filename })
  expect(errors).toEqual([])
  expect(descriptor.template?.ast).toBeTruthy()
  return descriptor.template!.ast as AstNode
}

function collectHelpTooltips(node: AstNode, result: AstNode[] = []): AstNode[] {
  if (node.type === 1 && node.tag === 'HelpTooltip') result.push(node)
  node.children?.forEach((child) => collectHelpTooltips(child, result))
  return result
}

function propExpression(node: AstNode, name: string): string {
  const prop = node.props?.find((candidate) => {
    if (candidate.type === 6) return candidate.name === name
    return candidate.type === 7 && candidate.name === 'bind' && candidate.arg?.content === name
  })
  if (!prop) return ''
  return (prop.type === 6 ? prop.value?.content : prop.exp?.content)?.trim() || ''
}

function getTooltipElement(): HTMLDivElement {
  const tooltip = document.body.querySelector('[role="tooltip"]')
  if (!(tooltip instanceof HTMLDivElement)) throw new Error('tooltip element not found')
  return tooltip
}

describe('admin GroupsView help tooltip contracts', () => {
  afterEach(() => {
    vi.restoreAllMocks()
    document.body.innerHTML = ''
  })

  it('uses named HelpTooltip triggers for every create and edit form hint', () => {
    const tooltips = collectHelpTooltips(readTemplate())

    expect(tooltips).toHaveLength(12)
    expect(
      tooltips
        .filter((tooltip) => !propExpression(tooltip, 'label'))
        .map((tooltip) => tooltip.loc?.start.line),
    ).toEqual([])

    const widths = tooltips.map((tooltip) => propExpression(tooltip, 'width-class'))
    expect(widths.filter((width) => width === 'w-[min(18rem,calc(100vw-2rem))]')).toHaveLength(10)
    expect(widths.filter((width) => width === 'w-[min(20rem,calc(100vw-2rem))]')).toHaveLength(2)

    const clickTooltips = tooltips.filter((tooltip) => propExpression(tooltip, 'trigger') === 'click')
    expect(clickTooltips).toHaveLength(2)
    expect(clickTooltips.every((tooltip) => propExpression(tooltip, 'close-label'))).toBe(true)

    expect(groupsViewSource).not.toContain('group-hover:pointer-events-auto')
    expect(groupsViewSource).not.toContain('pointer-events-none absolute bottom-full')
  })

  it('preserves the simple and rich help content in both forms', () => {
    const tooltipSources = collectHelpTooltips(readTemplate()).map((tooltip) => tooltip.loc?.source || '')
    const sourceFor = (key: string) => tooltipSources.filter((source) => source.includes(key))

    expect(sourceFor('admin.groups.copyAccounts.tooltip')).toHaveLength(2)
    expect(sourceFor('admin.groups.supportedScopes.tooltip')).toHaveLength(2)
    expect(sourceFor('admin.groups.mcpXml.tooltip')).toHaveLength(2)
    expect(sourceFor('admin.groups.claudeCode.tooltip')).toHaveLength(2)
    expect(sourceFor('admin.groups.modelRouting.tooltip')).toHaveLength(2)
    expect(sourceFor('admin.groups.exclusiveTooltip.description')).toHaveLength(2)
    expect(sourceFor('admin.groups.exclusiveTooltip.exampleContent')).toHaveLength(2)
  })

  it('opens a default trigger from focus, closes with Escape, and stays inside 320px', async () => {
    vi.spyOn(window, 'innerWidth', 'get').mockReturnValue(320)
    vi.spyOn(window, 'innerHeight', 'get').mockReturnValue(720)

    const wrapper = mount(HelpTooltip, {
      attachTo: document.body,
      props: {
        content: 'Supported model details',
        label: 'Supported models',
        widthClass: 'w-[min(18rem,calc(100vw-2rem))]',
      },
    })
    const trigger = wrapper.get('.group')
    const triggerButton = trigger.get('button')
    const tooltip = getTooltipElement()

    vi.spyOn(trigger.element, 'getBoundingClientRect').mockReturnValue({
      x: 300,
      y: 200,
      left: 300,
      right: 316,
      top: 200,
      bottom: 216,
      width: 16,
      height: 16,
      toJSON: () => ({}),
    })
    vi.spyOn(tooltip, 'getBoundingClientRect').mockReturnValue({
      x: 0,
      y: 0,
      left: 0,
      right: 288,
      top: 0,
      bottom: 80,
      width: 288,
      height: 80,
      toJSON: () => ({}),
    })

    expect(triggerButton.attributes('aria-label')).toBe('Supported models')
    await triggerButton.trigger('focusin')
    await nextTick()

    expect(tooltip.style.display).not.toBe('none')
    expect(tooltip.style.left).toBe('160px')
    expect(tooltip.classList.contains('w-[min(18rem,calc(100vw-2rem))]')).toBe(true)

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await nextTick()
    expect(tooltip.style.display).toBe('none')

    wrapper.unmount()
  })

  it('supports click activation and restores focus after Escape for rich help', async () => {
    const wrapper = mount(HelpTooltip, {
      attachTo: document.body,
      props: {
        trigger: 'click',
        label: 'Exclusive group',
        closeLabel: 'Close help',
        widthClass: 'w-[min(18rem,calc(100vw-2rem))]',
      },
      slots: {
        default: '<p>Exclusive group example</p>',
      },
    })
    const triggerButton = wrapper.get('button[aria-label="Exclusive group"]')
    const tooltip = getTooltipElement()

    await triggerButton.trigger('click')
    await nextTick()
    expect(tooltip.style.display).not.toBe('none')
    expect(tooltip.textContent).toContain('Exclusive group example')
    expect(tooltip.querySelector('button')?.getAttribute('aria-label')).toBe('Close help')

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await nextTick()
    expect(tooltip.style.display).toBe('none')
    expect(document.activeElement).toBe(triggerButton.element)

    wrapper.unmount()
  })
})
