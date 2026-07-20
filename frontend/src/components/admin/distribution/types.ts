export interface DistributionPickerOption {
  id: number
  email: string
  username: string
  meta?: string
  selectable?: boolean
  reason?: string
}

export type DistributionPickerSearch = (
  query: string,
) => Promise<DistributionPickerOption[]>
