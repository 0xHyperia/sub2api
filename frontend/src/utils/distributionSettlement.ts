export function calculateWithdrawalPreview(amount: number, feeRateBps: number, fixedFee: number) {
  const normalizedAmount = Math.max(0, Number.isFinite(amount) ? amount : 0)
  const fee = Math.max(0, normalizedAmount * Math.max(0, feeRateBps) / 10000 + Math.max(0, fixedFee))
  return { fee, payout: Math.max(0, normalizedAmount - fee) }
}

export function calculatePlatformCredit(amount: number, cnyPerPlatformUsd: number) {
  const normalizedAmount = Math.max(0, Number.isFinite(amount) ? amount : 0)
  const rate = Number.isFinite(cnyPerPlatformUsd) && cnyPerPlatformUsd > 0 ? cnyPerPlatformUsd : 1
  return normalizedAmount / rate
}
