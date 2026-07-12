export type ShellIconName =
  | 'badge'
  | 'bell'
  | 'chart'
  | 'clipboard'
  | 'cog'
  | 'creditCard'
  | 'cube'
  | 'database'
  | 'document'
  | 'dollar'
  | 'gift'
  | 'globe'
  | 'grid'
  | 'home'
  | 'key'
  | 'server'
  | 'shield'
  | 'sparkles'
  | 'terminal'
  | 'trendingUp'
  | 'user'
  | 'users'

export interface ShellNavItem {
  path: string
  label: string
  icon?: ShellIconName
  iconSvg?: string
  hideInSimpleMode?: boolean
  children?: ShellNavItem[]
  expandOnly?: boolean
  featureFlag?: () => boolean | undefined
}

export interface ShellNavSection {
  id: string
  label: string
  items: ShellNavItem[]
}

export function filterNavigationItems(
  items: ShellNavItem[],
  simpleMode: boolean
): ShellNavItem[] {
  const visible: ShellNavItem[] = []

  for (const item of items) {
    if (item.featureFlag?.() === false) continue
    if (simpleMode && item.hideInSimpleMode) continue

    if (item.children) {
      const children = filterNavigationItems(item.children, simpleMode)
      if (children.length === 0) continue
      visible.push({ ...item, children })
      continue
    }

    visible.push(item)
  }

  return visible
}

export type GroupClickAction = 'expand-sidebar' | 'toggle-group' | 'navigate-and-expand'

export function resolveGroupClickAction(
  sidebarCollapsed: boolean,
  expandOnly: boolean | undefined
): GroupClickAction {
  if (sidebarCollapsed) return 'expand-sidebar'
  return expandOnly ? 'toggle-group' : 'navigate-and-expand'
}
