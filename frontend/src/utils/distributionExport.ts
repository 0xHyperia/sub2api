import { saveAs } from 'file-saver'

export function saveDistributionExport(blob: Blob, prefix: string) {
  const timestamp = new Date().toISOString().slice(0, 19).replace(/[:T]/g, '-')
  saveAs(blob, `${prefix}-${timestamp}.csv`)
}
