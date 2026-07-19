/**
 * Common component types
 */

export interface Column {
  key: string
  label: string
  sortable?: boolean
  class?: string
  /** Hide this column from the generic mobile card without affecting the desktop table. */
  mobileHidden?: boolean
  formatter?: (value: any, row: any) => string
}
