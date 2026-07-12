import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import BackupView from '../BackupView.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'

const {
  getS3Config,
  getSchedule,
  listBackups,
  deleteBackup,
  restoreBackup,
} = vi.hoisted(() => ({
  getS3Config: vi.fn(),
  getSchedule: vi.fn(),
  listBackups: vi.fn(),
  deleteBackup: vi.fn(),
  restoreBackup: vi.fn(),
}))

vi.mock('@/api', () => ({
  adminAPI: {
    backup: {
      getS3Config,
      getSchedule,
      listBackups,
      deleteBackup,
      restoreBackup,
    },
  },
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn(),
    showWarning: vi.fn(),
  }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

const record = {
  id: 'backup-1',
  status: 'completed' as const,
  backup_type: 'manual',
  file_name: 'backup.sql.gz',
  s3_key: 'backups/backup.sql.gz',
  size_bytes: 1024,
  triggered_by: 'manual',
  started_at: '2026-07-12T00:00:00Z',
}

function mountView() {
  return mount(BackupView, {
    global: {
      stubs: {
        BaseDialog: true,
        ConfirmDialog: true,
        Toggle: true,
      },
    },
  })
}

function findButton(wrapper: ReturnType<typeof mountView>, text: string) {
  const button = wrapper.findAll('button').find((item) => item.text().includes(text))
  if (!button) throw new Error(`button not found: ${text}`)
  return button
}

function activeDialog(wrapper: ReturnType<typeof mountView>) {
  const dialog = wrapper.findAllComponents(ConfirmDialog).find((item) => item.props('show'))
  if (!dialog) throw new Error('active confirmation dialog not found')
  return dialog
}

describe('admin BackupView confirmations', () => {
  beforeEach(() => {
    getS3Config.mockReset()
    getSchedule.mockReset()
    listBackups.mockReset()
    deleteBackup.mockReset()
    restoreBackup.mockReset()

    getS3Config.mockResolvedValue({
      endpoint: '',
      region: 'auto',
      bucket: '',
      access_key_id: '',
      prefix: 'backups/',
      force_path_style: false,
    })
    getSchedule.mockResolvedValue({
      enabled: false,
      cron_expr: '0 2 * * *',
      retain_days: 14,
      retain_count: 10,
    })
    listBackups.mockResolvedValue({ items: [record] })
    deleteBackup.mockResolvedValue(undefined)
    restoreBackup.mockResolvedValue(record)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('deletes only after the destructive dialog is confirmed', async () => {
    const wrapper = mountView()
    await flushPromises()

    await findButton(wrapper, 'common.delete').trigger('click')
    let dialog = activeDialog(wrapper)
    expect(dialog.props('danger')).toBe(true)
    expect(dialog.props('message')).toBe('admin.backup.actions.deleteConfirm')
    expect(deleteBackup).not.toHaveBeenCalled()

    dialog.vm.$emit('cancel')
    await flushPromises()
    expect(deleteBackup).not.toHaveBeenCalled()

    await findButton(wrapper, 'common.delete').trigger('click')
    dialog = activeDialog(wrapper)
    dialog.vm.$emit('confirm')
    await flushPromises()

    expect(deleteBackup).toHaveBeenCalledOnce()
    expect(deleteBackup).toHaveBeenCalledWith('backup-1')
    wrapper.unmount()
  })

  it('asks for the restore password only after confirmation', async () => {
    const prompt = vi.fn(() => 'restore-password')
    vi.stubGlobal('prompt', prompt)
    const wrapper = mountView()
    await flushPromises()

    await findButton(wrapper, 'admin.backup.actions.restore').trigger('click')
    let dialog = activeDialog(wrapper)
    expect(dialog.props('danger')).toBe(true)
    expect(restoreBackup).not.toHaveBeenCalled()

    dialog.vm.$emit('cancel')
    await flushPromises()
    expect(prompt).not.toHaveBeenCalled()

    await findButton(wrapper, 'admin.backup.actions.restore').trigger('click')
    dialog = activeDialog(wrapper)
    dialog.vm.$emit('confirm')
    await flushPromises()

    expect(prompt).toHaveBeenCalledOnce()
    expect(restoreBackup).toHaveBeenCalledWith('backup-1', 'restore-password')
    wrapper.unmount()
  })
})
