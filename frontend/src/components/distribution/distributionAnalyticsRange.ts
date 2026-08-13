export type DistributionAnalyticsGranularity = 'day' | 'week' | 'month' | 'custom'

export interface DistributionAnalyticsRangeValue {
  granularity: DistributionAnalyticsGranularity
  preset: string
  date_from: string
  date_to: string
}

const partsInShanghai = (date = new Date()) => {
  const parts = new Intl.DateTimeFormat('en-US', {
    timeZone: 'Asia/Shanghai', year: 'numeric', month: '2-digit', day: '2-digit',
  }).formatToParts(date)
  const value = Object.fromEntries(parts.map(part => [part.type, part.value]))
  return { year: Number(value.year), month: Number(value.month), day: Number(value.day) }
}

export const formatDateOnly = (date: Date) => {
  const year = date.getUTCFullYear()
  const month = String(date.getUTCMonth() + 1).padStart(2, '0')
  const day = String(date.getUTCDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const shanghaiToday = () => {
  const { year, month, day } = partsInShanghai()
  return new Date(Date.UTC(year, month - 1, day))
}

const addDays = (date: Date, days: number) => new Date(date.getTime() + days * 86_400_000)

export function rangeForPreset(granularity: Exclude<DistributionAnalyticsGranularity, 'custom'>, index: number): DistributionAnalyticsRangeValue {
  const today = shanghaiToday()
  let from = today
  let to = today
  if (granularity === 'day') {
    from = to = addDays(today, -index)
  } else if (granularity === 'week') {
    const weekday = today.getUTCDay() || 7
    const currentMonday = addDays(today, 1 - weekday)
    from = addDays(currentMonday, -index * 7)
    to = index === 0 ? today : addDays(from, 6)
  } else {
    from = new Date(Date.UTC(today.getUTCFullYear(), today.getUTCMonth() - index, 1))
    to = index === 0 ? today : new Date(Date.UTC(from.getUTCFullYear(), from.getUTCMonth() + 1, 0))
  }
  return { granularity, preset: `${granularity}-${index}`, date_from: formatDateOnly(from), date_to: formatDateOnly(to) }
}

export const defaultDistributionAnalyticsRange = () => rangeForPreset('month', 0)

export const analyticsRangeParams = (value: DistributionAnalyticsRangeValue) => ({
  date_from: value.date_from,
  date_to: value.date_to,
})

export const formatAnalyticsRangeLabel = (value: DistributionAnalyticsRangeValue) => {
  const formatter = new Intl.DateTimeFormat(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
  const from = formatter.format(new Date(`${value.date_from}T00:00:00`))
  const to = formatter.format(new Date(`${value.date_to}T00:00:00`))
  return from === to ? from : `${from} – ${to}`
}
