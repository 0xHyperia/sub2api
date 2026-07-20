import { shallowRef } from 'vue'
import { getDistributionAccess, type DistributionAccess } from '@/api/distribution'

const access = shallowRef<DistributionAccess | null>(null)
let pending: Promise<DistributionAccess> | null = null

async function loadDistributionAccess(force = false) {
  if (access.value && !force) return access.value
  if (!pending || force) {
    pending = getDistributionAccess().then((value) => {
      access.value = value
      return value
    }).finally(() => {
      pending = null
    })
  }
  return pending
}

export function useDistributionAccess() {
  return { access, loadDistributionAccess }
}
