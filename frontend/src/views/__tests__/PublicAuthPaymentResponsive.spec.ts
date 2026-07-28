import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const testDir = dirname(fileURLToPath(import.meta.url))
const viewSource = (path: string) => readFileSync(resolve(testDir, '..', path), 'utf8')
const componentSource = (path: string) => readFileSync(resolve(testDir, '../..', 'components', path), 'utf8')

describe('public, auth, and payment responsive theme contracts', () => {
  it('keeps the custom home and download experiences deterministic in both themes', () => {
    const home = viewSource('HomeExperiment.vue')
    const download = viewSource('DownloadView.vue')

    expect(home).toContain(':data-theme="homeTheme"')
    expect(home).toContain('.usa-home[data-theme="dark"]')
    expect(home).toContain('@media (max-width: 620px)')
    expect(download).toContain(':data-theme="theme"')
    expect(download).toContain('.software-page[data-theme="dark"]')
    expect(download).toContain('@media (max-width: 620px)')
  })

  it('gives phone auth controls stable touch targets and stacked long-value actions', () => {
    const register = viewSource('auth/RegisterView.vue')
    const callback = viewSource('auth/OAuthCallbackView.vue')
    const pending = componentSource('auth/PendingOAuthCreateAccountForm.vue')
    const agreement = componentSource('auth/LoginAgreementPrompt.vue')

    expect(register).toContain('min-h-11 min-w-11 items-center justify-center')
    expect(callback.match(/flex flex-col gap-2 sm:flex-row/g)?.length).toBe(3)
    expect(callback.match(/btn btn-secondary w-full sm:w-auto/g)?.length).toBe(3)
    expect(pending).toContain('flex flex-col gap-2 sm:flex-row sm:gap-3')
    expect(pending).toContain('btn btn-secondary w-full shrink-0 sm:w-auto')
    expect(agreement).toContain('flex flex-wrap items-start gap-3')
    expect(agreement).not.toContain('ring-black/10')
  })

  it('uses dynamic viewport height for standalone authorization and payment states', () => {
    expect(viewSource('AppAuthorizationView.vue')).toContain('min-h-[100dvh]')
    expect(viewSource('user/PaymentResultView.vue')).toContain('min-h-[100dvh]')
    expect(viewSource('user/StripePaymentView.vue')).toContain("min-h-[100dvh] bg-canvas text-foreground")
    expect(viewSource('user/StripePopupView.vue')).toContain('min-h-[100dvh]')
  })

  it('preserves scannable QR white surfaces and third-party payment brand colors', () => {
    const qr = viewSource('user/PaymentQRCodeView.vue')
    const stripe = viewSource('user/StripePaymentView.vue')
    const popup = viewSource('user/StripePopupView.vue')

    expect(qr).toContain('bg-white')
    expect(qr).toContain("ctx.fillStyle = '#FFFFFF'")
    expect(stripe).toContain('bg-[#2BB741]')
    expect(popup).toContain("alipay: '#00AEEF'")
    expect(popup).toContain("wechat_pay: '#07C160'")
  })
})
