/** Round monetary display values with the conventional ROUND_HALF_UP rule. */
export function roundHalfUp(value: number, fractionDigits = 2): number {
  if (!Number.isFinite(value)) return value
  const factor = 10 ** fractionDigits
  const scaled = value * factor
  const correction = Number.EPSILON * Math.max(1, Math.abs(scaled))
  return Math.sign(scaled) * Math.floor(Math.abs(scaled) + 0.5 + correction) / factor
}

export function formatExchangeRate(value: number): string {
  if (!Number.isFinite(value) || value <= 0) return '—'
  return roundHalfUp(value, 2).toFixed(2)
}
