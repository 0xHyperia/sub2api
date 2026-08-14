export interface DistributionPickerOption {
  id: number
  email: string
  username: string
  meta?: string
  selectable?: boolean
  reason?: string
  depth?: 1 | 2
  status?: string
}

export type DistributionPickerSearch = (
  query: string,
) => Promise<DistributionPickerOption[]>
