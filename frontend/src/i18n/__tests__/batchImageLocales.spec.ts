import { describe, expect, it } from 'vitest'

import en from '../locales/en/batchImage'
import zh from '../locales/zh/batchImage'

const usa0MobileMessageKeys = [
  'selectAllJobs',
  'cancelTitle',
  'cancelAction',
  'deleteTitle',
  'deleteSelectedTitle',
  'deleteAction',
  'confirmCancel',
] as const

describe('batch image locale coverage', () => {
  it.each([
    ['en', en.batchImage.messages],
    ['zh', zh.batchImage.messages],
  ])('%s includes the USA0 mobile action messages', (_locale, messages) => {
    expect(Object.keys(messages)).toEqual(expect.arrayContaining(usa0MobileMessageKeys))
  })
})
